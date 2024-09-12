# API Reference

<ul>
  <li><a href="#user-content-Context">Context</a></li>
<ul>
    <li><a href="#user-content-Context_align">align</a></li>
    <li><a href="#user-content-Context_env">env</a></li>
    <li><a href="#user-content-Context_error">error</a></li>
    <li><a href="#user-content-Context_errorf">errorf</a></li>
    <li><a href="#user-content-Context_exist">exist</a></li>
    <li><a href="#user-content-Context_head">head</a></li>
    <li><a href="#user-content-Context_lang">lang</a></li>
    <li><a href="#user-content-Context_meta">meta</a></li>
    <li><a href="#user-content-Context_next">next</a></li>
    <li><a href="#user-content-Context_render">render</a></li>
    <li><a href="#user-content-Context_super">super</a></li>
    <li><a href="#user-content-Context_this">this</a></li>
    <li><a href="#user-content-Context_type">type</a></li>
</ul>
  <li><a href="#user-content-Object">Object</a></li>
<ul>
    <li><a href="#user-content-Object_ArrayType">ArrayType</a></li>
    <li><a href="#user-content-Object_Comment">Comment</a></li>
    <li><a href="#user-content-Object_Common">Common</a></li>
<ul>
      <li><a href="#user-content-Object_Common_Annotation">Annotation</a></li>
<ul>
        <li><a href="#user-content-Object_Common_Annotation_decl">decl</a></li>
        <li><a href="#user-content-Object_Common_Annotation_enum">enum</a></li>
        <li><a href="#user-content-Object_Common_Annotation_interface">interface</a></li>
        <li><a href="#user-content-Object_Common_Annotation_method">method</a></li>
        <li><a href="#user-content-Object_Common_Annotation_package">package</a></li>
        <li><a href="#user-content-Object_Common_Annotation_param">param</a></li>
        <li><a href="#user-content-Object_Common_Annotation_struct">struct</a></li>
</ul>
      <li><a href="#user-content-Object_Common_Annotations">Annotations</a></li>
      <li><a href="#user-content-Object_Common_Decl">Decl</a></li>
      <li><a href="#user-content-Object_Common_Fields">Fields</a></li>
      <li><a href="#user-content-Object_Common_List">List</a></li>
      <li><a href="#user-content-Object_Common_Node">Node</a></li>
      <li><a href="#user-content-Object_Common_NodeName">NodeName</a></li>
      <li><a href="#user-content-Object_Common_Symbol">Symbol</a></li>
      <li><a href="#user-content-Object_Common_Type">Type</a></li>
<ul>
        <li><a href="#user-content-Object_Common_Type_Kind">Kind</a></li>
</ul>
</ul>
    <li><a href="#user-content-Object_Const">Const</a></li>
    <li><a href="#user-content-Object_ConstName">ConstName</a></li>
    <li><a href="#user-content-Object_Consts">Consts</a></li>
    <li><a href="#user-content-Object_Decls">Decls</a></li>
    <li><a href="#user-content-Object_Doc">Doc</a></li>
    <li><a href="#user-content-Object_Enum">Enum</a></li>
    <li><a href="#user-content-Object_EnumMember">EnumMember</a></li>
    <li><a href="#user-content-Object_EnumMemberName">EnumMemberName</a></li>
    <li><a href="#user-content-Object_EnumMembers">EnumMembers</a></li>
    <li><a href="#user-content-Object_EnumType">EnumType</a></li>
    <li><a href="#user-content-Object_Enums">Enums</a></li>
    <li><a href="#user-content-Object_File">File</a></li>
    <li><a href="#user-content-Object_Import">Import</a></li>
    <li><a href="#user-content-Object_Imports">Imports</a></li>
    <li><a href="#user-content-Object_Interface">Interface</a></li>
    <li><a href="#user-content-Object_InterfaceMethod">InterfaceMethod</a></li>
    <li><a href="#user-content-Object_InterfaceMethodName">InterfaceMethodName</a></li>
    <li><a href="#user-content-Object_InterfaceMethodParam">InterfaceMethodParam</a></li>
    <li><a href="#user-content-Object_InterfaceMethodParamName">InterfaceMethodParamName</a></li>
    <li><a href="#user-content-Object_InterfaceMethodParamType">InterfaceMethodParamType</a></li>
    <li><a href="#user-content-Object_InterfaceMethodParams">InterfaceMethodParams</a></li>
    <li><a href="#user-content-Object_InterfaceMethodResult">InterfaceMethodResult</a></li>
    <li><a href="#user-content-Object_InterfaceMethods">InterfaceMethods</a></li>
    <li><a href="#user-content-Object_InterfaceType">InterfaceType</a></li>
    <li><a href="#user-content-Object_Interfaces">Interfaces</a></li>
    <li><a href="#user-content-Object_MapType">MapType</a></li>
    <li><a href="#user-content-Object_Package">Package</a></li>
    <li><a href="#user-content-Object_PrimitiveType">PrimitiveType</a></li>
    <li><a href="#user-content-Object_Struct">Struct</a></li>
    <li><a href="#user-content-Object_StructField">StructField</a></li>
    <li><a href="#user-content-Object_StructFieldName">StructFieldName</a></li>
    <li><a href="#user-content-Object_StructFieldType">StructFieldType</a></li>
    <li><a href="#user-content-Object_StructFields">StructFields</a></li>
    <li><a href="#user-content-Object_StructType">StructType</a></li>
    <li><a href="#user-content-Object_Structs">Structs</a></li>
    <li><a href="#user-content-Object_UsedType">UsedType</a></li>
    <li><a href="#user-content-Object_Value">Value</a></li>
    <li><a href="#user-content-Object_VectorType">VectorType</a></li>
