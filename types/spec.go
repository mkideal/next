package types

import (
	"strconv"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

type Specs struct {
	File    *File
	Consts  *ConstSpecs
	Enums   *EnumSpecs
	Structs *StructSpecs
}

type ConstSpecs struct {
	File *File
	List []*ValueSpec
}

type EnumSpecs struct {
	File *File
	List []*EnumSpec
}

type StructSpecs struct {
	File *File
	List []*StructSpec
}

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

	Doc         *Doc
	Comment     *Comment
	Annotations *AnnotationGroup
	Path        string
}

func newImportSpec(ctx *Context, _ *File, decl *Decl, src *ast.ImportSpec) *ImportSpec {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		ctx.addErrorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
		path = "!BAD-IMPORT-PATH!"
	}
	i := &ImportSpec{
		pos:     src.Pos(),
		decl:    decl,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Path:    path,
	}
	i.unresolved.annotations = src.Annotations
	return i
}

func (i *ImportSpec) String() string { return i.Path }

func (i *ImportSpec) File() *File { return i.importedFile }

func (i *ImportSpec) Decl() *Decl { return i.decl }

func (i *ImportSpec) resolve(ctx *Context, file *File, _ Scope) {
	i.Annotations = ctx.resolveAnnotationGroup(file, i.unresolved.annotations)
	if len(i.decl.Specs) == 1 {
		i.Doc = i.decl.Doc
		i.decl.Doc = nil
		i.Annotations = i.decl.Annotations
		i.decl.Annotations = nil
	}
}

type iotaValue struct {
	value int
	found bool
}

// ValueSpec represents an constant or enum member.
type ValueSpec struct {
	pos   token.Pos
	decl  *Decl
	name  string
	value constant.Value
	enum  struct {
		typ   *EnumSpec // parent enum type
		index int       // index in the enum type. start from 0
		iota  iotaValue // iota value
	}
	unresolved struct {
		annotations *ast.AnnotationGroup
		value       ast.Expr
	}

	Doc         *Doc
	Comment     *Comment
	Annotations *AnnotationGroup
}

func newValueSpec(_ *Context, _ *File, decl *Decl, src *ast.ValueSpec) *ValueSpec {
	v := &ValueSpec{
		pos:     src.Pos(),
		decl:    decl,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		name:    src.Name.Name,
	}
	v.unresolved.annotations = src.Annotations
	v.unresolved.value = src.Value
	return v
}

func (v *ValueSpec) Decl() *Decl { return v.decl }

func (v *ValueSpec) String() string {
	if v.value != nil {
		return v.name + "=" + v.value.String()
	}
	return v.name + "=?"
}

// @api(template/object) ValueSpec.Enum
func (v *ValueSpec) Enum() *EnumSpec {
	return v.enum.typ
}

// @api(template/object) ValueSpec.Value
func (v *ValueSpec) Value() constant.Value {
	return v.value
}

// @api(template/object) ValueSpec.Underlying
// Underlying returns the underlying value of the constant or enum member.
func (v *ValueSpec) Underlying() any {
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
	if v.enum.typ == nil && len(v.decl.Specs) == 1 {
		v.Doc = v.decl.Doc
		v.decl.Doc = nil
		v.Annotations = v.decl.Annotations
		v.decl.Annotations = nil
	}
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
		prev := v.enum.typ.Members.List[v.enum.index-1]
		prev.resolveValue(ctx, file, v.enum.typ, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.typ.Members.List[v.enum.index-v.enum.iota.value]

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
		return newEnumSpec(ctx, file, decl, src, t)
	case *ast.StructType:
		return newStructSpec(ctx, file, decl, src, t)
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

	Doc         *Doc
	Comment     *Comment
	Annotations *AnnotationGroup
	Name        string
	Type        Type
}

func newField(_ *Context, src *ast.Field) *Field {
	f := &Field{
		pos:     src.Pos(),
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Name:    src.Name.Name,
	}
	f.unresolved.annotations = src.Annotations
	f.unresolved.typ = src.Type
	return f
}

func (f *Field) resolve(ctx *Context, file *File, _ Scope) {
	f.Annotations = ctx.resolveAnnotationGroup(file, f.unresolved.annotations)
	f.Type = ctx.resolveType(file, f.unresolved.typ)
}

// Fields represents a list of struct fields.
type Fields struct {
	Struct *StructSpec
	List   []*Field
}

// StructSpec represents a struct type.
type StructSpec struct {
	decl       *Decl
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Type        *StructType
	Doc         *Doc
	Comment     *Comment
	Annotations *AnnotationGroup
	Fields      *Fields
}

func newStructSpec(ctx *Context, _ *File, decl *Decl, src *ast.TypeSpec, t *ast.StructType) Spec {
	s := &StructSpec{
		decl:    decl,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Type: &StructType{
			name: src.Name.Name,
			pos:  src.Pos(),
		},
	}
	s.unresolved.annotations = src.Annotations
	s.Fields = &Fields{Struct: s}
	for _, f := range t.Fields.List {
		s.Fields.List = append(s.Fields.List, newField(ctx, f))
	}
	return s
}

func (s *StructSpec) Decl() *Decl    { return s.decl }
func (s *StructSpec) String() string { return s.Type.name }

func (s *StructSpec) resolve(ctx *Context, file *File, scope Scope) {
	s.Annotations = ctx.resolveAnnotationGroup(file, s.unresolved.annotations)
	if len(s.decl.Specs) == 1 {
		s.Doc = s.decl.Doc
		s.decl.Doc = nil
		s.Annotations = s.decl.Annotations
		s.decl.Annotations = nil
	}
	for _, f := range s.Fields.List {
		f.resolve(ctx, file, scope)
	}
}

// EnumMembers represents a list of enum members.
type EnumMembers struct {
	Enum *EnumSpec
	List []*ValueSpec
}

// EnumSpec represents an enumeration type declaration.
type EnumSpec struct {
	decl       *Decl
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Type        *EnumType
	Doc         *Doc
	Comment     *Comment
	Annotations *AnnotationGroup
	Members     *EnumMembers
}

func newEnumSpec(ctx *Context, file *File, decl *Decl, src *ast.TypeSpec, t *ast.EnumType) *EnumSpec {
	e := &EnumSpec{
		decl:    decl,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Type: &EnumType{
			name: src.Name.Name,
			pos:  src.Pos(),
		},
	}
	e.Members = &EnumMembers{Enum: e}
	e.unresolved.annotations = src.Annotations
	for i, v := range t.Values {
		m := newValueSpec(ctx, file, decl, v)
		m.enum.typ = e
		m.enum.index = i
		e.Members.List = append(e.Members.List, m)
	}
	return e
}

func (e *EnumSpec) Decl() *Decl { return e.decl }

func (e *EnumSpec) resolve(ctx *Context, file *File, scope Scope) {
	e.Annotations = ctx.resolveAnnotationGroup(file, e.unresolved.annotations)
	if len(e.decl.Specs) == 1 {
		e.Doc = e.decl.Doc
		e.decl.Doc = nil
		e.Annotations = e.decl.Annotations
		e.decl.Annotations = nil
	}
	for _, m := range e.Members.List {
		m.resolve(ctx, file, e)
	}
}

func (e *EnumSpec) String() string { return e.Type.name }
