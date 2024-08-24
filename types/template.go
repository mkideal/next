package types

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gopherd/next/internal/templateutil"
)

// TemplateMeta represents the meta data of a template.
// It is used to generate the header of the generated file.
type TemplateMeta map[string]string

// Get returns the value of the given key.
func (m TemplateMeta) Get(key string) string {
	return m[key]
}

// parseMetadata extracts metadata from the content and returns the parsed metadata
// along with the modified content (with a simple metadata placeholder).
//
// The metadata is expected to be in the following format:
// {{/*
// name: <name>
// ignore: <true|false>
// authors: <author1>,<author2>,...
// date: <date>
// */}}
//
// Example:
// {{/*
// name: {{.Name}}.go
// ignore: {{eq "Test" .Name}}
// authors: gopherd, next
// date: 2024-01-01
// */}}
// parseMetadata extracts metadata from the content and returns the parsed metadata
// along with the modified content (with metadata replaced by placeholder and empty lines).
func parseMetadata(content string) (TemplateMeta, string, error) {
	startIndex := strings.Index(content, "{{/*")
	endIndex := strings.Index(content, "*/}}")

	if startIndex == -1 || endIndex == -1 || endIndex < startIndex {
		return nil, content, nil
	}

	metaContent := content[startIndex+4 : endIndex]
	meta := make(TemplateMeta)

	// Parse metadata
	lines := strings.Split(metaContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
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
			meta[key] = value
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

// parseTemplateFilename extracts the extension and object type from a template filename.
// Format: [[<name>.]<ext>.]<object_type>.nxp
// Examples: struct.nxp, hpp.struct.nxp, demo.hpp.struct.nxp
func parseTemplateFilename(filename string) (extension, objectType string, err error) {
	filename = filepath.Base(filename)
	parts := strings.Split(filename, ".")

	if len(parts) < 2 || parts[len(parts)-1] != templateExt[1:] {
		return "", "", fmt.Errorf("invalid filename %q: must end with %s", filename, templateExt)
	}

	// Remove the .nxp part
	parts = parts[:len(parts)-1]

	objectType = parts[len(parts)-1]
	if len(parts) > 1 {
		extension = parts[len(parts)-2]
	}

	switch objectType {
	case "package", "file", "const", "enum", "struct":
		return extension, objectType, nil
	default:
		return "", "", fmt.Errorf("invalid object type %q in %q, expected: package, file, struct, enum, or const", objectType, filename)
	}
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

// TemplateMeta is a map of key-value pairs used in templates.
type TemplateData[T Object] struct {
	context *Context
	obj     T
	dir     string
	ext     string
	lang    string

	types sync.Map

	Meta TemplateMeta
}

func (d *TemplateData[T]) funcs() template.FuncMap {
	return template.FuncMap{
		"this":   d.this,
		"typeof": d.typeof,
	}
}

func (d *TemplateData[T]) this() T {
	return d.obj
}

func (d *TemplateData[T]) typeof(t Type) (string, error) {
	if v, ok := d.types.Load(t); ok {
		return v.(string), nil
	}
	v, err := resolveLangType(d.context.flags.types, d.lang, t)
	if err != nil {
		return "", err
	}
	d.types.Store(t, v)
	return v, nil
}
