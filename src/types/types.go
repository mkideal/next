package types

import (
	"strconv"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// -------------------------------------------------------------------------

// Object represents an object in Next which can be rendered in a template like this: {{next Object}}
// @api(object/Object)
type Object interface {
	// getType returns the type of the object.
	getType() string
}

// All objects listed here implement the Object interface.
var _ Object = (List[Object])(nil)
var _ Object = (*Fields[Node, Object])(nil)
var _ Object = (*FieldName[Node])(nil)
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
var _ Object = (*DeclType[Decl])(nil)
var _ Object = (*Const)(nil)
var _ Object = (*Enum)(nil)
var _ Object = (*EnumMember)(nil)
var _ Object = (*Struct)(nil)
var _ Object = (*StructField)(nil)
var _ Object = (*StructFieldType)(nil)
var _ Object = (*Interface)(nil)
var _ Object = (*InterfaceMethod)(nil)
var _ Object = (*InterfaceMethodParam)(nil)
var _ Object = (*InterfaceMethodParamType)(nil)
var _ Object = (*InterfaceMethodResult)(nil)
var _ Object = (*CallStmt)(nil)

// Generic objects

// list objects: consts, enums, structs, interfaces
func (x List[T]) getType() string {
	var zero T
	return zero.getType() + "s"
}

// Fields objects: enum.members, struct.fields, interface.methods
func (x *Fields[Parent, Field]) getType() string {
	var zero Field
	return zero.getType() + "s"
}

// Name objects: const.name, enum.member.name, struct.field.name, interface.method.name, interface.method.param.name
func (x *FieldName[T]) getType() string {
	return x.field.getType() + ".name"
}

func (*File) getType() string    { return "file" }
func (*Doc) getType() string     { return "doc" }
func (*Comment) getType() string { return "comment" }
func (*Imports) getType() string { return "imports" }
func (*Import) getType() string  { return "import" }
func (*Decls) getType() string   { return "decls" }
func (x *Value) getType() string {
	if x.enum.typ == nil {
		return "const.value"
	}
	return "enum.member.value"
}

// Type objects

func (*UsedType) getType() string        { return "used.type" }
func (x *PrimitiveType) getType() string { return x.name + ".type" }
func (*ArrayType) getType() string       { return "array.type" }
func (*VectorType) getType() string      { return "vector.type" }
func (*MapType) getType() string         { return "map.type" }
func (x *DeclType[T]) getType() string   { return x.decl.getType() + ".type" }

// Decl objects

func (*Const) getType() string                    { return "const" }
func (*Enum) getType() string                     { return "enum" }
func (*EnumMember) getType() string               { return "enum.member" }
func (*Struct) getType() string                   { return "struct" }
func (*StructField) getType() string              { return "struct.field" }
func (*StructFieldType) getType() string          { return "struct.field.type" }
func (*Interface) getType() string                { return "interface" }
func (*InterfaceMethod) getType() string          { return "interface.method" }
func (*InterfaceMethodParam) getType() string     { return "interface.method.param" }
func (*InterfaceMethodParamType) getType() string { return "interface.method.param.type" }
func (*InterfaceMethodResult) getType() string    { return "interface.method.result" }

// Stmt objects

func (*CallStmt) getType() string { return "stmt.call" }

// -------------------------------------------------------------------------

// LocatedObject represents an object with a position.
type LocatedObject interface {
	Object

	// getPos returns the position of the object.
	getPos() token.Pos
}

func (x *File) getPos() token.Pos             { return x.pos }
func (x *commonNode[Self]) getPos() token.Pos { return x.pos }
func (x *Value) getPos() token.Pos            { return x.namePos }
func (x *DeclType[T]) getPos() token.Pos      { return x.pos }

// -------------------------------------------------------------------------

// Node represents a Next node.
type Node interface {
	LocatedObject

	// getName returns the name of the node.
	getName() string

	// getAnnotations returns the annotations for the node.
	getAnnotations() Annotations

	// File returns the file containing the node.
	File() *File

	// Package returns the package containing the node.
	// It's a shortcut for Node.File().Package().
	Package() *Package
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

func (x *File) getName() string             { return x.Name() }
func (x *commonNode[Self]) getName() string { return x.name.name }

func (x *File) getAnnotations() Annotations             { return x.Annotations }
func (x *commonNode[Self]) getAnnotations() Annotations { return x.Annotations }

func (x *File) File() *File { return x }
func (d *commonNode[Self]) File() *File {
	if d == nil {
		return nil
	}
	return d.file
}

func (x *File) Package() *Package {
	if x == nil {
		return nil
	}
	return x.pkg
}

func (x *commonNode[Self]) Package() *Package {
	if x == nil || x.file == nil {
		return nil
	}
	return x.file.pkg
}

// -------------------------------------------------------------------------
// Decl represents an top-level declaration in a file.
// @api(object/decl/Decl)
type Decl interface {
	Node

	declNode()
}

// All decls listed here implement the Decl interface.
var _ Decl = (*File)(nil)
var _ Decl = (*Const)(nil)
var _ Decl = (*Enum)(nil)
var _ Decl = (*Struct)(nil)
var _ Decl = (*Interface)(nil)

func (x *File) declNode()      {}
func (x *Const) declNode()     {}
func (x *Enum) declNode()      {}
func (x *Struct) declNode()    {}
func (x *Interface) declNode() {}

// builtinDecl represents a special declaration for built-in types.
type builtinDecl struct{}

var _ Decl = builtinDecl{}

func (builtinDecl) File() *File                 { return nil }
func (builtinDecl) Package() *Package           { return nil }
func (builtinDecl) getType() string             { return "<builtin.decl>" }
func (builtinDecl) getPos() token.Pos           { return token.NoPos }
func (builtinDecl) getName() string             { return "<builtin>" }
func (builtinDecl) getAnnotations() Annotations { return nil }
func (builtinDecl) declNode()                   {}

// -------------------------------------------------------------------------
// Types

// Type represents a Next type.
type Type interface {
	Object

	// Kind returns the kind of the type.
	// @api(object/Type/Kind)
	Kind() token.Kind

	// String returns the string representation of the type.
	// @api(object/Type/String)
	String() string

	// Decl returns the declaration of the type.
	Decl() Decl
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

func (x *UsedType) Kind() token.Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() token.Kind { return x.kind }
func (*ArrayType) Kind() token.Kind       { return token.KindArray }
func (*VectorType) Kind() token.Kind      { return token.KindVector }
func (*MapType) Kind() token.Kind         { return token.KindMap }
func (x *DeclType[T]) Kind() token.Kind   { return x.kind }

// UsedType represents a used type in a file.
// @api(object/Type/UsedType)
type UsedType struct {
	// File is the file containing the used type.
	// @api(object/Type/UsedType/File)
	File *File

	// Type is the underlying type.
	// @api(object/Type/UsedType/Type)
	Type Type

	// Node is the AST node of the used type.
	Node ast.Type
}

// Use uses a type in a file.
func Use(t Type, f *File, node ast.Type) *UsedType {
	return &UsedType{Type: t, File: f, Node: node}
}

// String returns the string representation of the used type.
func (u *UsedType) String() string { return u.Type.String() }

// PrimitiveType represents a primitive type.
type PrimitiveType struct {
	name string
	kind token.Kind
}

func (b *PrimitiveType) String() string { return b.name }

var primitiveTypes = func() map[string]*PrimitiveType {
	m := make(map[string]*PrimitiveType)
	for _, kind := range token.PrimitiveKinds {
		name := strings.ToLower(kind.String())
		m[name] = &PrimitiveType{kind: kind, name: name}
	}
	return m
}()

// ArrayType represents an array type.
type ArrayType struct {
	pos token.Pos

	ElemType Type
	N        int64
}

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatInt(a.N, 10) + ">"
}

// VectorType represents a vector type.
type VectorType struct {
	pos token.Pos

	ElemType Type
}

func (v *VectorType) String() string {
	return "vector<" + v.ElemType.String() + ">"
}

// MapType represents a map type.
type MapType struct {
	pos token.Pos

	KeyType  Type
	ElemType Type
}

func (m *MapType) String() string {
	return "map<" + m.KeyType.String() + "," + m.ElemType.String() + ">"
}

// DeclType represents a declaration type which is a type of a declaration: enum, struct, interface.
// @api(object/decl/DeclType)
type DeclType[T Decl] struct {
	pos  token.Pos
	kind token.Kind
	name string
	decl T
}

func newDeclType[T Decl](pos token.Pos, kind token.Kind, name string, decl T) *DeclType[T] {
	return &DeclType[T]{pos: pos, kind: kind, name: name, decl: decl}
}

// String returns the string representation of the declaration type.
// @api(object/decl/DeclType/String)
func (d *DeclType[T]) String() string { return d.name }
