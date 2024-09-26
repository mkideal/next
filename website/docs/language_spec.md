# Next Language Specification

## 1. Introduction

Next is a language designed for generating code in other languages or various types of files. This document defines the Next language and outlines its syntax and semantics.

## 2. Lexical Elements

### 2.1 Comments

Next supports two forms of comments:

1. Line comments start with `//` and continue until the end of the line.
2. General comments start with `/*` and end with `*/`.

### 2.2 Identifiers

Identifiers name program entities such as variables and types. An identifier is a sequence of one or more letters, digits, and underscores. The first character in an identifier must be a letter or an underscore.

```ebnf
identifier = ( letter | "_" ) { letter | unicode_digit | "_" } .
```

### 2.3 Keywords

The following keywords are reserved and may not be used as identifiers:

```
package   import    const     enum      struct    interface
```

### 2.4 Operators and Punctuation

```
+    &    &&   =    !=   (    )
-    |    ||   <<   <=   [    ]
*    !    ==   >>   >=   {    }
/    ,    ^    ;    !    <    >
%    .    &^
```

## 3. Source Code Representation

### 3.1 Source Files

Source files are encoded in UTF-8. The file extension is `.next`.

### 3.2 Package Declaration

Every Next source file begins with a package declaration:

```ebnf
PackageClause = { Annotation } "package" PackageName ";" .
PackageName    = identifier .
```

### 3.3 Import Declaration

Import declarations are used to include other files:

```ebnf
ImportDecl       = "import" ImportSpec ";" .
ImportSpec       = string_lit .
```

## 4. Declarations and Scope

### 4.1 Constants

Constant declarations use the `const` keyword:

```ebnf
ConstDecl      = { Annotation } "const" identifier "=" Expression ";" .
```

Constants can be numbers, strings, booleans, or any constant expressions, including built-in function calls.

### 4.2 Types

#### 4.2.1 Enum Types

Enum declarations use the `enum` keyword:

```ebnf
EnumDecl     = { Annotation } "enum" EnumSpec .
EnumSpec     = identifier "{" { { Annotation } identifier [ "=" Expression ] ";" } "}" .
```

Enums can use expressions containing `iota` for derivation, as well as bitwise operations and other complex expressions.

#### 4.2.2 Struct Types

Struct declarations use the `struct` keyword:

```ebnf
StructDecl     = { Annotation } "struct" StructSpec .
StructSpec     = identifier "{" { FieldDecl ";" } "}" .
FieldDecl      = { Annotation } Type identifier .
```

### 4.3 Annotations

Annotations can be added to packages, declarations, constants, enums (and their fields), structs (and their fields), and interfaces (and their methods). Multiple annotations can be used for each element:

```ebnf
Annotations   = { Annotation } .
Annotation    = "@" identifier [ "(" [ Parameters ] ")" ] .
Parameters    = NamedParam { "," NamedParam } .
NamedParam    = identifier [ "=" ConstantExpression ] .
```

Annotation parameters can only be constant expressions, which include literals, constant identifiers, and expressions composed of these.

## 5. Types

### 5.1 Primitive Types

- Boolean type: `bool`
- Integer types: `int`, `int8`, `int16`, `int32`, `int64`
- Floating-point types: `float32`, `float64`
- String types: `string`, `byte`, `bytes`
- Any: `any`

### 5.2 Composite Types

- Array type: `array<T, N>`, where T is the element type and N is the array length
- Vector type: `vector<T>`, where T is the element type
- Map type: `map<K, V>`, where K is the key type and V is the value type

```ebnf
Type = PrimitiveType | CompositeType .
PrimitiveType = "bool" | "int" | "int8" | "int16" | "int32" | "int64" |
                "float32" | "float64" | "string" | "byte" | "bytes" | "any" .
CompositeType = ArrayType | VectorType | MapType .
ArrayType = "array" "<" Type "," int_lit ">" .
VectorType = "vector" "<" Type ">" .
MapType = "map" "<" Type "," Type ">" .
```

## 6. Expressions

Expressions in Next are constant expressions evaluated at compile-time:

```ebnf
Expression     = UnaryExpr | Expression binary_op Expression | FunctionCall .
UnaryExpr      = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr    = Operand | PrimaryExpr Selector .
Operand        = Literal | QualifiedIdent | "(" Expression ")" .
QualifiedIdent = [ PackageName "." ] identifier .
Selector       = "." identifier .
FunctionCall   = identifier "(" [ ArgumentList ] ")" .
ArgumentList   = Expression { "," Expression } .

PackageName    = identifier .
binary_op      = "+" | "-" | "*" | "/" | "%" |
                 "&" | "|" | "^" | ">>" | "<<" | "&^" |
                 "<" | ">" | "<=" | ">=" | "==" | "!=" | "&&" | "||" .
unary_op       = "+" | "-" | "!" | "^" .
```

This expression syntax allows for package-qualified identifiers (e.g., `demo.X`, `demo.Color.Red`).

## 7. Built-in Variables and Functions

Next provides several built-in variables and functions:

