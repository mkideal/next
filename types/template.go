package types

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// TemplateMeta represents the meta data of a template.
// It is used to generate the header of the generated file.
//
// Example:
/*
---
name: {{.Name}}.go
ignore: {{eq "Test" .Name}}
authors: gopherd, next
date: 2024-01-01
---
*/
type TemplateMeta struct {
	Name    string   // generated file name
	Ignore  bool     // ignore this template
	Authors []string // authors of the template
	Date    string   // date of the template

	raw map[string]string // raw template meta data content
}

func (m *TemplateMeta) setValue(key, value string) error {
	switch key {
	case "name":
		m.Name = value
	case "ignore":
		m.Ignore = value == "true"
	case "authors":
		m.Authors = strings.Split(value, ",")
	case "date":
		m.Date = value
	}
	return nil
}

type Template interface {
	Execute(w io.Writer, data any) error
	GetMeta() map[string]string
	GetTemplate(name string) (Template, error)
}

func loadTemplate(engine, path string) (Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %q: %v", path, err)
	}
	return createTemplate(engine, string(content))
}

func executeTemplate(engine, name, content string, data any) (string, error) {
	tpl, err := createTemplate(engine, content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %v", name, err)
	}
	return buf.String(), nil
}

func createTemplate(engine, content string) (Template, error) {
	// TODO
	return nil, nil
}

// tplFile format: [<name>.]<node_type>.jet
func nodeTypeOfTemplateFile(tplFile string) (string, error) {
	parts := strings.Split(tplFile, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid template file name %q: expected [<name>.]<node_type>.jet", tplFile)
	}
	nodeType := parts[len(parts)-2]
	switch nodeType {
	case "package", "file", "struct", "protocol", "enum", "const":
		return nodeType, nil
	}
	return "", fmt.Errorf("invalid node type %q in template file name %q, expected one of [package, file, struct, protocol, enum, const]", nodeType, tplFile)
}
