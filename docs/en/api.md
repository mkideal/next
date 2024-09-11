# API Reference

<ul>
  <li><a href="#user-content-Function">Function</a></li>
<ul>
    <li><a href="#user-content-Function_align">align</a></li>
    <li><a href="#user-content-Function_env">env</a></li>
    <li><a href="#user-content-Function_error">error</a></li>
    <li><a href="#user-content-Function_errorf">errorf</a></li>
    <li><a href="#user-content-Function_exist">exist</a></li>
    <li><a href="#user-content-Function_head">head</a></li>
    <li><a href="#user-content-Function_lang">lang</a></li>
    <li><a href="#user-content-Function_meta">meta</a></li>
    <li><a href="#user-content-Function_next">next</a></li>
    <li><a href="#user-content-Function_render">render</a></li>
    <li><a href="#user-content-Function_super">super</a></li>
    <li><a href="#user-content-Function_this">this</a></li>
    <li><a href="#user-content-Function_type">type</a></li>
</ul>
  <li><a href="#user-content-Object">Object</a></li>
<ul>
    <li><a href="#user-content-Object_Annotation">Annotation</a></li>
    <li><a href="#user-content-Object_Annotations">Annotations</a></li>
    <li><a href="#user-content-Object_Comment">Comment</a></li>
    <li><a href="#user-content-Object_Const">Const</a></li>
    <li><a href="#user-content-Object_ConstName">ConstName</a></li>
    <li><a href="#user-content-Object_Consts">Consts</a></li>
    <li><a href="#user-content-Object_Decl">Decl</a></li>
    <li><a href="#user-content-Object_Decls">Decls</a></li>
    <li><a href="#user-content-Object_Doc">Doc</a></li>
    <li><a href="#user-content-Object_Enum">Enum</a></li>
    <li><a href="#user-content-Object_EnumMember">EnumMember</a></li>
    <li><a href="#user-content-Object_EnumMemberName">EnumMemberName</a></li>
    <li><a href="#user-content-Object_EnumMembers">EnumMembers</a></li>
    <li><a href="#user-content-Object_Enums">Enums</a></li>
    <li><a href="#user-content-Object_Fields">Fields</a></li>
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
    <li><a href="#user-content-Object_Interfaces">Interfaces</a></li>
    <li><a href="#user-content-Object_List">List</a></li>
    <li><a href="#user-content-Object_Node">Node</a></li>
    <li><a href="#user-content-Object_NodeName">NodeName</a></li>
    <li><a href="#user-content-Object_Object">Object</a></li>
    <li><a href="#user-content-Object_Package">Package</a></li>
    <li><a href="#user-content-Object_Struct">Struct</a></li>
    <li><a href="#user-content-Object_StructField">StructField</a></li>
    <li><a href="#user-content-Object_StructFieldName">StructFieldName</a></li>
    <li><a href="#user-content-Object_StructFieldType">StructFieldType</a></li>
    <li><a href="#user-content-Object_StructFields">StructFields</a></li>
    <li><a href="#user-content-Object_Structs">Structs</a></li>
    <li><a href="#user-content-Object_Symbol">Symbol</a></li>
    <li><a href="#user-content-Object_Type">Type</a></li>
<ul>
      <li><a href="#user-content-Object_Type_ArrayType">ArrayType</a></li>
      <li><a href="#user-content-Object_Type_EnumType">EnumType</a></li>
      <li><a href="#user-content-Object_Type_InterfaceType">InterfaceType</a></li>
      <li><a href="#user-content-Object_Type_MapType">MapType</a></li>
      <li><a href="#user-content-Object_Type_PrimitiveType">PrimitiveType</a></li>
      <li><a href="#user-content-Object_Type_StructType">StructType</a></li>
      <li><a href="#user-content-Object_Type_UsedType">UsedType</a></li>
      <li><a href="#user-content-Object_Type_VectorType">VectorType</a></li>
</ul>
    <li><a href="#user-content-Object_Value">Value</a></li>
</ul>
</ul>

<h2><a id="user-content-Function" target="_self">Function</a></h2>
<h3><a id="user-content-Function_align" target="_self">align</a></h3>

_align_ aligns the given text with the last line indent of the generated content.

<h3><a id="user-content-Function_env" target="_self">env</a></h3>

_env_ represents the environment variables defined in the command line with the flag `-D`.

