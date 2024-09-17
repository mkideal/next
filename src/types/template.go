package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/core/container/pair"
	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/text/templates"

	"github.com/next/next/src/fsutil"
)

// StubPrefix is the prefix for stub templates.
const StubPrefix = "next_stub_0041b8dcd21c3ad6ea2eb3a9f033a9861bc2873e6ab05106e8684ce1b961d4a7_"

// Meta represents the metadata of a entrypoint template file.
type Meta map[string]string

func (m Meta) lookup(key string) pair.Pair[string, bool] {
	if m == nil {
		return pair.Pair[string, bool]{}
	}
	v, ok := m[key]
	return pair.New(v, ok)
}

// ResolveMeta resolves the metadata of the given template for the given keys.
// If no key is given, it resolves all metadata except "this".
func ResolveMeta(tc *templateContext, t *template.Template, keys ...string) (Meta, error) {
	meta := make(Meta)
	if tc.meta == nil {
		tc.meta = make(Meta)
	}
	if len(keys) == 0 {
		templates := t.Templates()
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Root.Pos < templates[j].Root.Pos
		})
		for _, tt := range templates {
			key := tt.Name()
			if strings.HasPrefix(key, "meta/") && key != "meta/this" {
				key = strings.TrimPrefix(key, "meta/")
				var buf bytes.Buffer
				if err := tt.Execute(&buf, tc); err != nil {
					return meta, err
				}
				value := buf.String()
				meta[key] = value
				tc.meta[key] = value
			}
		}
		return meta, nil
	}
	for _, key := range keys {
		tt := t.Lookup("meta/" + key)
		if tt == nil {
			continue
		}
		var buf bytes.Buffer
		if err := tt.Execute(&buf, tc); err != nil {
			return meta, err
		}
		value := buf.String()
		meta[key] = value
		tc.meta[key] = value
	}
	return meta, nil
}

// createTemplate creates a new template from the given content.
func createTemplate(name, content string, funcs ...template.FuncMap) (*template.Template, error) {
	t := template.New(name).Funcs(templates.Funcs)
	for _, fs := range funcs {
		t = t.Funcs(fs)
	}
	return t.Parse(content)
}

