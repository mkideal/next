---
sidebar_position: 1
---

# Language Guide

## Language Syntax

### Lexical Elements

Next uses several common lexical elements:

- **Comments**: 
  - Line comments: `// This is a line comment`
  - Block comments: `/* This is a block comment */`
- **Identifiers**: Names must start with a letter and can contain letters and numbers.
- **Keywords**: Reserved words that have special meaning in Next.

### Keywords

Next reserves the following keywords:

```
package   import    const     enum      struct    interface
map       vector    array     bool      int       int8
int16     int32     int64     float32   float64   string
byte      bytes     any       true      false
```

:::caution
Avoid using these keywords as identifiers in your Next definitions.
:::

### Operators

Next supports various operators for different operations:

| Category | Operators |
|----------|-----------|
| Arithmetic | `+`, `-`, `*`, `/`, `%` |
| Logical | `&&`, `\|\|`, `!` |
| Comparison | `==`, `!=`, `<`, `>`, `<=`, `>=` |
| Bitwise | `&`, `\|`, `^`, `<<`, `>>`, `&^` |

## Data Types

Next offers a rich set of data types to model various kinds of information effectively.

### Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `bool` | Boolean value | `true`, `false` |
| `int`, `int8`, `int16`, `int32`, `int64` | Signed integers | `42`, `-7` |
| `float32`, `float64` | Floating-point numbers | `3.14`, `-0.01` |
| `string` | UTF-8 encoded string | `"Hello, World!"` |
| `byte`, `bytes` | Byte and byte arrays | `0x0A`, `[]byte{0x0A, 0x0B}` |
| `any` | Any type | - |

### Composite Types

Next also provides several composite types for more complex data structures:

1. **Array**: Fixed-size arrays
    ```next
    array<int, 5> scores;
    ```

2. **Vector**: Dynamic-size arrays
    ```next
    vector<string> names;
    ```

3. **Map**: Key-value pairs
    ```next
    map<string, int> ages;
    ```

4. **Enum**: Named constants
    ```next
    enum Color {
        Red = 1;
        Green = 2;
        Blue = 3;
    }
    ```

5. **Struct**: User-defined types
    ```next
    struct Person {
        string name;
        int age;
    }
    ```

6. **Interface**: Method signatures
    ```next
    interface Reader {
        read(bytes buffer) int;
    }
    ```

## Enumerations

Enums in Next allow you to define a set of named constants. They can be particularly useful for representing a fixed set of values, such as status codes or configuration options.

### Basic Enum Definition

```next
enum Status {
    Pending = 1;
    Approved = 2;
    Rejected = 3;
}
```

### Using `iota`

Next supports the `iota` keyword for auto-incrementing enum values:

```next
enum Color {
    Red = iota;    // 0
    Green;         // 1
    Blue;          // 2
}
```

You can also use `iota` with expressions:

```next
enum FilePermission {
    Read = 1 << iota;  // 1
    Write;             // 2
    Execute;           // 4
}
```

### Non-Integer Enums

Next also supports floating-point and string enums:

```next
enum MathConstants {
    Pi = 3.14159265358979323846;
    E = 2.71828182845904523536;
}

enum OperatingSystem {
    Windows = "windows";
    Linux = "linux";
    MacOS = "macos";
}
```

:::caution NOTE
When using non-integer enums, be aware that not all target languages may support them natively. Ensure your templates handle these cases appropriately.
:::

## Structs

Structs in Next allow you to define complex data types by grouping related fields together.

### Basic Struct Definition

```next
struct User {
    int64 id;
    string username;
    string email;
}
```

### Nested Structs

You can nest structs within each other:

```next
struct Address {
    string street;
    string city;
    string country;
}

struct Employee {
    User user;
    Address address;
    float64 salary;
}
```

### Using Annotations with Structs

Annotations can be used to provide additional metadata for structs and their fields:

```next
@message(type=101)
struct LoginRequest {
    string username;
    @sensitive string password;
    @next(optional) string twoFactorToken;
}
```

## Interfaces

Interfaces in Next define a set of method signatures that can be implemented by structs.

### Basic Interface Definition

