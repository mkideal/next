package types

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/gopherd/core/op"
	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/templateutil"
	"github.com/next/next/src/token"
)

// Imports holds a list of imports.
// @api(object/Imports)
type Imports struct {
	// File is the file containing the imports.
	File *File
	// List is the list of imports.
	// @api(object/Imports/List)
	List []*Import
}

func (i *Imports) resolve(ctx *Context, file *File) {
	for _, spec := range i.List {
		spec.target = ctx.lookupFile(file.Path, spec.Path)
		if spec.target == nil {
			ctx.addErrorf(spec.pos, "import file not found: %s", spec.Path)
		}
	}
}

// TrimmedList returns a list of unique imports sorted by package name.
// @api(object/Imports/TrimmedList)
func (i *Imports) TrimmedList() []*Import {
	var seen = make(map[string]bool)
	var pkgs []*Import
	for _, spec := range i.List {
		if seen[spec.target.pkg.name] {
			continue
		}
		seen[spec.target.pkg.name] = true
		pkgs = append(pkgs, spec)
	}
	slices.SortFunc(pkgs, func(a, b *Import) int {
		return cmp.Compare(a.target.pkg.name, b.target.pkg.name)
	})
	return pkgs
}

// Import represents a file import.
// @api(object/Import)
type Import struct {
	pos    token.Pos // position of the import declaration
	target *File     // imported file
	file   *File     // file containing the import

	// Doc is the [documentation](#Object.Doc) for the import declaration.
	// @api(object/Import/Doc)
	Doc *Doc

	// Comment is the [line comment](#Object.Comment) of the import declaration.
	// @api(object/import/Comment)
	Comment *Comment

	// Path is the import path.
	// @api(object/import/Path)
	Path string
}

func newImport(ctx *Context, file *File, src *ast.ImportDecl) *Import {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		ctx.addErrorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
		path = "!BAD-IMPORT-PATH!"
	}
	i := &Import{
		pos:     src.Pos(),
		file:    file,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Path:    path,
	}
	return i
}

// Target returns the imported file.
// @api(object/Import/Target
func (i *Import) Target() *File { return i.target }

// File returns the file containing the import.
// @api(object/Import/File)
func (i *Import) File() *File { return i.target }

func (i *Import) resolve(ctx *Context, file *File, _ Scope) {}

// Decl represents an decl node.
// @api(object/decl/Decl)
type Decl interface {
	Object

	// annotations returns the annotations for the declaration.
	annotations() AnnotationGroup

	// File returns the file containing the declaration.
	// @api(object/decl/File)
	File() *File

	// Package returns the package containing the declaration.
	// It's a shortcut for decl.File().Package().
	// @api(object/decl/Package)
	Package() *Package
}

func (d *PrimitiveType) Package() *Package        { return nil }
func (d *ArrayType) Package() *Package            { return nil }
func (d *VectorType) Package() *Package           { return nil }
func (d *MapType) Package() *Package              { return nil }
func (d *Const) Package() *Package                { return d.file.pkg }
func (d *Enum) Package() *Package                 { return d.file.pkg }
func (d *EnumMember) Package() *Package           { return d.file.pkg }
func (d *Struct) Package() *Package               { return d.file.pkg }
func (d *StructField) Package() *Package          { return d.file.pkg }
func (d *Interface) Package() *Package            { return d.file.pkg }
func (d *InterfaceMethod) Package() *Package      { return d.file.pkg }
func (d *InterfaceMethodParam) Package() *Package { return d.File().Package() }

func (d *PrimitiveType) annotations() AnnotationGroup { return nil }
func (d *ArrayType) annotations() AnnotationGroup     { return nil }
func (d *VectorType) annotations() AnnotationGroup    { return nil }
func (d *MapType) annotations() AnnotationGroup       { return nil }
func (d *File) annotations() AnnotationGroup          { return d.Annotations }

var _ Decl = (*PrimitiveType)(nil)
var _ Decl = (*ArrayType)(nil)
var _ Decl = (*VectorType)(nil)
var _ Decl = (*MapType)(nil)
var _ Decl = (*File)(nil)
var _ Decl = (*Const)(nil)
var _ Decl = (*Enum)(nil)
var _ Decl = (*EnumMember)(nil)
var _ Decl = (*Struct)(nil)
var _ Decl = (*StructField)(nil)
var _ Decl = (*Interface)(nil)
var _ Decl = (*InterfaceMethod)(nil)