</ul>
</ul>

<h2><a id="user-content-Context" target="_self">Context</a></h2>

_Context_ related methods and properties are used to retrieve information, perform operations, and generate code within the current code generator's context. These methods or properties are called directly by name, for example: 

```npl
{{head}}
{{next this}}
{{lang}}
{{exist meta.path}}
```

<h3><a id="user-content-Context_align" target="_self">align</a></h3>

_align_ aligns the given text with the same indent as the first line. 
Example (without align): 
```npl
	{{print "hello\nworld"}}
```

Output: 
```
	hello
world
```

To align it, you can use `align`: 
```npl
	{{align "hello\nworld"}}
```

Output: 

```
	hello
	world
```

It's useful when you want to align the generated content, especially for multi-line strings (e.g., comments).

<h3><a id="user-content-Context_env" target="_self">env</a></h3>

_env_ represents the environment variables defined in the command line with the flag `-D`. 
Example: 

```sh
next -D PROJECT_NAME=demo
```


```npl
{{env.PROJECT_NAME}}
```

<h3><a id="user-content-Context_error" target="_self">error</a></h3>

_error_ used to return an error message in the template. 
Example: 

```npl
{{error "Something went wrong"}}
```

<h3><a id="user-content-Context_errorf" target="_self">errorf</a></h3>

_errorf_ used to return a formatted error message in the template. 
Example: 

```npl
{{errorf "%s went wrong" "Something"}}
```

<h3><a id="user-content-Context_exist" target="_self">exist</a></h3>

_exist_ checks whether the given path exists. If the path is not absolute, it will be resolved relative to the current output directory for the current language by command line flag `-O`. 
Example: 
```npl
{{exist "path/to/file"}}
{{exist "/absolute/path/to/file"}}
{{exist meta.path}}
```

<h3><a id="user-content-Context_head" target="_self">head</a></h3>

_head_ outputs the header of the generated file. 
Example: 

```
{{head}}
```

Output (for c++): 
```
// Code generated by "next v0.0.1"; DO NOT EDIT.
```

Output (for c): 
```
/* Code generated by "next v0.0.1"; DO NOT EDIT. */
```

<h3><a id="user-content-Context_lang" target="_self">lang</a></h3>

_lang_ represents the current language to be generated. 
Example: 

```npl
{{lang}}
{{printf "%s_alias" lang}}
```

<h3><a id="user-content-Context_meta" target="_self">meta</a></h3>

