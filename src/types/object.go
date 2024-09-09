package types

import (
	"cmp"
	"path/filepath"
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

	// FullPath returns the full path of the imported file.
	// @api(object/Import/FullPath)
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

// Target returns the imported file.
// @api(object/Import/Target
func (i *Import) Target() *File { return i.target }

// File returns the file containing the import.
// @api(object/Import/File)
func (i *Import) File() *File { return i.file }

func (i *Import) resolve(ctx *Context, file *File, _ Scope) {}

// List represents a list of objects.
// @api(object/decl/List)
type List[T Object] []T

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

// FieldName represents an name of a field in a declaration:
// - Enum member name
// - Struct field name
// - Interface method name
// - Interface method parameter name
type FieldName[T Node] struct {
	pos   token.Pos
	name  string
	field T
}

func newFieldName[T Node](pos token.Pos, name string, field T) *FieldName[T] {
	return &FieldName[T]{pos: pos, name: name, field: field}
}

// Field returns the field node.
func (n *FieldName[T]) Field() T { return n.field }

// String returns the string representation of the object name.
// @api(object/decl/ObjectName/String)
func (n *FieldName[T]) String() string { return n.name }

// commonNode represents a common node.
type commonNode[Self Node] struct {
	self Self             // self represents the declaration object itself
	pos  token.Pos        // position of the declaration
	name *FieldName[Self] // name of the declaration
	file *File            // file containing the declaration

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

func newObject[Self Node](
	self Self, file *File,
	pos token.Pos, name string,
	doc *ast.CommentGroup, annotations *ast.AnnotationGroup,
) *commonNode[Self] {
	d := &commonNode[Self]{
		self: self,
		pos:  pos,
		name: newFieldName(pos, name, self),
		file: file,
		Doc:  newDoc(doc),
	}
	d.unresolved.annotations = annotations
	return d
}

func (d *commonNode[Self]) resolve(ctx *Context, file *File, scope Scope) {
	d.Annotations = ctx.resolveAnnotationGroup(file, d.self, d.unresolved.annotations)
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

func newValue(ctx *Context, file *File, name string, namePos token.Pos, src ast.Expr) *Value {
	v := &Value{
		namePos: namePos,
		name:    name,
	}
	if src != nil {
		file.addNode(ctx, src, v)
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

// IsEnum returns true if the value is an enum member.
func (v *Value) IsEnum() bool {
	return v.enum.typ != nil
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

// Any returns the underlying value of the constant or enum member.
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
	*commonNode[*Const]

	// value is the constant value.
	value *Value

	// Comment is the [line comment](#Object.Comment) of the constant declaration.
	// @api(object/decl/Const/Comment)
	Comment *Comment
}

func newConst(ctx *Context, file *File, src *ast.GenDecl[ast.Expr]) *Const {
	c := &Const{
		Comment: newComment(src.Comment),
	}
	file.addNode(ctx, src, c)
	c.commonNode = newObject(c, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	c.value = newValue(ctx, file, src.Name.Name, src.Name.NamePos, src.Spec)
	return c
}

func (c *Const) resolve(ctx *Context, file *File, scope Scope) {
	c.commonNode.resolve(ctx, file, scope)
	c.value.resolve(ctx, file, scope)
}

// Name returns the name object of the const declaration.
// @api(object/decl/Const/Name)
func (c *Const) Name() *FieldName[*Const] {
	return c.name
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
type Fields[D Node, F Object] struct {
	// Decl is the declaration object that contains the fields.
	// Decl may be an enum, struct, or interface.
	// @api(object/decl/Fields/Decl)
	Decl D

	// List is the list of fields in the declaration.
	// @api(object/decl/Fields/List)
	List []F
}

// Enum represents an enum declaration.
// @api(object/decl/Enum)
type Enum struct {
	*commonNode[*Enum]

	// Type is the enum type.
	// @api(object/decl/Enum/Type)
	Type *DeclType[*Enum]

	// Members is the list of enum members.
	// @api(object/decl/Enum/Members)
	Members *Fields[*Enum, *EnumMember]
}

func newEnum(ctx *Context, file *File, src *ast.GenDecl[*ast.EnumType]) *Enum {
	e := &Enum{}
	file.addNode(ctx, src, e)
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

// EnumMember represents an enum member object.
// @api(object/decl/Enum/EnumMember)
type EnumMember struct {
	*commonNode[*EnumMember]

	// value is the enum member value.
	value *Value

	// Decl is the enum that contains the member.
	// @api(object/decl/Enum/EnumMember/Decl)
	Decl *Enum

	// Comment is the [line comment](#Object.Comment) of the enum member declaration.
	// @api(object/decl/Enum/EnumMember/Comment)
	Comment *Comment
}

func newEnumMember(ctx *Context, file *File, e *Enum, src *ast.EnumMember, index int) *EnumMember {
	m := &EnumMember{
		Comment: newComment(src.Comment),
		Decl:    e,
	}
	file.addNode(ctx, src, m)
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

// Name returns the name object of the enum member.
// @api(object/decl/Enum/EnumMember/Name)
func (m *EnumMember) Name() *FieldName[*EnumMember] {
	return m.name
}

// Value returns the value object of the enum member.
// @api(object/decl/Enum/EnumMember/Value)
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
	*commonNode[*Struct]

	// lang is the current language to generate the struct.
	lang string

	// fields is the list of struct fields.
	fields *Fields[*Struct, *StructField]

	// Type is the struct type.
	// @api(object/decl/Struct/Type)
	Type *DeclType[*Struct]
}

func newStruct(ctx *Context, file *File, src *ast.GenDecl[*ast.StructType]) *Struct {
	s := &Struct{}
	file.addNode(ctx, src, s)
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

// Fields returns the list of struct fields.
func (s *Struct) Fields() *Fields[*Struct, *StructField] {
	return availableFields(s.fields, s.lang)
}

// Field represents a struct field declaration.
// @api(object/decl/StructField)
type StructField struct {
	*commonNode[*StructField]

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
	file.addNode(ctx, src, f)
	f.commonNode = newObject(f, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	f.Type = newStructFieldType(ctx, file, f, src.Type)
	return f
}

func (f *StructField) resolve(ctx *Context, file *File, scope Scope) {
	f.commonNode.resolve(ctx, file, scope)
	f.Type.resolve(ctx, file, scope)
}

// Name returns the name object of the struct field.
// @api(object/decl/StructField/Name)
func (f *StructField) Name() *FieldName[*StructField] {
	return f.name
}

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
	t.Type = ctx.resolveType(file, t.unresolved.typ, false)
}

// Interface represents an interface declaration.
// @api(object/decl/Interface)
type Interface struct {
	*commonNode[*Interface]

	// lang is the current language to generate the interface.
	lang string

	// methods is the list of interface methods.
	methods *Fields[*Interface, *InterfaceMethod]

	// Type is the interface type.
	// @api(object/decl/Interface/Type)
	Type *DeclType[*Interface]
}

func newInterface(ctx *Context, file *File, src *ast.GenDecl[*ast.InterfaceType]) *Interface {
	i := &Interface{}
	file.addNode(ctx, src, i)
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

// Methods returns the list of interface methods.
// @api(object/decl/Interface/Methods)
func (i *Interface) Methods() *Fields[*Interface, *InterfaceMethod] {
	return availableFields(i.methods, i.lang)
}

// InterfaceMethod represents an interface method declaration.
// @api(object/decl/InterfaceMethod)
type InterfaceMethod struct {
	*commonNode[*InterfaceMethod]

	// Decl is the interface that contains the method.
	// @api(object/decl/InterfaceMethod/Decl)
	Decl *Interface

	// Params is the list of method parameters.
	// @api(object/decl/InterfaceMethod/Params)
	Params *Fields[*InterfaceMethod, *InterfaceMethodParam]

	// Return is the method return type.
	// @api(object/decl/InterfaceMethod/Return)
	Return *InterfaceMethodResult

	// Comment is the [line comment](#Object.Comment) of the interface method declaration.
	// @api(object/decl/InterfaceMethod/Comment)
	Comment *Comment
}

func newInterfaceMethod(ctx *Context, file *File, i *Interface, src *ast.Method) *InterfaceMethod {
	m := &InterfaceMethod{
		Decl:    i,
		Comment: newComment(src.Comment),
	}
	file.addNode(ctx, src, m)
	m.commonNode = newObject(m, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	m.Params = &Fields[*InterfaceMethod, *InterfaceMethodParam]{Decl: m}
	for _, p := range src.Params.List {
		m.Params.List = append(m.Params.List, newInterfaceMethodParam(ctx, file, m, p))
	}
	m.Return = newInterfaceMethodReturn(ctx, file, m, src.Return)
	return m
}

func (m *InterfaceMethod) resolve(ctx *Context, file *File, scope Scope) {
	m.commonNode.resolve(ctx, file, scope)
	for _, p := range m.Params.List {
		p.resolve(ctx, file, scope)
	}
	if m.Return != nil {
		m.Return.resolve(ctx, file, scope)
	}
}

// Name returns the name object of the interface method.
// @api(object/decl/InterfaceMethod/Name)
func (m *InterfaceMethod) Name() *FieldName[*InterfaceMethod] {
	return m.name
}

// InterfaceMethodParam represents an interface method parameter declaration.
// @api(object/decl/InterfaceMethodParam)
type InterfaceMethodParam struct {
	*commonNode[*InterfaceMethodParam]

	// Method is the interface method that contains the parameter.
	// @api(object/decl/InterfaceMethodParam/Method)
	Method *InterfaceMethod

	// Type is the parameter type.
	// @api(object/decl/InterfaceMethodParam/Type)
	Type *InterfaceMethodParamType
}

func newInterfaceMethodParam(ctx *Context, file *File, m *InterfaceMethod, src *ast.MethodParam) *InterfaceMethodParam {
	p := &InterfaceMethodParam{
		Method: m,
	}
	file.addNode(ctx, src, p)
	p.commonNode = newObject(p, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	p.unresolved.annotations = src.Annotations
	p.Type = newInterfaceMethodParamType(ctx, file, p, src.Type)
	return p
}

func (p *InterfaceMethodParam) resolve(ctx *Context, file *File, scope Scope) {
	p.Annotations = ctx.resolveAnnotationGroup(file, p, p.unresolved.annotations)
	p.Type.resolve(ctx, file, scope)
}

// Name returns the name object of the interface method parameter.
func (p *InterfaceMethodParam) Name() *FieldName[*InterfaceMethodParam] {
	return p.name
}

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
	t.Type = ctx.resolveType(file, t.unresolved.typ, false)
}

// InterfaceMethodResult represents an interface method return type.
// @api(object/decl/InterfaceMethodResult)
type InterfaceMethodResult struct {
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

func newInterfaceMethodReturn(_ *Context, _ *File, m *InterfaceMethod, src ast.Type) *InterfaceMethodResult {
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
	s, ok := decl.getAnnotations().get("next").get("available").(string)
	return !ok || templateutil.ContainsWord(s, lang)
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