```next
interface Shape {
    area() float64;
    perimeter() float64;
}
```

### Interface with Error Handling

You can use annotations to indicate error handling in interfaces:

```next
interface FileSystem {
    @next(error)
    read(string path) bytes;
    
    @next(error)
    write(string path, bytes data);
}
```

## Annotations

Annotations in Next provide a powerful way to add metadata to your definitions. This metadata can be used to control code generation, add validation rules, or provide additional information for documentation.

### Built-in Annotations

Next includes `@next` as a built-in annotation, which is used to provide language-specific information:

```next
@next(
    go_package="github.com/example/project",
    cpp_package="example::project",
    java_package="com.example.project"
)
package myproject;
```

### Custom Annotations

You can define custom annotations for various purposes:

```next
@api(version="v1")
@authorize(roles="admin,user")
struct UserProfile {
    @validate(min=3, max=50)
    string username;
    
    @validate(pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
    string email;
}
```

:::tip
Custom annotations are processed by annotation solvers or within templates. Ensure you have the necessary logic to handle these annotations in your code generation process.
:::

## Template System

Next uses a powerful template system based on Go's `text/template` package. This system allows for complex logic and text manipulation during code generation.

### Basic Template Structure

```npl
{{- define "go/struct" -}}
type {{next .Type}} struct {
    {{- next .Fields}}
}
{{- end}}
```

### Template Inheritance

Next supports template inheritance, allowing you to extend or override base templates:

```npl
{{- define "cpp/struct" -}}
class {{next .Type}} {
public:
    {{next .Fields}}
};
{{- end}}
```

### Using `super` for Template Extension

The `super` keyword allows you to call the parent template within an overridden template:

```npl
{{- define "go/struct" -}}
{{super .}}
{{- with .Annotations.message.type}}

func ({{next $.Type}}) MessageType() int { return {{.}} }
{{- end}}
{{- end}}
```

## Command Line Interface

The Next command-line interface provides various options to control the code generation process.

### Basic Usage

```bash
next [options] [source_files...]
```

### Common Options

- `-O LANG=DIR`: Set output directory for a specific language
- `-M LANG.KEY=VALUE`: Define type mappings
- `-T LANG=PATH`: Specify custom templates
- `-X ANNOTATION=PROGRAM`: Define custom annotation solvers
- `-v LEVEL`: Set verbosity level (`0`=error, `1`=info, `2`=debug, `3`=trace)

### Example

```bash
next \
    -O go=./output/go \
    -M cpp.map="std::map<%K%,%V%>" \
    -O go=./gen/go -T go=./templates/go/custom.go.npl \
    -O cpp=./gen/cpp -T cpp=./templates/cpp/custom.cpp.npl \
    -X validate="./validate_solver" \
    -v 2
```

## Best Practices

To make the most of Next, consider the following best practices:

1. **Organize Your Next Files**: Group related definitions together and use meaningful file names.

2. **Use Annotations Effectively**: Leverage annotations to provide language-specific information and customize code generation.

3. **Create Modular Templates**: Break down your templates into reusable components for better maintainability.

4. **Leverage Type Mappings**: Use type mappings to ensure your generated code follows language-specific conventions.

5. **Implement Custom Annotation Solvers**: For complex annotations, create custom solvers to handle the logic separately from your templates.

6. **Version Control Your Templates**: Keep your templates and Next definitions under version control to track changes over time.

7. **Test Generated Code**: Regularly test the generated code across all target languages to ensure consistency and correctness.

8. **Document Your Annotations**: Provide clear documentation for any custom annotations you create, including their purpose and expected usage.

9. **Use Consistent Naming Conventions**: Adopt a consistent naming style across your Next definitions to improve readability and maintain consistency in generated code.

10. **Optimize for Performance**: When dealing with large projects, optimize your templates and annotation solvers for performance to reduce code generation time.

## Conclusion

Next provides a powerful and flexible way to generate code across multiple programming languages. By mastering its syntax, template system, and advanced features, you can significantly streamline your development process and ensure consistency across different parts of your project.

Remember that while Next offers many capabilities, the key to successful usage lies in understanding your project's specific needs and leveraging Next's features accordingly. Happy coding!