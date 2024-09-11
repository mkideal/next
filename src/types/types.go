package types

import (
	"strconv"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// -------------------------------------------------------------------------

// @api(Object/Object) represents an object in Next which can be rendered in a template like this: {{next Object}}
type Object interface {
	// getType returns the type of the object.
	getType() string
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
var _ Object = (*ConstName)(nil)
var _ Object = (*EnumMemberName)(nil)
var _ Object = (*StructFieldName)(nil)
var _ Object = (*InterfaceMethodName)(nil)
var _ Object = (*InterfaceMethodParamName)(nil)
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
func (x *Fields[D, F]) getType() string {
	var zero F
	return zero.getType() + "s"
}

// Name objects: const.name, enum.member.name, struct.field.name, interface.method.name, interface.method.param.name
func (x *NodeName[T]) getType() string {
	return x.node.getType() + ".name"
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

// @api(Object/Node) represents a Node in the AST. It's a special object that can be annotated with a documentation comment.
type Node interface {
	LocatedObject

	// getName returns the name of the node.
	getName() string

	// @api(Object/Node.File) represents the file containing the node.
	File() *File

	// @api(Object/Node.Package) represents the package containing the node.
	// It's a shortcut for Node.File().Package().
	Package() *Package

	// @api(Object/Node.Doc) represents the documentation comment for the node.
	Doc() *Doc

	// @api(Object/Node.Annotations) represents the annotations for the node.
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

func (x *File) getName() string             { return x.Name() }
func (x *commonNode[Self]) getName() string { return x.name.name }

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

// @api(Object/Decl) represents an top-level declaration in a file.
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

func (builtinDecl) getType() string          { return "<builtin.decl>" }
func (builtinDecl) getName() string          { return "<builtin>" }
func (builtinDecl) getPos() token.Pos        { return token.NoPos }
func (builtinDecl) File() *File              { return nil }
func (builtinDecl) Package() *Package        { return nil }
func (builtinDecl) Doc() *Doc                { return nil }
func (builtinDecl) Annotations() Annotations { return nil }
func (builtinDecl) declNode()                {}

// -------------------------------------------------------------------------
// Types

// @api(Object/Type) represents a Next type.
type Type interface {
	Object

	// @api(Object/Type.Kind) returns the kind of the type.
	Kind() token.Kind

	// @api(Object/Type.String) represents the string representation of the type.
	String() string

	// @api(Object/Type.Decl) represents the [declaration](#Decl) of the type.
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

// @api(Object/Type/UsedType) represents a used type in a file.
type UsedType struct {
	// @api(Object/UsedType.File) represents the file containing the used type.
	File *File

	// @api(Object/UsedType.Type) represents the used type.
	Type Type

	// node represents the AST node of the used type.
	node ast.Type
}

// Use uses a type in a file.
func Use(t Type, f *File, node ast.Type) *UsedType {
	return &UsedType{Type: t, File: f, node: node}
}

// String returns the string representation of the used type.
func (u *UsedType) String() string { return u.Type.String() }

// UsedTypeNode returns the AST node of the used type.
func UsedTypeNode(u *UsedType) ast.Type { return u.node }

// @api(Object/Type/PrimitiveType) represents a primitive type.
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

// @api(Object/Type/ArrayType) represents an array type.
type ArrayType struct {
	pos token.Pos

	ElemType Type
	N        int64
}

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatInt(a.N, 10) + ">"
}

// @api(Object/Type/VectorType) represents a vector type.
type VectorType struct {
	pos token.Pos

	ElemType Type
}

func (v *VectorType) String() string {
	return "vector<" + v.ElemType.String() + ">"
}

// @api(Object/Type/MapType) represents a map type.
type MapType struct {
	pos token.Pos

	KeyType  Type
	ElemType Type
}

func (m *MapType) String() string {
	return "map<" + m.KeyType.String() + "," + m.ElemType.String() + ">"
}

// DeclType represents a declaration type which is a type of a declaration: enum, struct, interface.
type DeclType[T Decl] struct {
	pos  token.Pos
	kind token.Kind
	name string
	decl T
}

// @api(Object/Type/EnumType) represents the type of an [enum](#Object/Enum) declaration.
type EnumType = DeclType[*Enum]

// @api(Object/Type/StructType) represents the type of a [struct](#Object/Struct) declaration.
type StructType = DeclType[*Struct]

// @api(Object/Type/InterfaceType) represents the type of an [interface](#Object/Interface) declaration.
type InterfaceType = DeclType[*Interface]

func newDeclType[T Decl](pos token.Pos, kind token.Kind, name string, decl T) *DeclType[T] {
	return &DeclType[T]{pos: pos, kind: kind, name: name, decl: decl}
}

func (d *DeclType[T]) String() string { return d.name }