_meta_ represents the metadata of a entrypoint template file by flag `-T`. To define a meta, you should define a template with the name `meta/<key>`. Currently, the following meta keys are used by the code generator: 
- `meta/this`: the current object to be rendered. See [this](#user-content-Context_this) for more details.
- `meta/path`: the output path for the current object. If the path is not absolute, it will be resolved relative to the current output directory for the current language by command line flag `-O`.
- `meta/skip`: whether to skip the current object.

Any other meta keys are user-defined. You can use them in the templates like `{{meta.<key>}}`. 
Example: 

```npl
{{- define "meta/this" -}}file{{- end -}}
{{- define "meta/path" -}}path/to/file{{- end -}}
{{- define "meta/skip" -}}{{exist meta.path}}{{- end -}}
{{- define "meta/custom" -}}custom value{{- end -}}
```

**The metadata will be resolved in the order of the template definition before rendering the entrypoint template.**

<h3><a id="user-content-Context_next" target="_self">next</a></h3>

_next_ executes the next template with the given [object](#user-content-Object). `{{next object}}` is equivalent to `{{render (object.Typeof) object}}`. 
Example: 

```npl
{{- /* Overrides "next/go/struct": add method 'MessageType' for each message after struct */ -}}
{{- define "go/struct"}}
{{- super .}}
{{- with .Annotations.message.type}}

func ({{next $.Type}}) MessageType() int { return {{.}} }
{{- end}}
{{- end -}}

{{next this}}
```

<h3><a id="user-content-Context_render" target="_self">render</a></h3>

_render_ executes the template with the given name and data.

<h3><a id="user-content-Context_super" target="_self">super</a></h3>

_super_ executes the super template with the given [object](#user-content-Object).

<h3><a id="user-content-Context_this" target="_self">this</a></h3>

_this_ represents the current [declaration](#user-content-Object_Common_Decl) object to be rendered. this defined in the template [meta](#user-content-meta) `meta/this`. Supported types are: 
- [file](#user-content-Object_File)
- [const](#user-content-Object_Const)
- [enum](#user-content-Object_Enum)
- [struct](#user-content-Object_Struct)
- [interface](#user-content-Object_Interface)

It's a [file](#user-content-Object_File) by default.

<h3><a id="user-content-Context_type" target="_self">type</a></h3>

_type_ outputs the string representation of the given [type](#user-content-Object_Common_Type) for the current language.

<h2><a id="user-content-Object" target="_self">Object</a></h2>

_Object_ is a generic object type. These objects can be used as parameters for the [next](#user-content-Context_next) function, like `{{next .}}`.

<h6><a id="user-content-Object_-Typeof" target="_self">.Typeof</a></h6>

_.Typeof_ returns the type name of the object. The type name is a string that represents the type of the object. Except for objects under [Common](#user-content-Object_Common), the type names of other objects are lowercase names separated by dots. For example, the type name of a `EnumMember` object is `enum.member`, and the type name of a `EnumMemberName` object is `enum.member.name`. These objects can be customized for code generation by defining templates. For example: 

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

{{- define "go/enum.member.name" -}}
{{.Node.Decl.Name}}_{{.}}
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

<h3><a id="user-content-Object_ArrayType" target="_self">ArrayType</a></h3>

_ArrayType_ represents an array [type](#user-content-Object_Common_Type).

<h6><a id="user-content-Object_ArrayType_-ElemType" target="_self">.ElemType</a></h6>

_.ElemType_ represents the element [type](#user-content-Object_Common_Type) of the array.

<h6><a id="user-content-Object_ArrayType_-N" target="_self">.N</a></h6>

_.N_ represents the number of elements in the array.

<h3><a id="user-content-Object_Comment" target="_self">Comment</a></h3>

_Comment_ represents a line comment or a comment group in Next source code. Use this in templates to access and format comments.

<h6><a id="user-content-Object_Comment_-String" target="_self">.String</a></h6>

_.String_ returns the full original comment text, including delimiters. 
Usage in templates: 
```npl
{{.Comment.String}}
```

<h6><a id="user-content-Object_Comment_-Text" target="_self">.Text</a></h6>

_.Text_ returns the content of the comment without comment delimiters. 
Usage in templates: 
```npl
{{.Comment.Text}}
```

<h3><a id="user-content-Object_Common" target="_self">Common</a></h3>

_Common_ contains some general types, including a generic type. Unless specifically stated, these objects cannot be directly called using the [next](#user-content-Context_next) function. The [Value](#user-content-Object_Common_Value) object represents a value, which can be either a constant value or an enum member's value. The object type for the former is `const.value`, and for the latter is `enum.member.value`.

<h4><a id="user-content-Object_Common_Annotation" target="_self">Annotation</a></h4>

_Annotation_ represents an annotation by `name` => value. 
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
```

The `next` annotation is used to pass information to the next compiler. It's a reserved annotation and should not be used for other purposes. The `next` annotation can be annotated to `package` statements, `const` declarations, `enum` declarations, `struct` declarations, `field` declarations, `interface` declarations, `method` declarations, and `parameter` declarations.

<h5><a id="user-content-Object_Common_Annotation_decl" target="_self">decl</a></h5>
<h6><a id="user-content-Object_Common_Annotation_decl_-available" target="_self">.available</a></h6>

The `@next(available="expression")` annotation for `file`, `const`, `enum`, `struct`, `field`, `interface`, `method` availability of the declaration. 
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

<h5><a id="user-content-Object_Common_Annotation_enum" target="_self">enum</a></h5>

The `next` annotation for `enum` declarations used to control the enum behavior.

<h6><a id="user-content-Object_Common_Annotation_enum_-type" target="_self">.type</a></h6>

_.type_ specifies the underlying type of the enum. 
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

<h5><a id="user-content-Object_Common_Annotation_interface" target="_self">interface</a></h5>

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

<h5><a id="user-content-Object_Common_Annotation_method" target="_self">method</a></h5>

The `next` annotation for `method` declarations used to control the method behavior.

<h6><a id="user-content-Object_Common_Annotation_method_-error" target="_self">.error</a></h6>

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

<h6><a id="user-content-Object_Common_Annotation_method_-mut" target="_self">.mut</a></h6>

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

<h5><a id="user-content-Object_Common_Annotation_package" target="_self">package</a></h5>

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

<h6><a id="user-content-Object_Common_Annotation_package_-go_imports" target="_self">.go_imports</a></h6>

_.go_imports_ represents a list of import paths for Go packages, separated by commas: `@next(go_imports="fmt.Printf,*io.Reader")`. Note: **`*` is required to import types.** 
Example: 
```next
@next(go_imports="fmt.Printf,*io.Reader")
package demo;
```

<h5><a id="user-content-Object_Common_Annotation_param" target="_self">param</a></h5>

The `next` annotation for `parameter` declarations used to control the parameter behavior.

<h6><a id="user-content-Object_Common_Annotation_param_-mut" target="_self">.mut</a></h6>

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

<h5><a id="user-content-Object_Common_Annotation_struct" target="_self">struct</a></h5>

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

<h4><a id="user-content-Object_Common_Annotations" target="_self">Annotations</a></h4>

_Annotations_ represents a group of annotations by `name` => [Annotation](#user-content-Object_Common_Annotation). 
Annotations is a map that stores multiple annotations for a given entity. The key is the annotation name (string), and the value is the corresponding [Annotation](#user-content-Object_Common_Annotation) object.

<h4><a id="user-content-Object_Common_Decl" target="_self">Decl</a></h4>

_Decl_ represents a top-level declaration in a file. 
All declarations are [nodes](#user-content-Object_Common_Node). Currently, the following declarations are supported: 
- [File](#user-content-Object_File)
- [Const](#user-content-Object_Const)
- [Enum](#user-content-Object_Enum)
- [Struct](#user-content-Object_Struct)
- [Interface](#user-content-Object_Interface)

<h4><a id="user-content-Object_Common_Fields" target="_self">Fields</a></h4>

_Fields_ represents a list of fields in a declaration.

<h6><a id="user-content-Object_Common_Fields_-Decl" target="_self">.Decl</a></h6>

_.Decl_ is the declaration object that contains the fields. 
Currently, it is one of following types: 
- [Enum](#user-content-Object_Enum)
- [Struct](#user-content-Object_Struct)
- [Interface](#user-content-Object_Interface)
- [InterfaceMethod](#user-content-Object_InterfaceMethod).

<h6><a id="user-content-Object_Common_Fields_-List" target="_self">.List</a></h6>

_.List_ is the list of fields in the declaration. 
Currently, the field object is one of following types: 
- [EnumMember](#user-content-Object_EnumMember)
- [StructField](#user-content-Object_StructField)
- [InterfaceMethod](#user-content-Object_InterfaceMethod).
- [InterfaceMethodParam](#user-content-Object_InterfaceMethodParam).

<h4><a id="user-content-Object_Common_List" target="_self">List</a></h4>

_List_ represents a list of objects.

<h6><a id="user-content-Object_Common_List_-List" target="_self">.List</a></h6>

_.List_ represents the list of objects. It is used to provide a uniform way to access.

<h4><a id="user-content-Object_Common_Node" target="_self">Node</a></h4>

_Node_ represents a Node in the Next AST. 
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

<h6><a id="user-content-Object_Common_Node_-Annotations" target="_self">.Annotations</a></h6>

_.Annotations_ represents the [annotations](#user-content-Annotation_Annotations) for the node.

<h6><a id="user-content-Object_Common_Node_-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the documentation comment for the node.

<h6><a id="user-content-Object_Common_Node_-File" target="_self">.File</a></h6>

_.File_ represents the file containing the node.

<h6><a id="user-content-Object_Common_Node_-Package" target="_self">.Package</a></h6>

_.Package_ represents the package containing the node. It's a shortcut for Node.File.Package.

<h4><a id="user-content-Object_Common_NodeName" target="_self">NodeName</a></h4>

_NodeName_ represents a name of a node in a declaration. 
Currently, the following types are supported: 
- [ConstName](#user-content-Object_ConstName)
- [EnumMemberName](#user-content-Object_EnumMemberName)
- [StructFieldName](#user-content-Object_StructFieldName)
- [InterfaceMethodName](#user-content-Object_InterfaceMethodName)
- [InterfaceMethodParamName](#user-content-Object_InterfaceMethodParamName)

<h6><a id="user-content-Object_Common_NodeName_-Node" target="_self">.Node</a></h6>

_.Node_ represents the [node](#user-content-Object_Common_Node) that contains the name.

<h6><a id="user-content-Object_Common_NodeName_-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the node name.

<h4><a id="user-content-Object_Common_Symbol" target="_self">Symbol</a></h4>

_Symbol_ represents a Next symbol: value(const, enum member), type(enum, struct, interface).

<h4><a id="user-content-Object_Common_Type" target="_self">Type</a></h4>

_Type_ represents a Next type. 
Currently, the following types are supported: 
- [UsedType](#user-content-Object_UsedType)
- [PrimitiveType](#user-content-Object_PrimitiveType)
- [ArrayType](#user-content-Object_ArrayType)
- [VectorType](#user-content-Object_VectorType)
- [MapType](#user-content-Object_MapType)
- [EnumType](#user-content-Object_EnumType)
- [StructType](#user-content-Object_StructType)
- [InterfaceType](#user-content-Object_InterfaceType)

<h6><a id="user-content-Object_Common_Type_-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the [declaration](#user-content-Decl) of the type.

<h6><a id="user-content-Object_Common_Type_-Kind" target="_self">.Kind</a></h6>

_.Kind_ returns the [kind](#user-content-Object_Common_Type_Kind) of the type.

<h6><a id="user-content-Object_Common_Type_-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the type.

<h5><a id="user-content-Object_Common_Type_Kind" target="_self">Kind</a></h5>

_Kind_ represents the type kind. Currently, the following kinds are supported: 
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

<h6><a id="user-content-Object_Common_Type_Kind_-Bits" target="_self">.Bits</a></h6>

_.Bits_ returns the number of bits for the type. If the type has unknown bits, it returns 0 (for example, `any`, `string`, `bytes`).

<h6><a id="user-content-Object_Common_Type_Kind_-Compatible" target="_self">.Compatible</a></h6>

_.Compatible_ returns the compatible type between two types. If the types are not compatible, it returns `KindInvalid`. If the types are the same, it returns the type. If the types are numeric, it returns the type with the most bits.

<h6><a id="user-content-Object_Common_Type_Kind_-IsFloat" target="_self">.IsFloat</a></h6>

_.IsFloat_ reports whether the type is a floating point.

<h6><a id="user-content-Object_Common_Type_Kind_-IsInteger" target="_self">.IsInteger</a></h6>

_.IsInteger_ reports whether the type is an integer.

<h6><a id="user-content-Object_Common_Type_Kind_-IsNumeric" target="_self">.IsNumeric</a></h6>

_.IsNumeric_ reports whether the type is a numeric type.

<h6><a id="user-content-Object_Common_Type_Kind_-IsString" target="_self">.IsString</a></h6>

_.IsString_ reports whether the type is a string.

<h6><a id="user-content-Object_Common_Type_Kind_-Valid" target="_self">.Valid</a></h6>

_.Valid_ reports whether the type is valid.

<h3><a id="user-content-Object_Const" target="_self">Const</a></h3>

_Const_ (extends [Decl](#user-content-Object_Common_Decl)) represents a const declaration.

<h6><a id="user-content-Object_Const_-Comment" target="_self">.Comment</a></h6>

_.Comment_ is the line [comment](#user-content-Object_Comment) of the constant declaration.

<h6><a id="user-content-Object_Const_-Name" target="_self">.Name</a></h6>

_.Name_ represents the [NodeName](#user-content-Object-NodeName) of the constant.

<h6><a id="user-content-Object_Const_-Type" target="_self">.Type</a></h6>

_.Type_ represents the type of the constant.

<h6><a id="user-content-Object_Const_-Value" target="_self">.Value</a></h6>

_.Value_ represents the [value object](#user-content-Object_Value) of the constant.

<h3><a id="user-content-Object_ConstName" target="_self">ConstName</a></h3>

_ConstName_ represents the [NodeName](#user-content-Object_Common_NodeName) of a [const](#user-content-Object_Const) declaration.

<h3><a id="user-content-Object_Consts" target="_self">Consts</a></h3>

_Consts_ represents a [list](#user-content-Object_Common_List) of const declarations.

<h3><a id="user-content-Object_Decls" target="_self">Decls</a></h3>

_Decls_ holds all declarations in a file.

<h6><a id="user-content-Object_Decls_-Consts" target="_self">.Consts</a></h6>

_.Consts_ represents the [list](#user-content-Object_Common_List) of [const](#user-content-Object_Const) declarations.

<h6><a id="user-content-Object_Decls_-Enums" target="_self">.Enums</a></h6>

_.Enums_ represents the [list](#user-content-Object_Common_List) of [enum](#user-content-Object_Enum) declarations.

<h6><a id="user-content-Object_Decls_-Interfaces" target="_self">.Interfaces</a></h6>

_.Interfaces_ represents the [list](#user-content-Object_Common_List) of [interface](#user-content-Object_Interface) declarations.

<h6><a id="user-content-Object_Decls_-Structs" target="_self">.Structs</a></h6>

_.Structs_ represents the [list](#user-content-Object_Common_List) of [struct](#user-content-Object_Struct) declarations.

<h3><a id="user-content-Object_Doc" target="_self">Doc</a></h3>

_Doc_ represents a documentation comment for a declaration in Next source code. Use this in templates to access and format documentation comments.

<h6><a id="user-content-Object_Doc_-Format" target="_self">.Format</a></h6>

_.Format_ formats the documentation comment for various output styles. 
Usage in templates: 

```npl
{{.Doc.Format "" " * " "/**\n" " */"}}
```

Example output: 

```c

/**
 * This is a documentation comment.
 * It can be multiple lines.
 */
```

<h6><a id="user-content-Object_Doc_-String" target="_self">.String</a></h6>

_.String_ returns the full original documentation comment, including delimiters. 
Usage in templates: 
```npl
{{.Doc.String}}
```

<h6><a id="user-content-Object_Doc_-Text" target="_self">.Text</a></h6>

_.Text_ returns the content of the documentation comment without comment delimiters. 
Usage in templates: 
```npl
{{.Doc.Text}}
```

<h3><a id="user-content-Object_Enum" target="_self">Enum</a></h3>

_Enum_ (extends [Decl](#user-content-Object_Common_Decl)) represents an enum declaration.

<h6><a id="user-content-Object_Enum_-MemberType" target="_self">.MemberType</a></h6>

_.MemberType_ represents the type of the enum members.

<h6><a id="user-content-Object_Enum_-Members" target="_self">.Members</a></h6>

_.Members_ is the list of enum members.

<h6><a id="user-content-Object_Enum_-Type" target="_self">.Type</a></h6>

_.Type_ is the enum type.

<h3><a id="user-content-Object_EnumMember" target="_self">EnumMember</a></h3>

_EnumMember_ (extends [Decl](#user-content-Object_Common_Decl)) represents an enum member object in an [enum](#user-content-Object_Enum) declaration.

<h6><a id="user-content-Object_EnumMember_-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the line [comment](#user-content-Object_Comment) of the enum member declaration.

<h6><a id="user-content-Object_EnumMember_-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the [enum](#user-content-Object_Enum) that contains the member.

<h6><a id="user-content-Object_EnumMember_-Name" target="_self">.Name</a></h6>

_.Name_ represents the [NodeName](#user-content-Object_Common_NodeName) of the enum member.

<h6><a id="user-content-Object_EnumMember_-Value" target="_self">.Value</a></h6>

_.Value_ represents the [value object](#user-content-Object_Value) of the enum member.

<h3><a id="user-content-Object_EnumMemberName" target="_self">EnumMemberName</a></h3>

_EnumMemberName_ represents the [NodeName](#user-content-Object_Common_NodeName) of an [enum member](#user-content-Object_EnumMember).

<h3><a id="user-content-Object_EnumMembers" target="_self">EnumMembers</a></h3>

_EnumMembers_ represents the [list](#user-content-Object_Common_Fields) of [enum members](#user-content-Object_EnumMember).

<h3><a id="user-content-Object_EnumType" target="_self">EnumType</a></h3>

_EnumType_ represents the [type](#user-content-Object_Common_Type) of an [enum](#user-content-Object_Enum) declaration.

<h3><a id="user-content-Object_Enums" target="_self">Enums</a></h3>

_Enums_ represents a [list](#user-content-Object_Common_List) of enum declarations.

<h3><a id="user-content-Object_File" target="_self">File</a></h3>

_File_ (extends [Decl](#user-content-Object_Common_Decl)) represents a Next source file.

<h6><a id="user-content-Object_File_-Decls" target="_self">.Decls</a></h6>

_.Decls_ returns the file's all top-level declarations.

<h6><a id="user-content-Object_File_-LookupLocalType" target="_self">.LookupLocalType</a></h6>

_.LookupLocalType_ looks up a type by name in the file's symbol table. If the type is not found, it returns an error. If the symbol is found but it is not a type, it returns an error.

<h6><a id="user-content-Object_File_-LookupLocalValue" target="_self">.LookupLocalValue</a></h6>

_.LookupLocalValue_ looks up a value by name in the file's symbol table. If the value is not found, it returns an error. If the symbol is found but it is not a value, it returns an error.

<h6><a id="user-content-Object_File_-Name" target="_self">.Name</a></h6>

_.Name_ represents the file name without the ".next" extension.

<h6><a id="user-content-Object_File_-Package" target="_self">.Package</a></h6>

_.Package_ represents the file's import declarations.

<h6><a id="user-content-Object_File_-Path" target="_self">.Path</a></h6>

_.Path_ represents the file full path.

<h3><a id="user-content-Object_Import" target="_self">Import</a></h3>

_Import_ represents a file import.

<h6><a id="user-content-Object_Import_-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the import declaration line [comment](#user-content-Object_Comment).

<h6><a id="user-content-Object_Import_-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the import declaration [documentation](#user-content-Object_Doc).

<h6><a id="user-content-Object_Import_-File" target="_self">.File</a></h6>

_.File_ represents the file containing the import declaration.

<h6><a id="user-content-Object_Import_-FullPath" target="_self">.FullPath</a></h6>

_.FullPath_ represents the full path of the import.

<h6><a id="user-content-Object_Import_-Path" target="_self">.Path</a></h6>

_.Path_ represents the import path.

<h6><a id="user-content-Object_Import_-Target" target="_self">.Target</a></h6>

_.Target_ represents the imported file.

<h3><a id="user-content-Object_Imports" target="_self">Imports</a></h3>

_Imports_ holds a list of imports.

<h6><a id="user-content-Object_Imports_-File" target="_self">.File</a></h6>

_.File_ represents the file containing the imports.

<h6><a id="user-content-Object_Imports_-List" target="_self">.List</a></h6>

_.List_ represents the list of [imports](#user-content-Object_Import).

<h6><a id="user-content-Object_Imports_-TrimmedList" target="_self">.TrimmedList</a></h6>

_.TrimmedList_ represents a list of unique imports sorted by package name.

<h3><a id="user-content-Object_Interface" target="_self">Interface</a></h3>

_Interface_ (extends [Decl](#user-content-Object_Common_Decl)) represents an interface declaration.

<h6><a id="user-content-Object_Interface_-Methods" target="_self">.Methods</a></h6>

_.Methods_ represents the list of interface methods.

<h6><a id="user-content-Object_Interface_-Type" target="_self">.Type</a></h6>

_.Type_ represents the interface type.

<h3><a id="user-content-Object_InterfaceMethod" target="_self">InterfaceMethod</a></h3>

_InterfaceMethod_ (extends [Node](#user-content-Object_Common_Node)) represents an interface method declaration.

<h6><a id="user-content-Object_InterfaceMethod_-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the line [comment](#user-content-Object_Comment) of the interface method declaration.

<h6><a id="user-content-Object_InterfaceMethod_-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the interface that contains the method.

<h6><a id="user-content-Object_InterfaceMethod_-Name" target="_self">.Name</a></h6>

_.Name_ represents the [NodeName](#user-content-Object_Common_NodeName) of the interface method.

<h6><a id="user-content-Object_InterfaceMethod_-Params" target="_self">.Params</a></h6>

_.Params_ represents the list of method parameters.

<h6><a id="user-content-Object_InterfaceMethod_-Result" target="_self">.Result</a></h6>

_.Result_ represents the return type of the method.

<h3><a id="user-content-Object_InterfaceMethodName" target="_self">InterfaceMethodName</a></h3>

_InterfaceMethodName_ represents the [NodeName](#user-content-Object_Common_NodeName) of an [interface method](#user-content-Object_InterfaceMethod).

<h3><a id="user-content-Object_InterfaceMethodParam" target="_self">InterfaceMethodParam</a></h3>

_InterfaceMethodParam_ (extends [Node](#user-content-Object_Common_Node)) represents an interface method parameter declaration.

<h6><a id="user-content-Object_InterfaceMethodParam_-Method" target="_self">.Method</a></h6>

_.Method_ represents the interface method that contains the parameter.

<h6><a id="user-content-Object_InterfaceMethodParam_-Name" target="_self">.Name</a></h6>

_.Name_ represents the [NodeName](#user-content-Object_Common_NodeName) of the interface method parameter.

<h6><a id="user-content-Object_InterfaceMethodParam_-Type" target="_self">.Type</a></h6>

_.Type_ represents the parameter type.

<h3><a id="user-content-Object_InterfaceMethodParamName" target="_self">InterfaceMethodParamName</a></h3>

_InterfaceMethodParamName_ represents the [NodeName](#user-content-Object_Common_NodeName) of an [interface method parameter](#user-content-Object_InterfaceMethodParam).

<h3><a id="user-content-Object_InterfaceMethodParamType" target="_self">InterfaceMethodParamType</a></h3>

_InterfaceMethodParamType_ represents an interface method parameter type.

<h6><a id="user-content-Object_InterfaceMethodParamType_-Param" target="_self">.Param</a></h6>

_.Param_ represents the interface method parameter that contains the type.

<h6><a id="user-content-Object_InterfaceMethodParamType_-Type" target="_self">.Type</a></h6>

_.Type_ represnts the underlying type of the parameter.

<h3><a id="user-content-Object_InterfaceMethodParams" target="_self">InterfaceMethodParams</a></h3>

_InterfaceMethodParams_ represents the [list](#user-content-Object_Common_Fields) of [interface method parameters](#user-content-Object_InterfaceMethodParam).

<h3><a id="user-content-Object_InterfaceMethodResult" target="_self">InterfaceMethodResult</a></h3>

_InterfaceMethodResult_ represents an interface method result.

<h6><a id="user-content-Object_InterfaceMethodResult_-Method" target="_self">.Method</a></h6>

_.Method_ represents the interface method that contains the result.

<h6><a id="user-content-Object_InterfaceMethodResult_-Type" target="_self">.Type</a></h6>

_.Type_ represents the underlying type of the result.

<h3><a id="user-content-Object_InterfaceMethods" target="_self">InterfaceMethods</a></h3>

_InterfaceMethods_ represents the [list](#user-content-Object_Common_Fields) of [interface methods](#user-content-Object_InterfaceMethod).

<h3><a id="user-content-Object_InterfaceType" target="_self">InterfaceType</a></h3>

_InterfaceType_ represents the [type](#user-content-Object_Common_Type) of an [interface](#user-content-Object_Interface) declaration.

<h3><a id="user-content-Object_Interfaces" target="_self">Interfaces</a></h3>

_Interfaces_ represents a [list](#user-content-Object_Common_List) of interface declarations.

<h3><a id="user-content-Object_MapType" target="_self">MapType</a></h3>

_MapType_ represents a map [type](#user-content-Object_Common_Type).

<h6><a id="user-content-Object_MapType_-ElemType" target="_self">.ElemType</a></h6>

_.ElemType_ represents the element [type](#user-content-Object_Common_Type) of the map.

<h6><a id="user-content-Object_MapType_-KeyType" target="_self">.KeyType</a></h6>

_.KeyType_ represents the key [type](#user-content-Object_Common_Type) of the map.

<h3><a id="user-content-Object_Package" target="_self">Package</a></h3>

_Package_ represents a Next package.

<h6><a id="user-content-Object_Package_-Annotations" target="_self">.Annotations</a></h6>

_.Annotations_ represents the package [annotations](#user-content-Object_Common_Annotations).

<h6><a id="user-content-Object_Package_-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the package [documentation](#user-content-Object_Doc).

<h6><a id="user-content-Object_Package_-Files" target="_self">.Files</a></h6>

_.Files_ represents the all declared files in the package.

<h6><a id="user-content-Object_Package_-In" target="_self">.In</a></h6>

_.In_ reports whether the package is the same as the given package. If the current package is nil, it always returns true. 
Example: 

```npl
{{- define "next/go/used.type" -}}
{{if not (.Type.Decl.File.Package.In .File.Package) -}}
{{.Type.Decl.File.Package.Name -}}.
{{- end -}}
{{next .Type}}
{{- end}}
```

<h6><a id="user-content-Object_Package_-Name" target="_self">.Name</a></h6>

_.Name_ represents the package name string.

<h6><a id="user-content-Object_Package_-Types" target="_self">.Types</a></h6>

_.Types_ represents the all declared types in the package.

<h3><a id="user-content-Object_PrimitiveType" target="_self">PrimitiveType</a></h3>

_PrimitiveType_ represents a primitive type.

<h3><a id="user-content-Object_Struct" target="_self">Struct</a></h3>

_Struct_ (extends [Decl](#user-content-Object_Common_Decl)) represents a struct declaration.

<h6><a id="user-content-Object_Struct_-Fields" target="_self">.Fields</a></h6>

_.Fields_ represents the list of struct fields.

<h6><a id="user-content-Object_Struct_-Type" target="_self">.Type</a></h6>

_.Type_ represents the struct type.

<h3><a id="user-content-Object_StructField" target="_self">StructField</a></h3>

_StructField_ (extends [Node](#user-content-Object_Common_Node)) represents a struct field declaration.

<h6><a id="user-content-Object_StructField_-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the line [comment](#user-content-Object_Comment) of the struct field declaration.

<h6><a id="user-content-Object_StructField_-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the struct that contains the field.

<h6><a id="user-content-Object_StructField_-Name" target="_self">.Name</a></h6>

_.Name_ represents the [NodeName](#user-content-Object_Common_NodeName) of the struct field.

<h6><a id="user-content-Object_StructField_-Type" target="_self">.Type</a></h6>

_.Type_ represents the [struct field type](#user-content-Object_StructFieldType).

<h3><a id="user-content-Object_StructFieldName" target="_self">StructFieldName</a></h3>

_StructFieldName_ represents the [NodeName](#user-content-Object_Common_NodeName) of a [struct field](#user-content-Object_StructField).

<h3><a id="user-content-Object_StructFieldType" target="_self">StructFieldType</a></h3>

_StructFieldType_ represents a struct field type.

<h6><a id="user-content-Object_StructFieldType_-Field" target="_self">.Field</a></h6>

_.Field_ represents the struct field that contains the type.

<h6><a id="user-content-Object_StructFieldType_-Type" target="_self">.Type</a></h6>

_.Type_ represents the underlying type of the struct field.

<h3><a id="user-content-Object_StructFields" target="_self">StructFields</a></h3>

_StructFields_ represents the [list](#user-content-Object_Common_Fields) of [struct fields](#user-content-Object_StructField).

<h3><a id="user-content-Object_StructType" target="_self">StructType</a></h3>

_StructType_ represents the [type](#user-content-Object_Common_Type) of a [struct](#user-content-Object_Struct) declaration.

<h3><a id="user-content-Object_Structs" target="_self">Structs</a></h3>

_Structs_ represents a [list](#user-content-Object_Common_List) of struct declarations.

<h3><a id="user-content-Object_UsedType" target="_self">UsedType</a></h3>

_UsedType_ represents a used type in a file.

<h6><a id="user-content-Object_UsedType_-File" target="_self">.File</a></h6>

_.File_ represents the file containing the used type.

<h6><a id="user-content-Object_UsedType_-Type" target="_self">.Type</a></h6>

_.Type_ represents the used type.

<h3><a id="user-content-Object_Value" target="_self">Value</a></h3>

_Value_ represents a constant value for a const declaration or an enum member.

<h6><a id="user-content-Object_Value_-Any" target="_self">.Any</a></h6>

_.Any_ represents the underlying value of the constant.

<h6><a id="user-content-Object_Value_-IsEnum" target="_self">.IsEnum</a></h6>

_.IsEnum_ returns true if the value is an enum member.

<h6><a id="user-content-Object_Value_-IsFirst" target="_self">.IsFirst</a></h6>

_.IsFirst_ reports whether the value is the first member of the enum type.

<h6><a id="user-content-Object_Value_-IsLast" target="_self">.IsLast</a></h6>

_.IsLast_ reports whether the value is the last member of the enum type.

<h6><a id="user-content-Object_Value_-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the value.

<h6><a id="user-content-Object_Value_-Type" target="_self">.Type</a></h6>

_.Type_ represents the [primitive type](#user-content-Object_PrimitiveType) of the value.

<h3><a id="user-content-Object_VectorType" target="_self">VectorType</a></h3>

_VectorType_ represents a vector [type](#user-content-Object_Common_Type).

<h6><a id="user-content-Object_VectorType_-ElemType" target="_self">.ElemType</a></h6>

_.ElemType_ represents the element [type](#user-content-Object_Common_Type) of the vector.