Example:

```shell
next -D PROJECT_NAME=demo
```

```npl
{{env.PROJECT_NAME}}
```

<h3><a id="user-content-Function_error" target="_self">error</a></h3>

_error_ used to return an error message in the template.

Example:

```npl
{{error "Something went wrong"}}
```

<h3><a id="user-content-Function_errorf" target="_self">errorf</a></h3>

_errorf_ used to return a formatted error message in the template.

Example:

```npl
{{errorf "%s went wrong" "Something"}}
```

<h3><a id="user-content-Function_exist" target="_self">exist</a></h3>

_exist_ checks whether the given path exists.
If the path is not absolute, it will be resolved relative to the current output directory
for the current language by command line flag `-O`.

<h3><a id="user-content-Function_head" target="_self">head</a></h3>

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

<h3><a id="user-content-Function_lang" target="_self">lang</a></h3>

_lang_ represents the current language to be generated.

<h3><a id="user-content-Function_meta" target="_self">meta</a></h3>

_meta_ represents the metadata of a entrypoint template.
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

<h3><a id="user-content-Function_next" target="_self">next</a></h3>

_next_ executes the next template with the given [object](#user-content-Object).

<h3><a id="user-content-Function_render" target="_self">render</a></h3>

_render_ executes the template with the given name and data.

<h3><a id="user-content-Function_super" target="_self">super</a></h3>

_super_ executes the super template with the given [object](#user-content-Object).

<h3><a id="user-content-Function_this" target="_self">this</a></h3>

_this_ represents the current [declaration](#user-content-Object_Decl) object to be rendered.
this defined in the template [meta](#user-content-meta) `meta/this`. Supported types are:

- [file](#user-content-Object_File)
- [const](#user-content-Object_Const)
- [enum](#user-content-Object_Enum)
- [struct](#user-content-Object_Struct)
- [interface](#user-content-Object_Interface)

It's a [file](#user-content-Object_File) by default.

<h3><a id="user-content-Function_type" target="_self">type</a></h3>

_type_ outputs the string representation of the given [type](#user-content-Object_Type) for the current language.

<h2><a id="user-content-Object" target="_self">Object</a></h2>
<h3><a id="user-content-Object_Annotation" target="_self">Annotation</a></h3>

_Annotation_ represents an annotation by `name` => value.

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

<h3><a id="user-content-Object_Annotations" target="_self">Annotations</a></h3>

_Annotations_ represents a group of annotations by `name` => [Annotation](#user-content-Object_Annotation).

<h3><a id="user-content-Object_Comment" target="_self">Comment</a></h3>

_Comment_ represents a line comment.

<h6><a id="user-content-Object_Comment-String" target="_self">.String</a></h6>

_.String_ returns the origin content of the comment in Next source code.

<h6><a id="user-content-Object_Comment-Text" target="_self">.Text</a></h6>

_.Text_ returns the content of the comment in Next source code.
The content is trimmed by the comment characters.
For example, the comment "// hello comment" will return "hello comment".

<h3><a id="user-content-Object_Const" target="_self">Const</a></h3>

_Const_ represents a const declaration.

<h6><a id="user-content-Object_Const-Comment" target="_self">.Comment</a></h6>

_.Comment_ is the line [comment](#user-content-Object_Comment) of the constant declaration.

<h6><a id="user-content-Object_Const-Name" target="_self">.Name</a></h6>

_.Name_ represents the [name object](#user-content-Object-NodeName) of the constant.

<h6><a id="user-content-Object_Const-Type" target="_self">.Type</a></h6>

_.Type_ represents the type of the constant.

<h6><a id="user-content-Object_Const-Value" target="_self">.Value</a></h6>

_.Value_ represents the [value object](#user-content-Object_Value) of the constant.

<h3><a id="user-content-Object_ConstName" target="_self">ConstName</a></h3>

_ConstName_ represents the [name object](#user-content-Object_NodeName) of a [const](#user-content-Object_Const) declaration.

<h3><a id="user-content-Object_Consts" target="_self">Consts</a></h3>

_Consts_ represents a [list](#user-content-Object_List) of const declarations.

<h3><a id="user-content-Object_Decl" target="_self">Decl</a></h3>

_Decl_ represents an top-level declaration in a file.

<h3><a id="user-content-Object_Decls" target="_self">Decls</a></h3>

_Decls_ holds all declarations in a file.

<h6><a id="user-content-Object_Decls-Consts" target="_self">.Consts</a></h6>

_.Consts_ represents the [list](#user-content-Object_List) of [const](#user-content-Object_Const) declarations.

<h6><a id="user-content-Object_Decls-Enums" target="_self">.Enums</a></h6>

_.Enums_ represents the [list](#user-content-Object_List) of [enum](#user-content-Object_Enum) declarations.

<h6><a id="user-content-Object_Decls-Interfaces" target="_self">.Interfaces</a></h6>

_.Interfaces_ represents the [list](#user-content-Object_List) of [interface](#user-content-Object_Interface) declarations.

<h6><a id="user-content-Object_Decls-Structs" target="_self">.Structs</a></h6>

_.Structs_ represents the [list](#user-content-Object_List) of [struct](#user-content-Object_Struct) declarations.

<h3><a id="user-content-Object_Doc" target="_self">Doc</a></h3>

_Doc_ represents a documentation comment for a declaration.

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

<h6><a id="user-content-Object_Doc-Format" target="_self">.Format</a></h6>

_.Format_ formats the documentation comment with the given prefix, ident, and begin and end strings.

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

<h6><a id="user-content-Object_Doc-String" target="_self">.String</a></h6>

_.String_ returns the origin content of the documentation comment in Next source code.

<h6><a id="user-content-Object_Doc-Text" target="_self">.Text</a></h6>

_.Text_ returns the content of the documentation comment in Next source code.
The content is trimmed by the comment characters.
For example, the comment "// hello comment" will return "hello comment".

<h3><a id="user-content-Object_Enum" target="_self">Enum</a></h3>

_Enum_ represents an enum declaration.

<h3><a id="user-content-Object_EnumMember" target="_self">EnumMember</a></h3>

_EnumMember_ represents an enum member object in an [enum](#user-content-Object_Enum) declaration.

<h6><a id="user-content-Object_EnumMember-Name" target="_self">.Name</a></h6>

_.Name_ represents the [name object](#user-content-Object_NodeName) of the enum member.

<h6><a id="user-content-Object_EnumMember-Value" target="_self">.Value</a></h6>

_.Value_ represents the [value object](#user-content-Object_Value) of the enum member.

<h3><a id="user-content-Object_EnumMemberName" target="_self">EnumMemberName</a></h3>

_EnumMemberName_ represents the [name object](#user-content-Object_NodeName) of an [enum member](#user-content-Object_EnumMember).

<h3><a id="user-content-Object_EnumMembers" target="_self">EnumMembers</a></h3>

_EnumMembers_ represents the [list](#user-content-Object_Fields) of [enum members](#user-content-Object_EnumMember).

<h3><a id="user-content-Object_Enums" target="_self">Enums</a></h3>

_Enums_ represents a [list](#user-content-Object_List) of enum declarations.

<h3><a id="user-content-Object_Fields" target="_self">Fields</a></h3>

_Fields_ represents a list of fields in a declaration.

<h6><a id="user-content-Object_Fields-Decl" target="_self">.Decl</a></h6>

_.Decl_ is the declaration object that contains the fields.
Decl may be an enum, struct, or interface.

<h6><a id="user-content-Object_Fields-List" target="_self">.List</a></h6>

_.List_ is the list of fields in the declaration.

<h3><a id="user-content-Object_File" target="_self">File</a></h3>

_File_ represents a Next source file.

<h6><a id="user-content-Object_File-Decls" target="_self">.Decls</a></h6>

_.Decls_ returns the file's all top-level declarations.

<h6><a id="user-content-Object_File-LookupLocalType" target="_self">.LookupLocalType</a></h6>

_.LookupLocalType_ looks up a type by name in the file's symbol table.
If the type is not found, it returns an error. If the symbol
is found but it is not a type, it returns an error.

<h6><a id="user-content-Object_File-LookupLocalValue" target="_self">.LookupLocalValue</a></h6>

_.LookupLocalValue_ looks up a value by name in the file's symbol table.
If the value is not found, it returns an error. If the symbol
is found but it is not a value, it returns an error.

<h6><a id="user-content-Object_File-Name" target="_self">.Name</a></h6>

_.Name_ represents the file name without the ".next" extension.

<h6><a id="user-content-Object_File-Package" target="_self">.Package</a></h6>

_.Package_ represents the file's import declarations.

<h6><a id="user-content-Object_File-Path" target="_self">.Path</a></h6>

_.Path_ represents the file full path.

<h3><a id="user-content-Object_Import" target="_self">Import</a></h3>

_Import_ represents a file import.

<h6><a id="user-content-Object_Import-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the import declaration [documentation](#user-content-Object_Doc).

<h6><a id="user-content-Object_Import-File" target="_self">.File</a></h6>

_.File_ represents the file containing the import declaration.

<h6><a id="user-content-Object_Import-FullPath" target="_self">.FullPath</a></h6>

_.FullPath_ represents the full path of the import.

<h6><a id="user-content-Object_Import-Target" target="_self">.Target</a></h6>

_.Target_ represents the imported file.

<h3><a id="user-content-Object_Imports" target="_self">Imports</a></h3>

_Imports_ holds a list of imports.

<h6><a id="user-content-Object_Imports-File" target="_self">.File</a></h6>

_.File_ represents the file containing the imports.

<h6><a id="user-content-Object_Imports-List" target="_self">.List</a></h6>

_.List_ represents the list of [imports](#user-content-Object_Import).

<h6><a id="user-content-Object_Imports-TrimmedList" target="_self">.TrimmedList</a></h6>

_.TrimmedList_ represents a list of unique imports sorted by package name.

<h3><a id="user-content-Object_Interface" target="_self">Interface</a></h3>

_Interface_ represents an interface declaration.

<h6><a id="user-content-Object_Interface-Methods" target="_self">.Methods</a></h6>

_.Methods_ represents the list of interface methods.

<h6><a id="user-content-Object_Interface-Type" target="_self">.Type</a></h6>

_.Type_ represents the interface type.

<h3><a id="user-content-Object_InterfaceMethod" target="_self">InterfaceMethod</a></h3>

_InterfaceMethod_ represents an interface method declaration.

<h6><a id="user-content-Object_InterfaceMethod-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the line [comment](#user-content-Object_Comment) of the interface method declaration.

<h6><a id="user-content-Object_InterfaceMethod-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the interface that contains the method.

<h6><a id="user-content-Object_InterfaceMethod-Name" target="_self">.Name</a></h6>

_.Name_ represents the [name object](#user-content-Object_NodeName) of the interface method.

<h6><a id="user-content-Object_InterfaceMethod-Params" target="_self">.Params</a></h6>

_.Params_ represents the list of method parameters.

<h6><a id="user-content-Object_InterfaceMethod-Result" target="_self">.Result</a></h6>

_.Result_ represents the return type of the method.

<h3><a id="user-content-Object_InterfaceMethodName" target="_self">InterfaceMethodName</a></h3>

_InterfaceMethodName_ represents the [name object](#user-content-Object_NodeName) of an [interface method](#user-content-Object_InterfaceMethod).

<h3><a id="user-content-Object_InterfaceMethodParam" target="_self">InterfaceMethodParam</a></h3>

_InterfaceMethodParam_ represents an interface method parameter declaration.

<h6><a id="user-content-Object_InterfaceMethodParam-Method" target="_self">.Method</a></h6>

_.Method_ represents the interface method that contains the parameter.

<h6><a id="user-content-Object_InterfaceMethodParam-Name" target="_self">.Name</a></h6>

_.Name_ represents the [name object](#user-content-Object_NodeName) of the interface method parameter.

<h6><a id="user-content-Object_InterfaceMethodParam-Type" target="_self">.Type</a></h6>

_.Type_ represents the parameter type.

<h3><a id="user-content-Object_InterfaceMethodParamName" target="_self">InterfaceMethodParamName</a></h3>

_InterfaceMethodParamName_ represents the [name object](#user-content-Object_NodeName) of an [interface method parameter](#user-content-Object_InterfaceMethodParam).

<h3><a id="user-content-Object_InterfaceMethodParamType" target="_self">InterfaceMethodParamType</a></h3>

_InterfaceMethodParamType_ represents an interface method parameter type.

<h6><a id="user-content-Object_InterfaceMethodParamType-Param" target="_self">.Param</a></h6>

_.Param_ represents the interface method parameter that contains the type.

<h6><a id="user-content-Object_InterfaceMethodParamType-Type" target="_self">.Type</a></h6>

_.Type_ represnts the underlying type of the parameter.

<h3><a id="user-content-Object_InterfaceMethodParams" target="_self">InterfaceMethodParams</a></h3>

_InterfaceMethodParams_ represents the [list](#user-content-Object_Fields) of [interface method parameters](#user-content-Object_InterfaceMethodParam).

<h3><a id="user-content-Object_InterfaceMethodResult" target="_self">InterfaceMethodResult</a></h3>

_InterfaceMethodResult_ represents an interface method result.

<h6><a id="user-content-Object_InterfaceMethodResult-Method" target="_self">.Method</a></h6>

_.Method_ represents the interface method that contains the result.

<h6><a id="user-content-Object_InterfaceMethodResult-Type" target="_self">.Type</a></h6>

_.Type_ represents the underlying type of the result.

<h3><a id="user-content-Object_InterfaceMethods" target="_self">InterfaceMethods</a></h3>

_InterfaceMethods_ represents the [list](#user-content-Object_Fields) of [interface methods](#user-content-Object_InterfaceMethod).

<h3><a id="user-content-Object_Interfaces" target="_self">Interfaces</a></h3>

_Interfaces_ represents a [list](#user-content-Object_List) of interface declarations.

<h3><a id="user-content-Object_List" target="_self">List</a></h3>

_List_ represents a list of objects.

<h6><a id="user-content-Object_List-List" target="_self">.List</a></h6>

_.List_ represents the list of objects. It is used to provide a uniform way to access.

<h3><a id="user-content-Object_Node" target="_self">Node</a></h3>

_Node_ represents a Node in the AST. It's a special object that can be annotated with a documentation comment.

<h6><a id="user-content-Object_Node-Annotations" target="_self">.Annotations</a></h6>

_.Annotations_ represents the annotations for the node.

<h6><a id="user-content-Object_Node-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the documentation comment for the node.

<h6><a id="user-content-Object_Node-File" target="_self">.File</a></h6>

_.File_ represents the file containing the node.

<h6><a id="user-content-Object_Node-Package" target="_self">.Package</a></h6>

_.Package_ represents the package containing the node.
It's a shortcut for Node.File().Package().

<h3><a id="user-content-Object_NodeName" target="_self">NodeName</a></h3>

_NodeName_ represents a name of a node in a declaration:
- Const name
- Enum member name
- Struct field name
- Interface method name
- Interface method parameter name

<h6><a id="user-content-Object_NodeName-Node" target="_self">.Node</a></h6>

_.Node_ represents the [node](#user-content-Object_Node) that contains the name.

<h6><a id="user-content-Object_NodeName-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the node name.

<h3><a id="user-content-Object_Object" target="_self">Object</a></h3>

_Object_ represents an object in Next which can be rendered in a template like this: {{next Object}}

<h3><a id="user-content-Object_Package" target="_self">Package</a></h3>

_Package_ represents a Next package.

<h6><a id="user-content-Object_Package-Annotations" target="_self">.Annotations</a></h6>

_.Annotations_ represents the package [annotations](#user-content-Object_Annotations).

<h6><a id="user-content-Object_Package-Doc" target="_self">.Doc</a></h6>

_.Doc_ represents the package [documentation](#user-content-Object_Doc).

<h6><a id="user-content-Object_Package-Files" target="_self">.Files</a></h6>

_.Files_ represents the all declared files in the package.

<h6><a id="user-content-Object_Package-In" target="_self">.In</a></h6>

_.In_ reports whether the package is the same as the given package.
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

<h6><a id="user-content-Object_Package-Name" target="_self">.Name</a></h6>

_.Name_ represents the package name string.

<h6><a id="user-content-Object_Package-Types" target="_self">.Types</a></h6>

_.Types_ represents the all declared types in the package.

<h3><a id="user-content-Object_Struct" target="_self">Struct</a></h3>

_Struct_ represents a struct declaration.

<h6><a id="user-content-Object_Struct-Fields" target="_self">.Fields</a></h6>

_.Fields_ represents the list of struct fields.

<h3><a id="user-content-Object_StructField" target="_self">StructField</a></h3>

_StructField_ represents a struct field declaration.

<h6><a id="user-content-Object_StructField-Name" target="_self">.Name</a></h6>

_.Name_ represents the [name object](#user-content-Object_NodeName) of the struct field.

<h3><a id="user-content-Object_StructFieldName" target="_self">StructFieldName</a></h3>

_StructFieldName_ represents the [name object](#user-content-Object_NodeName) of a [struct field](#user-content-Object_StructField).

<h3><a id="user-content-Object_StructFieldType" target="_self">StructFieldType</a></h3>

_StructFieldType_ represents a struct field type.

<h3><a id="user-content-Object_StructFields" target="_self">StructFields</a></h3>

_StructFields_ represents the [list](#user-content-Object_Fields) of [struct fields](#user-content-Object_StructField).

<h3><a id="user-content-Object_Structs" target="_self">Structs</a></h3>

_Structs_ represents a [list](#user-content-Object_List) of struct declarations.

<h3><a id="user-content-Object_Symbol" target="_self">Symbol</a></h3>

_Symbol_ represents a Next symbol: value(const, enum member), type(enum, struct, interface).

<h3><a id="user-content-Object_Type" target="_self">Type</a></h3>

_Type_ represents a Next type.

<h4><a id="user-content-Object_Type_ArrayType" target="_self">ArrayType</a></h4>

_ArrayType_ represents an array type.

<h4><a id="user-content-Object_Type_EnumType" target="_self">EnumType</a></h4>

_EnumType_ represents the type of an [enum](#user-content-Object_Enum) declaration.

<h4><a id="user-content-Object_Type_InterfaceType" target="_self">InterfaceType</a></h4>

_InterfaceType_ represents the type of an [interface](#user-content-Object_Interface) declaration.

<h4><a id="user-content-Object_Type_MapType" target="_self">MapType</a></h4>

_MapType_ represents a map type.

<h4><a id="user-content-Object_Type_PrimitiveType" target="_self">PrimitiveType</a></h4>

_PrimitiveType_ represents a primitive type.

<h4><a id="user-content-Object_Type_StructType" target="_self">StructType</a></h4>

_StructType_ represents the type of a [struct](#user-content-Object_Struct) declaration.

<h4><a id="user-content-Object_Type_UsedType" target="_self">UsedType</a></h4>

_UsedType_ represents a used type in a file.

<h4><a id="user-content-Object_Type_VectorType" target="_self">VectorType</a></h4>

_VectorType_ represents a vector type.

<h6><a id="user-content-Object_Type-Decl" target="_self">.Decl</a></h6>

_.Decl_ represents the [declaration](#user-content-Decl) of the type.

<h6><a id="user-content-Object_Type-Kind" target="_self">.Kind</a></h6>

_.Kind_ returns the kind of the type.

<h6><a id="user-content-Object_Type-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the type.

<h6><a id="user-content-Object_UsedType-File" target="_self">.File</a></h6>

_.File_ represents the file containing the used type.

<h6><a id="user-content-Object_UsedType-Type" target="_self">.Type</a></h6>

_.Type_ represents the used type.

<h3><a id="user-content-Object_Value" target="_self">Value</a></h3>

_Value_ represents a constant value for a const declaration or an enum member.

<h6><a id="user-content-Object_Value-Any" target="_self">.Any</a></h6>

_.Any_ represents the underlying value of the constant.

<h6><a id="user-content-Object_Value-IsEnum" target="_self">.IsEnum</a></h6>

_.IsEnum_ returns true if the value is an enum member.

<h6><a id="user-content-Object_Value-IsFirst" target="_self">.IsFirst</a></h6>

_.IsFirst_ reports whether the value is the first member of the enum type.

<h6><a id="user-content-Object_Value-IsLast" target="_self">.IsLast</a></h6>

_.IsLast_ reports whether the value is the last member of the enum type.

<h6><a id="user-content-Object_Value-String" target="_self">.String</a></h6>

_.String_ represents the string representation of the value.

<h6><a id="user-content-Object_Value-Type" target="_self">.Type</a></h6>

_.Type_ represents the [primitive type](#user-content-Object_Type_PrimitiveType) of the value.

<h6><a id="user-content-Object_import-Comment" target="_self">.Comment</a></h6>

_.Comment_ represents the import declaration line [comment](#user-content-Object_Comment).

<h6><a id="user-content-Object_import-Path" target="_self">.Path</a></h6>

_.Path_ represents the import path.

