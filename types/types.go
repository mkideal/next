package types

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopherd/core/op"
	"github.com/gopherd/next/token"
)

// Object represents a Next AST node.
type Object interface {
	ObjectType() string
}

func (*File) ObjectType() string             { return "file" }
func (*Comment) ObjectType() string          { return "comment" }
func (*Doc) ObjectType() string              { return "doc" }
func (*CallStmt) ObjectType() string         { return "stmt.call" }
func (*Imports) ObjectType() string          { return "imports" }
func (*ImportSpec) ObjectType() string       { return "import" }
func (x *ValueSpec) ObjectType() string      { return op.If(x.enum.typ != nil, "enum.member", "const") }
func (*Specs) ObjectType() string            { return "specs" }
func (*ConstSpecs) ObjectType() string       { return "consts" }
func (*EnumSpecs) ObjectType() string        { return "enums" }
func (*EnumSpec) ObjectType() string         { return "enum" }
func (EnumMembers) ObjectType() string       { return "enum.members" }
func (StructSpecs) ObjectType() string       { return "structs" }
func (*StructSpec) ObjectType() string       { return "struct" }
func (*Fields) ObjectType() string           { return "struct.fields" }
func (*Field) ObjectType() string            { return "struct.field" }
func (*FieldType) ObjectType() string        { return "struct.field.type" }
func (FieldName) ObjectType() string         { return "struct.field.name" }
func (*InterfaceSpecs) ObjectType() string   { return "interfaces" }
func (*InterfaceSpec) ObjectType() string    { return "interface" }
func (*Methods) ObjectType() string          { return "interface.methods" }
func (*Method) ObjectType() string           { return "interface.method" }
func (MethodName) ObjectType() string        { return "interface.method.name" }
func (MethodParams) ObjectType() string      { return "interface.method.params" }
func (*MethodParam) ObjectType() string      { return "interface.method.param" }
func (*MethodParamType) ObjectType() string  { return "interface.method.param.type" }
func (MethodParamName) ObjectType() string   { return "interface.method.param.name" }
func (*MethodReturnType) ObjectType() string { return "interface.method.return.type" }
func (*ArrayType) ObjectType() string        { return "type.array" }
func (*VectorType) ObjectType() string       { return "type.vector" }
func (*MapType) ObjectType() string          { return "type.map" }
func (*EnumType) ObjectType() string         { return "type.enum" }
func (*StructType) ObjectType() string       { return "type.struct" }
func (*InterfaceType) ObjectType() string    { return "type.interface" }
func (x *PrimitiveType) ObjectType() string  { return x.name }

// Node represents a Next AST node which may be a file, const, enum or struct to be generated.
type Node interface {
	Object
	Package() *Package
	Name() string
}

func (x *File) Package() *Package          { return x.pkg }
func (x *ValueSpec) Package() *Package     { return x.decl.file.pkg }
func (x *EnumSpec) Package() *Package      { return x.decl.file.pkg }
func (x *StructSpec) Package() *Package    { return x.decl.file.pkg }
func (x *InterfaceSpec) Package() *Package { return x.decl.file.pkg }

func (x *File) Name() string          { return strings.TrimSuffix(filepath.Base(x.Path), ".next") }
func (x *ValueSpec) Name() string     { return x.name }
func (x *EnumSpec) Name() string      { return x.Type.name }
func (x *StructSpec) Name() string    { return x.Type.name }
func (x *InterfaceSpec) Name() string { return x.Type.name }

// Symbol represents a Next symbol: value(constant, enum member), type(struct, enum)
type Symbol interface {
	Object
	symbolPos() token.Pos
	symbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

func (x *ValueSpec) symbolPos() token.Pos     { return x.pos }
func (x *EnumType) symbolPos() token.Pos      { return x.pos }
func (x *StructType) symbolPos() token.Pos    { return x.pos }
func (x *InterfaceType) symbolPos() token.Pos { return x.pos }

func (*ValueSpec) symbolType() string     { return ValueSymbol }
func (*EnumType) symbolType() string      { return TypeSymbol }
func (*StructType) symbolType() string    { return TypeSymbol }
func (*InterfaceType) symbolType() string { return TypeSymbol }

// Type represents a Next type.
type Type interface {
	Object
	Package() *Package
	Kind() Kind
	String() string
}

func (*PrimitiveType) Package() *Package   { return nil }
func (*ArrayType) Package() *Package       { return nil }
func (*VectorType) Package() *Package      { return nil }
func (*MapType) Package() *Package         { return nil }
func (x *EnumType) Package() *Package      { return x.spec.Package() }
func (x *StructType) Package() *Package    { return x.spec.Package() }
func (x *InterfaceType) Package() *Package { return x.spec.Package() }

func (x *PrimitiveType) Kind() Kind { return x.kind }
func (*ArrayType) Kind() Kind       { return Array }
func (*VectorType) Kind() Kind      { return Vector }
func (*MapType) Kind() Kind         { return Map }
func (*EnumType) Kind() Kind        { return Enum }
func (*StructType) Kind() Kind      { return Struct }

// Spec represents a specification: import, value(const, enum member), type(enum, struct)
type Spec interface {
	Object
	specNode()
	resolve(ctx *Context, file *File, scope Scope)
}

func (*ImportSpec) specNode()    {}
func (*ValueSpec) specNode()     {}
func (*EnumSpec) specNode()      {}
func (*StructSpec) specNode()    {}
func (*InterfaceSpec) specNode() {}

//-------------------------------------------------------------------------
// Types

// PrimitiveType represents a primitive type.
type PrimitiveType struct {
	pos  token.Pos
	name string
	kind Kind
}

func (b *PrimitiveType) String() string { return b.name }

var primitiveTypes = func() map[string]*PrimitiveType {
	m := make(map[string]*PrimitiveType)
	for _, kind := range primitiveKinds {
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

// EnumType represents an enum type.
type EnumType struct {
	name string
	pos  token.Pos
	spec *EnumSpec
}

func (e *EnumType) String() string { return e.name }

// StructType represents a struct type.
type StructType struct {
	name string
	pos  token.Pos
	spec *StructSpec
}

func (s *StructType) String() string { return s.name }

// InterfaceType represents an interface type.
type InterfaceType struct {
	name string
	pos  token.Pos
	spec *InterfaceSpec
}

func (i *InterfaceType) String() string { return i.name }
