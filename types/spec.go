package types

import (
	"strconv"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

// Spec represents a specification: import, value(constant, enum), struct
type Spec interface {
	Node
	specObject()

	Decl() *Decl
	resolve(ctx *Context, file *File, scope Scope)
}

func (*ImportSpec) specObject() {}
func (*ValueSpec) specObject()  {}
func (*EnumType) specObject()   {}
func (*StructType) specObject() {}

func newSpec(ctx *Context, file *File, decl *Decl, s ast.Spec) Spec {
	switch s := s.(type) {
	case *ast.ImportSpec:
		return newImportSpec(ctx, file, decl, s)
	case *ast.TypeSpec:
		return newTypeSpec(ctx, file, decl, s)
	case *ast.ValueSpec:
		return newValueSpec(ctx, file, decl, s)
	default:
		panic("unexpected spec type")
	}
}

// ImportSpec represents a file import.
type ImportSpec struct {
	pos          token.Pos
	importedFile *File
	decl         *Decl
	unresolved   struct {
		annotations *ast.AnnotationGroup
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Path         string
}

func newImportSpec(ctx *Context, _ *File, decl *Decl, src *ast.ImportSpec) *ImportSpec {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		ctx.errorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
		path = "!BAD-IMPORT-PATH!"
	}
	i := &ImportSpec{
		pos:     src.Pos(),
		decl:    decl,
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Path:    path,
	}
	i.unresolved.annotations = src.Annotations
	return i
}

func (i *ImportSpec) Pos() token.Pos { return i.pos }

func (i *ImportSpec) Decl() *Decl { return i.decl }

func (i *ImportSpec) resolve(ctx *Context, file *File, _ Scope) {
	i.Annotations = ctx.resolveAnnotationGroup(file, i.unresolved.annotations)
}

type iotaValue struct {
	value int
	found bool
}

// ValueSpec represents an constant or enum member.
type ValueSpec struct {
	pos   token.Pos
	decl  *Decl
	value constant.Value
	enum  struct {
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
}

func newValueSpec(_ *Context, _ *File, decl *Decl, src *ast.ValueSpec) *ValueSpec {
	v := &ValueSpec{
		pos:     src.Pos(),
		decl:    decl,
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	v.unresolved.annotations = src.Annotations
	v.unresolved.value = src.Value
	return v
}

func (v *ValueSpec) Pos() token.Pos { return v.pos }

func (v *ValueSpec) Decl() *Decl { return v.decl }

func (v *ValueSpec) String() string {
	if v.value != nil {
		return v.value.String()
	}
	return "<nil>"
}

func (v *ValueSpec) Value() any {
	if v.value == nil {
		return nil
	}
	switch v.value.Kind() {
	case constant.Int:
		if i, exactly := constant.Int64Val(v.value); exactly {
			return i
		}
		u, _ := constant.Uint64Val(v.value)
		return u
	case constant.Float:
		if f, exactly := constant.Float32Val(v.value); exactly {
			return f
		}
		f, _ := constant.Float64Val(v.value)
		return f
	case constant.Bool:
		return constant.BoolVal(v.value)
	case constant.String:
		return constant.StringVal(v.value)
	}
	return nil
}

func (v *ValueSpec) resolve(ctx *Context, file *File, scope Scope) {
	v.Annotations = ctx.resolveAnnotationGroup(file, v.unresolved.annotations)
	v.resolveValue(ctx, file, scope, make([]*ValueSpec, 0, 16))
}

func (v *ValueSpec) resolveValue(ctx *Context, file *File, scope Scope, refs []*ValueSpec) constant.Value {
	// If value already resolved, return it
	if v.value != nil {
		return v.value
	}

	// If enum type is nil, resolve constant value expression in which iota is not allowed
	if v.enum.typ == nil {
		v.value = ctx.recursiveResolveValue(file, scope, append(refs, v), v.unresolved.value, nil)
		return v.value
	}

	if v.unresolved.value != nil {
		// Resolve value expression
		v.value = ctx.recursiveResolveValue(file, v.enum.typ, append(refs, v), v.unresolved.value, &v.enum.iota)
	} else if v.enum.index == 0 {
		// First member of enum type has value 0 if not specified
		v.value = constant.MakeInt64(0)
	} else {
		// Resolve previous value
		prev := v.enum.typ.Members[v.enum.index-1]
		prev.resolveValue(ctx, file, v.enum.typ, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.typ.Members[v.enum.index-v.enum.iota.value]

		if start.value != nil && start.enum.iota.found {
			// If start value is specified and it has iota expression, resolve it with the current iota value
			v.value = ctx.recursiveResolveValue(file, v.enum.typ, append(refs, v), start.unresolved.value, &v.enum.iota)
		} else {
			// Otherwise, add 1 to the previous value
			v.value = constant.BinaryOp(prev.value, token.ADD, constant.MakeInt64(1))
		}
	}
	return v.value
}

func newTypeSpec(ctx *Context, file *File, decl *Decl, src *ast.TypeSpec) Spec {
	switch t := src.Type.(type) {
	case *ast.EnumType:
		return newEnumType(ctx, file, decl, src, t)
	case *ast.StructType:
		return newStructType(ctx, file, decl, src, t)
	default:
		panic("unexpected type")
	}
}

// Field represents a struct field.
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

// StructType represents a struct type.
type StructType struct {
	typ
	decl       *Decl
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Fields       []*Field
}

func (*StructType) typeNode() {}

func (*StructType) Kind() Kind     { return Struct }
func (*StructType) IsStruct() bool { return true }

func (s *StructType) String() string { return s.Name }

func (s *StructType) Decl() *Decl { return s.decl }

func newStructType(ctx *Context, _ *File, decl *Decl, src *ast.TypeSpec, t *ast.StructType) Spec {
	s := &StructType{
		typ:     typ{pos: src.Pos()},
		decl:    decl,
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	s.unresolved.annotations = src.Annotations
	for _, f := range t.Fields.List {
		s.Fields = append(s.Fields, newField(ctx, f))
	}
	return s
}

func (s *StructType) resolve(ctx *Context, file *File, scope Scope) {
	s.Annotations = ctx.resolveAnnotationGroup(file, s.unresolved.annotations)
	for _, f := range s.Fields {
		f.resolve(ctx, file, scope)
	}
}

// EnumType represents an enumeration type declaration.
type EnumType struct {
	typ
	unresolved struct {
		annotations *ast.AnnotationGroup
	}
	file *File
	decl *Decl

	Doc, Comment CommentGroup
	Annotations  AnnotationGroup
	Name         string
	Members      []*ValueSpec
}

func newEnumType(ctx *Context, file *File, decl *Decl, src *ast.TypeSpec, t *ast.EnumType) *EnumType {
	e := &EnumType{
		typ:     typ{pos: src.Pos()},
		file:    file,
		decl:    decl,
		Doc:     newCommentGroup(src.Doc),
		Comment: newCommentGroup(src.Comment),
		Name:    src.Name.Name,
	}
	e.unresolved.annotations = src.Annotations
	for i, v := range t.Values {
		m := newValueSpec(ctx, file, decl, v)
		m.enum.typ = e
		m.enum.index = i
		e.Members = append(e.Members, m)
	}
	return e
}

func (e *EnumType) Decl() *Decl { return e.decl }

func (e *EnumType) resolve(ctx *Context, file *File, scope Scope) {
	e.Annotations = ctx.resolveAnnotationGroup(file, e.unresolved.annotations)
	for _, m := range e.Members {
		m.resolve(ctx, file, e)
	}
}

func (*EnumType) typeNode() {}

func (e *EnumType) String() string { return e.Name }

func (*EnumType) Pos() token.Pos { return 0 }

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
