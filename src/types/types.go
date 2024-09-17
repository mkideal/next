//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.1 ./ ../../docs/en/api.md
package types

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// @api(Object/Common) contains some general types, including a generic type. Unless specifically stated,
// these objects cannot be directly called using the [next](#Context/next) function.
// The [Value](#Object/Common/Value) object represents a value, which can be either a constant
// value or an enum member's value. The object type for the former is `const.value`, and for
// the latter is `enum.member.value`.

// -------------------------------------------------------------------------

// @api(Object) is a generic object type. These objects can be used as parameters for the [next](#Context/next)
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
	// ```next
	//	package demo;
	//
	//	enum Color {
	//		Red = 1;
	//		Green = 2;
	//		Blue = 3;
	//	}
	// ```
	//
	// ```npl
	// {{- define "go/enum.member" -}}
	// const {{next .Name}} = {{.Value}}
	// {{- end}}
	//
	// {{- define "go/enum.member:name" -}}
	// {{.Decl.Name}}_{{.}}
	// {{- end}}
	// ```
	//
	// Output:
	// ```go
	// package demo
	//
	// type Color int
	//
	// const Color_Red = 1
	// const Color_Green = 2
	// const Color_Blue = 3
	// ```
	//
	// These two definitions will override the built-in template functions `next/go/enum.member` and `next/go/enum.member.name`.
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
var _ Object = (*Imports)(nil)
var _ Object = (*Import)(nil)
var _ Object = (*Decls)(nil)
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
var _ Object = (*InterfaceMethodParam)(nil)
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

