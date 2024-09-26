# Object {#user-content-Object}

`Object` is a generic object type. These objects can be used as parameters for the `Context/next` function, like `{{next .}}`.

###### .Typeof {#user-content-Object_-Typeof}

`.Typeof` returns the type name of the object. The type name is a string that represents the type of the object. Except for objects under [Common](#user-content-Object_Common), the type names of other objects are lowercase names separated by dots. For example, the type name of a `EnumMember` object is `enum.member`, and the type name of a `Enum` object is `enum`. These objects can be customized for code generation by defining templates. 
Example: 

```next
package demo;

enum Color {
	Red = 1;
	Green = 2;
	Blue = 3;
}
```


```npl
{{- define "go/enum.member" -}}
const {{next .Name}} = {{.Value}}
{{- end}}

{{- define "go/enum.member:name" -}}
{{.Decl.Name}}_{{.}}
{{- end}}
```

Output: 
```go
package demo

type Color int

const Color_Red = 1
const Color_Green = 2
const Color_Blue = 3
```

These two definitions will override the built-in template functions `next/go/enum.member` and `next/go/enum.member.name`.

## ArrayType {#user-content-Object_ArrayType}

`ArrayType` represents an array [type](#user-content-Object_Common_Type).

###### .ElemType {#user-content-Object_ArrayType_-ElemType}

`.ElemType` represents the element [type](#user-content-Object_Common_Type) of the array.

###### .N {#user-content-Object_ArrayType_-N}

`.N` represents the number of elements in the array.

## Comment {#user-content-Object_Comment}

`Comment` represents a line comment or a comment group in Next source code. Use this in templates to access and format comments.

###### .String {#user-content-Object_Comment_-String}

`.String` returns the full original comment text, including delimiters. 
Usage in templates: 
```npl
{{.Comment.String}}
```

###### .Text {#user-content-Object_Comment_-Text}

`.Text` returns the content of the comment without comment delimiters. 
Usage in templates: 
```npl
{{.Comment.Text}}
```

## Common {#user-content-Object_Common}

`Common` contains some general types, including a generic type. Unless specifically stated, these objects cannot be directly called using the `Context/next` function. The [Value](#user-content-Object_Common_Value) object represents a value, which can be either a constant value or an enum member's value. The object type for the former is `const.value`, and for the latter is `enum.member.value`.

### Annotation {#user-content-Object_Common_Annotation}

`Annotation` represents an annotation by `name` => value. 
Annotation is a map that stores the parameters of a single annotation. It allows for flexible parameter types, including strings, numbers, booleans and [types](#user-content-Object_Common_Type). 
Example: 
Next code: 

```next

@json(omitempty)
@event(name="Login")
@message(name="Login", type=100)
struct Login {}

@next(type=int8)
enum Color {
	Red = 1;
	Green = 2;
	Blue = 3;
}
```

Will be represented as: 

```npl
{{- define "go/struct" -}}
{{.Annotations.json.omitempty}}
{{.Annotations.event.name}}
{{.Annotations.message.name}}
{{.Annotations.message.type}}
{{- end}}

{{- define "go/enum" -}}
{{.Annotations.next.type}}
{{- end}}
```

Output: 
```
true
Login
Login
100
int8
```

The `next` annotation is used to pass information to the next compiler. It's a reserved annotation and should not be used for other purposes. The `next` annotation can be annotated to `package` statements, `const` declarations, `enum` declarations, `struct` declarations, `field` declarations, `interface` declarations, `method` declarations, and `parameter` declarations.

###### .Contains {#user-content-Object_Common_Annotation_-Contains}

`.Contains` reports whether the annotation contains the given parameter.

#### decl {#user-content-Object_Common_Annotation_decl}
###### .available {#user-content-Object_Common_Annotation_decl_-available}

The `@next(available="expression")` annotation for `file`, `const`, `enum`, `struct`, `field`, `interface`, `method` availability of the declaration. The `expression` is a boolean expression that can be used to control the availability of the declaration in the target language. Supported operators are `&`, `|`, `!`, `(`, `)`, and `true`, `false`. 
Example: 
```next
@next(available="c|cpp|java|go|csharp")
struct Point {
	int x;
	int y;

	@next(available="c | cpp | go")
	int z;

	@next(available="!java & !c")
	int w;
}
```

#### enum {#user-content-Object_Common_Annotation_enum}

The `next` annotation for `enum` declarations used to control the enum behavior.

###### .type {#user-content-Object_Common_Annotation_enum_-type}

`.type` specifies the underlying type of the enum. 
Example: 
```next
@next(type=int8)
enum Color {
	Red = 1;
	Green = 2;
	Blue = 3;
}
```

Output in Go: 
```go
type Color int8

const (
	ColorRed Color = 1
	ColorGreen Color = 2
	ColorBlue Color = 3
)
```

Output in C++: 
```cpp
enum class Color : int8_t {
	Red = 1,
	Green = 2,
	Blue = 3,
};
```

#### interface {#user-content-Object_Common_Annotation_interface}

The `next` annotation for `interface` declarations used to control the interface behavior. `L_alias` is a alias for the interface name in language `L`. It's used to reference an external type in the target language. 
Example: 
```next
@next(
	available="go|java",
	go_alias="net/http.Handler",
	java_alias="java.util.function.Function<com.sun.net.httpserver.HttpExchange, String>",
)
interface HTTPHandler {}

@next(available="go|java")
interface HTTPServer {
	@next(error)
	Handle(string path, HTTPHandler handler);
}
```

#### method {#user-content-Object_Common_Annotation_method}

The `next` annotation for `method` declarations used to control the method behavior.

###### .error {#user-content-Object_Common_Annotation_method_-error}

The `@next(error)` annotation used to indicate that the method returns an error or throws an exception. 
Example: 
```next
interface Parser {
	@next(error)
	parse(string s) int;
}
```

Output in Go: 
```go
type Parser interface {
	Parse(s string) (int, error)
}
```

Output in C++: 
```cpp
class Parser {
public:
	int parse(const std::string& s) const;
};
```

Output in Java: 
```java
interface Parser {
	int parse(String s) throws Exception;
}
```

###### .mut {#user-content-Object_Common_Annotation_method_-mut}

The `@next(mut)` annotation used to indicate that the method is a mutable method, which means it can modify the object's state. 
Example: 
```next
interface Writer {
	@next(error, mut)
	write(string data);
}
```

Output in Go: 
```go
type Writer interface {
	Write(data string) error
}
```

Output in C++: 
```cpp
class Writer {
public:
	void write(const std::string& data);
};
```

#### package {#user-content-Object_Common_Annotation_package}

The `next` annotation for `package` statements used to control the package behavior for specific languages. The `next` annotation can be used to set the package name, package path, and some other package-related information. 
For any language `L`, the `next` annotation for `package` statements is defined as `@next(L_package="package_info")`. 
Example: 
```next
@next(
	c_package="DEMO_",
	cpp_package="demo",
	java_package="com.exmaple.demo",
	go_package="github.com/next/demo",
	csharp_package="demo",
)
```


```npl
{{.Package.Annotations.next.c_package}}
{{.Package.Annotations.next.cpp_package}}
{{.Package.Annotations.next.java_package}}
{{.Package.Annotations.next.go_package}}
{{.Package.Annotations.next.csharp_package}}
```

There are some reserved keys for the `next` annotation for `package` statements.

###### .go_imports {#user-content-Object_Common_Annotation_package_-go_imports}

`.go_imports` represents a list of import paths for Go packages, separated by commas: `@next(go_imports="fmt.Printf,*io.Reader")`. Note: **`*` is required to import types.** 
Example: 
```next
@next(go_imports="fmt.Printf,*io.Reader")
package demo;
```

#### param {#user-content-Object_Common_Annotation_param}

The `next` annotation for `parameter` declarations used to control the parameter behavior.

###### .mut {#user-content-Object_Common_Annotation_param_-mut}

The `@next(mut)` annotation used to indicate that the parameter is mutable. 
Example: 
```next
interface Reader {
	@next(error);
	read(@next(mut) string data);
}
```

Output in Go: 
```go
type Reader interface {
	Read(data string) error
}
```

Output in C++: 
```cpp
class Reader {
public:
	void read(std::string& data);
};
```

#### struct {#user-content-Object_Common_Annotation_struct}

The `next` annotation for `struct` declarations used to control the struct behavior. `L_alias` is a alias for the struct name in language `L`. It's used to reference an external type in the target language. 
Example: 
```next
@next(rust_alias="u128")
struct uint128 {
	int64 low;
	int64 high;
}

@next(go_alias="complex128")
struct Complex {
	float64 real;
	float64 imag;
}

struct Contract {
	uint128 address;
	Complex complex;
}
```

This will don't generate the `uint128` struct in the `rust` language, but use `u128` instead. And in the `go` language, it will use `complex128` instead of `Complex`.

### Annotations {#user-content-Object_Common_Annotations}

`Annotations` represents a group of annotations by `name` => [Annotation](#user-content-Object_Common_Annotation). 
Annotations is a map that stores multiple annotations for a given entity. The key is the annotation name (string), and the value is the corresponding [Annotation](#user-content-Object_Common_Annotation) object.

###### .Contains {#user-content-Object_Common_Annotations_-Contains}

`.Contains` reports whether the annotations contain the given annotation.

### Decl {#user-content-Object_Common_Decl}

`Decl` represents a top-level declaration in a file. 
All declarations are [nodes](#user-content-Object_Common_Node). Currently, the following declarations are supported: 
- [Package](#user-content-Object_Package)
- [File](#user-content-Object_File)
- [Const](#user-content-Object_Const)
- [Enum](#user-content-Object_Enum)
- [Struct](#user-content-Object_Struct)
- [Interface](#user-content-Object_Interface)

###### .UsedKinds {#user-content-Object_Common_Decl_-UsedKinds}

`.UsedKinds` returns the used kinds in the declaration. Returns 0 if the declaration does not use any kinds. Otherwise, returns the OR of all used kinds. 
Example: 
```next
struct User {
	int64 id;
	string name;
	vector<string> emails;
	map<int, bool> flags;
}
```
The used kinds in the `User` struct are: `(1<<KindInt64) | (1<<KindString) | (1<<KindVector) | (1<<KindMap) | (1<<KindInt) | (1<<KindBool)`.

### Fields {#user-content-Object_Common_Fields}

`Fields` represents a list of fields in a declaration.

###### .Decl {#user-content-Object_Common_Fields_-Decl}

`.Decl` is the declaration object that contains the fields. 
Currently, it is one of following types: 
- [Enum](#user-content-Object_Enum)
- [Struct](#user-content-Object_Struct)
- [Interface](#user-content-Object_Interface)
- [InterfaceMethod](#user-content-Object_InterfaceMethod).

###### .List {#user-content-Object_Common_Fields_-List}

`.List` is the list of fields in the declaration. 
Currently, the field object is one of following types: 
- [EnumMember](#user-content-Object_EnumMember)
- [StructField](#user-content-Object_StructField)
- [InterfaceMethod](#user-content-Object_InterfaceMethod).
- [InterfaceMethodParam](#user-content-Object_InterfaceMethodParam).

### List {#user-content-Object_Common_List}

`List` represents a list of objects.

###### .List {#user-content-Object_Common_List_-List}

`.List` represents the list of objects. It is used to provide a uniform way to access.

### Node {#user-content-Object_Common_Node}

`Node` represents a Node in the Next AST. 
Currently, the following nodes are supported: 
- [File](#user-content-Object_File)
- [Const](#user-content-Object_Const)
- [Enum](#user-content-Object_Enum)
- [Struct](#user-content-Object_Struct)
- [Interface](#user-content-Object_Interface)
- [EnumMember](#user-content-Object_EnumMember)
- [StructField](#user-content-Object_StructField)
- [InterfaceMethod](#user-content-Object_InterfaceMethod)
- [InterfaceMethodParam](#user-content-Object_InterfaceMethodParam)

###### .Annotations {#user-content-Object_Common_Node_-Annotations}

`.Annotations` represents the [annotations](#user-content-Object_Common_Annotations) for the node.

###### .Doc {#user-content-Object_Common_Node_-Doc}

`.Doc` represents the documentation comment for the node.

###### .File {#user-content-Object_Common_Node_-File}

`.File` represents the file containing the node.

###### .Package {#user-content-Object_Common_Node_-Package}

`.Package` represents the package containing the node.

### Symbol {#user-content-Object_Common_Symbol}

`Symbol` represents a Next symbol: value(const, enum member), type(enum, struct, interface).

### Type {#user-content-Object_Common_Type}

`Type` represents a Next type. 
Currently, the following types are supported: 
- [UsedType](#user-content-Object_UsedType)
- [PrimitiveType](#user-content-Object_PrimitiveType)
- [ArrayType](#user-content-Object_ArrayType)
- [VectorType](#user-content-Object_VectorType)
- [MapType](#user-content-Object_MapType)
- [EnumType](#user-content-Object_EnumType)
- [StructType](#user-content-Object_StructType)
- [InterfaceType](#user-content-Object_InterfaceType)

###### .Decl {#user-content-Object_Common_Type_-Decl}

`.Decl` represents the [declaration](#user-content-Decl) of the type.

###### .Kind {#user-content-Object_Common_Type_-Kind}

`.Kind` returns the [kind](#user-content-Object_Common_Type_Kind) of the type.

###### .String {#user-content-Object_Common_Type_-String}

`.String` represents the string representation of the type.

###### .UsedKinds {#user-content-Object_Common_Type_-UsedKinds}

`.UsedKinds` returns the used kinds in the type.

###### .Value {#user-content-Object_Common_Type_-Value}

`.Value` returns the reflect value of the type.

#### Kind {#user-content-Object_Common_Type_Kind}

`Kind` represents the type kind. Currently, the following kinds are supported: 
- `bool`: true or false
- `int`: integer
- `int8`: 8-bit integer
- `int16`: 16-bit integer
- `int32`: 32-bit integer
- `int64`: 64-bit integer
- `float32`: 32-bit floating point
- `float64`: 64-bit floating point
- `byte`: byte
- `bytes`: byte slice
- `string`: string
- `any`: any object
- `map`: dictionary
- `vector`: vector of elements
- `array`: array of elements
- `enum`: enumeration
- `struct`: structure
- `interface`: interface

###### .Bits {#user-content-Object_Common_Type_Kind_-Bits}

`.Bits` returns the number of bits for the type. If the type has unknown bits, it returns 0 (for example, `any`, `string`, `bytes`).

###### .Compatible {#user-content-Object_Common_Type_Kind_-Compatible}

`.Compatible` returns the compatible type between two types. If the types are not compatible, it returns `KindInvalid`. If the types are the same, it returns the type. If the types are numeric, it returns the type with the most bits.

###### .IsAny {#user-content-Object_Common_Type_Kind_-IsAny}

`.IsAny` reports whether the type is any.

###### .IsArray {#user-content-Object_Common_Type_Kind_-IsArray}

`.IsArray` reports whether the type is an array.

###### .IsBool {#user-content-Object_Common_Type_Kind_-IsBool}

`.IsBool` reports whether the type is a boolean.

###### .IsByte {#user-content-Object_Common_Type_Kind_-IsByte}

`.IsByte` reports whether the type is a byte.

###### .IsBytes {#user-content-Object_Common_Type_Kind_-IsBytes}

`.IsBytes` reports whether the type is a byte slice.

###### .IsEnum {#user-content-Object_Common_Type_Kind_-IsEnum}

`.IsEnum` reports whether the type is an enumeration.

###### .IsFloat {#user-content-Object_Common_Type_Kind_-IsFloat}

`.IsFloat` reports whether the type is a floating point.

###### .IsInteger {#user-content-Object_Common_Type_Kind_-IsInteger}

`.IsInteger` reports whether the type is an integer.

###### .IsInterface {#user-content-Object_Common_Type_Kind_-IsInterface}

`.IsInterface` reports whether the type is an interface.

###### .IsMap {#user-content-Object_Common_Type_Kind_-IsMap}

`.IsMap` reports whether the type is a map.

###### .IsNumeric {#user-content-Object_Common_Type_Kind_-IsNumeric}

`.IsNumeric` reports whether the type is a numeric type.

###### .IsString {#user-content-Object_Common_Type_Kind_-IsString}

`.IsString` reports whether the type is a string.

###### .IsStruct {#user-content-Object_Common_Type_Kind_-IsStruct}

`.IsStruct` reports whether the type is a structure.

###### .IsVector {#user-content-Object_Common_Type_Kind_-IsVector}

`.IsVector` reports whether the type is a vector.

###### .Valid {#user-content-Object_Common_Type_Kind_-Valid}

`.Valid` reports whether the type is valid.

#### Kinds {#user-content-Object_Common_Type_Kinds}

`Kinds` represents the type kind set.

###### .Contains {#user-content-Object_Common_Type_Kinds_-Contains}

`.Contains` reports whether the type contains specific kind. The kind can be a `Kind` (or any integer) or a string representation of the [kind](#user-content-Object_Common_Type_Kind). If the kind is invalid, it returns an error. Currently, the following kinds are supported:

## Const {#user-content-Object_Const}

`Const` (extends [Decl](#user-content-Object_Common_Decl)) represents a const declaration.

###### .Comment {#user-content-Object_Const_-Comment}

`.Comment` is the line [comment](#user-content-Object_Comment) of the constant declaration.

###### .Type {#user-content-Object_Const_-Type}

`.Type` represents the type of the constant.

###### .Value {#user-content-Object_Const_-Value}

`.Value` represents the [value object](#user-content-Object_Value) of the constant.

## Consts {#user-content-Object_Consts}

`Consts` represents a [list](#user-content-Object_Common_List) of const declarations.

## Decls {#user-content-Object_Decls}

`Decls` holds all declarations in a file.

###### .Consts {#user-content-Object_Decls_-Consts}

`.Consts` represents the [list](#user-content-Object_Common_List) of [const](#user-content-Object_Const) declarations.

###### .Enums {#user-content-Object_Decls_-Enums}

`.Enums` represents the [list](#user-content-Object_Common_List) of [enum](#user-content-Object_Enum) declarations.

###### .Interfaces {#user-content-Object_Decls_-Interfaces}

`.Interfaces` represents the [list](#user-content-Object_Common_List) of [interface](#user-content-Object_Interface) declarations.

###### .Structs {#user-content-Object_Decls_-Structs}

`.Structs` represents the [list](#user-content-Object_Common_List) of [struct](#user-content-Object_Struct) declarations.

## Doc {#user-content-Object_Doc}

`Doc` represents a documentation comment for a declaration in Next source code. Use this in templates to access and format documentation comments.

###### .Format {#user-content-Object_Doc_-Format}

`.Format` formats the documentation comment for various output styles. 
Parameters: (_indent_ string[, _begin_ string[, _end_ string]]) 
Usage in templates: 

```npl
{{.Doc.Format "/// "}}
{{.Doc.Format " * " "/**\n" " */"}}
```

Example output: 

```c

/// This is a documentation comment.
/// It can be multiple lines.
/**
 * This is a documentation comment.
 * It can be multiple lines.
 */
```

###### .String {#user-content-Object_Doc_-String}

`.String` returns the full original documentation comment, including delimiters. 
Usage in templates: 
```npl
{{.Doc.String}}
```

###### .Text {#user-content-Object_Doc_-Text}

`.Text` returns the content of the documentation comment without comment delimiters. 
Usage in templates: 
```npl
{{.Doc.Text}}
```

## Enum {#user-content-Object_Enum}

`Enum` (extends [Decl](#user-content-Object_Common_Decl)) represents an enum declaration.

###### .MemberType {#user-content-Object_Enum_-MemberType}

`.MemberType` represents the type of the enum members.

###### .Members {#user-content-Object_Enum_-Members}

`.Members` is the list of enum members.

###### .Type {#user-content-Object_Enum_-Type}

`.Type` is the enum type.

## EnumMember {#user-content-Object_EnumMember}

`EnumMember` (extends [Decl](#user-content-Object_Common_Decl)) represents an enum member object in an [enum](#user-content-Object_Enum) declaration.

###### .Comment {#user-content-Object_EnumMember_-Comment}

`.Comment` represents the line [comment](#user-content-Object_Comment) of the enum member declaration.

###### .Decl {#user-content-Object_EnumMember_-Decl}

`.Decl` represents the [enum](#user-content-Object_Enum) that contains the member.

###### .Index {#user-content-Object_EnumMember_-Index}

`.Index` represents the index of the enum member in the enum type.

###### .Value {#user-content-Object_EnumMember_-Value}

`.Value` represents the [value object](#user-content-Object_Value) of the enum member.

## EnumMembers {#user-content-Object_EnumMembers}

`EnumMembers` represents the [list](#user-content-Object_Common_Fields) of [enum members](#user-content-Object_EnumMember).

## EnumType {#user-content-Object_EnumType}

`EnumType` represents the [type](#user-content-Object_Common_Type) of an [enum](#user-content-Object_Enum) declaration.

## Enums {#user-content-Object_Enums}

`Enums` represents a [list](#user-content-Object_Common_List) of enum declarations.

## File {#user-content-Object_File}

`File` (extends [Decl](#user-content-Object_Common_Decl)) represents a Next source file.

###### .Decls {#user-content-Object_File_-Decls}

`.Decls` returns the file's all top-level declarations.

###### .Imports {#user-content-Object_File_-Imports}

`.Imports` represents the file's import declarations.

###### .LookupLocalType {#user-content-Object_File_-LookupLocalType}

`.LookupLocalType` looks up a type by name in the file's symbol table. If the type is not found, it returns an error. If the symbol is found but it is not a type, it returns an error.

###### .LookupLocalValue {#user-content-Object_File_-LookupLocalValue}

`.LookupLocalValue` looks up a value by name in the file's symbol table. If the value is not found, it returns an error. If the symbol is found but it is not a value, it returns an error.

###### .Name {#user-content-Object_File_-Name}

`.Name` represents the file name without the ".next" extension.

###### .Path {#user-content-Object_File_-Path}

`.Path` represents the file full path.

## Import {#user-content-Object_Import}

`Import` represents a file import.

###### .Comment {#user-content-Object_Import_-Comment}

`.Comment` represents the import declaration line [comment](#user-content-Object_Comment).

###### .Doc {#user-content-Object_Import_-Doc}

`.Doc` represents the import declaration [documentation](#user-content-Object_Doc).

###### .File {#user-content-Object_Import_-File}

`.File` represents the file containing the import declaration.

###### .FullPath {#user-content-Object_Import_-FullPath}

`.FullPath` represents the full path of the import.

###### .Path {#user-content-Object_Import_-Path}

`.Path` represents the import path.

###### .Target {#user-content-Object_Import_-Target}

`.Target` represents the imported file.

## Imports {#user-content-Object_Imports}

`Imports` holds a list of imports.

###### .File {#user-content-Object_Imports_-File}

`.File` represents the file containing the imports.

###### .List {#user-content-Object_Imports_-List}

`.List` represents the list of [imports](#user-content-Object_Import).

###### .TrimmedList {#user-content-Object_Imports_-TrimmedList}

`.TrimmedList` represents a list of unique imports sorted by package name.

## Interface {#user-content-Object_Interface}

`Interface` (extends [Decl](#user-content-Object_Common_Decl)) represents an interface declaration.

###### .Methods {#user-content-Object_Interface_-Methods}

`.Methods` represents the list of interface methods.

###### .Type {#user-content-Object_Interface_-Type}

`.Type` represents the interface type.

## InterfaceMethod {#user-content-Object_InterfaceMethod}

`InterfaceMethod` (extends [Node](#user-content-Object_Common_Node)) represents an interface method declaration.

###### .Comment {#user-content-Object_InterfaceMethod_-Comment}

`.Comment` represents the line [comment](#user-content-Object_Comment) of the interface method declaration.

###### .Decl {#user-content-Object_InterfaceMethod_-Decl}

`.Decl` represents the interface that contains the method.

###### .Index {#user-content-Object_InterfaceMethod_-Index}

`.Index` represents the index of the interface method in the interface type.

###### .IsFirst {#user-content-Object_InterfaceMethod_-IsFirst}

`.IsFirst` reports whether the method is the first method in the interface type.

###### .IsLast {#user-content-Object_InterfaceMethod_-IsLast}

`.IsLast` reports whether the method is the last method in the interface type.

###### .Params {#user-content-Object_InterfaceMethod_-Params}

`.Params` represents the list of method parameters.

###### .Result {#user-content-Object_InterfaceMethod_-Result}

`.Result` represents the return type of the method.

## InterfaceMethodParam {#user-content-Object_InterfaceMethodParam}

`InterfaceMethodParam` (extends [Node](#user-content-Object_Common_Node)) represents an interface method parameter declaration.

###### .Index {#user-content-Object_InterfaceMethodParam_-Index}

`.Index` represents the index of the interface method parameter in the method.

###### .IsFirst {#user-content-Object_InterfaceMethodParam_-IsFirst}

`.IsFirst` reports whether the parameter is the first parameter in the method.

###### .IsLast {#user-content-Object_InterfaceMethodParam_-IsLast}

`.IsLast` reports whether the parameter is the last parameter in the method.

###### .Method {#user-content-Object_InterfaceMethodParam_-Method}

`.Method` represents the interface method that contains the parameter.

###### .Type {#user-content-Object_InterfaceMethodParam_-Type}

`.Type` represents the [type](#user-content-Object_Common_Type) of the parameter.

## InterfaceMethodParams {#user-content-Object_InterfaceMethodParams}

`InterfaceMethodParams` represents the [list](#user-content-Object_Common_Fields) of [interface method parameters](#user-content-Object_InterfaceMethodParam).

## InterfaceMethodResult {#user-content-Object_InterfaceMethodResult}

`InterfaceMethodResult` represents an interface method result.

###### .Method {#user-content-Object_InterfaceMethodResult_-Method}

`.Method` represents the interface method that contains the result.

###### .Type {#user-content-Object_InterfaceMethodResult_-Type}

`.Type` represents the underlying type of the result.

## InterfaceMethods {#user-content-Object_InterfaceMethods}

`InterfaceMethods` represents the [list](#user-content-Object_Common_Fields) of [interface methods](#user-content-Object_InterfaceMethod).

## InterfaceType {#user-content-Object_InterfaceType}

`InterfaceType` represents the [type](#user-content-Object_Common_Type) of an [interface](#user-content-Object_Interface) declaration.

## Interfaces {#user-content-Object_Interfaces}

`Interfaces` represents a [list](#user-content-Object_Common_List) of interface declarations.

## MapType {#user-content-Object_MapType}

`MapType` represents a map [type](#user-content-Object_Common_Type).

###### .ElemType {#user-content-Object_MapType_-ElemType}

`.ElemType` represents the element [type](#user-content-Object_Common_Type) of the map.

###### .KeyType {#user-content-Object_MapType_-KeyType}

`.KeyType` represents the key [type](#user-content-Object_Common_Type) of the map.

## Package {#user-content-Object_Package}

`Package` (extends [Decl](#user-content-Object_Common_Decl)) represents a Next package.

###### .Contains {#user-content-Object_Package_-Contains}

`.Contains` reports whether the package contains the given type. If the current package is nil, it always returns true. 
Example: 

```npl
{{- define "next/go/used.type" -}}
{{if not (.File.Package.Contains .Type) -}}
{{.Type.Decl.File.Package.Name -}}.
{{- end -}}
{{next .Type}}
{{- end}}
```

###### .Decls {#user-content-Object_Package_-Decls}

`.Decls` represents the top-level declarations in the package.

###### .Files {#user-content-Object_Package_-Files}

`.Files` represents the all declared files in the package.

###### .Imports {#user-content-Object_Package_-Imports}

`.Imports` represents the package's import declarations.

###### .Name {#user-content-Object_Package_-Name}

`.Name` represents the package name string.

###### .Types {#user-content-Object_Package_-Types}

`.Types` represents the all declared types in the package.

## PrimitiveType {#user-content-Object_PrimitiveType}

`PrimitiveType` represents a primitive type.

## Struct {#user-content-Object_Struct}

`Struct` (extends [Decl](#user-content-Object_Common_Decl)) represents a struct declaration.

###### .Fields {#user-content-Object_Struct_-Fields}

`.Fields` represents the list of struct fields.

###### .Type {#user-content-Object_Struct_-Type}

`.Type` represents the struct type.

## StructField {#user-content-Object_StructField}

`StructField` (extends [Node](#user-content-Object_Common_Node)) represents a struct field declaration.

###### .Comment {#user-content-Object_StructField_-Comment}

`.Comment` represents the line [comment](#user-content-Object_Comment) of the struct field declaration.

###### .Decl {#user-content-Object_StructField_-Decl}

`.Decl` represents the struct that contains the field.

###### .Index {#user-content-Object_StructField_-Index}

`.Index` represents the index of the struct field in the struct type.

###### .IsFirst {#user-content-Object_StructField_-IsFirst}

`.IsFirst` reports whether the field is the first field in the struct type.

###### .IsLast {#user-content-Object_StructField_-IsLast}

`.IsLast` reports whether the field is the last field in the struct type.

###### .Type {#user-content-Object_StructField_-Type}

`.Type` represents the [type](#user-content-Object_Common_Type) of the struct field.

## StructFields {#user-content-Object_StructFields}

`StructFields` represents the [list](#user-content-Object_Common_Fields) of [struct fields](#user-content-Object_StructField).

## StructType {#user-content-Object_StructType}

`StructType` represents the [type](#user-content-Object_Common_Type) of a [struct](#user-content-Object_Struct) declaration.

## Structs {#user-content-Object_Structs}

`Structs` represents a [list](#user-content-Object_Common_List) of struct declarations.

## UsedType {#user-content-Object_UsedType}

`UsedType` represents a used type in a file.

###### .File {#user-content-Object_UsedType_-File}

`.File` represents the file where the type is used.

###### .Type {#user-content-Object_UsedType_-Type}

`.Type` represents the used type.

## Value {#user-content-Object_Value}

`Value` represents a constant value for a const declaration or an enum member.

###### .Any {#user-content-Object_Value_-Any}

`.Any` represents the underlying value of the constant.

###### .IsEnum {#user-content-Object_Value_-IsEnum}

`.IsEnum` returns true if the value is an enum member.

###### .IsFirst {#user-content-Object_Value_-IsFirst}

`.IsFirst` reports whether the value is the first member of the enum type.

###### .IsLast {#user-content-Object_Value_-IsLast}

`.IsLast` reports whether the value is the last member of the enum type.

###### .String {#user-content-Object_Value_-String}

`.String` represents the string representation of the value.

###### .Type {#user-content-Object_Value_-Type}

`.Type` represents the [primitive type](#user-content-Object_PrimitiveType) of the value.

## VectorType {#user-content-Object_VectorType}

`VectorType` represents a vector [type](#user-content-Object_Common_Type).

###### .ElemType {#user-content-Object_VectorType_-ElemType}

`.ElemType` represents the element [type](#user-content-Object_Common_Type) of the vector.

