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
	Pos() token.Pos
	ObjectType() string
}

func (f *File) Pos() token.Pos       { return f.pos }
func (s *CallStmt) Pos() token.Pos   { return s.pos }
func (i *Imports) Pos() token.Pos    { return op.If(len(i.Specs) == 0, token.NoPos, i.Specs[0].pos) }
func (v *ValueSpec) Pos() token.Pos  { return v.pos }
func (i *ImportSpec) Pos() token.Pos { return i.pos }
func (e *EnumType) Pos() token.Pos   { return e.pos }
func (b *BasicType) Pos() token.Pos  { return b.pos }
func (a *ArrayType) Pos() token.Pos  { return a.pos }
func (v *VectorType) Pos() token.Pos { return v.pos }
func (m *MapType) Pos() token.Pos    { return m.pos }
func (s *StructType) Pos() token.Pos { return s.pos }
func (f *Field) Pos() token.Pos      { return f.pos }

func (*File) ObjectType() string        { return "file" }
func (*CallStmt) ObjectType() string    { return "stmt.call" }
func (*Imports) ObjectType() string     { return "imports" }
func (*ImportSpec) ObjectType() string  { return "import" }
func (*StructType) ObjectType() string  { return "struct" }
func (*Field) ObjectType() string       { return "struct.field" }
func (*EnumType) ObjectType() string    { return "enum" }
func (v *ValueSpec) ObjectType() string { return op.If(v.enum.typ != nil, "enum.member", "const") }
func (b *BasicType) ObjectType() string { return b.name }
func (*ArrayType) ObjectType() string   { return "array" }
func (*VectorType) ObjectType() string  { return "vector" }
func (*MapType) ObjectType() string     { return "map" }

// Node represents a Next AST node which may be a file, const, enum or struct to be generated.
type Node interface {
	Object
	Package() string
	Name() string
}

func (f *File) Package() string       { return f.pkg.name }
func (e *EnumType) Package() string   { return e.decl.file.pkg.name }
func (s *StructType) Package() string { return s.decl.file.pkg.name }
func (v *ValueSpec) Package() string  { return v.decl.file.pkg.name }

func (f *File) Name() string       { return strings.TrimSuffix(filepath.Base(f.Path), ".next") }
func (s *StructType) Name() string { return s.name }
func (e *EnumType) Name() string   { return e.name }
func (v *ValueSpec) Name() string  { return v.name }

// Symbol represents a Next symbol: value(constant, enum member), type(struct, enum)
type Symbol interface {
	Object
	SymbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

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
func (*EnumType) specNode()   {}
func (*StructType) specNode() {}

//-------------------------------------------------------------------------
// Builtin Types

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
