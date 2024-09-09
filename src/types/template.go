package types

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/op"

	"github.com/next/next/src/fsutil"
	"github.com/next/next/src/templateutil"
)

// metaValue represents a metadata value with the line number where it was defined.
type metaValue[T any] struct {
	content T
	line    int
}

func (m *metaValue[T]) value() T {
	var zero T
	if m == nil {
		return zero
	}
	return m.content
}

// templateMeta represents the meta data of a template.
// It is used to generate the header of the generated file.
type templateMeta[T any] map[string]*metaValue[T]

// Get returns the value of the given key.
func (m templateMeta[T]) Get(key string) *metaValue[T] {
	if m == nil {
		return nil
	}
	return m[key]
}

// resolveMeta resolves the metadata values by executing the metadata templates with the given data.
func resolveMeta[T Decl](metaTemplates templateMeta[*template.Template], data *templateContext[T]) (templateMeta[string], error) {
	if metaTemplates == nil {
		return nil, nil
	}
	meta := make(templateMeta[string])
	for k, t := range metaTemplates {
		v, err := executeTemplate(t.content, data)
		if err != nil {
			return nil, err
		}
		switch k {
		case "overwrite":
			if v != "true" && v != "false" {
				return nil, fmt.Errorf("invalid value %q for meta key %q, expected true or false", v, k)
			}
		}
		meta[k] = &metaValue[string]{
			content: v,
			line:    t.line,
		}
	}
	return meta, nil
}

// parseMeta extracts metadata from the content and returns the parsed metadata
// along with the modified content (with a simple metadata placeholder).
//
// The metadata is expected to be in the following format:
//
// {{/*
// # 'this' represents the type of the object to be generated,
// # default is 'file'
// this: [file|const|enum|struct|interface]
//
// # 'path' represents the output path of the generated file,
// # default is the object name with the current extension
// path: [relative/path/to/generated/file|/absolute/path/to/generated/file]
//
// # 'skip' represents whether to skip generating the file,
// # default is false
// skip: [true|false]
//
// # 'overwrite' represents whether to overwrite the existing file,
// # default is true
// overwrite: [true|false]
//
// # and other custom metadata key-value pairs
// # ...
// */}}
//
// Example:
// {{/*
// this: file
// path: {{this.Package.Name}}/{{this.Name}}.next.go
// skip: {{eq "Test" this.Name}}
// */}}
// parseMeta extracts metadata from the content and returns the parsed metadata
// along with the modified content (with metadata replaced by placeholder and empty lines).
func parseMeta(content string) (templateMeta[string], string, error) {
	const beginDelim = "{{/*"
	const endDelim = "*/}}"
	beginIndex := strings.Index(content, beginDelim)
	endIndex := strings.Index(content, endDelim)

	if beginIndex == -1 || endIndex == -1 || endIndex < beginIndex {
		return nil, content, nil
	}
	if strings.Index(content[:beginIndex], "\r") != -1 || strings.Index(content[:beginIndex], "\n") != -1 {
		return nil, content, nil
	}

	metaContent := content[beginIndex+len(beginDelim) : endIndex]
	meta := make(templateMeta[string])

	// Parse metadata
	lines := strings.Split(metaContent, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		key, value, err := parseMetaLine(line)
		if err != nil {
			return nil, content, err
		}
		if key != "" {
			if _, exists := meta[key]; exists {
				return nil, content, fmt.Errorf("duplicate meta key %q", key)
			}
			meta[key] = &metaValue[string]{
				content: value,
				line:    i + 1,
			}
		}
	}

	// Replace metadata block with placeholder
	replacement := "{{- /* meta */ -}}" + strings.Repeat("\n", strings.Count(metaContent, "\n"))
	modifiedContent := content[:beginIndex] + replacement + content[endIndex+len(endDelim):]

	return meta, modifiedContent, nil
}

// parseMetaLine parses a single metadata line into a key-value pair.
func parseMetaLine(line string) (string, string, error) {
	if line == "" {
		return "", "", nil
	}
	var key, value string
	index := strings.Index(line, ":")
	if index > 0 {
		key, value = strings.TrimSpace(line[:index]), strings.TrimSpace(line[index+1:])
	}
	if key == "" || value == "" {
		return "", "", fmt.Errorf("invalid meta line %q, expected key: value", line)
	}
	if len(value) >= 2 && value[0] == '"' {
		var err error
		value, err = strconv.Unquote(value)
		if err != nil {
			return "", "", fmt.Errorf("invalid meta value %q: %v", value, err)
		}
	}
	return key, value, nil
}

func createTemplates(file, content string, meta templateMeta[string], funcs template.FuncMap) (*template.Template, templateMeta[*template.Template], error) {
	t, err := createTemplate(file, content, funcs)
	if err != nil {
		return nil, nil, err
	}
	mt := make(templateMeta[*template.Template])
	for k, v := range meta {
		tt, err := createTemplate(k, v.content, funcs)
		if err != nil {
			return nil, nil, err
		}
		mt[k] = &metaValue[*template.Template]{
			content: tt,
			line:    v.line,
		}
	}
	return t, mt, nil
}

