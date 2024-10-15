package compile

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mkideal/next/src/ast"
	"github.com/mkideal/next/src/token"
)

// @api(Object/Common) contains some general types, including a generic type. Unless specifically stated,
// these objects cannot be directly called using the `Context/next` function.
// The [Value](#Object/Value) object represents a value, which can be either a constant
// value or an enum member's value. The object type for the former is `const.value`, and for
// the latter is `enum.member.value`.

// -------------------------------------------------------------------------

// @api(Object)
//
//	import Tabs from "@theme/Tabs";
//	import TabItem from "@theme/TabItem";
//
// `Object` is a generic object type. These objects can be used as parameters for the `Context/next`
// function, like `{{next .}}`.
type Object interface {
	// @api(Object.Typeof) returns the type name of the object.
	// The type name is a string that represents the type of the object.
	// Except for objects under [Common](#Object/Common), the type names of other objects are lowercase names
	// separated by dots. For example, the type name of a `EnumMember` object is `enum.member`, and the type name
	// of a `Enum` object is `enum`. These objects can be customized for code generation by defining templates.
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	enum Color {
	//		Red = 1;
	//		Green = 2;
	//		Blue = 3;
	//	}
	//	```
	//
	//	```npl
	//	{{- define "go/enum.member" -}}
	//	const {{render "enum.member:name" .Name}} = {{next .Value}}
	//	{{- end}}
	//
	//	{{- define "go/enum.member:name" -}}
	//	{{.Decl.Name}}_{{.}}
	//	{{- end}}
	//	```
	//
	// Output:
	//
	//	```go
	//	package demo
	//
	//	type Color int
	//
	//	const Color_Red = 1
	//	const Color_Green = 2
	//	const Color_Blue = 3
	//	```
	//
	// These two definitions will override the built-in template functions `next/go/enum.member` and `next/go/enum.member:name`.
	Typeof() string
}

// All objects listed here implement the Object interface.
var _ Object = (Consts)(nil)
var _ Object = (Enums)(nil)
var _ Object = (Structs)(nil)
var _ Object = (Interfaces)(nil)
var _ Object = (*EnumMembers)(nil)
var _ Object = (*StructFields)(nil)
var _ Object = (*InterfaceMethods)(nil)
var _ Object = (*InterfaceMethodParams)(nil)
var _ Object = (*File)(nil)
var _ Object = (*Doc)(nil)
var _ Object = (*Comment)(nil)
var _ Object = (*Imports[Decl])(nil)
var _ Object = (*Import)(nil)
var _ Object = (*Decls[Decl])(nil)
var _ Object = (*Value)(nil)
var _ Object = (*UsedType)(nil)
var _ Object = (*PrimitiveType)(nil)
var _ Object = (*ArrayType)(nil)
var _ Object = (*VectorType)(nil)
var _ Object = (*MapType)(nil)
var _ Object = (*EnumType)(nil)
var _ Object = (*StructType)(nil)
var _ Object = (*InterfaceType)(nil)
var _ Object = (*Const)(nil)
var _ Object = (*Enum)(nil)
var _ Object = (*EnumMember)(nil)
var _ Object = (*Struct)(nil)
var _ Object = (*StructField)(nil)
var _ Object = (*Interface)(nil)
var _ Object = (*InterfaceMethod)(nil)
var _ Object = (*InterfaceMethodParameter)(nil)
var _ Object = (*InterfaceMethodResult)(nil)
var _ Object = (*CallStmt)(nil)

// Generic objects

// list objects: consts, enums, structs, interfaces
func (x List[T]) Typeof() string {
	var zero T
	return zero.Typeof() + "s"
}

// Fields objects: enum.members, struct.fields, interface.methods
func (x *Fields[D, F]) Typeof() string {
	var zero F
	return zero.Typeof() + "s"
}

func (*Package) Typeof() string    { return "package" }
func (*File) Typeof() string       { return "file" }
func (*Doc) Typeof() string        { return "doc" }
func (*Comment) Typeof() string    { return "comment" }
func (*Imports[T]) Typeof() string { return "imports" }
func (*Import) Typeof() string     { return "import" }
func (*Decls[T]) Typeof() string   { return "decls" }
func (x *Value) Typeof() string {
	if x.enum.decl == nil {
		return "const.value"
	}
	return "enum.member.value"
}

// Type objects

func (*UsedType) Typeof() string        { return "used.type" }
func (x *PrimitiveType) Typeof() string { return "primitive.type" }
func (*ArrayType) Typeof() string       { return "array.type" }
func (*VectorType) Typeof() string      { return "vector.type" }
func (*MapType) Typeof() string         { return "map.type" }
func (x *DeclType[T]) Typeof() string   { return x.decl.Typeof() + ".type" }

// Decl objects

func (*Const) Typeof() string                    { return "const" }
func (*Enum) Typeof() string                     { return "enum" }
func (*EnumMember) Typeof() string               { return "enum.member" }
func (*Struct) Typeof() string                   { return "struct" }
func (*StructField) Typeof() string              { return "struct.field" }
func (*Interface) Typeof() string                { return "interface" }
func (*InterfaceMethod) Typeof() string          { return "interface.method" }
func (*InterfaceMethodParameter) Typeof() string { return "interface.method.parameter" }
func (*InterfaceMethodResult) Typeof() string    { return "interface.method.result" }

// Stmt objects

func (*CallStmt) Typeof() string { return "stmt.call" }

// -------------------------------------------------------------------------

// @api(Object/Common/Position) represents the position of an [LocatedObject](#Object/Common/LocatedObject) in a file.
type Position struct {
	pos token.Pos

	// @api(Object/Common/Position.Filename) represents the filename of the position.
	Filename string

	// @api(Object/Common/Position.Line) represents the line number of the position starting from 1.
	Line int

	// @api(Object/Common/Position.Column) represents the column number of the position starting from 1.
	Column int
}

// @api(Object/Common/Position.IsValid) reports whether the position is valid.
func (p Position) IsValid() bool { return p.pos.IsValid() }

// @api(Object/Common/Position) returns the string representation of the position, e.g., `demo.next:10:2`.
func (p Position) String() string {
	s := p.Filename
	if p.IsValid() {
		if s != "" {
			s += ":"
		}
		s += strconv.Itoa(p.Line)
		if p.Column != 0 {
			s += fmt.Sprintf(":%d", p.Column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

func positionFor(c *Compiler, pos token.Pos) Position {
	if pos.IsValid() {
		p := c.fset.Position(pos)
		return Position{
			pos:      pos,
			Filename: p.Filename,
			Line:     p.Line,
			Column:   p.Column,
		}
	}
	return Position{}
}

// @api(Object/Common/LocatedObject) represents an [Object](#Object) with a location in a file.
type LocatedObject interface {
	Object

	// @api(Object/Common/LocatedObject.Pos) represents the [Position](#Object/Common/Position) of the object.
	//
	// Example:
	//
	//	```next title="demo.next"
	//	package demo;
	//	const Name = "hei hei";
	//	```
	//
	//	```npl
	//	{{- define "meta/this" -}}const{{- end -}}
	//	{{this.Pos}}
	//	{{this.Value.Pos}}
	//	```
	//
	// Output:
	//
	//	```
	//	demo.next:2:1
	//	demo.next:2:14
	//	```
	Pos() Position

	// @api(Object/Common/LocatedObject.File) represents the file containing the object.
	File() *File

	// @api(Object/Common/LocatedObject.Package) represents the package containing the object.
	Package() *Package
}

func (x *File) Pos() Position             { return positionFor(x.compiler, x.pos) }
func (x *commonNode[Self]) Pos() Position { return positionFor(x.file.compiler, x.pos) }
func (x *Value) Pos() Position            { return positionFor(x.file.compiler, x.pos) }
func (x *DeclType[T]) Pos() Position      { return positionFor(x.file.compiler, x.pos) }

func (x *commonNode[Self]) Package() *Package { return x.File().Package() }
func (x *Value) Package() *Package            { return x.File().Package() }
func (x *DeclType[T]) Package() *Package      { return x.File().Package() }

// -------------------------------------------------------------------------

// @api(Object/Common/Node) represents a [LocatedObject](#Object/Common/LocatedObject) that is a node in a file.
//
// Currently, the following nodes are supported:
//
// - [File](#Object/File)
// - [Const](#Object/Const)
// - [Enum](#Object/Enum)
// - [Struct](#Object/Struct)
// - [Interface](#Object/Interface)
// - [EnumMember](#Object/EnumMember)
// - [StructField](#Object/StructField)
// - [InterfaceMethod](#Object/InterfaceMethod)
// - [InterfaceMethodParameter](#Object/InterfaceMethodParameter)
type Node interface {
	LocatedObject

	// @api(Object/Common/Node.Name) represents the name of the node.
	//
	// Example:
	//
	//	```next
	//	const x = 1;
	//	```
	//
	//	```npl
	//	{{.Name}}
	//	```
	//
	// Output:
	//
	//	```
	//	x
	//	```
	Name() string

	// @api(Object/Common/Node.NamePos) represents the position of the node name.
	//
	// Example:
	//
	//	```next title="demo.next"
	//	package demo;
	//	const x = 1;
	//	```
	//
	//	```npl
	//	{{.NamePos}}
	//	```
	//
	// Output:
	//
	//	```
	//	demo.next:2:7
	//	```
	NamePos() Position

	// @api(Object/Common/Node.Doc) represents the documentation comment for the node.
	// The documentation comment is a comment that appears before the node declaration.
	//
	// Example:
	//
	//	```next
	//	// This is a documentation comment for the node.
	//	// It can be multiple lines.
	//	const x = 1;
	//	```
	//
	//	```npl
	//	{{.Doc.Text}}
	//	```
	//
	// Output:
	//
	//	```
	//	This is a documentation comment for the node.
	//	It can be multiple lines.
	//	```
	Doc() *Doc

	// @api(Object/Common/Node.Annotations) represents the [Annotations](#Object/Common/Annotations) for the node.
	//
	// Example:
	//
	//	```next title="demo.next"
	//	package demo;
	//	@next(type=int8)
	//	@custom
	//	enum Color {
	//		Red = 1;
	//		Green = 2;
	//		Blue = 3;
	//	}
	//	```
	//
	//	```npl
	//	{{.Annotations.next.type}}
	//	{{.Annotations.next.Pos}}
	//	{{.Annotations.Has "custom"}}
	//	```
	//
	// Output:
	//
	//	```
	//	int8
	//	demo.next:2:1
	//	true
	//	```
	Annotations() Annotations
}

// All nodes listed here implement the Node interface.
var _ Node = (*File)(nil)
var _ Node = (*Const)(nil)
var _ Node = (*Enum)(nil)
var _ Node = (*Struct)(nil)
var _ Node = (*Interface)(nil)
var _ Node = (*EnumMember)(nil)
var _ Node = (*StructField)(nil)
var _ Node = (*InterfaceMethod)(nil)
var _ Node = (*InterfaceMethodParameter)(nil)

func (x *commonNode[Self]) Name() string { return x.name }

func (x *File) File() *File { return x }
func (x *commonNode[Self]) File() *File {
	if x == nil {
		return nil
	}
	return x.file
}

func (x *Package) NamePos() Position          { return positionFor(x.decls.compiler, x.File().namePos) }
func (x *File) NamePos() Position             { return positionFor(x.compiler, x.namePos) }
func (x *commonNode[Self]) NamePos() Position { return positionFor(x.file.compiler, x.namePos) }

// -------------------------------------------------------------------------

// @api(Object/Common/Decl) represents a top-level declaration [Node](#Object/Common/Node) in a file.
//
// Currently, the following declarations are supported:
//
// - [Package](#Object/Package)
// - [File](#Object/File)
// - [Const](#Object/Const)
// - [Enum](#Object/Enum)
// - [Struct](#Object/Struct)
// - [Interface](#Object/Interface)
type Decl interface {
	Node

	// @api(Object/Common/Decl.UsedKinds) returns the used kinds in the declaration.
	// Returns 0 if the declaration does not use any kinds.
	// Otherwise, returns the OR of all used kinds.
	//
	// Example:
	//
	//	```next
	//	struct User {
	//		int64 id;
	//		string name;
	//		vector<string> emails;
	//		map<int, bool> flags;
	//	}
	//	```
	// The used kinds in the `User` struct are: `(1<<KindInt64) | (1<<KindString) | (1<<KindVector) | (1<<KindMap) | (1<<KindInt) | (1<<KindBool)`.
	UsedKinds() Kinds

	declNode()
}

// All decls listed here implement the Decl interface.
var _ Decl = (*Package)(nil)
var _ Decl = (*File)(nil)
var _ Decl = (*Const)(nil)
var _ Decl = (*Enum)(nil)
var _ Decl = (*Struct)(nil)
var _ Decl = (*Interface)(nil)

func (x *Package) UsedKinds() Kinds {
	var kinds Kinds
	for _, f := range x.files {
		kinds |= f.UsedKinds()
	}
	return kinds
}

func (x *File) UsedKinds() Kinds {
	var kinds Kinds
	for _, d := range x.decls.consts {
		kinds |= d.UsedKinds()
	}
	for _, d := range x.decls.enums {
		kinds |= d.UsedKinds()
	}
	for _, d := range x.decls.structs {
		kinds |= d.UsedKinds()
	}
	for _, d := range x.decls.interfaces {
		kinds |= d.UsedKinds()
	}
	return kinds
}

func (x *Const) UsedKinds() Kinds { return x.Type().UsedKinds() }

func (x *Enum) UsedKinds() Kinds { return x.Type.UsedKinds() | x.MemberType.UsedKinds() }

func (x *Struct) UsedKinds() Kinds {
	var kinds = x.Type.UsedKinds()
	for _, f := range x.fields.List {
		kinds |= f.Type.UsedKinds()
	}
	return kinds
}
func (x *Interface) UsedKinds() Kinds {
	var kinds = x.Type.UsedKinds()
	for _, m := range x.methods.List {
		for _, p := range m.Params.List {
			kinds |= p.Type.UsedKinds()
		}
		if m.Result.Type != nil {
			kinds |= m.Result.Type.UsedKinds()
		}
	}
	return kinds
}

func (x *Package) declNode()   {}
func (x *File) declNode()      {}
func (x *Const) declNode()     {}
func (x *Enum) declNode()      {}
func (x *Struct) declNode()    {}
func (x *Interface) declNode() {}

// builtinDecl represents a special declaration for built-in types.
type builtinDecl struct {
	typ Type
}

var _ Decl = builtinDecl{}

func (builtinDecl) Typeof() string           { return "<builtin.decl>" }
func (builtinDecl) Name() string             { return "<builtin>" }
func (builtinDecl) NamePos() Position        { return Position{} }
func (builtinDecl) Pos() Position            { return Position{} }
func (builtinDecl) File() *File              { return nil }
func (builtinDecl) Package() *Package        { return nil }
func (builtinDecl) Doc() *Doc                { return nil }
func (builtinDecl) Annotations() Annotations { return nil }
func (builtinDecl) declNode()                {}
func (x builtinDecl) UsedKinds() Kinds       { return x.typ.UsedKinds() }

// -------------------------------------------------------------------------
// Types

// @api(Object/Common/Type) represents a Next type [Object](#Object).
//
// Currently, the following types are supported:
//
// - [UsedType](#Object/UsedType)
// - [PrimitiveType](#Object/PrimitiveType)
// - [ArrayType](#Object/ArrayType)
// - [VectorType](#Object/VectorType)
// - [MapType](#Object/MapType)
// - [EnumType](#Object/EnumType)
// - [StructType](#Object/StructType)
// - [InterfaceType](#Object/InterfaceType)
type Type interface {
	Object

	// @api(Object/Common/Type.Kind) returns the [Kind](#Object/Common/Type/Kind) of the type.
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	enum Color {
	//		Red = 1;
	//		Green = 2;
	//		Blue = 3;
	//	}
	//	```
	//
	//	```npl
	//	{{- define "cpp/enum" -}}
	//	{{.Type.Kind}}
	//	{{.MemberType.Kind}}
	//	{{end}}
	//	```
	//
	// Output:
	//
	//	```
	//	Enum
	//	Int32
	//	```
	Kind() Kind

	// @api(Object/Common/Type.UsedKinds) returns the used kinds in the type.
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	struct User {
	//		int64 id;
	//		string name;
	//		vector<string> emails;
	//		map<int, bool> flags;
	//	}
	//	```
	// The used kinds in the `User` struct are: `(1<<KindStruct) | (1<<KindInt64) | (1<<KindString) | (1<<KindVector) | (1<<KindMap) | (1<<KindInt) | (1<<KindBool)`.
	//
	//	```npl
	//	{{.UsedKinds.Has "struct"}}
	//	{{.UsedKinds.Has "int64"}}
	//	{{.UsedKinds.Has "string"}}
	//	{{.UsedKinds.Has "vector"}}
	//	{{.UsedKinds.Has "map"}}
	//	{{.UsedKinds.Has "int"}}
	//	{{.UsedKinds.Has "bool"}}
	//	{{.UsedKinds.Has "float32"}}
	//	```
	//
	// Output:
	//
	//	```
	//	true
	//	true
	//	true
	//	true
	//	true
	//	true
	//	true
	//	false
	//	```
	UsedKinds() Kinds

	// @api(Object/Common/Type.String) represents the string representation of the type.
	String() string

	// @api(Object/Common/Type.Decl) represents the [Decl](#Object/Common/Decl) of the type.
	// If the type is a built-in type, it returns a special declaration. Otherwise, it returns
	// the declaration of the type: [Enum](#Object/Enum), [Struct](#Object/Struct), or [Interface](#Object/Interface).
	Decl() Decl

	// @api(Object/Common/Type.Actual) represents the actual type.
	// You need to use the `Actual` method to get the actual type to access the specific fields of the type.
	//
	// For example, if you have a type `ArrayType`:
	//
	//	```npl
	//	{{.Type.Actual.ElemType}} {{/* Good */}}
	//	{{.Type.Actual.N}} {{/* Good */}}
	//
	//	// This will error
	//	{{.Type.ElemType}} {{/* Error */}}
	//	// This will error
	//	{{.Type.N}} {{/* Error */}}
	//	```
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	struct User {
	//		int64 id;
	//		string name;
	//		array<int, 3> codes;
	//	}
	//	```
	//
	//	```npl
	//	{{- define "c/struct.field" -}}
	//		{{- if .Type.Kind.IsArray -}}
	//			{{next .Type.Actual.ElemType}} {{.Name}}[{{.Type.Actual.N}}];
	//		{{- else -}}
	//			{{next .Type}} {{.Name}};
	//		{{- end}}
	//	{{- end}}
	//	```
	//
	// Output:
	//
	//	```c
	//	typedef struct User {
	//		int64_t id;
	//		char* name;
	//		int codes[3];
	//	} User;
	//	```
	//
	// :::tip
	//
	// In the example above, the `codes` field is an single-dimensional array of integers with a length of 3.
	// If we want to process the multi-dimensional array, we need to fix the template to recursively process the array type.
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	struct User {
	//		int64 id;
	//		string name;
	//		array<array<int, 3>, 2> codes;
	//	}
	//	```
	//
	//	```npl
	//	{{- define "c/struct.field" -}}
	//	{{next .Doc}}{{render "dict:struct.field.decl" (dict "type" .Type "name" (render "struct.field:name" .))}};{{next .Comment}}
	//	{{- end}}
	//
	//	{{- define "c/dict:struct.field.decl" -}}
	//	{{- $type := .type -}}
	//	{{- $name := .name -}}
	//	{{- if $type.Kind.IsArray -}}
	//	{{render "dict:struct.field.decl" (dict "type" $type.Actual.ElemType "name" (printf "%s[%d]" $name $type.Actual.N))}}
	//	{{- else -}}
	//	{{next $type}} {{$name}}
	//	{{- end}}
	//	{{- end}}
	//	```
	//
	// Output:
	//
	//	```c
	//	typedef struct User {
	//		int64_t id;
	//		char* name;
	//		int codes[2][3];
	//	} User;
	//	```
	//
	// :::
	Actual() reflect.Value
}

// All types listed here implement the Type interface.
var _ Type = (*UsedType)(nil)
var _ Type = (*PrimitiveType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*VectorType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*DeclType[Decl])(nil)

func (x *UsedType) Decl() Decl      { return x.Type.Decl() }
func (x *PrimitiveType) Decl() Decl { return builtinDecl{x} }
func (x *ArrayType) Decl() Decl     { return builtinDecl{x} }
func (x *VectorType) Decl() Decl    { return builtinDecl{x} }
func (x *MapType) Decl() Decl       { return builtinDecl{x} }
func (x *DeclType[T]) Decl() Decl   { return x.decl }

func (x *UsedType) Kind() Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() Kind { return x.kind }
func (*ArrayType) Kind() Kind       { return KindArray }
func (*VectorType) Kind() Kind      { return KindVector }
func (*MapType) Kind() Kind         { return KindMap }
func (x *DeclType[T]) Kind() Kind   { return x.kind }

func (x *UsedType) UsedKinds() Kinds      { return x.Type.UsedKinds() }
func (x *PrimitiveType) UsedKinds() Kinds { return x.kind.kinds() }
func (x *DeclType[T]) UsedKinds() Kinds   { return x.kind.kinds() }
func (x *ArrayType) UsedKinds() Kinds     { return x.Kind().kinds() | x.ElemType.UsedKinds() }
func (x *VectorType) UsedKinds() Kinds    { return x.Kind().kinds() | x.ElemType.UsedKinds() }
func (x *MapType) UsedKinds() Kinds {
	return x.Kind().kinds() | x.KeyType.UsedKinds() | x.ElemType.UsedKinds()
}

// @api(Object/UsedType) represents a used [Type](#Object/Common/Type) in a file.
type UsedType struct {
	depth int
	src   ast.Type
	node  Node

	// @api(Object/UsedType.Type) represents the underlying [Type](#Object/Common/Type).
	//
	// Example:
	//
	//	```next
	//	package demo;
	//
	//	struct User {/*...*/}
	//
	//	struct Group {
	//		// highlight-next-line
	//		User user;
	//	}
	//	```
	//
	// The type of the `user` field is a `UsedType` with the underlying type `StructType` of the `User` struct.
	Type Type
}

// Use uses a type in a file.
func Use(depth int, src ast.Type, node Node, t Type) *UsedType {
	return &UsedType{depth: depth, src: src, node: node, Type: t}
}

// @api(Object/UsedType.File) represents the [File](#Object/File) where the type is used.
func (u *UsedType) File() *File {
	return u.node.File()
}

// @api(Object/UsedType.Node) represents the node where the type is used.
//
// The node may be:
//
// - [StructField](#Object/StructField): for a struct field type
// - [InterfaceMethod](#Object/InterfaceMethod): for a method result type
// - [InterfaceMethodParameter](#Object/InterfaceMethodParameter): for a method parameter type
func (u *UsedType) Node() reflect.Value {
	return reflect.ValueOf(u.node)
}

func (u *UsedType) String() string { return u.Type.String() }

func (u *UsedType) Actual() reflect.Value { return u.Type.Actual() }

// UsedTypeNode returns the AST node of the used type.
func UsedTypeNode(u *UsedType) ast.Type { return u.src }

// @api(Object/PrimitiveType) represents a primitive type.
//
// Currently, the following primitive types are supported:
//
// - **int**
// - **int8**
// - **int16**
// - **int32**
// - **int64**
// - **float32**
// - **float64**
// - **bool**
// - **string**
// - **byte**
// - **bytes**
// - **any**
type PrimitiveType struct {
	name string
	kind Kind
}

func (b *PrimitiveType) String() string { return b.name }

func (b *PrimitiveType) Actual() reflect.Value { return reflect.ValueOf(b) }

var primitiveTypes = func() map[string]*PrimitiveType {
	m := make(map[string]*PrimitiveType)
	for _, kind := range primitiveKinds {
		name := kind.String()
		m[name] = &PrimitiveType{kind: kind, name: name}
	}
	return m
}()

// @api(Object/ArrayType) represents an fixed-size array [Type](#Object/Common/Type).
type ArrayType struct {
	pos     token.Pos
	lenExpr ast.Expr

	// @api(Object/ArrayType.ElemType) represents the element [Type](#Object/Common/Type) of the array.
	ElemType Type

	// @api(Object/ArrayType.N) represents the number of elements in the array.
	N int64
}

func LenExpr(a *ArrayType) ast.Expr { return a.lenExpr }

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatInt(a.N, 10) + ">"
}

func (a *ArrayType) Actual() reflect.Value { return reflect.ValueOf(a) }

// @api(Object/VectorType) represents a vector [Type](#Object/Common/Type).
type VectorType struct {
	pos token.Pos

	// @api(Object/VectorType.ElemType) represents the element [Type](#Object/Common/Type) of the vector.
	ElemType Type
}

func (v *VectorType) String() string {
	return "vector<" + v.ElemType.String() + ">"
}

func (v *VectorType) Actual() reflect.Value { return reflect.ValueOf(v) }

// @api(Object/MapType) represents a map [Type](#Object/Common/Type).
type MapType struct {
	pos token.Pos

	// @api(Object/MapType.KeyType) represents the key [Type](#Object/Common/Type) of the map.
	KeyType Type

	// @api(Object/MapType.ElemType) represents the element [Type](#Object/Common/Type) of the map.
	ElemType Type
}

func (m *MapType) String() string {
	return "map<" + m.KeyType.String() + "," + m.ElemType.String() + ">"
}

func (m *MapType) Actual() reflect.Value { return reflect.ValueOf(m) }

// DeclType represents a declaration type which is a type of a declaration: enum, struct, interface.
type DeclType[T Decl] struct {
	pos  token.Pos
	kind Kind
	name string
	file *File
	decl T
}

// @api(Object/EnumType) represents the [Type](#Object/Common/Type) of an [Enum](#Object/Enum) declaration.
type EnumType = DeclType[*Enum]

// @api(Object/StructType) represents the [Type](#Object/Common/Type) of a [Struct](#Object/Struct) declaration.
type StructType = DeclType[*Struct]

// @api(Object/InterfaceType) represents the [Type](#Object/Common/Type) of an [Interface](#Object/Interface) declaration.
type InterfaceType = DeclType[*Interface]

func newDeclType[T Decl](file *File, pos token.Pos, kind Kind, name string, decl T) *DeclType[T] {
	return &DeclType[T]{file: file, pos: pos, kind: kind, name: name, decl: decl}
}

func (d *DeclType[T]) String() string { return d.name }

func (d *DeclType[T]) Actual() reflect.Value { return reflect.ValueOf(d) }

func (d *DeclType[T]) File() *File { return d.file }
