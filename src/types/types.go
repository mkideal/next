package types

import (
	"strconv"
	"strings"

	"github.com/next/next/src/token"
)

// Node represents a Next AST node.
// @api(object/Node)
type Node interface {
	getType() string
}

func (*File) getType() string            { return "file" }
func (*Doc) getType() string             { return "doc" }
func (*Comment) getType() string         { return "comment" }
func (*Imports) getType() string         { return "imports" }
func (*Import) getType() string          { return "import" }
func (*UsedType) getType() string        { return "type.used" }
func (*ArrayType) getType() string       { return "type.array" }
func (*VectorType) getType() string      { return "type.vector" }
func (*MapType) getType() string         { return "type.map" }
func (x *PrimitiveType) getType() string { return "type." + x.name }

func (*Decls) getType() string                     { return "decls" }
func (*Const) getType() string                     { return "const" }
func (ConstName) getType() string                  { return "const.name" }
func (*Enum) getType() string                      { return "enum" }
func (EnumName) getType() string                   { return "enum.name" }
func (*EnumMember) getType() string                { return "enum.member" }
func (EnumMemberName) getType() string             { return "enum.member.name" }
func (*Struct) getType() string                    { return "struct" }
func (StructName) getType() string                 { return "struct.name" }
func (*StructField) getType() string               { return "struct.field" }
func (*StructFieldType) getType() string           { return "struct.field.type" }
func (StructFieldName) getType() string            { return "struct.field.name" }
func (*Interface) getType() string                 { return "interface" }
func (InterfaceName) getType() string              { return "interface.name" }
func (*InterfaceMethod) getType() string           { return "interface.method" }
func (InterfaceMethodName) getType() string        { return "interface.method.name" }
func (*InterfaceMethodParam) getType() string      { return "interface.method.param" }
func (InterfaceMethodParamName) getType() string   { return "interface.method.param.name" }
func (*InterfaceMethodParamType) getType() string  { return "interface.method.param.type" }
func (*InterfaceMethodReturnType) getType() string { return "interface.method.return.type" }

// list types: consts, enums, structs, interfaces
func (l List[T]) getType() string {
	var zero T
	return zero.getType() + "s"
}

func (v *Value) getType() string {
	if v.enum.typ == nil {
		return "const.value"
	}
	return "enum.member.value"
}

// fields types: enum.members, struct.fields, interface.methods
func (l *Fields[Parent, Field]) getType() string {
	return l.typename
}

// decl types: type.enum, type.struct, type.interface
func (t *DeclType[T]) getType() string {
	return "type." + t.decl.getType()
}

// Object represents a Next AST node which may be a file, const, enum, struct, interface to be generated.
type Object interface {
	Node

	getName() string
	getPos() token.Pos
}

var _ Object = (*File)(nil)
var _ Object = (*Value)(nil)
var _ Object = (*Const)(nil)
var _ Object = (*Enum)(nil)
var _ Object = (*Struct)(nil)
var _ Object = (*Interface)(nil)
var _ Object = (*EnumMember)(nil)
var _ Object = (*StructField)(nil)
var _ Object = (*InterfaceMethod)(nil)

func (x *File) getName() string             { return x.Name() }
func (x *Value) getName() string            { return x.name }
func (x *decl[Self, Name]) getName() string { return string(x.name) }

func (x *File) getPos() token.Pos             { return x.pos }
func (x *Value) getPos() token.Pos            { return x.namePos }
func (x *decl[Self, Name]) getPos() token.Pos { return x.pos }

// Type represents a Next type.
type Type interface {
	Node
	Kind() token.Kind
	String() string
	Decl() Decl
}

var _ Type = (*UsedType)(nil)
var _ Type = (*PrimitiveType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*VectorType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*DeclType[*Enum])(nil)
var _ Type = (*DeclType[*Struct])(nil)
var _ Type = (*DeclType[*Interface])(nil)

func (x *UsedType) Decl() Decl      { return x.Type.Decl() }
func (x *PrimitiveType) Decl() Decl { return builtinDecl }
func (*ArrayType) Decl() Decl       { return builtinDecl }
func (*VectorType) Decl() Decl      { return builtinDecl }
func (*MapType) Decl() Decl         { return builtinDecl }
func (x *DeclType[T]) Decl() Decl   { return x.decl }

func (x *UsedType) Kind() token.Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() token.Kind { return x.kind }
func (*ArrayType) Kind() token.Kind       { return token.Array }
func (*VectorType) Kind() token.Kind      { return token.Vector }
func (*MapType) Kind() token.Kind         { return token.Map }
func (x *DeclType[T]) Kind() token.Kind   { return x.kind }

//-------------------------------------------------------------------------
// Types

type UsedType struct {
	Type Type
	File *File
}

// Use uses a type in a file.
func Use(t Type, f *File) *UsedType {
	return &UsedType{Type: t, File: f}
}

func (u *UsedType) String() string { return u.Type.String() }

// PrimitiveType represents a primitive type.
type PrimitiveType struct {
	pos  token.Pos
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