// executeTemplate executes a template content with the given data.
func executeTemplate(t *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// templateContextInfo represents the context information of a template.
type templateContextInfo struct {
	context *Compiler
	lang    string // current language
	dir     string // output directory for current language
	ext     string // file extension for current language
}

// @api(Context) related methods and properties are used to retrieve information, perform operations,
// and generate code within the current code generator's context. These methods or properties are
// called directly by name, for example:
//
// ```npl
// {{head}}
// {{next this}}
// {{lang}}
// {{exist meta.path}}
// ```
type templateContext struct {
	templateContextInfo

	// buf is used to buffer the generated content.
	buf bytes.Buffer
	// current decl object to be rendered: File, Const, Enum, Struct, Interface
	decl reflect.Value

	// current entrypoint template
	entrypoint *template.Template

	cache struct {
		// cache for resolved types for current language: Type -> string
		types sync.Map
		// cache for loaded templates: filename -> *template.Template
		templates sync.Map
	}
	initiated bool
	funcs     template.FuncMap
	stack     []string
	pwds      []string
	maxStack  int
	meta      Meta
}

func newTemplateContext(info templateContextInfo) *templateContext {
	tc := &templateContext{
		templateContextInfo: info,
		maxStack:            100,
	}
	if x := os.Getenv("NEXT_MAX_STACK"); x != "" {
		maxStack, err := strconv.Atoi(x)
		if err == nil && maxStack > 0 {
			tc.maxStack = maxStack
		}
	}
	tc.funcs = template.FuncMap{
		// @api(Context/env) represents the environment variables defined in the command line with the flag `-D`.
		//
		// Example:
		// ```sh
		// $ next -D PROJECT_NAME=demo
		// ```
		//
		// ```npl
		// {{env.PROJECT_NAME}}
		// ```
		"env": tc.env,

		// @api(Context/this) represents the current [declaration](#Object/Common/Decl) object to be rendered.
		// this defined in the template [meta](#meta) `meta/this`. Supported types are:
		//
		// - [package](#Object/Package)
		// - [file](#Object/File)
		// - [const](#Object/Const)
		// - [enum](#Object/Enum)
		// - [struct](#Object/Struct)
		// - [interface](#Object/Interface)
		//
		// It's "file" by default.
		"this": tc.this,

		// @api(Context/lang) represents the current language to be generated.
		//
		// Example:
		// ```npl
		// {{lang}}
		// {{printf "%s_alias" lang}}
		// ```
		"lang": func() string { return tc.lang },

		// @api(Context/meta) represents the metadata of a entrypoint template file by flag `-T`.
		// To define a meta, you should define a template with the name `meta/<key>`.
		// Currently, the following meta keys are used by the code generator:
		//
		// - `meta/this`: the current object to be rendered. See [this](#Context/this) for more details.
		// - `meta/path`: the output path for the current object. If the path is not absolute, it will be resolved relative to the current output directory for the current language by command line flag `-O`.
		// - `meta/skip`: whether to skip the current object.
		//
		// Any other meta keys are user-defined. You can use them in the templates like `{{meta.<key>}}`.
		//
		// Example:
		// ```npl
		// {{- define "meta/this" -}}file{{- end -}}
		// {{- define "meta/path" -}}path/to/file{{- end -}}
		// {{- define "meta/skip" -}}{{exist meta.path}}{{- end -}}
		// {{- define "meta/custom" -}}custom value{{- end -}}
		// ```
		//
		// **The metadata will be resolved in the order of the template definition
		// before rendering the entrypoint template.**
		"meta": func() Meta { return tc.meta },

		// @api(Context/error) used to return an error message in the template.
		//
		// Example:
		// ```npl
		// {{error "Something went wrong"}}
		// ```
		"error": tc.error,

		// @api(Context/errorf) used to return a formatted error message in the template.
		//
		// - **Parameters**: (_format_: string, _args_: ...any)
		//
		// Example:
		// ```npl
		// {{errorf "%s went wrong" "Something"}}
		// ```
		"errorf": tc.errorf,

		// @api(Context/pwd) returns the current template file's directory.
		//
		// Example:
		// ```npl
		// {{pwd}}
		// ```
		"pwd": tc.pwd,

		// @api(Context/exist) checks whether the given path exists.
		// If the path is not absolute, it will be resolved relative to the current output directory
		// for the current language by command line flag `-O`.
		//
		// Example:
		// ```npl
		// {{exist "path/to/file"}}
		// {{exist "/absolute/path/to/file"}}
		// {{exist meta.path}}
		// ```
		"exist": tc.exist,

		// @api(Context/head) outputs the header of the generated file.
		//
		// Example:
		// ```
		// {{head}}
		// ```
		//
		// Output (for c++):
		// ```
		// // Code generated by "next v0.0.1"; DO NOT EDIT.
		// ```
		//
		// Output (for c):
		// ```
		// /* Code generated by "next v0.0.1"; DO NOT EDIT. */
		// ```
		"head": tc.head,

		// @api(Context/align) aligns the given text with the same indent as the first line.
		//
		// Example (without align):
		// ```npl
		//		{{print "hello\nworld"}}
		// ```
		//
		// Output:
		// ```
		//		hello
		//	world
		// ```
		//
		// To align it, you can use `align`:
		// ```npl
		//		{{align "hello\nworld"}}
		// ```
		//
		// Output:
		//
		// ```
		//		hello
		//		world
		// ```
		//
		// It's useful when you want to align the generated content, especially for multi-line strings (e.g., comments).
		"align": tc.align,

		// @api(Context/load) loads a template file. It will execute the template immediately but ignore the output.
		// It's useful when you want to load a template file and import the templates it needs.
		//
		// Example:
		// ```npl
		// {{load "path/to/template.npl"}}
		// ```
		"load": tc.load,

		// @api(Context/type) outputs the string representation of the given [type](#Object/Common/Type) for the current language.
		// The type function will lookup the type mapping in the command line flag `-M` and return the corresponding type. If
		// the type is not found, it will lookup <LANG>.map file (e.g., cpp.map) for the type mapping. If the type is still not found,
		// it will return an error.
		"type": tc.type_,

		// @api(Context/next) executes the next template with the given [object](#Object).
		// `{{next object}}` is equivalent to `{{render (object.Typeof) object}}`.
		//
		// Example:
		// ```npl
		// {{- /* Overrides "next/go/struct": add method 'MessageType' for each message after struct */ -}}
		// {{- define "go/struct"}}
		// {{- super .}}
		// {{- with .Annotations.message.type}}
		//
		// func ({{next $.Type}}) MessageType() int { return {{.}} }
		// {{- end}}
		// {{- end -}}
		//
		// {{next this}}
		// ```
		"next": tc.next,

		// @api(Context/super) executes the super template with the given [object](#Object).
		// super is used to call the parent template in the current template. It's useful when
		// you want to extend the parent template. The super template looks up the template with
		// the following priority:
		//
		// ```mermaid
		// %%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#f0f8ff', 'primaryBorderColor': '#7eb0d5', 'lineColor': '#5a9bcf', 'primaryTextColor': '#333333' }}}%%
		// flowchart LR
		//     A["<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"] --> |<span style='color:#5a9bcf'>super</span>| B["<span style='color:#67D7E5'>lang</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//     B --> |<span style='color:#5a9bcf'>super</span>| C["<span style='color:#58B7FF'>next</span>/<span style='color:#67D7E5'>lang</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//     C --> |<span style='color:#5a9bcf'>super</span>| D["<span style='color:#58B7FF'>next</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//
		//     classDef default fill:#f0f8ff,stroke:#7eb0d5,stroke-width:1.5px,rx:12,ry:12;
		//     linkStyle default stroke:#5a9bcf,stroke-width:1.5px;
		// ```
		//
		// e.g.,
		//
		// - `struct` -> `go/struct` -> `next/go/struct` -> `next/struct`
		// - `struct:foo` -> `go/struct:foo` -> `next/go/struct:foo` -> `next/struct:foo`
		//
		// Example:
		// ```npl
		// {{- /* Overrides "next/go/struct": add method 'MessageType' for each message after struct */ -}}
		// {{- define "go/struct"}}
		// {{- super .}}
		// {{- with .Annotations.message.type}}
		//
		// func ({{next $.Type}}) MessageType() int { return {{.}} }
		// {{- end}}
		// {{- end -}}
		// ```
		"super": tc.super,

		// @api(Context/render) executes the template with the given name and data.
		//
		// - **Parameters**: (_name_: string, _data_: any[, _lang_: string])
		//
		// `name` is the template name to be executed.
		// `lang` is the current language by default if not specified.
		//
		// `name` has a specific format. When the corresponding template is not found, it will look up
		// the parent template according to the rules of [super](#Context/super).
		//
		// Example:
		// ```npl
		// {{render "go/struct" this}}
		// {{render "struct" this "go"}}
		// ```
		"render": tc.render,
	}
	return tc
}

func (tc *templateContext) reset(t *template.Template, d reflect.Value) error {
	tc.entrypoint = t
	tc.decl = d
	tc.meta = make(Meta)
	tc.buf.Reset()
	return tc.lazyInit()
}

func (tc *templateContext) pushPwd(pwd string) {
	tc.pwds = append(tc.pwds, pwd)
}

func (tc *templateContext) popPwd() error {
	if len(tc.pwds) == 0 {
		return errors.New("empty pwd stack")
	}
	tc.pwds = tc.pwds[:len(tc.pwds)-1]
	return nil
}

func (tc *templateContext) lazyInit() error {
	if tc.initiated {
		return nil
	}
	tc.initiated = true

	var files []string
	for _, dir := range tc.context.searchDirs {
		var err error
		files, err = fsutil.AppendFiles(files, filepath.Join(dir, "next"+templateExt), templateExt, false)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		files, err = fsutil.AppendFiles(files, filepath.Join(dir, tc.lang+templateExt), templateExt, false)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		for i := range files {
			files[i], err = filepath.Abs(files[i])
			if err != nil {
				return err
			}
		}
	}

	// Load built-in templates
	if _, err := tc.loadTemplate("builtin/next"+templateExt, tc.context.builtin); err != nil {
		return err
	}
	if _, err := tc.loadTemplate("builtin/"+tc.lang+templateExt, tc.context.builtin); err != nil {
		return err
	}

	// Load custom templates
	for _, file := range files {
		if _, err := tc.loadTemplate(file, nil); err != nil {
			return err
		}
	}
	return nil
}

func (tc *templateContext) this() reflect.Value {
	return tc.decl
}

func (tc *templateContext) env() flags.Map {
	return tc.context.flags.envs
}

func (tc *templateContext) error(msg string) (string, error) {
	return "", errors.New(msg)
}

func (tc *templateContext) errorf(format string, args ...any) (string, error) {
	return "", fmt.Errorf(format, args...)
}

func (tc *templateContext) pwd() (string, error) {
	if len(tc.pwds) == 0 {
		return "", errors.New("empty pwd stack")
	}
	return tc.pwds[len(tc.pwds)-1], nil
}

func (tc *templateContext) exist(name string) (bool, error) {
	if name != "" && name[0] != '/' {
		name = filepath.Join(tc.dir, name)
	}
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (tc *templateContext) load(path string) (string, error) {
	if len(tc.pwds) == 0 {
		return "", errors.New("empty pwd stack")
	}
	path = filepath.Join(tc.pwds[len(tc.pwds)-1], path)
	if _, err := tc.loadTemplate(path, nil); err != nil {
		return "", err
	}
	return "", nil
}

func (tc *templateContext) type_(t any, langs ...string) (string, error) {
	if len(langs) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	lang := tc.lang
	if len(langs) == 1 {
		lang = langs[0]
	}
	if v, ok := tc.cache.types.Load(t); ok {
		return v.(string), nil
	}
	v, err := tc.resolveLangType(lang, t)
	if err != nil {
		return "", err
	}
	tc.cache.types.Store(t, v)
	return v, nil
}

func (tc *templateContext) resolveLangType(lang string, t any) (result string, err error) {
	mappings := tc.context.flags.mappings
	defer func() {
		if err == nil && strings.Contains(result, "box(") {
			// replace box(...) with the actual type
			for k, v := range mappings {
				if strings.HasPrefix(k, lang+".box(") {
					dot := strings.Index(k, ".")
					if dot > 0 && k[:dot] == lang {
						result = strings.ReplaceAll(result, k[dot+1:], v)
					}
				}
			}
			result = removeBox(result)
		}
	}()
	for {
		x, ok := t.(*UsedType)
		if !ok {
			break
		}
		t = x.Type
	}
	switch t := t.(type) {
	case string:
		p, ok := mappings[lang+"."+t]
		if !ok {
			return "", fmt.Errorf("type %q not found", t)
		}
		return p, nil

	case *PrimitiveType:
		p, ok := mappings[lang+"."+t.name]
		if !ok {
			return "", fmt.Errorf("type %q not found", t.name)
		}
		return p, nil

	case *MapType:
		p, ok := mappings[lang+".map"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "map")
		}
		if strings.Contains(p, "%K%") {
			k, err := tc.next(t.KeyType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%K%", k)
		}
		if strings.Contains(p, "%V%") {
			v, err := tc.next(t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%V%", v)
		}
		return p, nil

	case *VectorType:
		p, ok := mappings[lang+".vector"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "vector")
		}
		if strings.Contains(p, "%T%") {
			e, err := tc.next(t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T%", e)
		}
		return p, nil

	case *ArrayType:
		p, ok := mappings[lang+".array"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "array")
		}
		if strings.Contains(p, "%T%") {
			e, err := tc.next(t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T%", e)
		}
		if strings.Contains(p, "%N%") {
			p = strings.ReplaceAll(p, "%N%", strconv.FormatInt(t.N, 10))
		}
		return p, nil

	default:
		if t, ok := t.(Type); ok {
			name := t.String()
			p, ok := mappings[lang+"."+name]
			if ok {
				return p, nil
			}
			return name, nil
		}
		return "", fmt.Errorf("unsupported type %T, expected Type or string", t)
	}
}

func (tc *templateContext) head() string {
	p, ok := tc.context.flags.mappings[tc.lang+".comment"]
	if !ok {
		return ""
	}

	return strings.ReplaceAll(p, "%T%", `Code generated by "next `+builder.Info().Version+`"; DO NOT EDIT.`)
}

func (tc *templateContext) loadTemplate(path string, fs fs.FS) (*template.Template, error) {
	var content []byte
	var err error

	if fs != nil {
		f, err := fs.Open(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			tc.context.Infof("template %q not found", path)
			return nil, nil
		}
		defer f.Close()
		content, err = io.ReadAll(f)
	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve template %q absolute path: %w", path, err)
		}
		if v, ok := tc.cache.templates.Load(path); ok {
			return v.(*template.Template), nil
		}
		content, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	t, err := createTemplate(path, string(content), tc.funcs)
	if err != nil {
		return nil, err
	}
	templates := t.Templates()
	for i := range templates {
		tt := templates[i]
		name := tt.Name()
		if t2 := tc.entrypoint.Lookup(name); t2 != nil {
			return nil, fmt.Errorf("template %q already defined at %s", name, t2.ParseName)
		}
		if fs != nil {
			if _, err := tc.entrypoint.AddParseTree(name, tt.Tree); err != nil {
				return nil, err
			}
			continue
		}
		stubName := StubPrefix + hex.EncodeToString([]byte(name))
		stubContent := fmt.Sprintf(`{{%s .}}`, stubName)
		stubFuncs := template.FuncMap{
			stubName: func(data any) (string, error) {
				return tc.execute(tt, path, data)
			},
		}
		stub, err := createTemplate(stubName, stubContent, tc.funcs, stubFuncs)
		if err != nil {
			return nil, err
		}
		tc.entrypoint.Funcs(stubFuncs)
		if _, err := tc.entrypoint.AddParseTree(name, stub.Tree); err != nil {
			return nil, err
		}
	}
	tc.cache.templates.Store(path, t)
	// Execute the template but ignore the result. This is to ensure that the template is valid
	// and immediately imports the templates it needs.
	tc.pushPwd(filepath.Dir(path))
	defer tc.popPwd()
	if err := t.Execute(io.Discard, tc); err != nil {
		return nil, err
	}
	return t, nil
}

func (tc *templateContext) execute(tt *template.Template, path string, data any) (string, error) {
	if path != "" {
		tc.pushPwd(filepath.Dir(path))
		defer tc.popPwd()
	}
	start := tc.buf.Len()
	if err := tt.Execute(&tc.buf, data); err != nil {
		return "", err
	}
	result := tc.buf.String()[start:]
	tc.buf.Truncate(start)
	return result, nil
}

const sep = "/"

func (tc *templateContext) parseTemplateNames(lang string, name string) []string {
	op.Assert(lang != "next")
	priority := 1
	parts := strings.Split(name, sep)
	switch len(parts) {
	case 1:
	case 2:
		if parts[0] == "next" {
			name = parts[1]
			priority = 3
		} else if parts[0] == lang {
			name = parts[1]
		} else {
			return nil
		}
	case 3:
		if parts[0] == "next" && parts[1] == lang {
			name = parts[2]
			priority = 2
		} else {
			return nil
		}
	default:
		return nil
	}
	var names []string
	// <lang>.<name>
	if priority <= 1 {
		names = append(names, lang+sep+name)
	}
	// next.<lang>.<name>
	if priority <= 2 {
		names = append(names, "next"+sep+lang+sep+name)
	}
	// next.<name>
	if priority <= 3 {
		names = append(names, "next"+sep+name)
	}
	return names
}

// lookupTemplate looks up the template with the given name.
// If exactly is true, it only looks up the template with the given name.
// Otherwise, it looks up the template with the following priority:
//
// 1. <lang>/<name>
// 2. next/<lang>/<name>
// 3. next/<name>
func (tc *templateContext) lookupTemplate(names []string) (*template.Template, error) {
	for i := range names {
		if tt := tc.entrypoint.Lookup(names[i]); tt != nil {
			return tt, nil
		}
	}
	return nil, &TemplateNotFoundError{Name: names[0]}
}

func (tc *templateContext) next(obj Object, langs ...string) (string, error) {
	if len(langs) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	lang := tc.lang
	if len(langs) == 1 {
		lang = langs[0]
	}
	names := tc.parseTemplateNames(lang, obj.Typeof())
	return tc.nextWithNames(names, obj)
}

func (tc *templateContext) nextWithNames(names []string, obj Object) (string, error) {
	result, err := tc.renderWithNames(names, obj)
	if err != nil && IsTemplateNotFoundError(err) {
		if t, ok := obj.(Type); ok {
			return tc.type_(t)
		}
		return "", err
	}
	return result, err
}

func (tc *templateContext) super(obj Object, langs ...string) (string, error) {
	if len(langs) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	lang := tc.lang
	if len(langs) == 1 {
		lang = langs[0]
	}
	if len(tc.stack) == 0 {
		return "", fmt.Errorf("super not allowed in the template")
	}
	name := tc.stack[len(tc.stack)-1]
	tc.context.Tracef("super template %q", name)
	names := tc.parseTemplateNames(lang, name)
	if len(names) < 2 {
		return "", fmt.Errorf("invalid template name %q", name)
	}
	return tc.nextWithNames(names[1:], obj)
}

func (tc *templateContext) render(name string, data any, langs ...string) (result string, err error) {
	if len(langs) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	lang := tc.lang
	if len(langs) == 1 {
		lang = langs[0]
	}
	names := tc.parseTemplateNames(lang, name)
	if len(names) == 0 {
		return "", fmt.Errorf("invalid template name %q", name)
	}
	return tc.renderWithNames(names, data)
}

func (tc *templateContext) renderWithNames(names []string, data any) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	if len(tc.stack) >= tc.maxStack {
		return "", fmt.Errorf("npl: stack overflow")
	}
	tt, err := tc.lookupTemplate(names)
	if err != nil {
		return "", err
	}
	tc.stack = append(tc.stack, tt.Name())
	tc.context.Tracef("rendering template %q", tt.Name())
	defer func() {
		tc.stack = tc.stack[:len(tc.stack)-1]
	}()
	return tc.execute(tt, "", data)
}

func (tc *templateContext) lastLine() string {
	content := tc.buf.Bytes()
	if len(content) == 0 {
		return ""
	}

	lastLineEnd := len(content)
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == '\n' {
			lastLineEnd = i + 1
			break
		}
	}

	return string(content[lastLineEnd:])
}

func (tc *templateContext) lastIndent() string {
	content := tc.buf.Bytes()
	if len(content) == 0 {
		return ""
	}

	lastLineBegin := len(content)
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == '\n' {
			lastLineBegin = i + 1
			break
		}
	}

	indentEnd := lastLineBegin
	for indentEnd < len(content) && (content[indentEnd] == ' ' || content[indentEnd] == '\t') {
		indentEnd++
	}

	return string(content[lastLineBegin:indentEnd])
}

func (tc *templateContext) align(s string) string {
	return tc.alignWith(s, "")
}

func (tc *templateContext) alignWith(text, prefix string) string {
	indent := strings.Map(func(r rune) rune {
		if r != ' ' && r != '\t' {
			return ' '
		}
		return r
	}, tc.lastLine())
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return prefix + text
	}
	lines[0] = prefix + lines[0]
	for i := 1; i < len(lines); i++ {
		if lines[i] != "" || i+1 == len(lines) {
			lines[i] = indent + prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
