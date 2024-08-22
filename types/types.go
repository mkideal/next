package types

import (
	"strconv"
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

type Scope interface {
	ParentScope() Scope
	LookupLocalSymbol(name string) Symbol
}

func lookupSymbol(scope Scope, name string) Symbol {
	for s := scope; s != nil; s = s.ParentScope() {
		if sym := s.LookupLocalSymbol(name); sym != nil {
			return sym
		}
	}
	return nil
}

func lookupType(scope Scope, name string) (Type, error) {
	return expectTypeSymbol(name, lookupSymbol(scope, name))
}

func lookupValue(scope Scope, name string) (*ValueSpec, error) {
	return expectValueSymbol(name, lookupSymbol(scope, name))
}

func expectTypeSymbol(name string, s Symbol) (Type, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if t, ok := s.(Type); ok {
		return t, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "type", Got: s.symbolType()}
}

func expectValueSymbol(name string, s Symbol) (*ValueSpec, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*ValueSpec); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "value", Got: s.symbolType()}
}

//-------------------------------------------------------------------------
// Types

// Type represents a Next type.
type Type interface {
	Node
	typeNode()

	String() string
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
	name string
	kind Kind
}

func (*BasicType) typeNode()        {}
func (b *BasicType) String() string { return b.name }
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
	typ

	ElemType Type
	N        uint64
}

func (*ArrayType) typeNode() {}

func (a *ArrayType) String() string {
	return "array<" + a.ElemType.String() + "," + strconv.FormatUint(a.N, 10) + ">"
}

func (*ArrayType) Kind() Kind    { return Array }
func (*ArrayType) IsArray() bool { return true }

// VectorType represents a vector type.
type VectorType struct {
	typ

	ElemType Type
}

func (*VectorType) typeNode() {}

func (v *VectorType) String() string {
	return "vector<" + v.ElemType.String() + ">"
}

func (*VectorType) Kind() Kind     { return Vector }
func (*VectorType) IsVector() bool { return true }

// MapType represents a map type.
type MapType struct {
	typ

	KeyType  Type
	ElemType Type
}

func (*MapType) typeNode() {}

func (m *MapType) String() string {
	return "map<" + m.KeyType.String() + "," + m.ElemType.String() + ">"
}

func (*MapType) Kind() Kind  { return Map }
func (*MapType) IsMap() bool { return true }