func (*PrimitiveType) File() *File { return nil }
func (*ArrayType) File() *File     { return nil }
func (*VectorType) File() *File    { return nil }
func (*MapType) File() *File       { return nil }

// List represents a list of objects.
// @api(object/decl/List)
type List[T Node] []T

// List returns the list of objects. It is used to provide a uniform way to access.
// @api(object/decl/List/List)
func (l List[T]) List() []T {
	return l
}

// Decls holds all declarations in a file.
// @api(object/decl/Decls)
type Decls struct {
	consts     List[*Const]
	enums      List[*Enum]
	structs    List[*Struct]
	interfaces List[*Interface]

	lang string
}

func (d *Decls) resolve(ctx *Context, file *File) {
	if d == nil {
		return
	}
	for _, c := range d.consts {
		c.resolve(ctx, file, file)
	}
	for _, e := range d.enums {
		e.resolve(ctx, file, file)
	}
	for _, s := range d.structs {
		s.resolve(ctx, file, file)
	}
	for _, i := range d.interfaces {
		i.resolve(ctx, file, file)
	}
}

// Consts returns the list of constant declarations.
// @api(object/decl/Consts)
func (d *Decls) Consts() List[*Const] {
	if d == nil {
		return nil
	}
	return availableList(d.consts, d.lang)
}

// Enums returns the list of enum declarations.
// @api(object/decl/Enums)
func (d *Decls) Enums() List[*Enum] {
	if d == nil {
		return nil
	}
	return availableList(d.enums, d.lang)
}

// Structs returns the list of struct declarations.
// @api(object/decl/Structs)
func (d *Decls) Structs() List[*Struct] {
	if d == nil {
		return nil
	}
	return availableList(d.structs, d.lang)
}

// Interfaces returns the list of interface declarations.
// @api(object/decl/Interfaces)
func (d *Decls) Interfaces() List[*Interface] {
	if d == nil {
		return nil
	}
	return availableList(d.interfaces, d.lang)
}

// decl represents a declaration header (const, enum, struct, interface).
type decl[Self Decl, Name ~string] struct {
	self Self      // self represents the declaration object itself
	pos  token.Pos // position of the declaration
	name Name      // name of the declaration
	file *File     // file containing the declaration

	// unresolved is the unresolved declaration data. It is resolved in the
	// resolve method.
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	// Doc is the [documentation](#Object.Doc) for the declaration.
	// @api(object/decl/Doc)
	Doc *Doc

	// Annotations is the [annotations](#Object.Annotations) for the declaration.
	// @api(object/decl/Annotations)
	Annotations AnnotationGroup
}

func newDecl[Self Decl, Name ~string](
	self Self, file *File,
	pos token.Pos, name Name,
	doc *ast.CommentGroup, annotations *ast.AnnotationGroup,
) *decl[Self, Name] {
	d := &decl[Self, Name]{
		self: self,
		pos:  pos,
		name: Name(name),
		file: file,
		Doc:  newDoc(doc),
	}
	d.unresolved.annotations = annotations
	return d
}

// Name returns the name of the declaration.
// @api(object/decl/Name)
func (d *decl[Self, Name]) Name() Name {
	return d.name
}

// File returns the file containing the declaration.
// @api(object/decl/File)
func (d *decl[Self, Name]) File() *File {
	if d == nil {
		return nil
	}
	return d.file
}

func (d *decl[Self, Name]) resolve(ctx *Context, file *File, scope Scope) {
	d.Annotations = ctx.resolveAnnotationGroup(file, d.self, d.unresolved.annotations)
}

func (d *decl[Self, Name]) annotations() AnnotationGroup {
	return d.Annotations
}

// iotaValue represents the iota value of an enum member.
type iotaValue struct {
	value int
	found bool
}

// Value represents a constant value.
// @api(object/decl/Value)
type Value struct {
	namePos    token.Pos
	name       string
	typ        *PrimitiveType
	val        constant.Value
	unresolved struct {
		value ast.Expr
	}
	enum struct {
		typ   *Enum     // parent enum
		index int       // index in the enum type. start from 0
		iota  iotaValue // iota value
	}
}

