package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

// Spec represents a specification: import, value(constant, enum), struct, protocol
type Spec interface {
	Node
	resolve(ctx *Context, file *File, scope Scope)
	specObject()
}

func (*ImportSpec) specObject()   {}
func (*ValueSpec) specObject()    {}
func (*EnumType) specObject()     {}
func (*StructType) specObject()   {}
func (*ProtocolType) specObject() {}

func newSpec(ctx *Context, file *File, s ast.Spec) Spec {
	switch s := s.(type) {
	case *ast.ImportSpec:
		return newImportSpec(ctx, file, s)
	case *ast.TypeSpec:
		return newTypeSpec(ctx, file, s)
	case *ast.ValueSpec:
		return newValueSpec(ctx, file, s)
	default:
		panic("unexpected spec type")
	}
}

// ImportSpec represents a file import.
type ImportSpec struct {
	pos        token.Pos
	file       *File
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Path         string
}

func newImportSpec(_ *Context, _ *File, src *ast.ImportSpec) *ImportSpec {
	i := &ImportSpec{
		pos:     src.Pos(),
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Path:    src.Path.Value,
	}
	i.unresolved.annotations = src.Annotations
	return i
}

func (i *ImportSpec) Pos() token.Pos { return i.pos }

func (i *ImportSpec) resolve(ctx *Context, file *File, _ Scope) {
	i.Annotations = ctx.resolveAnnotationGroup(file, i.unresolved.annotations)
}

type iotaValue struct {
	value int
	found bool
}

// ValueSpec represents an constant or enum member.
type ValueSpec struct {
	pos  token.Pos
	enum struct {
		typ   *EnumType // parent enum type
		index int       // index in the enum type. start from 0
		iota  iotaValue // iota value
	}
	unresolved struct {
		annotations *ast.AnnotationGroup
		value       ast.Expr
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Value        constant.Value
}

func newValueSpec(_ *Context, _ *File, src *ast.ValueSpec) *ValueSpec {
	v := &ValueSpec{
		pos:     src.Pos(),
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	v.unresolved.annotations = src.Annotations
	v.unresolved.value = src.Value
	return v
}

func (v *ValueSpec) Pos() token.Pos { return v.pos }

func (v *ValueSpec) resolve(ctx *Context, file *File, scope Scope) {
	v.Annotations = ctx.resolveAnnotationGroup(file, v.unresolved.annotations)
	v.resolveValue(ctx, file, scope, make([]*ValueSpec, 0, 16))
}

func (v *ValueSpec) resolveValue(ctx *Context, file *File, scope Scope, refs []*ValueSpec) constant.Value {
	// If value already resolved, return it
	if v.Value != nil {
		return v.Value
	}

	// If enum type is nil, resolve constant value expression in which iota is not allowed
	if v.enum.typ == nil {
		v.Value = ctx.recusiveResolveValue(file, scope, append(refs, v), v.unresolved.value, nil)
		return v.Value
	}

	if v.unresolved.value != nil {
		// Resolve value expression
		v.Value = ctx.recusiveResolveValue(file, v.enum.typ, append(refs, v), v.unresolved.value, &v.enum.iota)
	} else if v.enum.index == 0 {
		// First member of enum type has value 0 if not specified
		v.Value = constant.MakeInt64(0)
	} else {
		// Resolve previous value
		prev := v.enum.typ.Members[v.enum.index-1]
		prev.resolveValue(ctx, file, v.enum.typ, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.typ.Members[v.enum.index-v.enum.iota.value]

		if start.Value != nil && start.enum.iota.found {
			// If start value is specified and it has iota expression, resolve it with the current iota value
			v.Value = ctx.recusiveResolveValue(file, v.enum.typ, append(refs, v), start.unresolved.value, &v.enum.iota)
		} else {
			// Otherwise, add 1 to the previous value
			v.Value = constant.BinaryOp(prev.Value, token.ADD, constant.MakeInt64(1))
		}
	}
	return v.Value
}

func newTypeSpec(ctx *Context, file *File, src *ast.TypeSpec) Spec {
	switch t := src.Type.(type) {
	case *ast.EnumType:
		return newEnumType(ctx, file, src, t)
	case *ast.StructType:
		return newBeanType(ctx, file, src, t)
	default:
		panic("unexpected type")
	}
}

// Field represents a struct/protocol field.
type Field struct {
	pos        token.Pos
	unresolved struct {
		annotations *ast.AnnotationGroup
		typ         ast.Type
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Type         Type
}

func newField(_ *Context, src *ast.Field) *Field {
	f := &Field{
		pos:     src.Pos(),
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	f.unresolved.annotations = src.Annotations
	f.unresolved.typ = src.Type
	return f
}

func (f *Field) Pos() token.Pos { return f.pos }

func (f *Field) resolve(ctx *Context, file *File, _ Scope) {
	f.Annotations = ctx.resolveAnnotationGroup(file, f.unresolved.annotations)
	f.Type = ctx.resolveType(file, f.unresolved.typ)
}

// beanType represents a an struct/protocol type.
type beanType struct {
	typ
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Fields       []*Field
}

func (*beanType) IsBean() bool { return true }

func newBeanType(ctx *Context, _ *File, src *ast.TypeSpec, t *ast.StructType) Spec {
	b := &beanType{
		typ:     typ{pos: src.Pos()},
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	b.unresolved.annotations = src.Annotations
	for _, f := range t.Fields.List {
		b.Fields = append(b.Fields, newField(ctx, f))
	}
	if src.Keyword == token.STRUCT {
		return &StructType{*b}
	}
	return &ProtocolType{*b}
}

func (b *beanType) resolve(ctx *Context, file *File, scope Scope) {
	b.Annotations = ctx.resolveAnnotationGroup(file, b.unresolved.annotations)
	for _, f := range b.Fields {
		f.resolve(ctx, file, scope)
	}
}

// StructType represents a struct type.
type StructType struct {
	beanType
}

func (*StructType) Kind() Kind     { return Struct }
func (*StructType) IsStruct() bool { return true }

// ProtocolType represents a protocol type declaration.
type ProtocolType struct {
	beanType
}

func (*ProtocolType) Kind() Kind       { return Protocol }
func (*ProtocolType) IsProtocol() bool { return true }

// EnumType represents an enumeration type declaration.
type EnumType struct {
	typ
	unresolved struct {
		annotations *ast.AnnotationGroup
	}
	file *File

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Members      []*ValueSpec
}

func newEnumType(ctx *Context, file *File, src *ast.TypeSpec, t *ast.EnumType) *EnumType {
	e := &EnumType{
		typ:     typ{pos: src.Pos()},
		file:    file,
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	e.unresolved.annotations = src.Annotations
	for i, v := range t.Values {
		m := newValueSpec(ctx, file, v)
		m.enum.typ = e
		m.enum.index = i
		e.Members = append(e.Members, m)
	}
	return e
}

func (e *EnumType) resolve(ctx *Context, file *File, scope Scope) {
	e.Annotations = ctx.resolveAnnotationGroup(file, e.unresolved.annotations)
	for _, m := range e.Members {
		m.resolve(ctx, file, e)
	}
}

func (*EnumType) Kind() Kind   { return Enum }
func (*EnumType) IsEnum() bool { return true }

func (e *EnumType) ParentScope() Scope { return e.file }

func (e *EnumType) LookupLocalSymbol(name string) Symbol {
	for _, m := range e.Members {
		if m.Name == name {
			return m
		}
	}
	return nil
}
