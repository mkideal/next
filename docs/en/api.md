# API Reference

<ul>
  <li><a href="#user-content-Functions">Functions</a></li>
<ul>
    <li><a href="#user-content-Functions_ENV">ENV</a></li>
    <li><a href="#user-content-Functions_algin">algin</a></li>
    <li><a href="#user-content-Functions_error">error</a></li>
    <li><a href="#user-content-Functions_errorf">errorf</a></li>
    <li><a href="#user-content-Functions_exist">exist</a></li>
    <li><a href="#user-content-Functions_head">head</a></li>
    <li><a href="#user-content-Functions_lang">lang</a></li>
    <li><a href="#user-content-Functions_meta">meta</a></li>
    <li><a href="#user-content-Functions_next">next</a></li>
    <li><a href="#user-content-Functions_render">render</a></li>
    <li><a href="#user-content-Functions_super">super</a></li>
    <li><a href="#user-content-Functions_this">this</a></li>
    <li><a href="#user-content-Functions_type">type</a></li>
</ul>
  <li><a href="#user-content-Meta">Meta</a></li>
  <li><a href="#user-content-Objects">Objects</a></li>
<ul>
    <li><a href="#user-content-Objects_Annotation">Annotation</a></li>
    <li><a href="#user-content-Objects_Annotations">Annotations</a></li>
    <li><a href="#user-content-Objects_Comment">Comment</a></li>
    <li><a href="#user-content-Objects_Const">Const</a></li>
    <li><a href="#user-content-Objects_ConstName">ConstName</a></li>
    <li><a href="#user-content-Objects_Consts">Consts</a></li>
    <li><a href="#user-content-Objects_Decl">Decl</a></li>
    <li><a href="#user-content-Objects_Decls">Decls</a></li>
    <li><a href="#user-content-Objects_Doc">Doc</a></li>
    <li><a href="#user-content-Objects_Enum">Enum</a></li>
    <li><a href="#user-content-Objects_EnumMember">EnumMember</a></li>
    <li><a href="#user-content-Objects_EnumMemberName">EnumMemberName</a></li>
    <li><a href="#user-content-Objects_EnumMembers">EnumMembers</a></li>
    <li><a href="#user-content-Objects_Enums">Enums</a></li>
    <li><a href="#user-content-Objects_Fields">Fields</a></li>
    <li><a href="#user-content-Objects_File">File</a></li>
    <li><a href="#user-content-Objects_Import">Import</a></li>
    <li><a href="#user-content-Objects_Imports">Imports</a></li>
    <li><a href="#user-content-Objects_Interface">Interface</a></li>
    <li><a href="#user-content-Objects_InterfaceMethod">InterfaceMethod</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodName">InterfaceMethodName</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodParam">InterfaceMethodParam</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodParamName">InterfaceMethodParamName</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodParamType">InterfaceMethodParamType</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodParams">InterfaceMethodParams</a></li>
    <li><a href="#user-content-Objects_InterfaceMethodResult">InterfaceMethodResult</a></li>
    <li><a href="#user-content-Objects_InterfaceMethods">InterfaceMethods</a></li>
    <li><a href="#user-content-Objects_Interfaces">Interfaces</a></li>
    <li><a href="#user-content-Objects_List">List</a></li>
    <li><a href="#user-content-Objects_Node">Node</a></li>
    <li><a href="#user-content-Objects_NodeName">NodeName</a></li>
    <li><a href="#user-content-Objects_Object">Object</a></li>
    <li><a href="#user-content-Objects_Package">Package</a></li>
    <li><a href="#user-content-Objects_Struct">Struct</a></li>
    <li><a href="#user-content-Objects_StructField">StructField</a></li>
    <li><a href="#user-content-Objects_StructFieldName">StructFieldName</a></li>
    <li><a href="#user-content-Objects_StructFieldType">StructFieldType</a></li>
    <li><a href="#user-content-Objects_StructFields">StructFields</a></li>
    <li><a href="#user-content-Objects_Structs">Structs</a></li>
    <li><a href="#user-content-Objects_Symbol">Symbol</a></li>
    <li><a href="#user-content-Objects_Type">Type</a></li>
