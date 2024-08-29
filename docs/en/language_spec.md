# Next Language Specification

## 1. Introduction

Next is a language designed for generating code in other languages or various types of files. This document defines the Next language and outlines its syntax and semantics.

## 2. Lexical Elements

### 2.1 Comments

Next supports two forms of comments:

1. Line comments start with `//` and continue until the end of the line.
2. General comments start with `/*` and end with `*/`.

### 2.2 Identifiers

Identifiers name program entities such as variables and types. An identifier is a sequence of one or more letters and digits. The first character in an identifier must be a letter.

```
identifier = letter { letter | unicode_digit } .
```

### 2.3 Keywords

The following keywords are reserved and may not be used as identifiers:

```
package   import    const     enum      struct
```

### 2.4 Operators and Punctuation

```
+    &    &&   =    !=   (    )
-    |    ||   <<   <=   [    ]
*    !    ==   >>   >=   {    }
/    ,    ^    ;    !    <    >
%    .    &^
```

**Note**: Square brackets `[]` are currently unused but reserved for potential future syntax extensions.

## 3. Source Code Representation

### 3.1 Source Files

Source files are encoded in UTF-8. The file extension is `.next`.

### 3.2 Package Declaration

Every Next source file begins with a package declaration:

```
PackageClause = "package" PackageName ";" .
PackageName    = identifier .
```

Example:
```next
package demo;
```

### 3.3 Import Declaration

Import declarations are used to include other files:

```
ImportDecl       = "import" ImportSpec ";" .
ImportSpec       = string_lit .
```

Example:

```next
import "./a.next";
```

After importing a file, you can use constants, enums, structs, protocols, etc., defined in that file's package.

## 4. Declarations and Scope

### 4.1 Constants

Constant declarations use the `const` keyword:

```
ConstDecl      = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
ConstSpec      = identifier "=" Expression ";" .
```

Constants can be numbers, strings, booleans, or any constant expressions.

Example:

```next
const V0 = 1;
const V1 = 100.5;
const V2 = 1000_000; // equivalent to 1000000
const V3 = V1 + V2; // expression referencing other constants

enum Errno {
    OK = 0,
}

const (
    A = 1;
    B = 2.0;
    C = false;
    D = "hello";
    E = Errno.OK;  // enum field reference
)
```

### 4.2 Types

#### 4.2.1 Enum Types

Enum declarations use the `enum` keyword:

```
EnumDecl     = "enum" ( EnumSpec | "(" { EnumSpec ";" } ")" ) .
EnumSpec     = identifier "{" { identifier [ "=" Expression ] "," } "}" .
```

Enums can use expressions containing `iota` for derivation. **Note that only enums can use iota derivation; const definitions cannot.**

Example:

```next
enum Color {
    Red = 1,
    Green = 2,
    Blue = 3,
}

enum Errno {
    OK = iota,  // 0
    Internal,   // 1
    BadRequest, // 2

    UserNotFound = iota + 100, // 100
    ProviderNotFound,          // 101
}

enum (
    EnumA {
        A1 = 0,
        A2 = 1,
    }

    EnumB {
        B1 = 0,
        B2 = 1,
    }
)
```

Enum fields can be referenced using `EnumName.FieldName` syntax and can be used in constant definitions and constant expressions.

#### 4.2.2 Struct Types

Struct declarations use the `struct` keyword:

```
StructDecl     = "struct" ( StructSpec | "(" { StructSpec ";" } ")" ) .
StructSpec     = identifier "{" { FieldDecl ";" } "}" .
FieldDecl      = identifier Type .
```

Example:

```next
struct Location {
    string country;
    string city;
    int zipCode;
}

struct (
    StructA {
        int field1;
        bool field2;
    }

    StructB {
        StructA a;
        vector<StructA> as;
        string field1;
        int field2;
    }
)
```

### 4.3 Annotations

Annotations can be added to packages, any declarations, constants, enums (and their fields), structs (and their fields), and protocols (and their fields). Annotations start with the `@` symbol:

```
Annotation     = "@" identifier [ "(" [ Parameters ] ")" ] ";" .
Parameters     = PositionalParams | NamedParams .
PositionalParams = Expression { "," Expression } .
NamedParams    = NamedParam { "," NamedParam } .
NamedParam     = identifier "=" Expression .
```

Annotations support forms with no parameters, with parameters (including named and anonymous parameters).

Example:

```next
@next(
    go_package = "github.com/username/repo/a",
    cpp_namespace = "repo::a", // trailing comma is optional
)
package demo;

@protocol(type=100)
struct LoginRequest {
    @required
    string token;
    string ip;
}

@json(omitempty)
struct User {
    @key
    int id;

    @json(name="nick_name")
    string nickname;

    @json(ignore)
    string password;
}
```

