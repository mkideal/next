---
sidebar_position: 2
---

# Annotation Guide

Annotations in Next provide a powerful mechanism to customize code generation, add metadata to declarations, and control language-specific behavior. This comprehensive guide covers the syntax, usage, and best practices for annotations in Next, with numerous examples to illustrate each concept.

## Annotation Syntax

Annotations in Next start with the `@` symbol and can be applied to packages, declarations, constants, enums (and their fields), structs (and their fields), and interfaces (and their methods).

The general syntax for annotations is:

```next
@annotation_name[(parameter1[=value1], parameter2[=value2], ...)]
```

Annotations can take three forms:

1. No parameters: `@annotation_name`
2. With parameters: `@annotation_name(param1=value1, param2=value2)`
3. With boolean parameters (no value implies `true`): `@annotation_name(param1, param2=false)`

Examples:

```next
@deprecated
@json(name="user_id")
@validate(minLength=3, maxLength=50)
@auth(required)  // Equivalent to @auth(required=true)
```

:::tip
For boolean parameters, `@annotation(param)` is equivalent to `@annotation(param=true)`.
:::

## Built-in Annotations

### @next

The `@next` annotation is a built-in, reserved annotation in Next. It's used to pass information to the Next compiler and should not be used as a custom annotation.

#### Package-level @next

The `@next` annotation at the package level allows you to specify language-specific package information. Each language has its own parameter, all using the `_package` suffix:

```next
@next(
    go_package="github.com/username/repo/demo",
    cpp_package="repo::demo",
    c_package="DEMO_",
    java_package="com.example.demo",
    csharp_package="Company.Demo",
    go_imports="fmt.Printf, *net/http.Client, *encoding/json.Decoder"
)
package demo;
```

Let's break down each `*_package` parameter:

1. `go_package`: 
   - Specifies the full import path for the Go package.
   - Example: `"github.com/username/repo/demo"`

2. `cpp_package`: 
   - Defines the C++ namespace for the generated code.
   - Example: `"repo::demo"` will generate code in the `repo::demo` namespace.

3. `c_package`: 
   - Provides a prefix for C code generation, typically used for function and type names.
   - Example: `"DEMO_"` might generate functions like `DEMO_Function()`.

4. `java_package`: 
   - Sets the Java package for the generated code.
   - Example: `"com.example.demo"` will place Java classes in the `com.example.demo` package.

5. `csharp_package`: 
   - Defines the C# namespace for the generated code.
   - Example: `"Company.Demo"` will place C# classes in the `Company.Demo` namespace.

Additional parameters:

