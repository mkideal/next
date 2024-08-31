package types

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopherd/core/op"

	"github.com/next/next/src/token"
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
func (*UsedType) ObjectType() string         { return "type.used" }
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

func (x *File) Package() *Package {
	if x == nil {
		return nil
	}
	return x.pkg
}
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

// Spec represents a specification: import, value(const, enum member), type(enum, struct)
type Spec interface {
	Object
	Decl() *Decl

	resolve(ctx *Context, file *File, scope Scope)
}

func (x *ImportSpec) Decl() *Decl    { return x.decl }
func (x *ValueSpec) Decl() *Decl     { return x.decl }
func (e *EnumSpec) Decl() *Decl      { return e.decl }
func (s *StructSpec) Decl() *Decl    { return s.decl }
func (i *InterfaceSpec) Decl() *Decl { return i.decl }

// Type represents a Next type.
type Type interface {
	Object
	Kind() Kind
	String() string
	Spec() Spec
	Package() *Package
	In(*Package) bool
}

func (x *UsedType) Kind() Kind      { return x.Type.Kind() }
func (x *PrimitiveType) Kind() Kind { return x.kind }
func (*ArrayType) Kind() Kind       { return Array }
func (*VectorType) Kind() Kind      { return Vector }
func (*MapType) Kind() Kind         { return Map }
func (*EnumType) Kind() Kind        { return Enum }
func (*StructType) Kind() Kind      { return Struct }
func (*InterfaceType) Kind() Kind   { return Interface }

func (x *UsedType) Spec() Spec      { return x.Type.Spec() }
func (*PrimitiveType) Spec() Spec   { return globalBuiltinSpec }
func (*ArrayType) Spec() Spec       { return globalBuiltinSpec }
func (*VectorType) Spec() Spec      { return globalBuiltinSpec }
func (*MapType) Spec() Spec         { return globalBuiltinSpec }
func (x *EnumType) Spec() Spec      { return x.spec }
func (x *StructType) Spec() Spec    { return x.spec }
func (x *InterfaceType) Spec() Spec { return x.spec }

func (x *UsedType) Package() *Package      { return packageOfType(x) }
func (x *PrimitiveType) Package() *Package { return packageOfType(x) }
func (x *ArrayType) Package() *Package     { return packageOfType(x) }
func (x *VectorType) Package() *Package    { return packageOfType(x) }
func (x *MapType) Package() *Package       { return packageOfType(x) }
func (x *EnumType) Package() *Package      { return packageOfType(x) }
func (x *StructType) Package() *Package    { return packageOfType(x) }
func (x *InterfaceType) Package() *Package { return packageOfType(x) }

func (x *UsedType) In(pkg *Package) bool      { return typeInPackage(x, pkg) }
func (x *PrimitiveType) In(pkg *Package) bool { return typeInPackage(x, pkg) }
func (x *ArrayType) In(pkg *Package) bool     { return typeInPackage(x, pkg) }
func (x *VectorType) In(pkg *Package) bool    { return typeInPackage(x, pkg) }
func (x *MapType) In(pkg *Package) bool       { return typeInPackage(x, pkg) }
func (x *EnumType) In(pkg *Package) bool      { return typeInPackage(x, pkg) }
func (x *StructType) In(pkg *Package) bool    { return typeInPackage(x, pkg) }
func (x *InterfaceType) In(pkg *Package) bool { return typeInPackage(x, pkg) }

func packageOfType(t Type) *Package {
	if t == nil {
		return nil
	}
	return t.Spec().Decl().File().Package()
}

func typeInPackage(t Type, pkg *Package) bool {
	if t == nil {
		return true
	}
	return t.Package().In(pkg)
}

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
