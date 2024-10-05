# Template Guide

This guide provides a detailed overview of writing and using templates in the Next project, based on the actual `.npl` files provided.

## Introduction to NPL Files

NPL (**N**ext Tem**p**late **L**anguage based on Go's [text/template](https://pkg.go.dev/text/template/)) files are used to define templates for code generation in the Next project. These templates control how different programming constructs are translated into various target languages.

:::tip

See [Builtin Template Files](https://github.com/gopherd/next/tree/main/builtin) for more information.

:::

## Basic Template Structure

A typical npl file consists of the following elements:

1. Metadata definitions (for entrypoint template file via command line flags `-T`)
2. Template definitions or overrides
3. Main content generation

Example:

```npl
{{/* Define metadata in entrypoint template file */}}
{{- define "meta/this"}}file{{end -}}
{{- define "meta/path"}}{{this.Package.Name}}/{{this.Name}}.next.h{{end -}}

{{/* Overrides next/cpp/import template */}}
{{- define "cpp/import" -}}
#include "../{{.File.Package.Name}}/{{.File.Name}}.next.h"
{{- end -}}

{{head}}

{{next this}}
```

## Metadata Definitions

Metadata is crucial for controlling the template behavior. Common metadata includes:

```npl
{{- define "meta/this"}}file{{end -}}
{{- define "meta/path"}}{{this.Package.Name}}/{{this.Name}}.next.h{{end -}}
{{- define "meta/skip"}}{{exist meta.path}}{{end -}}
```

- `meta/this`: Specifies the type of entity being generated (e.g., "file", "struct", "enum", see [API/Context/this](/docs/api/latest/context#user-content-Context_this) for details)
- `meta/path`: Defines the output path for the generated file
- `meta/skip`: Provides a condition to skip generation

## Template Definitions and Overrides

Templates are defined or overridden using the `define` keyword:

```npl
{{- define "next/cpp/struct" -}}
{{next .Doc}}class {{next .Type}} {
public:
    {{next .Type}}() = default;
    ~{{next .Type}}() = default;
    {{next .Fields}}
};
{{- end}}
```

:::caution NOTE
Template name starting with `next/` is reserved for builtin language supporting. For example, if you plan to write a extension for a new language `LANG`，you should define templates like `next/LANG/struct`.
:::

To extend an existing template, use the `super` keyword:

```npl
{{- define "cpp/struct" -}}
{{super .}}
{{- with .Annotations.message.type}}

    static int message_type() { return {{.}}; }
{{- end}}
{{- end -}}
```

## Language-Specific Templates

Each supported language typically has its own npl file (e.g., go.npl, cpp.npl, java.npl). These files contain language-specific template definitions.

Example from go.npl:

```npl
{{- define "next/go/file" -}}
package {{.Package.Name}}
{{super . -}}
{{- end}}

{{- define "next/go/struct" -}}
{{next .Doc}}type {{next .Type}} struct {
    {{- next .Fields}}
}
{{- end}}
```

Example from cpp.npl:

```npl
{{- define "next/cpp/file" -}}
#pragma once

{{next .Imports -}}
{{render "file:namespace.begin" . -}}
{{render "file:forward.declarations" . -}}
{{next .Decls -}}
{{render "file:namespace.end" . -}}
{{- end}}
```

## Common Template Patterns

1. File Generation:
    ```npl
    {{- define "next/go/file" -}}
    package {{.Package.Name}}
    {{super . -}}
    {{- end}}
    ```

2. Struct Generation:
    ```npl
    {{- define "next/go/struct" -}}
    {{next .Doc}}type {{next .Type}} struct {
        {{- next .Fields}}
    }
    {{- end}}
    ```

3. Enum Generation:
    ```npl
    {{- define "next/go/enum" -}}
    {{next .Doc}}type {{next .Type}} {{render "enum:member.type" .}}

    const (
    {{- next .Members}}
    )
    {{- end}}
    ```

4. Interface Generation:
    ```npl
    {{- define "next/go/interface" -}}
    {{next .Doc}}type {{next .Type}} interface {
        {{- next .Methods}}
    }
    {{- end}}
    ```

## Advanced Techniques

1. Conditional Logic:
    ```npl
    {{- if .Annotations.message.type}}
    static int message_type() { return {{.Annotations.message.type}}; }
    {{- end}}
    ```

2. Custom Rendering:
    ```npl
    {{render "file:namespace.begin" . -}}
    {{render "file:forward.declarations" . -}}
    ```

3. Nested Templates:
    ```npl
    {{- define "next/go/imports:decl" -}}
    {{_}}
    import "strconv"
    {{- if .File.Annotations.next.go_imports}}
    {{- range (.File.Annotations.next.go_imports | split "," | map (trim | split "." | first | trimPrefix "*") | sort | uniq)}}
    import "{{.}}"
    {{- end}}
    {{- end}}
    {{- end}}
    ```

4. Using Built-in Functions:
    ```npl
    {{.Name | camelCase}}
    {{.Name | snakeCase | upper}}
    ```

## Best Practices

1. Use consistent naming conventions for templates (e.g., `<lang>/<element>`, `next/<lang>/<element>`)
2. Leverage metadata to control template behavior
3. Create modular, reusable template components
4. Use comments to explain complex logic
5. Handle language-specific differences in separate files
6. Use `super` when extending templates to maintain base functionality
7. Utilize built-in functions for string manipulations and other common operations

By following these guidelines and patterns, you can create efficient, maintainable, and powerful templates for the Next project. Remember to always consider the specific requirements of each target language when writing your templates.