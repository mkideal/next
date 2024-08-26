package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"

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
func resolveMeta[T Object](metaTemplates templateMeta[*template.Template], data *templateContext[T]) (templateMeta[string], error) {
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

// templateContext represents the context of a template.
type templateContext[T Object] struct {
	context *Context

	lang string // current language
	dir  string // output directory for current language
	ext  string // file extension for current language

	// current object to be rendered: Package, File, ValueSpec, EnumType, StructType
	obj T

	// cache for resolved types for current language: Type -> string
	types sync.Map
}

func (tc *templateContext[T]) funcs() template.FuncMap {
	return template.FuncMap{
		"this":   tc.this,
		"typeof": tc.typeof,
		"head":   tc.head,
		"next":   tc.next,
	}
}

// @api(context): this
// `this` returns the current [object](#Object) to be rendered.
//
// Example:
// > ```
// > {{this.Package}}
// > {{this.Name}}
// > ```
func (tc *templateContext[T]) this() T {
	return tc.obj
}

// @api(context): typeof (Type)
// `typeof` outputs the string representation of the given [type](#Type) for the current language.
//
// Example:
//
// ```
// {{/*
// this: struct
// */}}
//
// {{range this.Fields}}
// {{typeof .Type}}
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
func (tc *templateContext[T]) typeof(t Type) (string, error) {
	if v, ok := tc.types.Load(t); ok {
		return v.(string), nil
	}
	v, err := resolveLangType(tc.context.flags.types, tc.lang, t)
	if err != nil {
		return "", err
	}
	tc.types.Store(t, v)
	return v, nil
}

// @api(context): head
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
// // Auto-generated by "next", DO NOT EDIT.
// ```
//
// Output (for c):
// ```
// /* Auto-generated by "next", DO NOT EDIT. */
// ```
func (tc *templateContext[T]) head() string {
	p, ok := tc.context.flags.types[tc.lang+".comment(%S%)"]
	if !ok {
		return ""
	}
	return strings.ReplaceAll(p, "%S%", `Auto-generated by "next", DO NOT EDIT.`)
}

// @api(context): next (value, options...)
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
func (tc *templateContext[T]) next(value any, options ...string) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	switch v := value.(type) {
	case *Package:
		return tc.nextPackage(v, options...)
	case *File:
		return tc.nextFile(v, options...)
	case *Decl:
		return tc.nextDecl(v, options...)
	case *ValueSpec:
		return tc.nextValueSpec(v, options...)
	case *EnumType:
		return tc.nextEnumType(v, options...)
	case *StructType:
		return tc.nextStructType(v, options...)
	case *Field:
		return tc.nextField(v, options...)
	default:
		return nil, fmt.Errorf("unsupported object type %T", value)
	}
}

func (tc *templateContext[T]) nextPackage(pkg *Package, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextFile(file *File, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextDecl(decl *Decl, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextValueSpec(spec *ValueSpec, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextEnumType(enum *EnumType, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextStructType(strct *StructType, options ...string) (any, error) {
	return nil, nil
}

func (tc *templateContext[T]) nextField(field *Field, options ...string) (any, error) {
	return nil, nil
}