<ul>
      <li><a href="#user-content-Objects_Type_ArrayType">ArrayType</a></li>
      <li><a href="#user-content-Objects_Type_EnumType">EnumType</a></li>
      <li><a href="#user-content-Objects_Type_InterfaceType">InterfaceType</a></li>
      <li><a href="#user-content-Objects_Type_MapType">MapType</a></li>
      <li><a href="#user-content-Objects_Type_PrimitiveType">PrimitiveType</a></li>
      <li><a href="#user-content-Objects_Type_StructType">StructType</a></li>
      <li><a href="#user-content-Objects_Type_UsedType">UsedType</a></li>
      <li><a href="#user-content-Objects_Type_VectorType">VectorType</a></li>
</ul>
    <li><a href="#user-content-Objects_Value">Value</a></li>
</ul>
</ul>

<h2><a id="user-content-Functions" target="_self">Functions</a></h2>
<h3><a id="user-content-Functions_ENV" target="_self">ENV</a></h3>

ENV represents the environment variables defined in the command line with the flag `-D`.

Example:

```shell
next -D PROJECT_NAME=demo
```

```npl
{{ENV.PROJECT_NAME}}
```

<h3><a id="user-content-Functions_algin" target="_self">algin</a></h3>

`align` aligns the given text with the last line indent of the generated content.

<h3><a id="user-content-Functions_error" target="_self">error</a></h3>

error used to return an error message in the template.

Example:

```npl
{{error "Something went wrong"}}
```

<h3><a id="user-content-Functions_errorf" target="_self">errorf</a></h3>

errorf used to return a formatted error message in the template.

Example:

```npl
{{errorf "%s went wrong" "Something"}}
```

<h3><a id="user-content-Functions_exist" target="_self">exist</a></h3>

exist checks whether the given path exists.
If the path is not absolute, it will be resolved relative to the current output directory
for the current language by command line flag `-O`.

<h3><a id="user-content-Functions_head" target="_self">head</a></h3>

`head` outputs the header of the generated file.

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

<h3><a id="user-content-Functions_lang" target="_self">lang</a></h3>

lang represents the current language to be generated.

<h3><a id="user-content-Functions_meta" target="_self">meta</a></h3>

