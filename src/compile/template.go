package compile

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"go/constant"
	"io"
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
	"github.com/gopherd/core/text/document"
	"github.com/gopherd/core/text/templates"

	"github.com/next/next/src/internal/fsutil"
	"github.com/next/next/src/internal/stringutil"
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

func validMetaKey(key string) error {
	if !stringutil.IsIdentifer(key) {
		return fmt.Errorf("must be a valid identifier")
	}
	if strings.HasPrefix(key, "_") || key == "this" || key == "path" || key == "skip" {
		return nil
	}
	return fmt.Errorf("must be %q, %q, %q, or starts with %q for custom key (e.g., %q)", "this", "path", "skip", "_", "_"+key)
}

func templatePos(content string, pos int, lookahead string) document.Position {
	if lookahead != "" {
		index := strings.LastIndex(content[:pos], lookahead)
		if index >= 0 {
			pos = index + len(lookahead)
		}
	}
	p := document.PositionFor(content, pos)
	if p.IsValid() {
		p.Character--
	}
	return p
}

// resolveMeta resolves the metadata of the given template for the given keys.
// If no key is given, it resolves all metadata except "this".
func resolveMeta(tc *templateContext, t *template.Template, content string, keys ...string) (Meta, error) {
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
				if err := validMetaKey(key); err != nil {
					pos := templatePos(content, int(tt.Root.Pos), "meta/")
					return meta, fmt.Errorf("%s: invalid meta key %q: %w", document.FormatPosition(t.ParseName, pos), key, err)
				}
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

// templateContextInfo represents the context information of a template.
type templateContextInfo struct {
	compiler *Compiler
	lang     string // current language
	dir      string // output directory for current language
	ext      string // file extension for current language
}

type buffer struct {
	lang string
	id   string
	data []byte
}

func (b *buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *buffer) String() string {
	return string(b.data)
}

func (b *buffer) Bytes() []byte {
	return b.data
}

func (b *buffer) Reset() {
	b.data = b.data[:0]
}

type stack struct {
	pwd string
	buf *buffer
}

// @api(Context) related methods and properties are used to retrieve information, perform operations,
// and generate code within the current code generator's context. These methods or properties are
// called directly by name, for example:
//
//	```npl
//	{{head}}
//	{{next this}}
//	{{lang}}
//	{{exist meta.path}}
//	```
type templateContext struct {
	templateContextInfo

	// execution stacks
	stacks   []*stack
	trace    []string
	maxStack int

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
	meta      Meta
}

func newTemplateContext(info templateContextInfo) *templateContext {
	tc := &templateContext{
		templateContextInfo: info,
		maxStack:            100,
	}
	if x := os.Getenv(NEXTMAXSTACK); x != "" {
		maxStack, err := strconv.Atoi(x)
		if err == nil && maxStack > 0 {
			tc.maxStack = maxStack
		}
	}
	tc.funcs = template.FuncMap{
		// @api(Context/env) represents the environment variables defined in the command line with the flag `-D`.
		//
		// Example:
		//
		//	```sh
		//	$ next -D PROJECT_NAME=demo
		//	```
		//
		//	```npl
		//	{{env.PROJECT_NAME}}
		//	```
		"env": tc.env,

		// @api(Context/make) converts the given value to a constant value.
		//
		// Example:
		//
		//	```npl
		//	{{make 1}}
		//	{{make 1.0}}
		//	{{make "hello"}}
		//	{{printf "hello"}}
		//	```
		//
		//	Output:
		//
		//	```
		//	1
		//	1.0
		//	"hello"
		//	hello
		//	```
		"make": func(x any) (constant.Value, error) {
			if x == nil {
				return constant.MakeUnknown(), fmt.Errorf("invalid value %v", x)
			}
			rv := reflect.ValueOf(x)
			if rv.CanInt() {
				return constant.MakeInt64(rv.Int()), nil
			}
			if rv.CanUint() {
				return constant.MakeUint64(rv.Uint()), nil
			}
			if rv.CanFloat() {
				return constant.MakeFloat64(rv.Float()), nil
			}
			v := constant.Make(x)
			if v.Kind() == constant.Unknown {
				return v, fmt.Errorf("invalid value %v", x)
			}
			return v, nil
		},

		// @api(Context/this) represents the current [Decl](#Object/Common/Decl) object to be rendered.
		// this defined in the template [meta](#meta) `meta/this`. Supported types are:
		//
		// - **package**: [Package](#Object/Package)
		// - **file**: [File](#Object/File)
		// - **const**: [Const](#Object/Const)
		// - **enum**: [Enum](#Object/Enum)
		// - **struct**: [Struct](#Object/Struct)
		// - **interface**: [Interface](#Object/Interface)
		//
		// It's "file" by default.
		"this": tc.this,

		// @api(Context/lang) represents the current language to be generated.
		//
		// Example:
		//
		//	```npl
		//	{{lang}}
		//	{{printf "%s_alias" lang}}
		//	```
		"lang": func() string { return tc.lang },

		// @api(Context/indent) adds an indent to each non-empty line of the given text.
		// The indent is defined in the command line flag `-M <lang>.indent` or `LANG.map` file (e.g., cpp.map).
		//
		// Example:
		//
		//	```npl
		//	{{- define "next/enum" -}}
		//	enum {{next .Type}} {
		//	    {{- next .Members | indent | blockspace}}
		//	}
		//	{{end}}
		//	```
		"indent": tc.indent,

		// @api(Context/meta) represents the metadata of entrypoint template file by flag `-T`.
		// To define a meta, you should define a template with the name `meta/<key>`.
		// Currently, the following meta keys are used by the code generator:
		//
		// - `meta/this`: the current object to be rendered. See [this](#Context/this) for details.
		// - `meta/path`: the output path for the current object. If the path is not absolute, it will be resolved relative to the current output directory for the current language by command line flag `-O`.
		// - `meta/skip`: whether to skip the current object.
		//
		// You can use them in the templates like `{{meta.<key>}}`.
		//
		// :::tip
		//
		// User-defined meta key **MUST** be prefixed with `_`, e.g., `_custom_key`.
		//
		// :::
		//
		// Example:
		//
		//	```npl
		//	{{- define "meta/this" -}}file{{- end -}}
		//	{{- define "meta/path" -}}path/to/file{{- end -}}
		//	{{- define "meta/skip" -}}{{exist meta.path}}{{- end -}}
		//	{{- define "meta/_custom_key" -}}custom value{{- end -}}
		//
		//	{{meta._custom_key}}
		//	```
		//
		// :::note
		//
		// The metadata will be resolved in the order of the template definition
		// before rendering the entrypoint template.
		//
		// :::
		"meta": func() Meta { return tc.meta },

		// @api(Context/debug) outputs a debug message in the console.
		//
		// Example:
		//
		//	```npl
		//	{{debug "Hello, world"}}
		//	{{debug "Hello, %s" "world"}}
		//	```
		//
		// :::tip
		//
		// It's useful when you want to print debug information in the console during code generation.
		//
		// :::
		"debug": tc.debug,

		// @api(Context/error) used to return an error message in the template.
		//
		// Example:
		//
		//	```npl
		//	{{error "Something went wrong"}}
		//	{{error "%s went wrong" "Something"}}
		//	```
		//
		// :::tip
		//
		// Using `.Pos` to get the current object's position in source file is a good practice.
		//
		//	```npl
		//	{{- define "next/protobuf/enum" -}}
		//	{{- if not .MemberType.Kind.IsInteger -}}
		//	{{error "%s: enum type must be an integer" .Pos}}
		//	{{- end}}
		//	{{- end}}
		//	```
		//
		// :::
		"error": tc.error,

		// @api(Context/pwd) returns the current template file's directory.
		//
		// Example:
		//
		//	```npl
		//	{{pwd}}
		//	```
		"pwd": tc.pwd,

		// @api(Context/exist) checks whether the given path exists.
		// If the path is not absolute, it will be resolved relative to the current output directory
		// for the current language by command line flag `-O`.
		//
		// Example:
		//
		//	```npl
		//	{{exist "path/to/file"}}
		//	{{exist "/absolute/path/to/file"}}
		//	{{exist meta.path}}
		//	```
		"exist": tc.exist,

		// @api(Context/head) outputs the header of the generated file.
		//
		// Example:
		//
		//	```npl
		//	{{head}}
		//	```
		//
		// Output (for c++):
		//
		//	```cpp
		//	// Code generated by "next 0.0.1"; DO NOT EDIT.
		//	```
		//
		// Output (for c):
		//
		//	```c
		//	/* Code generated by "next 0.0.1"; DO NOT EDIT. */
		//	```
		"head": tc.head,

		// @api(Context/align) aligns the given text with the same indent as the first line.
		//
		// Example (without align):
		//
		//	```npl
		//		{{print "hello\nworld"}}
		//	```
		//
		// Output:
		//
		//	```
		//		hello
		//	world
		//	```
		//
		// To align it, you can use `align`:
		//
		//	```npl
		//		{{align "hello\nworld"}}
		//	```
		//
		// Output:
		//
		//	```
		//		hello
		//		world
		//	```
		//
		// It's useful when you want to align the generated content, especially for multi-line strings (e.g., comments).
		"align": tc.align,

		// @api(Context/load) loads a template file. It will execute the template immediately but ignore the output.
		// It's useful when you want to load a template file and import the templates it needs.
		//
		// Example:
		//
		//	```npl
		//	{{load "path/to/template.npl"}}
		//	```
		"load": tc.load,

		// @api(Context/type) outputs the string representation of the given [Type](#Object/Common/Type) for the current language.
		// The type function will lookup the type mapping in the command line flag `-M` and return the corresponding type. If
		// the type is not found, it will lookup LANG.map file (e.g., cpp.map) for the type mapping. If the type is still not found,
		// it will return an error.
		"type": tc.type_,

		// @api(Context/next) executes the next template with the given [Object](#Object).
		// `{{next object}}` is equivalent to `{{render (object.Typeof) object}}`.
		//
		// Example:
		//
		//	```npl
		//	{{- /* Overrides "next/go/struct": add method 'MessageType' for each message after struct */ -}}
		//	{{- define "go/struct"}}
		//	{{- super .}}
		//	{{- with .Annotations.message.type}}
		//
		//	func ({{next $.Type}}) MessageType() int { return {{.}} }
		//	{{- end}}
		//	{{- end -}}
		//
		//	{{next this}}
		//	```
		"next": tc.next,

		// @api(Context/super) executes the super template with the given [Object](#Object).
		// super is used to call the parent template in the current template. It's useful when
		// you want to extend the parent template. The super template looks up the template with
		// the following priority:
		//
		//	```mermaid
		//	%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#f0f8ff', 'primaryBorderColor': '#7eb0d5', 'lineColor': '#5a9bcf', 'primaryTextColor': '#333333' }}}%%
		//	flowchart LR
		//	    A["<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"] --> |<span style='color:#5a9bcf'>super</span>| B["<span style='color:#67D7E5'>lang</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//	    B --> |<span style='color:#5a9bcf'>super</span>| C["<span style='color:#58B7FF'>next</span>/<span style='color:#67D7E5'>lang</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//	    C --> |<span style='color:#5a9bcf'>super</span>| D["<span style='color:#58B7FF'>next</span>/<span style='color:#9095FF'>type</span><span style='color:#E59C00'>[:name]</span>"]
		//
		//	    classDef default fill:#f0f8ff,stroke:#7eb0d5,stroke-width:1.5px,rx:12,ry:12;
		//	    linkStyle default stroke:#5a9bcf,stroke-width:1.5px;
		//	```
		//
		// e.g.,
		//
		// - `struct` -> `go/struct` -> `next/go/struct` -> `next/struct`
		// - `struct:foo` -> `go/struct:foo` -> `next/go/struct:foo` -> `next/struct:foo`
		//
		// Example:
		//
		//	```npl
		//	{{- /* Overrides "next/go/struct": add method 'MessageType' for each message after struct */ -}}
		//	{{- define "go/struct"}}
		//	{{- super .}}
		//	{{- with .Annotations.message.type}}
		//
		//	func ({{next $.Type}}) MessageType() int { return {{.}} }
		//	{{- end}}
		//	{{- end -}}
		//	```
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
		//
		//	```npl
		//	{{render "struct" this}}
		//	{{render "struct" this "go"}}
		//	```
		"render": tc.render,
	}
	return tc
}

func (tc *templateContext) reset(t *template.Template, d reflect.Value) error {
	tc.entrypoint = t
	tc.decl = d
	tc.meta = make(Meta)
	tc.stacks = tc.stacks[:0]
	return tc.lazyInit()
}

var nextBuffer int

func (tc *templateContext) push(pwd string) *stack {
	nextBuffer++
	tc.stacks = append(tc.stacks, &stack{pwd: pwd, buf: &buffer{
		lang: tc.lang,
		id:   fmt.Sprintf("%d:%d", len(tc.stacks), nextBuffer),
	}})
	return tc.pc()
}

func (tc *templateContext) pop() error {
	if len(tc.stacks) == 0 {
		return errors.New("no stacks")
	}
	tc.stacks = tc.stacks[:len(tc.stacks)-1]
	return nil
}

func (tc *templateContext) pc() *stack {
	return tc.stacks[len(tc.stacks)-1]
}

func (tc *templateContext) lazyInit() error {
	if tc.initiated {
		return nil
	}
	tc.initiated = true

	var files []string
	for _, dir := range tc.compiler.searchDirs {
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
	if _, err := tc.loadTemplate("builtin/next"+templateExt, tc.compiler.builtin); err != nil {
		return err
	}
	if _, err := tc.loadTemplate("builtin/"+tc.lang+templateExt, tc.compiler.builtin); err != nil {
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

func (tc *templateContext) indent(args ...reflect.Value) (string, error) {
	var times int
	var s string
	switch len(args) {
	case 1:
		if args[0].Kind() != reflect.String {
			return "", fmt.Errorf("invalid argument type %q, expected string", args[0].Kind())
		}
		times = 1
		s = args[0].String()
	case 2:
		if args[0].CanInt() {
			times = int(args[0].Int())
		} else if args[0].CanUint() {
			times = int(args[0].Uint())
		} else {
			return "", fmt.Errorf("invalid argument type %q, first argument must be int", args[0].Kind())
		}
		if times < 0 {
			return "", fmt.Errorf("invalid argument %d, first argument must be non-negative", times)
		}
		if args[1].Kind() != reflect.String {
			return "", fmt.Errorf("invalid argument type %q, second argument must be string", args[1].Kind())
		}
		s = args[1].String()
	default:
		return "", fmt.Errorf("wrong number of arguments: expected <int, string> or <string>")
	}
	if s == "" {
		return "", nil
	}
	var indent string
	if times > 0 {
		indent = tc.compiler.options.Mapping.Get(tc.lang + ".indent")
		if indent == "" || indent == "tab" {
			indent = "\t"
		} else {
			i, err := strconv.Atoi(indent)
			if i < 0 || err != nil {
				return s, fmt.Errorf("invalid indent %q of language %q", indent, tc.lang)
			}
			indent = strings.Repeat(" ", i)
		}
		indent = strings.Repeat(indent, times)
	}
	return tc.alignWithIndent(indent, s), nil
}

func (tc *templateContext) this() reflect.Value {
	return tc.decl
}

func (tc *templateContext) env() flags.Map {
	return tc.compiler.options.Env
}

func (tc *templateContext) debug(msg string, args ...any) string {
	tc.compiler.Debug(msg, args...)
	return ""
}

func (tc *templateContext) error(msg string, args ...any) (string, error) {
	if len(args) == 0 {
		return "", errors.New(msg)
	}
	return "", fmt.Errorf(msg, args...)
}

func (tc *templateContext) pwd() (string, error) {
	if len(tc.stacks) == 0 {
		return "", errors.New("no stacks")
	}
	return tc.pc().pwd, nil
}

func (tc *templateContext) exist(name string) bool {
	if name != "" && name[0] != '/' {
		name = filepath.Join(tc.dir, name)
	}
	return tc.compiler.platform.IsExist(name)
}

func (tc *templateContext) load(path string) (string, error) {
	if len(tc.stacks) == 0 {
		return "", errors.New("no stacks")
	}
	pwd := tc.pc().pwd
	path = filepath.Join(pwd, path)
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
	mappings := tc.compiler.options.Mapping
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
			return "", fmt.Errorf("%s: type %q not found", tc.lang, t)
		}
		return p, nil

	case *PrimitiveType:
		p, ok := mappings[lang+"."+t.name]
		if !ok {
			return "", fmt.Errorf("%s: type %q not found", tc.lang, t.name)
		}
		return p, nil

	case *MapType:
		p, ok := mappings[lang+".map"]
		if !ok {
			return "", fmt.Errorf("%s: type %q not found", tc.lang, "map")
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
			return "", fmt.Errorf("%s: type %q not found", tc.lang, "vector")
		}
		if strings.Contains(p, "%T%") {
			e, err := tc.next(t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T%", e)
		}
		if strings.Contains(p, "%T.E%") {
			e, err := tc.next(finalElementType(t.ElemType))
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T.E%", e)
		}
		return p, nil

	case *ArrayType:
		p, ok := mappings[lang+".array"]
		if !ok {
			return "", fmt.Errorf("%s: type %q not found", tc.lang, "array")
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
		if strings.Contains(p, "%T.E%") {
			e, err := tc.next(finalElementType(t.ElemType))
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T.E%", e)
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

func finalElementType(t Type) Type {
	for {
		switch x := t.(type) {
		case *VectorType:
			t = x.ElemType
		case *ArrayType:
			t = x.ElemType
		case *UsedType:
			t = x.Type
		default:
			return t
		}
	}
}

func (tc *templateContext) head() string {
	p, ok := tc.compiler.options.Mapping[tc.lang+".comment"]
	if !ok {
		return ""
	}
	header := tc.compiler.options.Head
	if header == "" {
		header = `Code generated by "next ` + strings.TrimPrefix(builder.Info().Version, "v") + `"; DO NOT EDIT.`
	}
	return strings.ReplaceAll(p, "%T%", header)
}

func (tc *templateContext) loadTemplate(path string, fs FileSystem) (*template.Template, error) {
	var content []byte
	var err error

	if fs != nil {
		var f io.ReadCloser
		f, err = fs.Open(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			tc.compiler.Trace("template %q not found", path)
			return nil, nil
		}
		if abs, err := fs.Abs(path); err == nil {
			path = abs
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
		content, err = tc.compiler.platform.ReadFile(path)
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
				return tc.execute(tt, filepath.Dir(path), data)
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
	tc.push(filepath.Dir(path))
	defer tc.pop()
	if err := t.Execute(io.Discard, tc); err != nil {
		return nil, err
	}
	return t, nil
}

func (tc *templateContext) execute(tt *template.Template, path string, data any) (string, error) {
	stack := tc.push(path)
	defer tc.pop()
	buf := stack.buf
	if err := tt.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
	if t, ok := obj.(*UsedType); ok && t.depth == 0 {
		alias, ok := t.node.Annotations().get("next").get(tc.lang + "_alias").(string)
		if ok && alias != "" {
			return alias, nil
		}
	}
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
	if len(tc.trace) == 0 {
		return "", fmt.Errorf("super not allowed in the template")
	}
	name := tc.trace[len(tc.trace)-1]
	tc.compiler.Trace("super template %q", name)
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
	if len(tc.trace) >= tc.maxStack {
		return "", fmt.Errorf("npl: stack overflow")
	}
	tt, err := tc.lookupTemplate(names)
	if err != nil {
		return "", err
	}
	tc.trace = append(tc.trace, tt.Name())
	tc.compiler.Trace("rendering template %q", tt.Name())
	defer func() {
		tc.trace = tc.trace[:len(tc.trace)-1]
	}()
	return tc.execute(tt, tc.pc().pwd, data)
}

func (tc *templateContext) lastLine() string {
	buf := tc.pc().buf
	content := buf.Bytes()
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

	return string(content[lastLineBegin:])
}

func (tc *templateContext) align(text string) string {
	lastLine := tc.lastLine()
	indent := strings.Map(func(r rune) rune {
		if r != ' ' && r != '\t' {
			return ' '
		}
		return r
	}, lastLine)
	return tc.alignWithIndent(indent, text)
}

func (tc *templateContext) alignWithIndent(indent, text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] != "" {
			lines[i] = indent + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