// createTemplate creates a new template from the given content.
func createTemplate(name, content string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(name).Funcs(templateutil.Funcs).Funcs(funcs).Parse(content)
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

// templateContext represents the context of a template.
type templateContext[T Decl] struct {
	templateContextInfo

	// buf is used to buffer the generated content.
	buf bytes.Buffer
	// current decl object to be rendered: File, Const, Enum, Struct, Interface
	decl T

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
}

func newTemplateContext[T Decl](ctx templateContextInfo) *templateContext[T] {
	tc := &templateContext[T]{
		templateContextInfo: ctx,
		dontOverrides:       make(map[string]bool),
	}
	tc.funcs = template.FuncMap{
		"ENV":    tc.env,
		"this":   tc.this,
		"lang":   tc.getLang,
		"error":  tc.error,
		"errorf": tc.errorf,
		"head":   tc.head,
		"align":  tc.align,
		"type":   tc.type_,
		"next":   tc.next,
		"super":  tc.super,
		"render": tc.render,
	}
	return tc
}

func (tc *templateContext[T]) init() error {
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

func (tc *templateContext[T]) reset(obj T) {
	tc.decl = obj
	tc.buf.Reset()
}

// @api(template/context): this
// `this` returns the current [object](#Object) to be rendered.
//
// Example:
// > ```
// > {{this.Package.Name}}
// > {{this.Name}}
// > ```
func (tc *templateContext[T]) this() T {
	return tc.decl
}

func (tc *templateContext[T]) env() flags.Map {
	return tc.context.flags.envs
}

func (tc *templateContext[T]) getLang() string {
	return tc.lang
}

func (tc *templateContext[T]) error(msg string) (string, error) {
	return "", errors.New(msg)
}

func (tc *templateContext[T]) errorf(format string, args ...any) (string, error) {
	return "", fmt.Errorf(format, args...)
}

// @api(template/context): type (Type)
// `type` outputs the string representation of the given [type](#Type) for the current language.
//
// Example:
//
// ```
// {{/*
// this: struct
// */}}
//
// {{range this.Fields}}
// {{type .Type}}
// {{end}}
// ```
//
// Output (for c++):
//
// ```
// int
// std::string
// std::map<std::string,int>
// ```
func (tc *templateContext[T]) type_(t any, langs ...string) (string, error) {
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

func (tc *templateContext[T]) resolveLangType(lang string, t any) (result string, err error) {
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

// @api(template/context): head
// `head` outputs the header of the generated file.
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
func (tc *templateContext[T]) head() string {
	p, ok := tc.context.flags.mappings[tc.lang+".comment(%S%)"]
	if !ok {
		return ""
	}

	return strings.ReplaceAll(p, "%S%", `Code generated by "next `+builder.Info().Version+`"; DO NOT EDIT.`)
}

func (tc *templateContext[T]) loadTemplate(fs fs.FS, filename string) (*template.Template, error) {
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

func (tc *templateContext[T]) parseTemplateNames(lang string, name string) []string {
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
func (tc *templateContext[T]) lookupTemplate(names []string) (*template.Template, error) {
	for i := range names {
		if tt := tc.entrypoint.Lookup(names[i]); tt != nil {
			return tt, nil
		}
	}
	return nil, &TemplateNotFoundError{Name: names[0]}
}

// @api(template/context): next (object)
// next executes the next template with the given (object)[#Object].
//
// Example:
//
// ```
// {{/*
// this: struct
// */}}
//
// {{next this}}
//
// {{range this.Fields}}
// {{next .}}
// {{end}}
// ```
func (tc *templateContext[T]) next(obj Object, langs ...string) (string, error) {
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

func (tc *templateContext[T]) nextWithNames(names []string, obj Object) (string, error) {
	result, err := tc.renderWithNames(names, obj)
	if err != nil && IsTemplateNotFoundError(err) {
		if t, ok := obj.(Type); ok {
			return tc.type_(t)
		}
		return "", err
	}
	return result, err
}

// @api(template/context): super (object)
func (tc *templateContext[T]) super(obj Object, langs ...string) (string, error) {
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

// @api(template/context): render (name, data)
// render executes the template with the given name and data.
func (tc *templateContext[T]) render(name string, data any, langs ...string) (result string, err error) {
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

func (tc *templateContext[T]) renderWithNames(names []string, data any) (result string, err error) {
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

func (tc *templateContext[T]) lastLine() string {
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

func (tc *templateContext[T]) lastIndent() string {
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

func (tc *templateContext[T]) align(s string) string {
	return tc.alignWith(s, "")
}

func (tc *templateContext[T]) alignWith(text, prefix string) string {
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