meta represents the [meta](#user-content-meta) data of the current template.

<h3><a id="user-content-Functions_next" target="_self">next</a></h3>

`next` executes the next template with the given [object](#user-content-Objects).

<h3><a id="user-content-Functions_render" target="_self">render</a></h3>

`render` executes the template with the given name and data.

<h3><a id="user-content-Functions_super" target="_self">super</a></h3>

`super` executes the super template with the given [object](#user-content-Objects).

<h3><a id="user-content-Functions_this" target="_self">this</a></h3>

this represents the current [declaration](#user-content-Objects_Decl) object to be rendered.
`this` defined in the template [meta](#user-content-meta) `meta/this`. Supported types are:

- [file](#user-content-Objects_File)
- [const](#user-content-Objects_Const)
- [enum](#user-content-Objects_Enum)
- [struct](#user-content-Objects_Struct)
- [interface](#user-content-Objects_Interface)

It's a [file](#user-content-Objects_File) by default.

<h3><a id="user-content-Functions_type" target="_self">type</a></h3>

`type` outputs the string representation of the given [type](#user-content-Objects_Type) for the current language.

<h2><a id="user-content-Meta" target="_self">Meta</a></h2>

Meta represents the metadata of a entrypoint template.
To define a meta, you should define a template with the name "meta/<key>".
Currently, the following meta keys are supported:

- "meta/this": the current object to be rendered.
- "meta/path": the output path for the current object.
- "meta/skip": whether to skip the current object.

Any other meta keys are user-defined. You can use them in the templates like `{{meta.<key>}}`.

Example:

```npl
{{- define "meta/this" -}}file{{- end -}}
{{- define "meta/path" -}}/path/to/file{{- end -}}
{{- define "meta/skip" -}}{{exist meta.path}}{{- end -}}
{{- define "meta/custom" -}}custom value{{- end -}}
```

All meta templates should be defined in the entrypoint template.
The meta will be resolved in the order of the template definition
before rendering the entrypoint template.

<h2><a id="user-content-Objects" target="_self">Objects</a></h2>
<h3><a id="user-content-Objects_Annotation" target="_self">Annotation</a></h3>

Annotation represents an annotation by `name` => value.

Example:

Next code:
```next
@json(omitempty)
@event(name="Login")
@message(name="Login", type=100)
```

Will be represented as:
```npl
{{.json.omitempty}}
{{.event.name}}
{{.message.name}}
{{.message.type}}
```

Output:
```
true
Login
Login
100
```

<h3><a id="user-content-Objects_Annotations" target="_self">Annotations</a></h3>

Annotations represents a group of annotations by `name` => [Annotation](#user-content-Objects_Annotation).

<h3><a id="user-content-Objects_Comment" target="_self">Comment</a></h3>

Comment represents a line comment.

<h5><a id="user-content-Objects_Comment-String" target="_self">.String</a></h5>

String returns the origin content of the comment in Next source code.

<h5><a id="user-content-Objects_Comment-Text" target="_self">.Text</a></h5>

Text returns the content of the comment in Next source code.
The content is trimmed by the comment characters.
For example, the comment "// hello comment" will return "hello comment".

<h3><a id="user-content-Objects_Const" target="_self">Const</a></h3>

Const represents a const declaration.

<h5><a id="user-content-Objects_Const-Comment" target="_self">.Comment</a></h5>

Comment is the line [comment](#user-content-Objects_Comment) of the constant declaration.

<h5><a id="user-content-Objects_Const-Name" target="_self">.Name</a></h5>

Name represents the [name object](#user-content-Objects-NodeName) of the constant.

<h5><a id="user-content-Objects_Const-Type" target="_self">.Type</a></h5>

Type represents the type of the constant.

<h5><a id="user-content-Objects_Const-Value" target="_self">.Value</a></h5>

Value represents the [value object](#user-content-Objects_Value) of the constant.

<h3><a id="user-content-Objects_ConstName" target="_self">ConstName</a></h3>

`ConstName` represents the [name object](#user-content-Objects_NodeName) of a [const](#user-content-Objects_Const) declaration.

<h3><a id="user-content-Objects_Consts" target="_self">Consts</a></h3>

`Consts` represents a [list](#user-content-Objects_List) of const declarations.

<h3><a id="user-content-Objects_Decl" target="_self">Decl</a></h3>

Decl represents an top-level declaration in a file.

<h3><a id="user-content-Objects_Decls" target="_self">Decls</a></h3>

Decls holds all declarations in a file.

<h5><a id="user-content-Objects_Decls-Consts" target="_self">.Consts</a></h5>

`Consts` represents the [list](#user-content-Objects_List) of [const](#user-content-Objects_Const) declarations.

<h5><a id="user-content-Objects_Decls-Enums" target="_self">.Enums</a></h5>

`Enums` represents the [list](#user-content-Objects_List) of [enum](#user-content-Objects_Enum) declarations.

<h5><a id="user-content-Objects_Decls-Interfaces" target="_self">.Interfaces</a></h5>

`Interfaces` represents the [list](#user-content-Objects_List) of [interface](#user-content-Objects_Interface) declarations.

<h5><a id="user-content-Objects_Decls-Structs" target="_self">.Structs</a></h5>

`Structs` represents the [list](#user-content-Objects_List) of [struct](#user-content-Objects_Struct) declarations.

<h3><a id="user-content-Objects_Doc" target="_self">Doc</a></h3>

Doc represents a documentation comment for a declaration.

Example:

```next

// This is a documentation comment.
// It can be multiple lines.
struct User {
	// This is a field documentation comment.
	// It can be multiple lines.
	name string
}
```

<h5><a id="user-content-Objects_Doc-Format" target="_self">.Format</a></h5>

Format formats the documentation comment with the given prefix, ident, and begin and end strings.

Example:

```next

// This is a documentation comment.
// It can be multiple lines.
struct User {
	// This is a field documentation comment.
	// It can be multiple lines.
	name string
}
```

```npl
{{- define "next/c/doc" -}}
{{.Format "" " * " "/**\n" " */" | align}}
{{- end}}
```

Output:

```c

/**
 * This is a documentation comment.
 * It can be multiple lines.
 */
typedef struct {
	/**
	 * This is a field documentation comment.
	 * It can be multiple lines.
	 */
	char *name;
} User;
```

<h5><a id="user-content-Objects_Doc-String" target="_self">.String</a></h5>

String returns the origin content of the documentation comment in Next source code.

<h5><a id="user-content-Objects_Doc-Text" target="_self">.Text</a></h5>

Text returns the content of the documentation comment in Next source code.
The content is trimmed by the comment characters.
For example, the comment "// hello comment" will return "hello comment".

<h3><a id="user-content-Objects_Enum" target="_self">Enum</a></h3>

Enum represents an enum declaration.

<h5><a id="user-content-Objects_Enum-Members" target="_self">.Members</a></h5>

Members is the list of enum members.

<h5><a id="user-content-Objects_Enum-Type" target="_self">.Type</a></h5>

Type is the enum type.

<h3><a id="user-content-Objects_EnumMember" target="_self">EnumMember</a></h3>

EnumMember represents an enum member object in an [enum](#user-content-Objects_Enum) declaration.

<h5><a id="user-content-Objects_EnumMember-Comment" target="_self">.Comment</a></h5>

Comment represents the line [comment](#user-content-Objects_Comment) of the enum member declaration.

<h5><a id="user-content-Objects_EnumMember-Decl" target="_self">.Decl</a></h5>

Decl represents the [enum](#user-content-Objects_Enum) that contains the member.

<h5><a id="user-content-Objects_EnumMember-Name" target="_self">.Name</a></h5>

Name represents the [name object](#user-content-Objects_NodeName) of the enum member.

<h5><a id="user-content-Objects_EnumMember-Value" target="_self">.Value</a></h5>

Value represents the [value object](#user-content-Objects_Value) of the enum member.

<h3><a id="user-content-Objects_EnumMemberName" target="_self">EnumMemberName</a></h3>

`EnumMemberName` represents the [name object](#user-content-Objects_NodeName) of an [enum member](#user-content-Objects_EnumMember).

<h3><a id="user-content-Objects_EnumMembers" target="_self">EnumMembers</a></h3>

`EnumMembers` represents the [list](#user-content-Objects_Fields) of [enum members](#user-content-Objects_EnumMember).

<h3><a id="user-content-Objects_Enums" target="_self">Enums</a></h3>

`Enums` represents a [list](#user-content-Objects_List) of enum declarations.

<h3><a id="user-content-Objects_Fields" target="_self">Fields</a></h3>

Fields represents a list of fields in a declaration.

<h5><a id="user-content-Objects_Fields-Decl" target="_self">.Decl</a></h5>

Decl is the declaration object that contains the fields.
Decl may be an enum, struct, or interface.

<h5><a id="user-content-Objects_Fields-List" target="_self">.List</a></h5>

List is the list of fields in the declaration.

<h3><a id="user-content-Objects_File" target="_self">File</a></h3>

File represents a Next source file.

<h5><a id="user-content-Objects_File-Decls" target="_self">.Decls</a></h5>

Decls returns the file's all top-level declarations.

<h5><a id="user-content-Objects_File-LookupLocalType" target="_self">.LookupLocalType</a></h5>

LookupLocalType looks up a type by name in the file's symbol table.
If the type is not found, it returns an error. If the symbol
is found but it is not a type, it returns an error.

<h5><a id="user-content-Objects_File-LookupLocalValue" target="_self">.LookupLocalValue</a></h5>

LookupLocalValue looks up a value by name in the file's symbol table.
If the value is not found, it returns an error. If the symbol
is found but it is not a value, it returns an error.

<h5><a id="user-content-Objects_File-Name" target="_self">.Name</a></h5>

Name represents the file name without the ".next" extension.

<h5><a id="user-content-Objects_File-Package" target="_self">.Package</a></h5>

Imports represents the file's import declarations.

<h5><a id="user-content-Objects_File-Path" target="_self">.Path</a></h5>

Path represents the file full path.

<h3><a id="user-content-Objects_Import" target="_self">Import</a></h3>

Import represents a file import.

<h5><a id="user-content-Objects_Import-Doc" target="_self">.Doc</a></h5>

Doc represents the import declaration [documentation](#user-content-Objects_Doc).

<h5><a id="user-content-Objects_Import-File" target="_self">.File</a></h5>

File represents the file containing the import declaration.

<h5><a id="user-content-Objects_Import-FullPath" target="_self">.FullPath</a></h5>

FullPath represents the full path of the import.

<h5><a id="user-content-Objects_Import-Target" target="_self">.Target</a></h5>

Target represents the imported file.

<h3><a id="user-content-Objects_Imports" target="_self">Imports</a></h3>

Imports holds a list of imports.

<h5><a id="user-content-Objects_Imports-File" target="_self">.File</a></h5>

File represents the file containing the imports.

<h5><a id="user-content-Objects_Imports-List" target="_self">.List</a></h5>

List represents the list of [imports](#user-content-Objects_Import).

<h5><a id="user-content-Objects_Imports-TrimmedList" target="_self">.TrimmedList</a></h5>

TrimmedList represents a list of unique imports sorted by package name.

<h3><a id="user-content-Objects_Interface" target="_self">Interface</a></h3>

Interface represents an interface declaration.

<h5><a id="user-content-Objects_Interface-Methods" target="_self">.Methods</a></h5>

Methods represents the list of interface methods.

<h5><a id="user-content-Objects_Interface-Type" target="_self">.Type</a></h5>

Type represents the interface type.

<h3><a id="user-content-Objects_InterfaceMethod" target="_self">InterfaceMethod</a></h3>

InterfaceMethod represents an interface method declaration.

<h5><a id="user-content-Objects_InterfaceMethod-Comment" target="_self">.Comment</a></h5>

Comment represents the line [comment](#user-content-Objects_Comment) of the interface method declaration.

<h5><a id="user-content-Objects_InterfaceMethod-Decl" target="_self">.Decl</a></h5>

Decl represents the interface that contains the method.

<h5><a id="user-content-Objects_InterfaceMethod-Name" target="_self">.Name</a></h5>

Name represents the [name object](#user-content-Objects_NodeName) of the interface method.

<h5><a id="user-content-Objects_InterfaceMethod-Params" target="_self">.Params</a></h5>

Params represents the list of method parameters.

<h5><a id="user-content-Objects_InterfaceMethod-Result" target="_self">.Result</a></h5>

Result represents the return type of the method.

<h3><a id="user-content-Objects_InterfaceMethodName" target="_self">InterfaceMethodName</a></h3>

`InterfaceMethodName` represents the [name object](#user-content-Objects_NodeName) of an [interface method](#user-content-Objects_InterfaceMethod).

<h3><a id="user-content-Objects_InterfaceMethodParam" target="_self">InterfaceMethodParam</a></h3>

InterfaceMethodParam represents an interface method parameter declaration.

<h5><a id="user-content-Objects_InterfaceMethodParam-Method" target="_self">.Method</a></h5>

Method represents the interface method that contains the parameter.

<h5><a id="user-content-Objects_InterfaceMethodParam-Name" target="_self">.Name</a></h5>

Name represents the [name object](#user-content-Objects_NodeName) of the interface method parameter.

<h5><a id="user-content-Objects_InterfaceMethodParam-Type" target="_self">.Type</a></h5>

Type represents the parameter type.

<h3><a id="user-content-Objects_InterfaceMethodParamName" target="_self">InterfaceMethodParamName</a></h3>

`InterfaceMethodParamName` represents the [name object](#user-content-Objects_NodeName) of an [interface method parameter](#user-content-Objects_InterfaceMethodParam).

<h3><a id="user-content-Objects_InterfaceMethodParamType" target="_self">InterfaceMethodParamType</a></h3>

InterfaceMethodParamType represents an interface method parameter type.

<h5><a id="user-content-Objects_InterfaceMethodParamType-Param" target="_self">.Param</a></h5>

Param represents the interface method parameter that contains the type.

<h5><a id="user-content-Objects_InterfaceMethodParamType-Type" target="_self">.Type</a></h5>

Type represnts the underlying type of the parameter.

<h3><a id="user-content-Objects_InterfaceMethodParams" target="_self">InterfaceMethodParams</a></h3>

`InterfaceMethodParams` represents the [list](#user-content-Objects_Fields) of [interface method parameters](#user-content-Objects_InterfaceMethodParam).

<h3><a id="user-content-Objects_InterfaceMethodResult" target="_self">InterfaceMethodResult</a></h3>

InterfaceMethodResult represents an interface method result.

<h5><a id="user-content-Objects_InterfaceMethodResult-Method" target="_self">.Method</a></h5>

Method represents the interface method that contains the result.

<h5><a id="user-content-Objects_InterfaceMethodResult-Type" target="_self">.Type</a></h5>

Type represents the underlying type of the result.

<h3><a id="user-content-Objects_InterfaceMethods" target="_self">InterfaceMethods</a></h3>

`InterfaceMethods` represents the [list](#user-content-Objects_Fields) of [interface methods](#user-content-Objects_InterfaceMethod).

<h3><a id="user-content-Objects_Interfaces" target="_self">Interfaces</a></h3>

`Interfaces` represents a [list](#user-content-Objects_List) of interface declarations.

<h3><a id="user-content-Objects_List" target="_self">List</a></h3>

List represents a list of objects.

<h5><a id="user-content-Objects_List-List" target="_self">.List</a></h5>

List represents the list of objects. It is used to provide a uniform way to access.

<h3><a id="user-content-Objects_Node" target="_self">Node</a></h3>

Node represents a Node in the AST. It's a special object that can be annotated with a documentation comment.

<h5><a id="user-content-Objects_Node-Annotations" target="_self">.Annotations</a></h5>

Annotations represents the annotations for the node.

<h5><a id="user-content-Objects_Node-Doc" target="_self">.Doc</a></h5>

Doc represents the documentation comment for the node.

<h5><a id="user-content-Objects_Node-File" target="_self">.File</a></h5>

File represents the file containing the node.

<h5><a id="user-content-Objects_Node-Package" target="_self">.Package</a></h5>

Package represents the package containing the node.
It's a shortcut for Node.File().Package().

<h3><a id="user-content-Objects_NodeName" target="_self">NodeName</a></h3>

NodeName represents a name of a node in a declaration:
- Const name
- Enum member name
- Struct field name
- Interface method name
- Interface method parameter name

<h5><a id="user-content-Objects_NodeName-Node" target="_self">.Node</a></h5>

Node represents the [node](#user-content-Objects_Node) that contains the name.

<h5><a id="user-content-Objects_NodeName-String" target="_self">.String</a></h5>

String represents the string representation of the node name.

<h3><a id="user-content-Objects_Object" target="_self">Object</a></h3>

`Object` represents an object in Next which can be rendered in a template like this: {{next Object}}

<h3><a id="user-content-Objects_Package" target="_self">Package</a></h3>

Package represents a Next package.

<h5><a id="user-content-Objects_Package-Annotations" target="_self">.Annotations</a></h5>

Annotations represents the package [annotations](#user-content-Objects_Annotations).

<h5><a id="user-content-Objects_Package-Doc" target="_self">.Doc</a></h5>

Doc represents the package [documentation](#user-content-Objects_Doc).

<h5><a id="user-content-Objects_Package-Files" target="_self">.Files</a></h5>

Files represents the all declared files in the package.

<h5><a id="user-content-Objects_Package-In" target="_self">.In</a></h5>

In reports whether the package is the same as the given package.
If the current package is nil, it always returns true.

Example:

```npl
{{- define "next/go/used.type" -}}
{{if not (.Type.Decl.File.Package.In .File.Package) -}}
{{.Type.Decl.File.Package.Name -}}.
{{- end -}}
{{next .Type}}
{{- end}}
```

<h5><a id="user-content-Objects_Package-Name" target="_self">.Name</a></h5>

Name represents the package name string.

<h5><a id="user-content-Objects_Package-Types" target="_self">.Types</a></h5>

Types represents the all declared types in the package.

<h3><a id="user-content-Objects_Struct" target="_self">Struct</a></h3>

Struct represents a struct declaration.

<h5><a id="user-content-Objects_Struct-Fields" target="_self">.Fields</a></h5>

Fields represents the list of struct fields.

<h5><a id="user-content-Objects_Struct-Type" target="_self">.Type</a></h5>

Type represents the struct type.

<h3><a id="user-content-Objects_StructField" target="_self">StructField</a></h3>

StructField represents a struct field declaration.

<h5><a id="user-content-Objects_StructField-Comment" target="_self">.Comment</a></h5>

Comment represents the line [comment](#user-content-Objects_Comment) of the struct field declaration.

<h5><a id="user-content-Objects_StructField-Decl" target="_self">.Decl</a></h5>

Decl represents the struct that contains the field.

<h5><a id="user-content-Objects_StructField-Name" target="_self">.Name</a></h5>

Name represents the [name object](#user-content-Objects_NodeName) of the struct field.

<h5><a id="user-content-Objects_StructField-Type" target="_self">.Type</a></h5>

Type represents the [struct field type](#user-content-StructFieldType).

<h3><a id="user-content-Objects_StructFieldName" target="_self">StructFieldName</a></h3>

`StructFieldName` represents the [name object](#user-content-Objects_NodeName) of a [struct field](#user-content-Objects_StructField).

<h3><a id="user-content-Objects_StructFieldType" target="_self">StructFieldType</a></h3>

StructFieldType represents a struct field type.

<h5><a id="user-content-Objects_StructFieldType-Field" target="_self">.Field</a></h5>

Field represents the struct field that contains the type.

<h5><a id="user-content-Objects_StructFieldType-Type" target="_self">.Type</a></h5>

Type represents the underlying type of the struct field.

<h3><a id="user-content-Objects_StructFields" target="_self">StructFields</a></h3>

`StructFields` represents the [list](#user-content-Objects_Fields) of [struct fields](#user-content-Objects_StructField).

<h3><a id="user-content-Objects_Structs" target="_self">Structs</a></h3>

`Structs` represents a [list](#user-content-Objects_List) of struct declarations.

<h3><a id="user-content-Objects_Symbol" target="_self">Symbol</a></h3>

Symbol represents a Next symbol: value(const, enum member), type(enum, struct, interface).

<h3><a id="user-content-Objects_Type" target="_self">Type</a></h3>

Type represents a Next type.

<h4><a id="user-content-Objects_Type_ArrayType" target="_self">ArrayType</a></h4>

ArrayType represents an array type.

<h4><a id="user-content-Objects_Type_EnumType" target="_self">EnumType</a></h4>

`EnumType` represents the type of an [enum](#user-content-Objects_Enum) declaration.

<h4><a id="user-content-Objects_Type_InterfaceType" target="_self">InterfaceType</a></h4>

`InterfaceType` represents the type of an [interface](#user-content-Objects_Interface) declaration.

<h4><a id="user-content-Objects_Type_MapType" target="_self">MapType</a></h4>

MapType represents a map type.

<h4><a id="user-content-Objects_Type_PrimitiveType" target="_self">PrimitiveType</a></h4>

PrimitiveType represents a primitive type.

<h4><a id="user-content-Objects_Type_StructType" target="_self">StructType</a></h4>

`StructType` represents the type of a [struct](#user-content-Objects_Struct) declaration.

<h4><a id="user-content-Objects_Type_UsedType" target="_self">UsedType</a></h4>

UsedType represents a used type in a file.

<h4><a id="user-content-Objects_Type_VectorType" target="_self">VectorType</a></h4>

VectorType represents a vector type.

<h5><a id="user-content-Objects_Type-Decl" target="_self">.Decl</a></h5>

Decl represents the [declaration](#user-content-Decl) of the type.

<h5><a id="user-content-Objects_Type-Kind" target="_self">.Kind</a></h5>

Kind returns the kind of the type.

<h5><a id="user-content-Objects_Type-String" target="_self">.String</a></h5>

String represents the string representation of the type.

<h5><a id="user-content-Objects_UsedType-File" target="_self">.File</a></h5>

`File` represents the file containing the used type.

<h5><a id="user-content-Objects_UsedType-Type" target="_self">.Type</a></h5>

`Type` represents the used type.

<h3><a id="user-content-Objects_Value" target="_self">Value</a></h3>

Value represents a constant value for a const declaration or an enum member.

<h5><a id="user-content-Objects_Value-Any" target="_self">.Any</a></h5>

Any represents the underlying value of the constant.

<h5><a id="user-content-Objects_Value-IsEnum" target="_self">.IsEnum</a></h5>

IsEnum returns true if the value is an enum member.

<h5><a id="user-content-Objects_Value-IsFirst" target="_self">.IsFirst</a></h5>

IsFirst reports whether the value is the first member of the enum type.

<h5><a id="user-content-Objects_Value-IsLast" target="_self">.IsLast</a></h5>

IsLast reports whether the value is the last member of the enum type.

<h5><a id="user-content-Objects_Value-String" target="_self">.String</a></h5>

String represents the string representation of the value.

<h5><a id="user-content-Objects_Value-Type" target="_self">.Type</a></h5>

Type represents the type of the value.

<h5><a id="user-content-Objects_import-Comment" target="_self">.Comment</a></h5>

Comment represents the import declaration line [comment](#user-content-Objects_Comment).

<h5><a id="user-content-Objects_import-Path" target="_self">.Path</a></h5>

Path represents the import path.