**@next is a built-in annotation in Next. Its usage follows specific parameter definitions, and its effects are interpreted internally by Next. @next should not be used as a custom annotation.**

## 5. Types

### 5.1 Primitive Types

- Boolean type: `bool`
- Integer types: `int`, `int8`, `int16`, `int32`, `int64`
- Floating-point types: `float32`, `float64`
- String types: `string`, `byte`, `bytes`

**Note: Unsigned integer types are not supported.**

### 5.2 Composite Types

- Array type: `array<T, N>`, where T is the element type and N is the array length
- Vector type: `vector<T>`, where T is the element type
- Map type: `map<K, V>`, where K is the key type and V is the value type

## 6. Properties and Fields

Field declarations in structs and protocols follow this syntax:

```
FieldDecl = identifier Type ";" .
```

Fields can have annotations.

## 7. Expressions

In the Next language, all expressions are constant expressions evaluated at compile-time. Expressions are used to compute values and follow this syntax:

```
Expression     = UnaryExpr | Expression binary_op Expression | FunctionCall .
UnaryExpr      = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr    = Operand | PrimaryExpr Selector .
Operand        = Literal | identifier | EnumFieldRef | "(" Expression ")" .
Selector       = "." identifier .
EnumFieldRef   = identifier "." identifier .
FunctionCall   = identifier "(" [ ArgumentList ] ")" .
ArgumentList   = Expression { "," Expression } .

binary_op     = "+" | "-" | "*" | "/" | "%" |
                "&" | "|" | "^" | ">>" | "<<" | "&^" |
                "<" | ">" | "<=" | ">=" | "==" | "!=" | "&&" | "||" .
unary_op      = "+" | "-" | "!" | "^" .
```

Expressions can include:

1. Literals (numbers, strings, booleans)
2. Identifiers (constant names, enum names, etc.)
3. Enum field references
4. Binary operations (arithmetic, bitwise, logical, comparison operations, etc.)
5. Unary operations (plus/minus signs, logical not, bitwise not, etc.)
6. Parenthesized grouping
7. Field selection
8. Function calls (currently only built-in functions are supported)

All expressions are constant expressions evaluated at compile-time. They can contain literals, constant identifiers, enum field references, and operations on these.

The Next language's expressions do not support indexing operations.

Example:

```next
42                  // Numeric literal
"hello"             // String literal
true                // Boolean literal
X                   // Constant identifier
Color.Red           // Enum field reference
x + y               // Binary addition
-z                  // Unary minus
(x + y) * z         // Parenthesized expression
min(x, y, z)        // Function call
max(a, b)           // Function call
len("hello")        // Function call
```

All expressions used in constant declarations must be valid constant expressions that can be evaluated at compile-time.

## 8. Built-in Variables and Functions

| Function or Variable | Usage Description |
|----------------------|-------------------|
| **iota** | Used in enum declarations to declare built-in incrementing variable |
| **int**(`x: bool\|int\|float`) | Convert bool, int, or float to integer |
| **float**(`x: bool\|int\|float`) | Convert bool, int, or float to floating-point number |
| **bool**(`x: any`) | Convert bool, int, float, or string to bool |
| **min**(`x: any`, `y: any...`) | Get the minimum value of one or more values |
| **max**(`x: any`, `y: any...`) | Get the maximum value of one or more values |
| **abs**(`x: int\|float`) | Get the absolute value |
| **len**(`s: string`) | Calculate the length of a string |
| **sprint**(`args: any...`) | Output arguments as a string, if all arguments are strings, separate them with spaces |
| **sprintf**(`fmt: string`, `args: any...`) | Formatted string output |
| **sprintln**(`args: any...`) | Similar to `sprint`, but adds a newline at the end |
| **print**(`args: any...`) | Debug output information, automatically adds a newline if content doesn't have one |
| **printf**(`fmt: string`, `args: any...`) | Debug output formatted information, automatically adds a newline if content doesn't have one |
| **error**(`args: any...`) | Output error message, requires at least one argument |
| **assert**(`cond: bool`, `args: any...`) | Assert if true |
| **assert_eq**(`got: any`, `want: any`, `args: any...`) | Assert if equal |
| **assert_ne**(`got: any`, `want: any`, `args: any...`) | Assert if not equal |
| **assert_lt**(`x: any`, `y: any`, `args: any...`) | Assert if `x` is less than `y` |
| **assert_le**(`x: any`, `y: any`, `args: any...`) | Assert if `x` is less than or equal to `y` |
| **assert_gt**(`x: any`, `y: any`, `args: any...`) | Assert if `x` is greater than `y` |
| **assert_ge**(`x: any`, `y: any`, `args: any...`) | Assert if `x` is greater than or equal to `y` |

## 9. Lexical Conventions

- Package names use lower camel case (all lowercase is recommended).
- Constants, enum members, and type names (structs, enums, protocols) use upper camel case.