func newValue(_ *Context, _ *File, name string, namePos token.Pos, src ast.Expr) *Value {
	v := &Value{
		namePos: namePos,
		name:    name,
	}
	v.unresolved.value = src
	return v
}

func (v *Value) resolve(ctx *Context, file *File, scope Scope) {
	v.val = v.resolveValue(ctx, file, scope, make([]*Value, 0, 16))
	switch v.Any().(type) {
	case int64:
		v.typ = primitiveTypes["int64"]
	case float32:
		v.typ = primitiveTypes["float32"]
	case float64:
		v.typ = primitiveTypes["float64"]
	case bool:
		v.typ = primitiveTypes["bool"]
	case string:
		v.typ = primitiveTypes["string"]
	default:
		ctx.addErrorf(v.namePos, "unexpected constant type: %T", v.Any())
	}
}

func (v *Value) resolveValue(ctx *Context, file *File, scope Scope, refs []*Value) constant.Value {
	// If value already resolved, return it
	if v.val != nil {
		return v.val
	}

	// If enum type is nil, resolve constant value expression in which iota is not allowed
	if v.enum.typ == nil {
		v.val = ctx.recursiveResolveValue(file, scope, append(refs, v), v.unresolved.value, nil)
		return v.val
	}

	if v.unresolved.value != nil {
		// Resolve value expression
		v.val = ctx.recursiveResolveValue(file, v.enum.typ, append(refs, v), v.unresolved.value, &v.enum.iota)
	} else if v.enum.index == 0 {
		// First member of enum type has value 0 if not specified
		v.val = constant.MakeInt64(0)
	} else {
		// Resolve previous value
		prev := v.enum.typ.Members.List[v.enum.index-1]
		prev.value.resolveValue(ctx, file, v.enum.typ, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.value.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.typ.Members.List[v.enum.index-v.enum.iota.value]

		if start.value.val != nil && start.value.enum.iota.found {
			// If start value is specified and it has iota expression, resolve it with the current iota value
			v.val = ctx.recursiveResolveValue(file, v.enum.typ, append(refs, v), start.value.unresolved.value, &v.enum.iota)
		} else {
			// Otherwise, add 1 to the previous value
			v.val = constant.BinaryOp(prev.value.val, token.ADD, constant.MakeInt64(1))
		}
	}
	return v.val
}

// Type returns the type of the constant value.
// @api(object/decl/Value/Type)
func (v *Value) Type() *PrimitiveType {
	return v.typ
}

// String returns the string representation of the constant value.
// @api(object/decl/Value/String)
func (v *Value) String() string {
	return v.val.String()
}

// Underlying returns the underlying value of the constant or enum member.
// @api(object/decl/Value/Any)
func (v *Value) Any() any {
	if v.val == nil {
		return nil
	}
	switch v.val.Kind() {
	case constant.Int:
		if i, exactly := constant.Int64Val(v.val); exactly {
			return i
		}
	case constant.Float:
		if f, exactly := constant.Float32Val(v.val); exactly {
			return f
		}
		f, _ := constant.Float64Val(v.val)
		return f
	case constant.Bool:
		return constant.BoolVal(v.val)
	case constant.String:
		return constant.StringVal(v.val)
	}
	return nil
}

// Const represents a constant declaration.
// @api(object/decl/Const)
type Const struct {
	*decl[*Const, ConstName]

	value *Value // Value is the constant value.

	// Comment is the [line comment](#Object.Comment) of the constant declaration.
	// @api(object/decl/Const/Comment)
	Comment *Comment
}

// ConstName represents a constant name.
// @api(object/decl/ConstName)
type ConstName string

func newConst(ctx *Context, file *File, src *ast.GenDecl[ast.Expr]) *Const {
	c := &Const{
		Comment: newComment(src.Comment),
	}
	c.decl = newDecl(c, file, src.Pos(), ConstName(src.Name.Name), src.Doc, src.Annotations)
	c.value = newValue(ctx, file, string(c.name), src.Name.NamePos, src.Spec)
	return c
}

func (c *Const) resolve(ctx *Context, file *File, scope Scope) {
	c.decl.resolve(ctx, file, scope)
	c.value.resolve(ctx, file, scope)
}

// Type returns the type of the constant value.
// @api(object/decl/Const/Type)
func (c *Const) Type() *PrimitiveType {
	return c.value.Type()
}

// Value returns the constant value.
// @api(object/decl/Const/Value)
func (c *Const) Value() *Value {
	return c.value
}

// Fields represents a list of fields.
// @api(object/decl/Fields)
type Fields[D Object, F Node] struct {
	// typename is the type name of the fields:
	// - enum.members
	// - struct.fields
	// - interface.methods
	typename string

	// Decl is the declaration object that contains the fields.
	// Decl may be an enum, struct, or interface.
	// @api(object/decl/Fields/Decl)
	Decl D

	// List is the list of fields in the declaration.
	// @api(object/decl/Fields/List)
	List []F
}

// DeclType represents a declaration type.
// @api(object/decl/DeclType)
type DeclType[T Decl] struct {
	pos  token.Pos
	name string
	kind token.Kind

	// Decl is the declaration object that contains the type.
	// @api(object/decl/DeclType/Decl)
	decl T
}

func newDeclType[T Decl](pos token.Pos, name string, kind token.Kind, decl T) *DeclType[T] {
	return &DeclType[T]{pos: pos, name: name, kind: kind, decl: decl}
}

// String returns the string representation of the declaration type.
// @api(object/decl/DeclType/String)
func (d *DeclType[T]) String() string { return d.name }

// Enum represents an enum declaration.
// @api(object/decl/Enum)
type Enum struct {
	*decl[*Enum, EnumName]

	// Type is the enum type.
	// @api(object/decl/Enum/Type)
	Type *DeclType[*Enum]

	// Members is the list of enum members.
	// @api(object/decl/Enum/Members)
	Members *Fields[*Enum, *EnumMember]
}

// EnumName represents an enum name.
// @api(object/decl/EnumName)
type EnumName string

func newEnum(ctx *Context, file *File, src *ast.GenDecl[*ast.EnumType]) *Enum {
	e := &Enum{}
	e.decl = newDecl(e, file, src.Pos(), EnumName(src.Name.Name), src.Doc, src.Annotations)
	e.Type = newDeclType(src.Pos(), src.Name.Name, token.Enum, e)
	e.Members = &Fields[*Enum, *EnumMember]{typename: "enum.members", Decl: e}
	for i, m := range src.Spec.Members.List {
		e.Members.List = append(e.Members.List, newEnumMember(ctx, file, e, m, i))
	}
	return e
}

func (e *Enum) resolve(ctx *Context, file *File, scope Scope) {
	e.decl.resolve(ctx, file, scope)
	for _, m := range e.Members.List {
		m.resolve(ctx, file, scope)
	}
}

// EnumMember represents an enum member declaration.
// @api(object/decl/Enum/EnumMember)
type EnumMember struct {
	*decl[*EnumMember, EnumMemberName]

	// value is the enum member value.
	value *Value

	// Decl is the enum that contains the member.
	// @api(object/decl/Enum/EnumMember/Decl)
	Decl *Enum

	// Comment is the [line comment](#Object.Comment) of the enum member declaration.
	// @api(object/decl/Enum/EnumMember/Comment)
	Comment *Comment
}

// EnumMemberName represents an enum member name.
// @api(object/decl/Enum/EnumMemberName)
type EnumMemberName string

func newEnumMember(ctx *Context, file *File, e *Enum, src *ast.EnumMember, index int) *EnumMember {
	m := &EnumMember{
		Comment: newComment(src.Comment),
		Decl:    e,
	}
	m.decl = newDecl(m, file, src.Pos(), EnumMemberName(src.Name.Name), src.Doc, src.Annotations)
	m.value = newValue(ctx, file, string(m.name), src.Name.NamePos, src.Value)
	m.value.enum.typ = e
	m.value.enum.index = index
	return m
}

func (m *EnumMember) resolve(ctx *Context, file *File, scope Scope) {
	m.decl.resolve(ctx, file, scope)
	m.value.resolve(ctx, file, scope)
}

func (m *EnumMember) Value() *Value {
	return m.value
}

// IsFirst returns true if the value is the first member of the enum type.
// @api(object/decl/Value/IsFirst)
func (m *EnumMember) IsFirst() bool {
	return m.value.enum.typ != nil && m.value.enum.index == 0
}

// IsLast returns true if the value is the last member of the enum type.
// @api(object/decl/Value/IsLast)
func (m *EnumMember) IsLast() bool {
	return m.value.enum.typ != nil && m.value.enum.index == len(m.value.enum.typ.Members.List)-1
}

// Struct represents a struct declaration.
// @api(object/decl/Struct)
type Struct struct {
	*decl[*Struct, StructName]

	// lang is the current language to generate the struct.
	lang string

	// fields is the list of struct fields.
	fields *Fields[*Struct, *StructField]

	// Type is the struct type.
	// @api(object/decl/Struct/Type)
	Type *DeclType[*Struct]
}

// StructName represents a struct name.
// @api(object/decl/StructName)
type StructName string

func newStruct(ctx *Context, file *File, src *ast.GenDecl[*ast.StructType]) *Struct {
	s := &Struct{}
	s.decl = newDecl(s, file, src.Pos(), StructName(src.Name.Name), src.Doc, src.Annotations)
	s.Type = newDeclType(src.Pos(), src.Name.Name, token.Struct, s)
	s.fields = &Fields[*Struct, *StructField]{typename: "struct.fields", Decl: s}
	for _, f := range src.Spec.Fields.List {
		s.fields.List = append(s.fields.List, newStructField(ctx, file, s, f))
	}
	return s
}

func (s *Struct) resolve(ctx *Context, file *File, scope Scope) {
	s.decl.resolve(ctx, file, scope)
	for _, f := range s.fields.List {
		f.resolve(ctx, file, scope)
	}
}

// Fields returns the list of struct fields.
func (s *Struct) Fields() *Fields[*Struct, *StructField] {
	return availableFields(s.fields, s.lang)
}

// Field represents a struct field declaration.
// @api(object/decl/StructField)
type StructField struct {
	*decl[*StructField, StructFieldName]

	// Decl is the struct that contains the field.
	// @api(object/decl/StructField/Decl)
	Decl *Struct

	// Type is the field type.
	// @api(object/decl/StructField/Type)
	Type *StructFieldType

	// Comment is the [line comment](#Object.Comment) of the struct field declaration.
	// @api(object/decl/StructField/Comment)
	Comment *Comment
}

func newStructField(ctx *Context, file *File, s *Struct, src *ast.StructField) *StructField {
	f := &StructField{Decl: s}
	f.decl = newDecl(f, file, src.Pos(), StructFieldName(src.Name.Name), src.Doc, src.Annotations)
	f.Type = newStructFieldType(ctx, file, f, src.Type)
	return f
}

func (f *StructField) resolve(ctx *Context, file *File, scope Scope) {
	f.decl.resolve(ctx, file, scope)
	f.Type.resolve(ctx, file, scope)
}

// StructFieldName represents a struct field name.
// @api(object/decl/StructFieldName)
type StructFieldName string

// StructFieldType represents a struct field type.
// @api(object/decl/StructFieldType)
type StructFieldType struct {
	unresolved struct {
		typ ast.Type
	}

	// Type is the field type.
	// @api(object/decl/StructFieldType/Type)
	Type Type

	// Field is the struct field that contains the type.
	// @api(object/decl/StructFieldType/Field)
	Field *StructField
}

func newStructFieldType(_ *Context, _ *File, f *StructField, src ast.Type) *StructFieldType {
	t := &StructFieldType{Field: f}
	t.unresolved.typ = src
	return t
}

func (t *StructFieldType) resolve(ctx *Context, file *File, scope Scope) {
	t.Type = ctx.resolveType(file, t.unresolved.typ)
}

// Interface represents an interface declaration.
// @api(object/decl/Interface)
type Interface struct {
	*decl[*Interface, InterfaceName]

	// lang is the current language to generate the interface.
	lang string

	// methods is the list of interface methods.
	methods *Fields[*Interface, *InterfaceMethod]

	// Type is the interface type.
	// @api(object/decl/Interface/Type)
	Type *DeclType[*Interface]
}

// InterfaceName represents an interface name.
// @api(object/decl/InterfaceName)
type InterfaceName string

func newInterface(ctx *Context, file *File, src *ast.GenDecl[*ast.InterfaceType]) *Interface {
	i := &Interface{}
	i.decl = newDecl(i, file, src.Pos(), InterfaceName(src.Name.Name), src.Doc, src.Annotations)
	i.Type = newDeclType(src.Pos(), src.Name.Name, token.Interface, i)
	i.methods = &Fields[*Interface, *InterfaceMethod]{typename: "interface.methods", Decl: i}
	for _, m := range src.Spec.Methods.List {
		i.methods.List = append(i.methods.List, newInterfaceMethod(ctx, file, i, m))
	}
	return i
}

func (i *Interface) resolve(ctx *Context, file *File, scope Scope) {
	i.decl.resolve(ctx, file, scope)
	for _, m := range i.methods.List {
		m.resolve(ctx, file, scope)
	}
}

// Methods returns the list of interface methods.
// @api(object/decl/Interface/Methods)
func (i *Interface) Methods() *Fields[*Interface, *InterfaceMethod] {
	return availableFields(i.methods, i.lang)
}

// InterfaceMethod represents an interface method declaration.
// @api(object/decl/InterfaceMethod)
type InterfaceMethod struct {
	*decl[*InterfaceMethod, InterfaceMethodName]

	// Decl is the interface that contains the method.
	// @api(object/decl/InterfaceMethod/Decl)
	Decl *Interface

	// Params is the list of method parameters.
	// @api(object/decl/InterfaceMethod/Params)
	Params *Fields[*InterfaceMethod, *InterfaceMethodParam]

	// Return is the method return type.
	// @api(object/decl/InterfaceMethod/Return)
	Return *InterfaceMethodReturn

	// Comment is the [line comment](#Object.Comment) of the interface method declaration.
	// @api(object/decl/InterfaceMethod/Comment)
	Comment *Comment
}

func newInterfaceMethod(ctx *Context, file *File, i *Interface, src *ast.Method) *InterfaceMethod {
	m := &InterfaceMethod{
		Decl:    i,
		Comment: newComment(src.Comment),
	}
	m.decl = newDecl(m, file, src.Pos(), InterfaceMethodName(src.Name.Name), src.Doc, src.Annotations)
	m.Params = &Fields[*InterfaceMethod, *InterfaceMethodParam]{typename: "interface.method.params", Decl: m}
	for _, p := range src.Params.List {
		m.Params.List = append(m.Params.List, newInterfaceMethodParam(ctx, file, m, p))
	}
	m.Return = newInterfaceMethodReturn(ctx, file, m, src.Return)
	return m
}

func (m *InterfaceMethod) resolve(ctx *Context, file *File, scope Scope) {
	m.decl.resolve(ctx, file, scope)
	for _, p := range m.Params.List {
		p.resolve(ctx, file, scope)
	}
	if m.Return != nil {
		m.Return.resolve(ctx, file, scope)
	}
}

// InterfaceMethodName represents an interface method name.
// @api(object/decl/InterfaceMethodName)
type InterfaceMethodName string

// InterfaceMethodParam represents an interface method parameter declaration.
// @api(object/decl/InterfaceMethodParam)
type InterfaceMethodParam struct {
	pos        token.Pos
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	// Annotations is the [annotations](#Object.Annotations) for the parameter.
	// @api(object/decl/InterfaceMethodParam/Annotations)
	Annotations AnnotationGroup

	// Method is the interface method that contains the parameter.
	// @api(object/decl/InterfaceMethodParam/Method)
	Method *InterfaceMethod

	// Name is the parameter name.
	// @api(object/decl/InterfaceMethodParam/Name)
	Name InterfaceMethodParamName

	// Type is the parameter type.
	// @api(object/decl/InterfaceMethodParam/Type)
	Type *InterfaceMethodParamType
}

func newInterfaceMethodParam(ctx *Context, file *File, m *InterfaceMethod, src *ast.MethodParam) *InterfaceMethodParam {
	p := &InterfaceMethodParam{
		pos:    src.Pos(),
		Method: m,
	}
	p.unresolved.annotations = src.Annotations
	p.Name = InterfaceMethodParamName(src.Name.Name)
	p.Type = newInterfaceMethodParamType(ctx, file, p, src.Type)
	return p
}

func (p *InterfaceMethodParam) resolve(ctx *Context, file *File, scope Scope) {
	p.Annotations = ctx.resolveAnnotationGroup(file, p, p.unresolved.annotations)
	p.Type.resolve(ctx, file, scope)
}

func (p *InterfaceMethodParam) File() *File {
	return p.Method.File()
}

func (p *InterfaceMethodParam) annotations() AnnotationGroup {
	return p.Annotations
}

// InterfaceMethodParamName represents an interface method parameter name.
// @api(object/decl/InterfaceMethodParamName)
type InterfaceMethodParamName string

// InterfaceMethodParamType represents an interface method parameter type.
// @api(object/decl/InterfaceMethodParamType)
type InterfaceMethodParamType struct {
	unresolved struct {
		typ ast.Type
	}

	// Param is the interface method parameter that contains the type.
	// @api(object/decl/InterfaceMethodParamType/Param)
	Param *InterfaceMethodParam

	// Type is the parameter type.
	// @api(object/decl/InterfaceMethodParamType/Type)
	Type Type
}

func newInterfaceMethodParamType(_ *Context, _ *File, p *InterfaceMethodParam, src ast.Type) *InterfaceMethodParamType {
	t := &InterfaceMethodParamType{Param: p}
	t.unresolved.typ = src
	return t
}

func (t *InterfaceMethodParamType) resolve(ctx *Context, file *File, scope Scope) {
	t.Type = ctx.resolveType(file, t.unresolved.typ)
}

// InterfaceMethodReturn represents an interface method return type.
// @api(object/decl/InterfaceMethodReturn)
type InterfaceMethodReturn struct {
	unresolved struct {
		typ ast.Type
	}

	// Method is the interface method that contains the return type.
	// @api(object/decl/InterfaceMethodReturn/Method)
	Method *InterfaceMethod

	// Type is the return type.
	// @api(object/decl/InterfaceMethodReturn/Type)
	Type Type
}

func newInterfaceMethodReturn(_ *Context, _ *File, m *InterfaceMethod, src ast.Type) *InterfaceMethodReturn {
	t := &InterfaceMethodReturn{Method: m}
	t.unresolved.typ = src
	return t
}

func (t *InterfaceMethodReturn) resolve(ctx *Context, file *File, scope Scope) {
	if t.unresolved.typ == nil {
		return
	}
	t.Type = ctx.resolveType(file, t.unresolved.typ)
}

// isAvailable reports whether the declaration is available in the current language.
func isAvailable(decl Decl, lang string) bool {
	s, ok := decl.annotations().get("next").get("available").Value().(string)
	return !ok || templateutil.ContainsWord(s, lang)
}

// available returns the declaration if it is available in the current language.
func available[T Decl](decl T, lang string) (T, bool) {
	op.Assertf(lang != "", "language must not be empty")
	if !isAvailable(decl, lang) {
		return decl, false
	}
	switch decl := any(decl).(type) {
	case *File:
		decl.decls.lang = lang
	case *Struct:
		decl.lang = lang
	case *Interface:
		decl.lang = lang
	}
	return decl, true
}

// availableFields returns the list of fields that are available in the current language.
func availableFields[D, F Decl](fields *Fields[D, F], lang string) *Fields[D, F] {
	op.Assertf(lang != "", "language must not be empty")
	for i, f := range fields.List {
		if isAvailable(f, lang) {
			continue
		}
		list := make([]F, 0, len(fields.List))
		list = append(list, fields.List[:i]...)
		for j := i + 1; j < len(fields.List); j++ {
			if isAvailable(fields.List[j], lang) {
				list = append(list, fields.List[j])
			}
		}
		return &Fields[D, F]{typename: fields.typename, Decl: fields.Decl, List: list}
	}
	return fields
}

// availableList returns the list of declarations that are available in the current language.
func availableList[T Decl](list List[T], lang string) List[T] {
	op.Assertf(lang != "", "language must not be empty")
	availables := make([]T, 0, len(list))
	for i, d := range list {
		if _, ok := available(d, lang); ok {
			availables = append(availables, list[i])
		}
	}
	return availables
}
