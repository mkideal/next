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

func (*File) ObjectType() string        { return "file" }
func (*Comment) ObjectType() string     { return "comment" }
func (*Doc) ObjectType() string         { return "doc" }
func (*CallStmt) ObjectType() string    { return "stmt.call" }
func (*Imports) ObjectType() string     { return "imports" }
func (*ImportSpec) ObjectType() string  { return "import" }
func (v *ValueSpec) ObjectType() string { return op.If(v.enum.typ != nil, "enum.member", "const") }
func (*Specs) ObjectType() string       { return "specs" }
func (*ConstSpecs) ObjectType() string  { return "consts" }
func (*EnumSpecs) ObjectType() string   { return "enums" }
func (*EnumSpec) ObjectType() string    { return "enum" }
func (EnumMembers) ObjectType() string  { return "enum.members" }
func (StructSpecs) ObjectType() string  { return "structs" }
func (*StructSpec) ObjectType() string  { return "struct" }
func (*Fields) ObjectType() string      { return "struct.fields" }
func (*Field) ObjectType() string       { return "struct.field" }
func (b *BasicType) ObjectType() string { return b.name }
func (*ArrayType) ObjectType() string   { return "type.array" }
func (*VectorType) ObjectType() string  { return "type.vector" }
func (*MapType) ObjectType() string     { return "type.map" }
func (*EnumType) ObjectType() string    { return "type.enum" }
func (*StructType) ObjectType() string  { return "type.struct" }

// Node represents a Next AST node which may be a file, const, enum or struct to be generated.
type Node interface {
	Object
	Package() *Package
	Name() string
}

func (f *File) Package() *Package       { return f.pkg }
func (v *ValueSpec) Package() *Package  { return v.decl.file.pkg }
func (e *EnumSpec) Package() *Package   { return e.decl.file.pkg }
func (s *StructSpec) Package() *Package { return s.decl.file.pkg }

func (f *File) Name() string       { return strings.TrimSuffix(filepath.Base(f.Path), ".next") }
func (v *ValueSpec) Name() string  { return v.name }
func (e *EnumSpec) Name() string   { return e.Type.name }
func (s *StructSpec) Name() string { return s.Type.name }

// Symbol represents a Next symbol: value(constant, enum member), type(struct, enum)
type Symbol interface {
	Object
	Pos() token.Pos
	SymbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

func (v *ValueSpec) Pos() token.Pos  { return v.pos }
func (e *EnumType) Pos() token.Pos   { return e.pos }
func (s *StructType) Pos() token.Pos { return s.pos }

func (*ValueSpec) SymbolType() string  { return ValueSymbol }
func (*EnumType) SymbolType() string   { return TypeSymbol }
func (*StructType) SymbolType() string { return TypeSymbol }

// Type represents a Next type.
type Type interface {
	Object
	Kind() Kind
	String() string
}

func (b *BasicType) Kind() Kind { return b.kind }
func (*ArrayType) Kind() Kind   { return Array }
func (*VectorType) Kind() Kind  { return Vector }
func (*MapType) Kind() Kind     { return Map }
func (*EnumType) Kind() Kind    { return Enum }
func (*StructType) Kind() Kind  { return Struct }

// Spec represents a specification: import, value(const, enum member), type(enum, struct)
type Spec interface {
	Object
	specNode()
	resolve(ctx *Context, file *File, scope Scope)
}

func (*ImportSpec) specNode() {}
func (*ValueSpec) specNode()  {}
func (*EnumSpec) specNode()   {}
func (*StructSpec) specNode() {}

//-------------------------------------------------------------------------
// Types

// BasicType represents a basic type.
type BasicType struct {
	pos  token.Pos
	name string
	kind Kind
}

func (b *BasicType) String() string { return b.name }

var basicTypes = map[string]*BasicType{
	"int":     {kind: Int, name: "int"},
	"int8":    {kind: Int8, name: "int8"},
	"int16":   {kind: Int16, name: "int16"},
	"int32":   {kind: Int32, name: "int32"},
	"int64":   {kind: Int64, name: "int64"},
	"float32": {kind: Float32, name: "float32"},
	"float64": {kind: Float64, name: "float64"},
	"bool":    {kind: Bool, name: "bool"},
	"string":  {kind: String, name: "string"},
	"byte":    {kind: Byte, name: "byte"},
	"bytes":   {kind: Bytes, name: "bytes"},
}

// ArrayType represents an array type.
type ArrayType struct {
	pos token.Pos

	ElemType Type
	N        uint64
}

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatUint(a.N, 10) + ">"
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
