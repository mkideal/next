package types

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// TemplateMeta represents the meta data of a template.
// It is used to generate the header of the generated file.
type TemplateMeta map[string]string

// Get returns the value of the given key.
func (m TemplateMeta) Get(key string) string {
	return m[key]
}

// parseMetadata parses the metadata from the content of a template.
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
func parseMetadata(content string) (TemplateMeta, error) {
	scanner := bufio.NewScanner(bytes.NewBufferString(content))

	if !scanner.Scan() {
		return nil, nil
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "{{/*" {
		return nil, nil
	}

	meta := make(TemplateMeta)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "*/}}" {
			break
		}

		key, value, err := parseMetaLine(line)
		if err != nil {
			return nil, err
		}
		if key != "" {
			if _, exists := meta[key]; exists {
				return nil, fmt.Errorf("duplicate meta key %q", key)
			}
			meta[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning template: %w", err)
	}

	return meta, nil
}

func parseMetaLine(line string) (string, string, error) {
	if line == "" || strings.HasPrefix(line, "//") {
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

// loadTemplate loads a template from a file.
func loadTemplate(path string) (*template.Template, TemplateMeta, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read template file %q: %v", path, err)
	}
	return createTemplate(path, string(content), true)
}

// executeTemplate executes a template content with the given data.
func executeTemplate(name, content string, data any) (string, error) {
	tpl, _, err := createTemplate(name, content, false)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %v", name, err)
	}
	return buf.String(), nil
}

// createTemplate creates a new template from the given content.
// If hasMeta is true, the content is expected to contain metadata.
func createTemplate(name, content string, hasMeta bool) (*template.Template, TemplateMeta, error) {
	var metadata TemplateMeta
	var err error
	if hasMeta {
		metadata, err = parseMetadata(content)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse metadata for template %q: %w", name, err)
		}
	}

	t, err := template.New(name).Parse(content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse template %q: %w", name, err)
	}

	return t, metadata, nil
}

// template functions

// TemplateMeta is a map of key-value pairs used in templates.
type TemplateData[T Object] struct {
	T    T
	Lang string
	Meta TemplateMeta

	context *Context
	types   sync.Map
}

func (d *TemplateData[T]) TypeOf(t Type) (string, error) {
	if v, ok := d.types.Load(t); ok {
		return v.(string), nil
	}
	v, err := resolveLangType(d.context.flags.types, d.Lang, t)
	if err != nil {
		return "", err
	}
	d.types.Store(t, v)
	return v, nil
}
