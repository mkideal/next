package types

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/next/internal/fsutil"
	"github.com/gopherd/next/internal/templateutil"
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
func resolveMeta[T Object](metaTemplates templateMeta[*template.Template], data *templateData[T]) (templateMeta[string], error) {
	if metaTemplates == nil {
		return nil, nil
	}
	meta := make(templateMeta[string])
	for k, t := range metaTemplates {
		v, err := executeTemplate(t.content, data)
		if err != nil {
			return nil, fmt.Errorf("failed to execute template %q: %v", v, err)
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
// this: [package|file|const|enum|struct]
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
// path: {{this.Package}}/{{this.Name}}.next.go
// skip: {{eq "Test" this.Name}}
// */}}
// parseMeta extracts metadata from the content and returns the parsed metadata
// along with the modified content (with metadata replaced by placeholder and empty lines).
func parseMeta(content string) (templateMeta[string], string, error) {
	startIndex := strings.Index(content, "{{/*")
	endIndex := strings.Index(content, "*/}}")

	if startIndex == -1 || endIndex == -1 || endIndex < startIndex {
		return nil, content, nil
	}

	metaContent := content[startIndex+4 : endIndex]
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

	// If no valid metadata was found, return an error
	if len(meta) == 0 {
		return nil, content, fmt.Errorf("no valid metadata found")
	}

	// Replace metadata block with placeholder
	replacement := "{{- /* meta */ -}}" + strings.Repeat("\n", strings.Count(metaContent, "\n"))
	modifiedContent := content[:startIndex] + replacement + content[endIndex+4:]

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
		return "", fmt.Errorf("failed to execute template: %v", err)
	}
	return buf.String(), nil
}

type templateContext struct {
	context *Context
	lang    string // current language
	dir     string // output directory for current language
	ext     string // file extension for current language
}

// templateData represents the context of a template.
type templateData[T Object] struct {
	templateContext

	// current entrypoint template
	entrypoint *template.Template
	// current object to be rendered: Package, File, ValueSpec, EnumType, StructType
	obj T

	cache struct {
		// cache for resolved types for current language: Type -> string
		types sync.Map
		// cache for loaded templates: filename -> *template.Template
		templates sync.Map
	}
	dontOverrides map[string]bool
	initiated     bool
	funcs         template.FuncMap
}

func newTemplateData[T Object](ctx templateContext) *templateData[T] {
	tc := &templateData[T]{
		templateContext: ctx,
		dontOverrides:   make(map[string]bool),
	}
	tc.funcs = template.FuncMap{
		"this":   tc.this,
		"type":   tc.type_,
		"head":   tc.head,
		"next":   tc.next,
		"render": tc.render,
	}
	return tc
}

func (tc *templateData[T]) init() error {
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
		files, err = fsutil.AppendFiles(files, filepath.Join(dir, templatesDir, tc.lang), templateExt, false)
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
	for _, file := range files {
		if _, err := tc.loadTemplate(file); err != nil {
			return err
		}
	}
	return nil
}

// @api(template/context): this
// `this` returns the current [object](#Object) to be rendered.
//
// Example:
// > ```
// > {{this.Package}}
// > {{this.Name}}
// > ```
func (tc *templateData[T]) this() T {
	return tc.obj
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
func (tc *templateData[T]) type_(t Type) (string, error) {
	if v, ok := tc.cache.types.Load(t); ok {
		return v.(string), nil
	}
	v, err := resolveLangType(tc.context.flags.types, tc.lang, t)
	if err != nil {
		return "", err
	}
	tc.cache.types.Store(t, v)
	return v, nil
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
func (tc *templateData[T]) head() string {
	p, ok := tc.context.flags.types[tc.lang+".comment(%S%)"]
	if !ok {
		return ""
	}

	return strings.ReplaceAll(p, "%S%", `Code generated by "next `+builder.Info().Version+`"; DO NOT EDIT.`)
}

func (tc *templateData[T]) loadTemplate(filename string) (*template.Template, error) {
	if v, ok := tc.cache.templates.Load(filename); ok {
		return v.(*template.Template), nil
	}
	content, err := os.ReadFile(filename)
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

func (tc *templateData[T]) lookupTemplate(name string, withNext bool) (*template.Template, error) {
	if tt := tc.entrypoint.Lookup(name); tt != nil {
		return tt, nil
	}
	if withNext {
		if tt := tc.entrypoint.Lookup("next." + name); tt != nil {
			return tt, nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

// @api(template/context): next (value, options...)
// next executes the next template with the given object and [options](#Options) as key=value pairs.
//
// Example:
//
// ```
// {{/*
// this: struct
// */}}
//
// {{next this "key1=value1" "key2=value2"}}
//
// {{range this.Fields}}
// {{next .}}
// {{end}}
// ```
func (tc *templateData[T]) next(value Node, options ...string) (result string, err error) {
	return tc.render(fmt.Sprintf("%s.%s", tc.lang, value.nodeType()), value, options...)
}

func (tc *templateData[T]) render(name string, data any, options ...string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	tt, err := tc.lookupTemplate(name, true)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := tt.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %v", name, err)
	}
	return sb.String(), nil
}