- **go_imports**: Specifies additional imports for Go (see [go_imports](#go_imports) section for details).

Examples of usage in different languages:

```go
// Go
package demo

// The actual import path "github.com/username/repo/demo" is specified in the build system or go.mod file

// ...
```

```cpp
// C++
namespace repo {
namespace demo {
    // ...
}
}
```

```c
// C
PREFIX_Function();
// ...
```

```java
// Java
package com.example.demo;

// ...
```

```csharp
// C#
namespace Company.Demo
{
    // ...
}
```

:::tip
Use these language-specific package parameters to ensure your generated code fits well with the target language's package or namespace conventions.
:::

:::note
The exact effect of these parameters may depend on the code generation templates used. Always refer to your project's specific code generation rules for the most accurate information.
:::

##### go_imports

The `go_imports` parameter in the `@next` annotation for packages specifies additional imports for Go:

```next
@next(
    go_package="github.com/username/repo/demo",
    go_imports="fmt.Printf, *net/http.Client, *encoding/json.Decoder, time.Now, *context.Context"
)
package demo;
```

The `go_imports` parameter accepts a comma-separated list of import paths and items. Use the following syntax:

- For functions, variables, or constants: `pkg.Item`
- For types: `*pkg.Type` (note the `*` prefix for types)

:::caution
When specifying types in `go_imports`, always prefix with `*` for types to generate correct unused declarations.
:::

This will generate the following in your Go file:

```go
import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

var _ = fmt.Printf
var _ = (*http.Client)(nil)
var _ = (*json.Decoder)(nil)
var _ = time.Now
var _ = (*context.Context)(nil)
```

#### Enum-level @next

```next
@next(type=int8)
enum Color {
    Red = 1;
    Green = 2;
    Blue = 3;
}
```

#### Struct and Interface-level @next

```next
@next(
    go_alias="net/http.Handler",
    java_alias="java.util.function.Function<com.sun.net.httpserver.HttpExchange, String>",
    cpp_alias="std::function<void(const HttpRequest&, HttpResponse&)>",
    csharp_alias="System.Func<HttpContext, Task>"
)
interface HTTPHandler {}

@next(
    go_alias="time.Time",
    java_alias="java.time.Instant",
    cpp_alias="std::chrono::system_clock::time_point",
    csharp_alias="System.DateTime"
)
struct Timestamp {
    int64 seconds;
    int32 nanos;
}
```

:::important
When a type is aliased using `<lang>_alias`, the Next compiler will not generate code for that type in the specified language. Instead, it will use the aliased type directly.
:::

#### Method-level @next

The `@next` annotation can be applied to interface methods to indicate specific behaviors or characteristics. Here are the key parameters for method-level `@next` annotations:

1. `error`: Indicates that the method may return an error or throw an exception.
2. `mut`: When applied to the method, indicates that the method is mutable (non-const in C++).

When applied to a parameter, `mut` indicates that the parameter is mutable.

Examples:

```next
interface FileSystem {
    @next(error, mut)
    read(@next(mut) bytes buffer) int;

    @next(error)
    write(string filename, bytes data) bool;

    @next(mut)
    seek(int offset, int whence) int;
}

interface Database {
    @next(error)
    query(string sql) Result;

    @next(error, mut)
    execute(@next(mut) Transaction tx) bool;
}
```

In these examples:

- `read` method:
  - May return an error (`error`)
  - Is mutable (`mut`)
  - Has a mutable `buffer` parameter

- `write` method:
  - May return an error (`error`)
  - Is immutable (no `mut`)

- `seek` method:
  - Is mutable (`mut`)
  - Does not indicate error handling

- `query` method:
  - May return an error (`error`)
  - Is immutable (no `mut`)

- `execute` method:
  - May return an error (`error`)
  - Is mutable (`mut`)
  - Has a mutable `tx` parameter

The effects of these annotations vary by target language:

- In Go:
  - `error` methods will return an additional `error` type
  - `mut` has no effect (Go doesn't have const methods)

- In C++:
  - `error` methods will be declared to throw exceptions
  - `mut` methods will be non-const, `mut` parameters will be non-const references

- In Java:
  - `error` methods will be declared to throw exceptions
  - `mut` has no direct effect but can be used in custom code generation

- In Rust:
  - `error` methods will return a `Result` type
  - `mut` methods will take `&mut self`, `mut` parameters will be `&mut` references

:::tip
Use `error` to clearly indicate methods that can fail, allowing for proper error handling in generated code.
Use `mut` judiciously to express intent about state modification, especially useful in languages with const-correctness like C++.
:::

### available

The `available` parameter in the `@next` annotation controls conditional code generation for specific programming languages:

```next
@next(available="go|java")
interface HTTPServer {
    @next(error)
    handle(string path, HTTPHandler handler);
}

@next(available="!cpp & !java", go_alias="complex128)
struct Complex {
    float64 real;
    float64 imag;
}

@next(available="c | cpp | go")
struct LowLevelBuffer {
    bytes data;
    int size;
}
```

The `available` parameter accepts a boolean expression that can include language identifiers, operators (`&`, `|`, `!`, `(`, `)`), and the literals `true` and `false`.

:::tip
Use `available` to generate language-specific interfaces, structs, or other declarations without needing to maintain separate files for each language. This allows you to write your Next code once and generate appropriate output for multiple target languages.
:::

:::caution
The `available` parameter is strictly for specifying target programming languages. It should not be used for specifying platforms, operating systems, or other non-language conditions.
:::

## Custom Annotations

Custom annotations can be created for specific needs and processed in custom templates or plugins. Here are some examples:

```next
@deprecated(instead="Use NewUser()")
struct User {
    int id;
    @json(omitempty)
    string name;
}

@message(type=101, req)
struct LoginRequest {
    string username;
    string password;
}

@auth(required)
@rate_limit(requests=100, per="minute")
interface UserService {
    getUser(int id) User;
}

@db(table="products", primaryKey="id")
struct Product {
    int id;
    string name;
    float price;
    @validate(min=0, max=100)
    int percentage;
}

@async
@retry(maxAttempts=3, delay="1s")
interface DataFetcher {
    fetchData(string url) bytes;
}
```

## Using Annotations in Templates

Annotations can be accessed and used within templates to customize code generation. Here are some examples:

### Checking for Annotation Presence

```npl
{{- if .Annotations.Contains "deprecated" -}}
// Deprecated: {{.Annotations.deprecated}}
{{- end -}}
```

### Accessing Annotation Parameters

```npl
{{- if .Annotations.Contains "json" -}}
`json:"{{or (.Annotations.json.name) (.Name | camelCase)}}{{if .Annotations.json.omitempty}},omitempty{{end}}"`
{{- end -}}
```

### Handling Aliased Types

```npl
{{- if .Annotations.next.go_alias -}}
{{.Annotations.next.go_alias}}
{{- else -}}
{{.Name}}
{{- end -}}
```

### Processing Custom Annotations

```npl
{{- if .Annotations.Contains "validate" -}}
@Min({{.Annotations.validate.min}})
@Max({{.Annotations.validate.max}})
{{- end -}}

{{- if .Annotations.Contains "db" -}}
@Table(name = "{{.Annotations.db.table}}")
@Id("{{.Annotations.db.primaryKey}}")
{{- end -}}

{{- if .Annotations.Contains "async" -}}
@Async
{{- end -}}
{{- if .Annotations.Contains "retry" -}}
@Retry(maxAttempts = {{.Annotations.retry.maxAttempts}}, delay = {{.Annotations.retry.delay}})
{{- end -}}
```

### Generating Documentation from Annotations

```npl
{{- if .Doc -}}
/**
{{.Doc.Format " * " | align}}
{{- if .Annotations.Contains "deprecated"}}
 * @deprecated {{.Annotations.deprecated}}
{{- end}}
{{- if .Annotations.Contains "param"}}
{{- range $name, $desc := .Annotations.param}}
 * @param {{$name}} {{$desc}}
{{- end}}
{{- end}}
{{- if .Annotations.Contains "return"}}
 * @return {{.Annotations.return}}
{{- end}}
 */
{{- end -}}
```

## Best Practices

1. Use annotations consistently across your project.
2. Leverage language-specific annotations to optimize generated code for each target language.
3. Document custom annotations thoroughly, especially if they're used by other team members or in open-source projects.
4. Use the `@next` annotation judiciously, as it has special meaning in the Next compiler.
5. When using `available`, consider the maintainability of your code across different language targets.
6. For `go_imports`, group related imports together and order them logically.
7. In templates, always check if an annotation exists before trying to access its values to avoid errors.
8. Use custom annotations to encapsulate complex generation logic, making your templates cleaner and more maintainable.
9. When creating custom annotations, choose clear and descriptive names that indicate their purpose.
10. Consider the impact of annotations on code readability and maintain a balance between expressiveness and simplicity.

:::tip
While annotations are powerful, overuse can lead to cluttered and hard-to-read code. Strike a balance between expressiveness and simplicity.
:::