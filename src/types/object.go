package types

import (
	"cmp"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/gopherd/core/op"
	"github.com/gopherd/core/text"
	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// @template(object/Imports)
// Imports holds a list of imports.
type Imports struct {
	// @template(object/Imports.File)
	// File represents the file containing the imports.
	File *File

	// @template(object/Imports.List)
	// List represents the list of [imports](#object/Import).
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

// @template(object/Imports.TrimmedList)
// TrimmedList represents a list of unique imports sorted by package name.
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

// @template(object/Import)
// Import represents a file import.
type Import struct {
	pos    token.Pos // position of the import declaration
	target *File     // imported file
	file   *File     // file containing the import

	// @template(object/Import.Doc)
	// Doc represents the import declaration [documentation](#Object.Doc).
	Doc *Doc

	// @template(object/import.Comment)
	// Comment represents the import declaration line [comment](#Object.Comment).
	Comment *Comment

	// @template(object/import.Path)
	// Path represents the import path.
	Path string

	// @template(object/Import.FullPath)
	// FullPath represents the full path of the import.
	FullPath string
}

func newImport(ctx *Context, file *File, src *ast.ImportDecl) *Import {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		ctx.addErrorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
		path = "!BAD-IMPORT-PATH!"
	}
	i := &Import{
		pos:      src.Pos(),
		file:     file,
		Doc:      newDoc(src.Doc),
		Comment:  newComment(src.Comment),
		Path:     path,
		FullPath: path,
	}
	if len(path) > 0 && path[0] != '/' {
		var err error
		path, err = filepath.Abs(filepath.Join(filepath.Dir(file.Path), path))
		if err != nil {
			ctx.addErrorf(token.NoPos, "failed to get absolute path of %s: %v", i.Path, err)
		} else {
			i.FullPath = path
		}
	}
	return i
}

// @template(object/Import.Target)
// Target represents the imported file.
func (i *Import) Target() *File { return i.target }

// @template(object/Import.File)
// File represents the file containing the import declaration.
func (i *Import) File() *File { return i.file }

func (i *Import) resolve(ctx *Context, file *File, _ Scope) {}

// @template(object/List)
// List represents a list of objects.
type List[T Object] []T

// @template(object/List.List)
// List represents the list of objects. It is used to provide a uniform way to access.
func (l List[T]) List() []T {
	return l
}

// @template(object/Decls)
// Decls holds all declarations in a file.
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

// @template(object/Decls.Consts)
// Consts represents the list of const declarations.
func (d *Decls) Consts() List[*Const] {
	if d == nil {
		return nil
	}
	return availableList(d.consts, d.lang)
}

// @template(object/Decls.Enums)
// Enums represents the list of enum declarations.
func (d *Decls) Enums() List[*Enum] {
	if d == nil {
		return nil
	}
	return availableList(d.enums, d.lang)
}

// @template(object/Decls.Structs)
// Structs represents the list of struct declarations.
func (d *Decls) Structs() List[*Struct] {
	if d == nil {
		return nil
	}
	return availableList(d.structs, d.lang)
}

// @template(object/Decls.Interfaces)
// Interfaces represents the list of interface declarations.
func (d *Decls) Interfaces() List[*Interface] {
	if d == nil {
		return nil
	}
	return availableList(d.interfaces, d.lang)
}

// @template(object/NodeName)
// NodeName represents a name of a field in a declaration:
// - Enum member name
// - Struct field name
// - Interface method name
// - Interface method parameter name
type NodeName[T Node] struct {
	pos   token.Pos
	name  string
	field T
}

func newNodeName[T Node](pos token.Pos, name string, field T) *NodeName[T] {
	return &NodeName[T]{pos: pos, name: name, field: field}
}

// @template(object/NodeName.Field)
// Field represents the field object that contains the name.
func (n *NodeName[T]) Field() T { return n.field }

// @template(object/NodeName.String)
// String represents the string representation of the field name.
func (n *NodeName[T]) String() string { return n.name }

// commonNode represents a common node.
type commonNode[Self Node] struct {
	self Self            // self represents the declaration object itself
	pos  token.Pos       // position of the declaration
	name *NodeName[Self] // name of the declaration
	file *File           // file containing the declaration

	// unresolved is the unresolved declaration data. It is resolved in the
	// resolve method.
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	doc         *Doc
	annotations Annotations
}

func newObject[Self Node](
	self Self, file *File,
	pos token.Pos, name string,
	doc *ast.CommentGroup, annotations *ast.AnnotationGroup,
) *commonNode[Self] {
	d := &commonNode[Self]{
		self: self,
		pos:  pos,
		name: newNodeName(pos, name, self),
		file: file,
		doc:  newDoc(doc),
	}
	d.unresolved.annotations = annotations
	return d
}

func (d *commonNode[Self]) resolve(ctx *Context, file *File, scope Scope) {
	d.annotations = ctx.resolveAnnotationGroup(file, d.self, d.unresolved.annotations)
}

func (d *commonNode[Self]) Doc() *Doc {
	return d.doc
}

func (d *commonNode[Self]) Annotations() Annotations {
	return d.annotations
}

// iotaValue represents the iota value of an enum member.
type iotaValue struct {
	value int
	found bool
}

// @template(object/Value)
// Value represents a constant value for a const declaration or an enum member.
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

func newValue(ctx *Context, file *File, name string, namePos token.Pos, src ast.Expr) *Value {
	v := &Value{
		namePos: namePos,
		name:    name,
	}
	if src != nil {
		file.addObject(ctx, src, v)
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

// @template(object/Value.IsEnum)
// IsEnum returns true if the value is an enum member.
func (v *Value) IsEnum() bool {
	return v.enum.typ != nil
}

// @template(object/Value.Type)
// Type represents the type of the value.
func (v *Value) Type() *PrimitiveType {
	return v.typ
}

// @template(object/Value.String)
// String represents the string representation of the value.
func (v *Value) String() string {
	return v.val.String()
}

// @template(object/Value.Any)
// Any represents the underlying value of the constant.
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

// @template(object/Const)
// Const represents a const declaration.
type Const struct {
	*commonNode[*Const]

	// value is the constant value.
	value *Value

	// @template(object/Const.Comment)
	// Comment is the line [comment](#object/Comment) of the constant declaration.
	Comment *Comment
}

func newConst(ctx *Context, file *File, src *ast.GenDecl[ast.Expr]) *Const {
	c := &Const{
		Comment: newComment(src.Comment),
	}
	file.addObject(ctx, src, c)
	c.commonNode = newObject(c, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	c.value = newValue(ctx, file, src.Name.Name, src.Name.NamePos, src.Spec)
	return c
}

func (c *Const) resolve(ctx *Context, file *File, scope Scope) {
	c.commonNode.resolve(ctx, file, scope)
	c.value.resolve(ctx, file, scope)
}

// @template(object/Const.Name)
// Name represents the [name object](#object.NodeName) of the constant.
func (c *Const) Name() *NodeName[*Const] {
	return c.name
}

// @template(object/Const.Type)
// Type represents the type of the constant.
func (c *Const) Type() *PrimitiveType {
	return c.value.Type()
}

// @template(object/Const.Value)
// Value represents the [value object](#object/Value) of the constant.
func (c *Const) Value() *Value {
	return c.value
}

// @template(object/Fields)
// Fields represents a list of fields in a declaration.
type Fields[D Node, F Object] struct {
	// @template(object/Fields.Decl)
	// Decl is the declaration object that contains the fields.
	// Decl may be an enum, struct, or interface.
	Decl D

	// @template(object/Fields.List)
	// List is the list of fields in the declaration.
	List []F
}

// @template(object/Enum)
// Enum represents an enum declaration.
type Enum struct {
	*commonNode[*Enum]

	// @template(object/Enum.Type)
	// Type is the enum type.
	Type *DeclType[*Enum]

	// @template(object/Enum.Members)
	// Members is the list of enum members.
	Members *Fields[*Enum, *EnumMember]
}

func newEnum(ctx *Context, file *File, src *ast.GenDecl[*ast.EnumType]) *Enum {
	e := &Enum{}
	file.addObject(ctx, src, e)
	e.commonNode = newObject(e, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	e.Type = newDeclType(src.Pos(), token.KindEnum, src.Name.Name, e)
	e.Members = &Fields[*Enum, *EnumMember]{Decl: e}
	for i, m := range src.Spec.Members.List {
		e.Members.List = append(e.Members.List, newEnumMember(ctx, file, e, m, i))
	}
	return e
}

func (e *Enum) resolve(ctx *Context, file *File, scope Scope) {
	e.commonNode.resolve(ctx, file, scope)
	for _, m := range e.Members.List {
		m.resolve(ctx, file, scope)
	}
}

// @template(object/EnumMember)
// EnumMember represents an enum member object in an [enum](#object/Enum) declaration.
type EnumMember struct {
	*commonNode[*EnumMember]

	// value is the enum member value.
	value *Value

	// @template(object/EnumMember.Decl)
	// Decl represents the [enum](#object/Enum) that contains the member.
	Decl *Enum

	// @template(object/EnumMember.Comment)
	// Comment represents the line [comment](#object/Comment) of the enum member declaration.
	Comment *Comment
}

func newEnumMember(ctx *Context, file *File, e *Enum, src *ast.EnumMember, index int) *EnumMember {
	m := &EnumMember{
		Comment: newComment(src.Comment),
		Decl:    e,
	}
	file.addObject(ctx, src, m)
	m.commonNode = newObject(m, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	m.value = newValue(ctx, file, src.Name.Name, src.Name.NamePos, src.Value)
	m.value.enum.typ = e
	m.value.enum.index = index
	return m
}

func (m *EnumMember) resolve(ctx *Context, file *File, scope Scope) {
	m.commonNode.resolve(ctx, file, scope)
	m.value.resolve(ctx, file, scope)
}

// @template(object/EnumMember.Name)
// Name represents the [name object](#object/NodeName) of the enum member.
func (m *EnumMember) Name() *NodeName[*EnumMember] {
	return m.name
}

// @template(object/EnumMember.Value)
// Value represents the [value object](#object/Value) of the enum member.
func (m *EnumMember) Value() *Value {
	return m.value
}

// @template(object/Value.IsFirst)
// IsFirst reports whether the value is the first member of the enum type.
func (m *EnumMember) IsFirst() bool {
	return m.value.enum.typ != nil && m.value.enum.index == 0
}

// @template(object/Value.IsLast)
// IsLast reports whether the value is the last member of the enum type.
func (m *EnumMember) IsLast() bool {
	return m.value.enum.typ != nil && m.value.enum.index == len(m.value.enum.typ.Members.List)-1
}

// @template(object/Struct)
// Struct represents a struct declaration.
type Struct struct {
	*commonNode[*Struct]

	// lang is the current language to generate the struct.
	lang string

	// fields is the list of struct fields.
	fields *Fields[*Struct, *StructField]

	// @template(object/Struct.Type)
	// Type represents the struct type.
	Type *DeclType[*Struct]
}

func newStruct(ctx *Context, file *File, src *ast.GenDecl[*ast.StructType]) *Struct {
	s := &Struct{}
	file.addObject(ctx, src, s)
	s.commonNode = newObject(s, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	s.Type = newDeclType(src.Pos(), token.KindStruct, src.Name.Name, s)
	s.fields = &Fields[*Struct, *StructField]{Decl: s}
	for _, f := range src.Spec.Fields.List {
		s.fields.List = append(s.fields.List, newStructField(ctx, file, s, f))
	}
	return s
}

func (s *Struct) resolve(ctx *Context, file *File, scope Scope) {
	s.commonNode.resolve(ctx, file, scope)
	for _, f := range s.fields.List {
		f.resolve(ctx, file, scope)
	}
}

// @template(object/Struct.Fields)
// Fields represents the list of struct fields.
func (s *Struct) Fields() *Fields[*Struct, *StructField] {
	return availableFields(s.fields, s.lang)
}

// @template(object/StructField)
// StructField represents a struct field declaration.
type StructField struct {
	*commonNode[*StructField]

	// @template(object/StructField.Decl)
	// Decl represents the struct that contains the field.
	Decl *Struct

	// @template(object/StructField.Type)
	// Type represents the [struct field type](#StructFieldType).
	Type *StructFieldType

	// @template(object/StructField.Comment)
	// Comment represents the line [comment](#object/Comment) of the struct field declaration.
	Comment *Comment
}

func newStructField(ctx *Context, file *File, s *Struct, src *ast.StructField) *StructField {
	f := &StructField{Decl: s}
	file.addObject(ctx, src, f)
	f.commonNode = newObject(f, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	f.Type = newStructFieldType(ctx, file, f, src.Type)
	return f
}

func (f *StructField) resolve(ctx *Context, file *File, scope Scope) {
	f.commonNode.resolve(ctx, file, scope)
	f.Type.resolve(ctx, file, scope)
}

// @template(object/StructField.Name)
// Name represents the [name object](#object/NodeName) of the struct field.
func (f *StructField) Name() *NodeName[*StructField] {
	return f.name
}

// @template(object/decl/StructFieldType)
// StructFieldType represents a struct field type.
type StructFieldType struct {
	unresolved struct {
		typ ast.Type
	}

	// @template(object/StructFieldType.Type)
	// Type represents the underlying type of the struct field.
	Type Type

	// @template(object/StructFieldType.Field)
	// Field represents the struct field that contains the type.
	Field *StructField
}

func newStructFieldType(_ *Context, _ *File, f *StructField, src ast.Type) *StructFieldType {
	t := &StructFieldType{Field: f}
	t.unresolved.typ = src
	return t
}

func (t *StructFieldType) resolve(ctx *Context, file *File, scope Scope) {
	t.Type = ctx.resolveType(file, t.unresolved.typ, false)
}

// @template(object/Interface)
// Interface represents an interface declaration.
type Interface struct {
	*commonNode[*Interface]

	// lang is the current language to generate the interface.
	lang string

	// methods is the list of interface methods.
	methods *Fields[*Interface, *InterfaceMethod]

	// @template(object/Interface.Type)
	// Type represents the interface type.
	Type *DeclType[*Interface]
}

func newInterface(ctx *Context, file *File, src *ast.GenDecl[*ast.InterfaceType]) *Interface {
	i := &Interface{}
	file.addObject(ctx, src, i)
	i.commonNode = newObject(i, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	i.Type = newDeclType(src.Pos(), token.KindInterface, src.Name.Name, i)
	i.methods = &Fields[*Interface, *InterfaceMethod]{Decl: i}
	for _, m := range src.Spec.Methods.List {
		i.methods.List = append(i.methods.List, newInterfaceMethod(ctx, file, i, m))
	}
	return i
}

func (i *Interface) resolve(ctx *Context, file *File, scope Scope) {
	i.commonNode.resolve(ctx, file, scope)
	for _, m := range i.methods.List {
		m.resolve(ctx, file, scope)
	}
}

// @template(object/Interface.Methods)
// Methods represents the list of interface methods.
func (i *Interface) Methods() *Fields[*Interface, *InterfaceMethod] {
	return availableFields(i.methods, i.lang)
}

// @template(object/InterfaceMethod)
// InterfaceMethod represents an interface method declaration.
type InterfaceMethod struct {
	*commonNode[*InterfaceMethod]

	// @template(object/InterfaceMethod.Decl)
	// Decl represents the interface that contains the method.
	Decl *Interface

	// @template(object/InterfaceMethod.Params)
	// Params represents the list of method parameters.
	Params *Fields[*InterfaceMethod, *InterfaceMethodParam]

	// @template(object/InterfaceMethod.Result)
	// Result represents the return type of the method.
	Result *InterfaceMethodResult

	// @template(object/InterfaceMethod.Comment)
	// Comment represents the line [comment](#object/Comment) of the interface method declaration.
	Comment *Comment
}

func newInterfaceMethod(ctx *Context, file *File, i *Interface, src *ast.Method) *InterfaceMethod {
	m := &InterfaceMethod{
		Decl:    i,
		Comment: newComment(src.Comment),
	}
	file.addObject(ctx, src, m)
	m.commonNode = newObject(m, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	m.Params = &Fields[*InterfaceMethod, *InterfaceMethodParam]{Decl: m}
	for _, p := range src.Params.List {
		m.Params.List = append(m.Params.List, newInterfaceMethodParam(ctx, file, m, p))
	}
	m.Result = newInterfaceMethodResult(ctx, file, m, src.Result)
	return m
}

func (m *InterfaceMethod) resolve(ctx *Context, file *File, scope Scope) {
	m.commonNode.resolve(ctx, file, scope)
	for _, p := range m.Params.List {
		p.resolve(ctx, file, scope)
	}
	if m.Result != nil {
		m.Result.resolve(ctx, file, scope)
	}
}

// @template(object/InterfaceMethod.Name)
// Name represents the [name object](#object/NodeName) of the interface method.
func (m *InterfaceMethod) Name() *NodeName[*InterfaceMethod] {
	return m.name
}

// @template(object/InterfaceMethodParam)
// InterfaceMethodParam represents an interface method parameter declaration.
type InterfaceMethodParam struct {
	*commonNode[*InterfaceMethodParam]

	// @template(object/InterfaceMethodParam.Method)
	// Method represents the interface method that contains the parameter.
	Method *InterfaceMethod

	// @template(object/InterfaceMethodParam.Type)
	// Type represents the parameter type.
	Type *InterfaceMethodParamType
}

func newInterfaceMethodParam(ctx *Context, file *File, m *InterfaceMethod, src *ast.MethodParam) *InterfaceMethodParam {
	p := &InterfaceMethodParam{
		Method: m,
	}
	file.addObject(ctx, src, p)
	p.commonNode = newObject(p, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	p.unresolved.annotations = src.Annotations
	p.Type = newInterfaceMethodParamType(ctx, file, p, src.Type)
	return p
}

func (p *InterfaceMethodParam) resolve(ctx *Context, file *File, scope Scope) {
	p.annotations = ctx.resolveAnnotationGroup(file, p, p.unresolved.annotations)
	p.Type.resolve(ctx, file, scope)
}

// @template(object/InterfaceMethodParam.Name)
// Name represents the [name object](#object/NodeName) of the interface method parameter.
func (p *InterfaceMethodParam) Name() *NodeName[*InterfaceMethodParam] {
	return p.name
}

// @template(object/InterfaceMethodParamType)
// InterfaceMethodParamType represents an interface method parameter type.
type InterfaceMethodParamType struct {
	unresolved struct {
		typ ast.Type
	}

	// @template(object/InterfaceMethodParamType.Param)
	// Param represents the interface method parameter that contains the type.
	Param *InterfaceMethodParam

	// @template(object/InterfaceMethodParamType.Type)
	// Type represnts the underlying type of the parameter.
	Type Type
}

func newInterfaceMethodParamType(_ *Context, _ *File, p *InterfaceMethodParam, src ast.Type) *InterfaceMethodParamType {
	t := &InterfaceMethodParamType{Param: p}
	t.unresolved.typ = src
	return t
}

func (t *InterfaceMethodParamType) resolve(ctx *Context, file *File, scope Scope) {
	t.Type = ctx.resolveType(file, t.unresolved.typ, false)
}

// @template(object/InterfaceMethodResult)
// InterfaceMethodResult represents an interface method result.
type InterfaceMethodResult struct {
	unresolved struct {
		typ ast.Type
	}

	// @template(object/InterfaceMethodResult.Method)
	// Method represents the interface method that contains the result.
	Method *InterfaceMethod

	// @template(object/InterfaceMethodResult.Type)
	// Type represents the underlying type of the result.
	Type Type
}

func newInterfaceMethodResult(_ *Context, _ *File, m *InterfaceMethod, src ast.Type) *InterfaceMethodResult {
	t := &InterfaceMethodResult{Method: m}
	t.unresolved.typ = src
	return t
}

func (t *InterfaceMethodResult) resolve(ctx *Context, file *File, scope Scope) {
	if t.unresolved.typ == nil {
		return
	}
	t.Type = ctx.resolveType(file, t.unresolved.typ, false)
}

// isAvailable reports whether the declaration is available in the current language.
func isAvailable(decl Node, lang string) bool {
	s, ok := decl.Annotations().get("next").get("available").(string)
	return !ok || text.ContainsWord(s, lang)
}

// available returns the declaration if it is available in the current language.
func available[T Node](obj T, lang string) (T, bool) {
	op.Assertf(lang != "", "language must not be empty")
	if !isAvailable(obj, lang) {
		return obj, false
	}
	switch decl := any(obj).(type) {
	case *File:
		decl.decls.lang = lang
	case *Struct:
		decl.lang = lang
	case *Interface:
		decl.lang = lang
	}
	return obj, true
}

// availableFields returns the list of fields that are available in the current language.
func availableFields[D, F Node](fields *Fields[D, F], lang string) *Fields[D, F] {
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
		return &Fields[D, F]{Decl: fields.Decl, List: list}
	}
	return fields
}

// availableList returns the list of declarations that are available in the current language.
func availableList[T Node](list List[T], lang string) List[T] {
	op.Assertf(lang != "", "language must not be empty")
	availables := make([]T, 0, len(list))
	for i, d := range list {
		if _, ok := available(d, lang); ok {
			availables = append(availables, list[i])
		}
	}
	return availables
}