| Function or Variable | Usage Description |
|----------------------|-------------------|
| **iota** | Used in enum declarations to declare built-in incrementing variable |
| **int**(`x: bool\|int\|float`) | Convert bool, int, or float to integer |
| **float**(`x: bool\|int\|float`) | Convert bool, int, or float to floating-point number |
| **bool**(`x`) | Convert bool, int, float, or string to bool |
| **min**(`x, y...`) | Get the minimum value of one or more values |
| **max**(`x, y...`) | Get the maximum value of one or more values |
| **abs**(`x: int\|float`) | Get the absolute value |
| **len**(`s: string`) | Calculate the length of a string |
| **sprint**(`args...`) | Output arguments as a string, if all arguments are strings, separate them with spaces |
| **sprintf**(`fmt, args...`) | Formatted string output |
| **sprintln**(`args...`) | Similar to `sprint`, but adds a newline at the end |
| **print**(`args...`) | Debug output information, automatically adds a newline if content doesn't have one |
| **printf**(`fmt, args...`) | Debug output formatted information, automatically adds a newline if content doesn't have one |
| **error**(`args...`) | Output error message, requires at least one argument |
| **assert**(`cond, args...`) | Assert if true |
| **assert_eq**(`x, y, args...`) | Assert if `x` equals to `y` |
| **assert_ne**(`x, y, args...`) | Assert if `x` does not equal to `y` |
| **assert_lt**(`x, y, args...`) | Assert if `x` is less than `y` |
| **assert_le**(`x, y, args...`) | Assert if `x` is less than or equal to `y` |
| **assert_gt**(`x, y, args...`) | Assert if `x` is greater than `y` |
| **assert_ge**(`x, y, args...`) | Assert if `x` is greater than or equal to `y` |

## 8. Interfaces

Interfaces define a set of method signatures:

```ebnf
InterfaceDecl     = { Annotation } "interface" InterfaceSpec .
InterfaceSpec     = identifier "{" { MethodDecl ";" } "}" .
MethodDecl        = { Annotation } identifier "(" [ ParameterList ] ")" [ Type ] .
ParameterList     = ParameterDecl { "," ParameterDecl } .
ParameterDecl     = { Annotation } Type identifier .
```

## 9. Lexical Conventions

- Package names use lower camel case (all lowercase is recommended).
- Constants, enum members, and type names (structs, enums, interfaces) use upper camel case.

## 10. EBNF Syntax Specification

```ebnf
(* Top-level constructs *)
SourceFile = PackageClause {ImportDecl} {TopLevelDecl} .

PackageClause = {Annotation} "package" PackageName ";" .
PackageName = identifier .

ImportDecl = "import" ImportSpec ";" .
ImportSpec = string_lit .

TopLevelDecl = ConstDecl | EnumDecl | StructDecl | InterfaceDecl .

(* Declarations *)
ConstDecl = {Annotation} "const" identifier "=" Expression ";" .

EnumDecl = {Annotation} "enum" identifier "{" {EnumMember} "}" .
EnumMember = {Annotation} identifier ["=" Expression] ";" .

StructDecl = {Annotation} "struct" identifier "{" {FieldDecl} "}" .
FieldDecl = {Annotation} Type identifier ";" .

InterfaceDecl = {Annotation} "interface" identifier "{" {MethodDecl} "}" .
MethodDecl = {Annotation} identifier "(" [ParameterList] ")" [Type] ";" .
ParameterList = ParameterDecl {"," ParameterDecl} .
ParameterDecl = {Annotation} Type identifier .

(* Types *)
Type = PrimitiveType | CompositeType .
PrimitiveType = "bool" | "int" | "int8" | "int16" | "int32" | "int64" |
                "float32" | "float64" | "string" | "byte" | "bytes" | "any" .
CompositeType = ArrayType | VectorType | MapType .
ArrayType = "array" "<" Type "," int_lit ">" .
VectorType = "vector" "<" Type ">" .
MapType = "map" "<" Type "," Type ">" .

(* Expressions *)
Expression = UnaryExpr | Expression binary_op Expression | FunctionCall .
UnaryExpr = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr = Operand | PrimaryExpr Selector .
Operand = Literal | QualifiedIdent | "(" Expression ")" .
QualifiedIdent = [PackageName "."] identifier .
Selector = "." identifier .
FunctionCall = identifier "(" [ArgumentList] ")" .
ArgumentList = Expression {"," Expression} .

binary_op = "+" | "-" | "*" | "/" | "%" |
            "&" | "|" | "^" | "<<" | ">>" | "&^" |
            "==" | "!=" | "<" | "<=" | ">" | ">=" |
            "&&" | "||" .
unary_op = "+" | "-" | "!" | "^" .

(* Annotations *)
Annotation = "@" identifier ["(" [Parameters] ")"] .
Parameters = NamedParam {"," NamedParam} .
NamedParam = identifier ["=" ConstantExpression] .

(* Lexical elements *)
identifier = ( letter | "_" ) { letter | unicode_digit | "_" } .

int_lit = decimal_lit | octal_lit | hex_lit .
decimal_lit = "0" | (non_zero_digit {decimal_digit}) .
octal_lit = "0" {octal_digit} .
hex_lit = "0" ("x" | "X") hex_digit {hex_digit} .

float_lit = decimals "." [decimals] [exponent] |
            decimals exponent |
            "." decimals [exponent] .
decimals = decimal_digit {decimal_digit} .
exponent = ("e" | "E") ["+" | "-"] decimals .

string_lit = raw_string_lit | interpreted_string_lit .
raw_string_lit = "`" {unicode_char} "`" .
interpreted_string_lit = '"' {unicode_value | byte_value} '"' .

(* Built-in functions and variables *)
BuiltInFunction = "int" | "float" | "bool" | "min" | "max" | "abs" | "len" |
                  "sprint" | "sprintf" | "sprintln" | "print" | "printf" |
                  "error" | "assert" | "assert_eq" | "assert_ne" |
                  "assert_lt" | "assert_le" | "assert_gt" | "assert_ge" .
BuiltInVariable = "iota" .
```