package types

import (
	"strings"

	"github.com/gopherd/next/token"
)

func splitSymbolName(name string) (ns, sym string) {
	if i := strings.Index(name, "."); i >= 0 {
		return name[:i], name[i+1:]
	}
	return "", name
}

func joinSymbolName(syms ...string) string {
	return strings.Join(syms, ".")
}

type Node interface {
	Pos() token.Pos
}

// Symbol represents a Next symbol: value(constant, enum member), type(struct, protocol, enum)
type Symbol interface {
	Node
	symbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

func (*ValueSpec) symbolType() string    { return ValueSymbol }
func (*EnumType) symbolType() string     { return TypeSymbol }
func (*StructType) symbolType() string   { return TypeSymbol }
func (*ProtocolType) symbolType() string { return TypeSymbol }

//-------------------------------------------------------------------------
// Types

// Type represents a Next type.
type Type interface {
	Node
	Kind() Kind
	IsBool() bool
	IsInteger() bool
	IsString() bool
	IsByte() bool
	IsBytes() bool
	IsVector() bool
	IsArray() bool
	IsMap() bool
	IsEnum() bool
	IsStruct() bool
	IsProtocol() bool
	IsBean() bool // IsStruct() || IsProtocol()
}

type typ struct {
	pos token.Pos
}

func (t *typ) Pos() token.Pos { return t.pos }
func (*typ) Kind() Kind       { return Invalid }
func (*typ) IsBool() bool     { return false }
func (*typ) IsInteger() bool  { return false }
func (*typ) IsString() bool   { return false }
func (*typ) IsByte() bool     { return false }
func (*typ) IsBytes() bool    { return false }
func (*typ) IsVector() bool   { return false }
func (*typ) IsArray() bool    { return false }
func (*typ) IsMap() bool      { return false }
func (*typ) IsEnum() bool     { return false }
func (*typ) IsStruct() bool   { return false }
func (*typ) IsProtocol() bool { return false }
func (*typ) IsBean() bool     { return false }

// BasicType represents a basic type.
type BasicType struct {
	typ
	kind Kind
}

func (b *BasicType) Kind() Kind     { return b.kind }
func (b *BasicType) IsBool() bool   { return b.kind == Bool }
func (b *BasicType) IsString() bool { return b.kind == String }
func (b *BasicType) IsByte() bool   { return b.kind == Byte }
func (b *BasicType) IsBytes() bool  { return b.kind == Bytes }

func (b *BasicType) IsInteger() bool {
	switch b.kind {
	case Int, Int8, Int16, Int32, Int64:
		return true
	}
	return false
}

// ArrayType represents an array type.
type ArrayType struct {
	typ

	ElemType Type
	N        uint64
}

func (*ArrayType) Kind() Kind    { return Array }
func (*ArrayType) IsArray() bool { return true }

// VectorType represents a vector type.
type VectorType struct {
	typ

	ElemType Type
}

func (*VectorType) Kind() Kind     { return Vector }
func (*VectorType) IsVector() bool { return true }

// MapType represents a map type.
type MapType struct {
	typ

	KeyType  Type
	ElemType Type
}

func (*MapType) Kind() Kind  { return Map }
func (*MapType) IsMap() bool { return true }
