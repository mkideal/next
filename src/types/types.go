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
func (x *PrimitiveType) getName() string    { return x.name }
func (x *ArrayType) getName() string        { return x.String() }
func (x *VectorType) getName() string       { return x.String() }
func (x *MapType) getName() string          { return x.String() }

func (x *File) getPos() token.Pos             { return x.pos }
func (x *Value) getPos() token.Pos            { return x.namePos }
func (x *decl[Self, Name]) getPos() token.Pos { return x.pos }
func (x *PrimitiveType) getPos() token.Pos    { return x.pos }
func (x *ArrayType) getPos() token.Pos        { return x.pos }
func (x *VectorType) getPos() token.Pos       { return x.pos }
func (x *MapType) getPos() token.Pos          { return x.pos }

// -------------------------------------------------------------------------
// Types

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
func (x *PrimitiveType) Decl() Decl { return x }
func (x *ArrayType) Decl() Decl     { return x }
func (x *VectorType) Decl() Decl    { return x }
func (x *MapType) Decl() Decl       { return x }
func (x *DeclType[T]) Decl() Decl   { return x.decl }

func (x *UsedType) Kind() token.Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() token.Kind { return x.kind }
func (*ArrayType) Kind() token.Kind       { return token.Array }
func (*VectorType) Kind() token.Kind      { return token.Vector }
func (*MapType) Kind() token.Kind         { return token.Map }
func (x *DeclType[T]) Kind() token.Kind   { return x.kind }

// UsedType represents a used type in a file.
// @api(object/Type/UsedType)
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

// -------------------------------------------------------------------------
// Symbol represents a Next symbol: value(const, enum member), type(enum, struct, interface).
type Symbol interface {
	Node
	symbolPos() token.Pos
	symbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

func (x *Value) symbolPos() token.Pos       { return x.namePos }
func (t *DeclType[T]) symbolPos() token.Pos { return t.pos }

func (*Value) symbolType() string         { return ValueSymbol }
func (t *DeclType[T]) symbolType() string { return TypeSymbol }

func splitSymbolName(name string) (ns, sym string) {
	if i := strings.Index(name, "."); i >= 0 {
		return name[:i], name[i+1:]
	}
	return "", name
}

func joinSymbolName(syms ...string) string {
	return strings.Join(syms, ".")
}

// Scope represents a symbol scope.
type Scope interface {
	ParentScope() Scope
	LookupLocalSymbol(name string) Symbol
}

func (f *File) ParentScope() Scope { return &fileParentScope{f} }
func (e *Enum) ParentScope() Scope { return e.file }

func (f *File) LookupLocalSymbol(name string) Symbol { return f.symbols[name] }
func (e *Enum) LookupLocalSymbol(name string) Symbol {
	for _, m := range e.Members.List {
		if string(m.name) == name {
			return m.value
		}
	}
	return nil
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

func lookupValue(scope Scope, name string) (*Value, error) {
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

func expectValueSymbol(name string, s Symbol) (*Value, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*Value); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "value", Got: s.symbolType()}
}

type fileParentScope struct {
	f *File
}

func (s *fileParentScope) ParentScope() Scope {
	return nil
}

func (s *fileParentScope) LookupLocalSymbol(name string) Symbol {
	var files []*File
	pkg, name := splitSymbolName(name)
	for i := range s.f.imports.List {
		if s.f.imports.List[i].target.pkg.name == pkg {
			files = append(files, s.f.imports.List[i].target)
		}
	}
	for _, file := range files {
		if s := file.symbols[name]; s != nil {
			return s
		}
	}
	return nil
}
