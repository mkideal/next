package types

import (
	"bytes"
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

// Meta represents the metadata of a template.
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
func createTemplate(name, content string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(name).Funcs(templates.Funcs).Funcs(funcs).Parse(content)
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
	context *Context
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
// {{next this)}}
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
	dontOverrides map[string]bool
	initiated     bool
	funcs         template.FuncMap
	stack         []string
	meta          Meta
}

func newTemplateContext(info templateContextInfo) *templateContext {
	tc := &templateContext{
		templateContextInfo: info,
		dontOverrides:       make(map[string]bool),
	}
	tc.funcs = template.FuncMap{
		// @api(Context/env) represents the environment variables defined in the command line with the flag `-D`.
		//
		// Example:
		//
		// ```sh
		// next -D PROJECT_NAME=demo
		// ```
		//
		// ```npl
		// {{env.PROJECT_NAME}}
		// ```
		"env": tc.env,

		// @api(Context/this) represents the current [declaration](#Object/Common/Decl) object to be rendered.
		// this defined in the template [meta](#meta) `meta/this`. Supported types are:
		//
		// - [file](#Object/File)
		// - [const](#Object/Const)
		// - [enum](#Object/Enum)
		// - [struct](#Object/Struct)
		// - [interface](#Object/Interface)
		//
		// It's a [file](#Object/File) by default.
		"this": tc.this,

		// @api(Context/lang) represents the current language to be generated.
		"lang": func() string { return tc.lang },

		// @api(Context/meta) represents the metadata of a entrypoint template.
		// To define a meta, you should define a template with the name `meta/<key>`.
		// Currently, the following meta keys are supported:
		//
		// - `meta/this`: the current object to be rendered.
		// - `meta/path`: the output path for the current object.
		// - `meta/skip`: whether to skip the current object.
		//
		// Any other meta keys are user-defined. You can use them in the templates like `{{meta.<key>}}`.
		//
		// Example:
		//
		// ```npl
		// {{- define "meta/this" -}}file{{- end -}}
		// {{- define "meta/path" -}}/path/to/file{{- end -}}
		// {{- define "meta/skip" -}}{{exist meta.path}}{{- end -}}
		// {{- define "meta/custom" -}}custom value{{- end -}}
		// ```
		//
		// All meta templates should be defined in the entrypoint template.
		// The meta will be resolved in the order of the template definition
		// before rendering the entrypoint template.
		"meta": func() Meta { return tc.meta },

		// @api(Context/error) used to return an error message in the template.
		//
		// Example:
		//
		// ```npl
		// {{error "Something went wrong"}}
		// ```
		"error": tc.error,

		// @api(Context/errorf) used to return a formatted error message in the template.
		//
		// Example:
		//
		// ```npl
		// {{errorf "%s went wrong" "Something"}}
		// ```
		"errorf": tc.errorf,

		// @api(Context/exist) checks whether the given path exists.
		// If the path is not absolute, it will be resolved relative to the current output directory
		// for the current language by command line flag `-O`.
		"exist": tc.exist,

		// @api(Context/head) outputs the header of the generated file.
		//
		// Example:
		//
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

		// @api(Context/align) aligns the given text with the last line indent of the generated content.
		"align": tc.align,

		// @api(Context/type) outputs the string representation of the given [type](#Object/Common/Type) for the current language.
		"type": tc.type_,

		// @api(Context/next) executes the next template with the given [object](#Object).
		"next": tc.next,

		// @api(Context/super) executes the super template with the given [object](#Object).
		"super": tc.super,

		// @api(Context/render) executes the template with the given name and data.
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

func (tc *templateContext) lazyInit() error {
	if tc.initiated {
		return nil
	}
	tc.initiated = true

	for _, tt := range tc.entrypoint.Templates() {
		if _, ok := tc.dontOverrides[tt.Name()]; ok {
			return fmt.Errorf("template %q is already defined", tt.Name())
		}
		tc.dontOverrides[tt.Name()] = true
	}

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
	if _, err := tc.loadTemplate(tc.context.builtin, "builtin/next"+templateExt); err != nil {
		return err
	}
	if _, err := tc.loadTemplate(tc.context.builtin, "builtin/"+tc.lang+templateExt); err != nil {
		return err
	}

	// Load custom templates
	for _, file := range files {
		if _, err := tc.loadTemplate(nil, file); err != nil {
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
		p, ok := mappings[lang+".map<%K%,%V%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "map<%K%,%V%>")
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
		p, ok := mappings[lang+".vector<%T%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "vector<%T%>")
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
		p, ok := mappings[lang+".array<%T%,%N%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "array<%T%,%N%>")
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
	p, ok := tc.context.flags.mappings[tc.lang+".comment(%S%)"]
	if !ok {
		return ""
	}

	return strings.ReplaceAll(p, "%S%", `Code generated by "next `+builder.Info().Version+`"; DO NOT EDIT.`)
}

func (tc *templateContext) loadTemplate(fs fs.FS, filename string) (*template.Template, error) {
	if v, ok := tc.cache.templates.Load(filename); ok {
		return v.(*template.Template), nil
	}
	tc.context.Printf("loading template %q", filename)

	var content []byte
	var err error

	if fs != nil {
		f, err := fs.Open(filename)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			tc.context.Infof("template %q not found", filename)
			return nil, nil
		}
		defer f.Close()
		content, err = io.ReadAll(f)
	} else {
		content, err = os.ReadFile(filename)
	}
	if err != nil {
		return nil, err
	}
	t, err := createTemplate(filename, string(content), tc.funcs)
	if err != nil {
		return nil, err
	}
	for _, tt := range t.Templates() {
		if tc.dontOverrides[tt.Name()] {
			continue
		}
		if _, err := tc.entrypoint.AddParseTree(tt.Name(), tt.Tree); err != nil {
			return nil, err
		}
	}
	tc.cache.templates.Store(filename, t)
	return t, nil
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
	names := tc.parseTemplateNames(lang, obj.getType())
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
	tt, err := tc.lookupTemplate(names)
	if err != nil {
		return "", err
	}
	tc.stack = append(tc.stack, tt.Name())
	tc.context.Tracef("rendering template %q", tt.Name())
	defer func() {
		tc.stack = tc.stack[:len(tc.stack)-1]
	}()
	start := tc.buf.Len()
	if err := tt.Execute(&tc.buf, data); err != nil {
		return "", err
	}
	result = tc.buf.String()[start:]
	tc.buf.Truncate(start)
	return result, nil
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