func (*Package) Typeof() string { return "package" }
func (*File) Typeof() string    { return "file" }
func (*Doc) Typeof() string     { return "doc" }
func (*Comment) Typeof() string { return "comment" }
func (*Imports) Typeof() string { return "imports" }
func (*Import) Typeof() string  { return "import" }
func (*Decls) Typeof() string   { return "decls" }
func (x *Value) Typeof() string {
	if x.enum.typ == nil {
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

func (*Const) Typeof() string                 { return "const" }
func (*Enum) Typeof() string                  { return "enum" }
func (*EnumMember) Typeof() string            { return "enum.member" }
func (*Struct) Typeof() string                { return "struct" }
func (*StructField) Typeof() string           { return "struct.field" }
func (*Interface) Typeof() string             { return "interface" }
func (*InterfaceMethod) Typeof() string       { return "interface.method" }
func (*InterfaceMethodParam) Typeof() string  { return "interface.method.param" }
func (*InterfaceMethodResult) Typeof() string { return "interface.method.result" }

// Stmt objects

func (*CallStmt) Typeof() string { return "stmt.call" }

// -------------------------------------------------------------------------

// LocatedObject represents an object with a position.
type LocatedObject interface {
	Object

	// Pos returns the position of the object.
	Pos() token.Pos

	// @api(Object/Common/Node.File) represents the file containing the node.
	File() *File

	// @api(Object/Common/Node.Package) represents the package containing the node.
	Package() *Package
}

func (x *File) Pos() token.Pos             { return x.pos }
func (x *commonNode[Self]) Pos() token.Pos { return x.pos }
func (x *Value) Pos() token.Pos            { return x.namePos }
func (x *DeclType[T]) Pos() token.Pos      { return x.pos }

func (x *commonNode[Self]) Package() *Package { return x.File().Package() }
func (x *Value) Package() *Package            { return x.File().Package() }
func (x *DeclType[T]) Package() *Package      { return x.File().Package() }

// -------------------------------------------------------------------------

// @api(Object/Common/Node) represents a Node in the Next AST.
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
// - [InterfaceMethodParam](#Object/InterfaceMethodParam)
type Node interface {
	LocatedObject

	// Name returns the name of the node.
	Name() string

	// @api(Object/Common/Node.Doc) represents the documentation comment for the node.
	Doc() *Doc

	// @api(Object/Common/Node.Annotations) represents the [annotations](#Annotation/Annotations) for the node.
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
var _ Node = (*InterfaceMethodParam)(nil)

func (x *commonNode[Self]) Name() string { return x.name }

func (x *File) File() *File { return x }
func (x *commonNode[Self]) File() *File {
	if x == nil {
		return nil
	}
	return x.file
}

// -------------------------------------------------------------------------

// @api(Object/Common/Decl) represents a top-level declaration in a file.
//
// All declarations are [nodes](#Object/Common/Node). Currently, the following declarations are supported:
//
// - [Package](#Object/Package)
// - [File](#Object/File)
// - [Const](#Object/Const)
// - [Enum](#Object/Enum)
// - [Struct](#Object/Struct)
// - [Interface](#Object/Interface)
type Decl interface {
	Node

	declNode()
}

// All decls listed here implement the Decl interface.
var _ Decl = (*Package)(nil)
var _ Decl = (*File)(nil)
var _ Decl = (*Const)(nil)
var _ Decl = (*Enum)(nil)
var _ Decl = (*Struct)(nil)
var _ Decl = (*Interface)(nil)

func (x *Package) declNode()   {}
func (x *File) declNode()      {}
func (x *Const) declNode()     {}
func (x *Enum) declNode()      {}
func (x *Struct) declNode()    {}
func (x *Interface) declNode() {}

// builtinDecl represents a special declaration for built-in types.
type builtinDecl struct{}

var _ Decl = builtinDecl{}

func (builtinDecl) Typeof() string           { return "<builtin.decl>" }
func (builtinDecl) Name() string             { return "<builtin>" }
func (builtinDecl) Pos() token.Pos           { return token.NoPos }
func (builtinDecl) File() *File              { return nil }
func (builtinDecl) Package() *Package        { return nil }
func (builtinDecl) Doc() *Doc                { return nil }
func (builtinDecl) Annotations() Annotations { return nil }
func (builtinDecl) declNode()                {}

// -------------------------------------------------------------------------
// Types

// @api(Object/Common/Type) represents a Next type.
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

	// @api(Object/Common/Type.Kind) returns the [kind](#Object/Common/Type/Kind) of the type.
	Kind() Kind

	// @api(Object/Common/Type.String) represents the string representation of the type.
	String() string

	// @api(Object/Common/Type.Decl) represents the [declaration](#Decl) of the type.
	Decl() Decl

	// @api(Object/Common/Type.Value) returns the reflect value of the type.
	Value() reflect.Value
}

// All types listed here implement the Type interface.
var _ Type = (*UsedType)(nil)
var _ Type = (*PrimitiveType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*VectorType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*DeclType[Decl])(nil)

func (x *UsedType) Decl() Decl      { return x.Type.Decl() }
func (x *PrimitiveType) Decl() Decl { return builtinDecl{} }
func (x *ArrayType) Decl() Decl     { return builtinDecl{} }
func (x *VectorType) Decl() Decl    { return builtinDecl{} }
func (x *MapType) Decl() Decl       { return builtinDecl{} }
func (x *DeclType[T]) Decl() Decl   { return x.decl }

func (x *UsedType) Kind() Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() Kind { return x.kind }
func (*ArrayType) Kind() Kind       { return KindArray }
func (*VectorType) Kind() Kind      { return KindVector }
func (*MapType) Kind() Kind         { return KindMap }
func (x *DeclType[T]) Kind() Kind   { return x.kind }

// @api(Object/UsedType) represents a used type in a file.
type UsedType struct {
	node ast.Type

	// @api(Object/UsedType.File) represents the file where the type is used.
	File *File

	// @api(Object/UsedType.Type) represents the used type.
	Type Type
}

// Use uses a type in a file.
func Use(t Type, f *File, node ast.Type) *UsedType {
	return &UsedType{Type: t, File: f, node: node}
}

func (u *UsedType) String() string { return u.Type.String() }

func (u *UsedType) Value() reflect.Value { return u.Type.Value() }

// UsedTypeNode returns the AST node of the used type.
func UsedTypeNode(u *UsedType) ast.Type { return u.node }

// @api(Object/PrimitiveType) represents a primitive type.
type PrimitiveType struct {
	name string
	kind Kind
}

func (b *PrimitiveType) String() string { return b.name }

func (b *PrimitiveType) Value() reflect.Value { return reflect.ValueOf(b) }

func (b *PrimitiveType) ToPrimitive() *PrimitiveType { return b }

var primitiveTypes = func() map[string]*PrimitiveType {
	m := make(map[string]*PrimitiveType)
	for _, kind := range primitiveKinds {
		name := strings.ToLower(kind.String())
		m[name] = &PrimitiveType{kind: kind, name: name}
	}
	return m
}()

// @api(Object/ArrayType) represents an array [type](#Object/Common/Type).
type ArrayType struct {
	pos token.Pos

	// @api(Object/ArrayType.ElemType) represents the element [type](#Object/Common/Type) of the array.
	ElemType Type

	// @api(Object/ArrayType.N) represents the number of elements in the array.
	N int64
}

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatInt(a.N, 10) + ">"
}

func (a *ArrayType) Value() reflect.Value { return reflect.ValueOf(a) }

// @api(Object/VectorType) represents a vector [type](#Object/Common/Type).
type VectorType struct {
	pos token.Pos

	// @api(Object/VectorType.ElemType) represents the element [type](#Object/Common/Type) of the vector.
	ElemType Type
}

func (v *VectorType) String() string {
	return "vector<" + v.ElemType.String() + ">"
}

func (v *VectorType) Value() reflect.Value { return reflect.ValueOf(v) }

// @api(Object/MapType) represents a map [type](#Object/Common/Type).
type MapType struct {
	pos token.Pos

	// @api(Object/MapType.KeyType) represents the key [type](#Object/Common/Type) of the map.
	KeyType Type

	// @api(Object/MapType.ElemType) represents the element [type](#Object/Common/Type) of the map.
	ElemType Type
}

func (m *MapType) String() string {
	return "map<" + m.KeyType.String() + "," + m.ElemType.String() + ">"
}

func (m *MapType) Value() reflect.Value { return reflect.ValueOf(m) }

// DeclType represents a declaration type which is a type of a declaration: enum, struct, interface.
type DeclType[T Decl] struct {
	pos  token.Pos
	kind Kind
	name string
	file *File
	decl T
}

// @api(Object/EnumType) represents the [type](#Object/Common/Type) of an [enum](#Object/Enum) declaration.
type EnumType = DeclType[*Enum]

// @api(Object/StructType) represents the [type](#Object/Common/Type) of a [struct](#Object/Struct) declaration.
type StructType = DeclType[*Struct]

// @api(Object/InterfaceType) represents the [type](#Object/Common/Type) of an [interface](#Object/Interface) declaration.
type InterfaceType = DeclType[*Interface]

func newDeclType[T Decl](file *File, pos token.Pos, kind Kind, name string, decl T) *DeclType[T] {
	return &DeclType[T]{file: file, pos: pos, kind: kind, name: name, decl: decl}
}

func (d *DeclType[T]) String() string { return d.name }

func (d *DeclType[T]) Value() reflect.Value { return reflect.ValueOf(d) }

func (d *DeclType[T]) File() *File { return d.file }
