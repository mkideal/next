package types

import (
	"cmp"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gopherd/core/container/iters"
	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// @api(Object/Imports) holds a list of imports.
type Imports struct {
	// @api(Object/Imports.File) represents the file containing the imports.
	File *File

	// @api(Object/Imports.List) represents the list of [imports](#Object/Import).
	List []*Import
}

func (i *Imports) resolve(c *Compiler, file *File) {
	for _, spec := range i.List {
		spec.target = c.lookupFile(file.Path, spec.Path)
		if spec.target == nil {
			c.addErrorf(spec.pos, "import file not found: %s", spec.Path)
		}
	}
}

// @api(Object/Imports.TrimmedList) represents a list of unique imports sorted by package name.
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

// @api(Object/Import) represents a file import.
type Import struct {
	pos    token.Pos // position of the import declaration
	target *File     // imported file
	file   *File     // file containing the import

	// @api(Object/Import.Doc) represents the import declaration [documentation](#Object/Doc).
	Doc *Doc

	// @api(Object/Import.Comment) represents the import declaration line [comment](#Object/Comment).
	Comment *Comment

	// @api(Object/Import.Path) represents the import path.
	Path string

	// @api(Object/Import.FullPath) represents the full path of the import.
	FullPath string
}

func newImport(c *Compiler, file *File, src *ast.ImportDecl) *Import {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		c.addErrorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
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
			c.addErrorf(token.NoPos, "failed to get absolute path of %s: %v", i.Path, err)
		} else {
			i.FullPath = path
		}
	}
	return i
}

// @api(Object/Import.Target) represents the imported file.
func (i *Import) Target() *File { return i.target }

// @api(Object/Import.File) represents the file containing the import declaration.
func (i *Import) File() *File { return i.file }

func (i *Import) resolve(c *Compiler, file *File, _ Scope) {}

// @api(Object/Common/List) represents a list of objects.
type List[T Object] []T

// @api(Object/Common/List.List) represents the list of objects. It is used to provide a uniform way to access.
func (l List[T]) List() []T {
	return l
}

// @api(Object/Consts) represents a [list](#Object/Common/List) of const declarations.
type Consts = List[*Const]

// @api(Object/Enums) represents a [list](#Object/Common/List) of enum declarations.
type Enums = List[*Enum]

// @api(Object/Structs) represents a [list](#Object/Common/List) of struct declarations.
type Structs = List[*Struct]

// @api(Object/Interfaces) represents a [list](#Object/Common/List) of interface declarations.
type Interfaces = List[*Interface]

// @api(Object/Decls) holds all declarations in a file.
type Decls struct {
	consts     Consts
	enums      Enums
	structs    Structs
	interfaces Interfaces

	lang string
}

func (d *Decls) resolve(c *Compiler, file *File) {
	if d == nil {
		return
	}
	for _, x := range d.consts {
		x.resolve(c, file, file)
	}
	for _, x := range d.enums {
		x.resolve(c, file, file)
	}
	for _, x := range d.structs {
		x.resolve(c, file, file)
	}
	for _, x := range d.interfaces {
		x.resolve(c, file, file)
	}
}

// @api(Object/Decls.Consts) represents the [list](#Object/Common/List) of [const](#Object/Const) declarations.
func (d *Decls) Consts() Consts {
	if d == nil {
		return nil
	}
	return availableList(d.consts, d.lang)
}

// @api(Object/Decls.Enums) represents the [list](#Object/Common/List) of [enum](#Object/Enum) declarations.
func (d *Decls) Enums() Enums {
	if d == nil {
		return nil
	}
	return availableList(d.enums, d.lang)
}

// @api(Object/Decls.Structs) represents the [list](#Object/Common/List) of [struct](#Object/Struct) declarations.
func (d *Decls) Structs() Structs {
	if d == nil {
		return nil
	}
	return availableList(d.structs, d.lang)
}

// @api(Object/Decls.Interfaces) represents the [list](#Object/Common/List) of [interface](#Object/Interface) declarations.
func (d *Decls) Interfaces() Interfaces {
	if d == nil {
		return nil
	}
	return availableList(d.interfaces, d.lang)
}

// @api(Object/Common/NodeName) represents a name of a node in a declaration.
//
// Currently, the following types are supported:
//
// - [ConstName](#Object/ConstName)
// - [EnumMemberName](#Object/EnumMemberName)
// - [StructFieldName](#Object/StructFieldName)
// - [InterfaceMethodName](#Object/InterfaceMethodName)
// - [InterfaceMethodParamName](#Object/InterfaceMethodParamName)
type NodeName[T Node] struct {
	pos  token.Pos
	name string
	node T
}

// @api(Object/ConstName) represents the [NodeName](#Object/Common/NodeName) of a [const](#Object/Const) declaration.
type ConstName = NodeName[*Const]

// @api(Object/EnumMemberName) represents the [NodeName](#Object/Common/NodeName) of an [enum member](#Object/EnumMember).
type EnumMemberName = NodeName[*EnumMember]

// @api(Object/StructFieldName) represents the [NodeName](#Object/Common/NodeName) of a [struct field](#Object/StructField).
type StructFieldName = NodeName[*StructField]

// @api(Object/InterfaceMethodName) represents the [NodeName](#Object/Common/NodeName) of an [interface method](#Object/InterfaceMethod).
type InterfaceMethodName = NodeName[*InterfaceMethod]

// @api(Object/InterfaceMethodParamName) represents the [NodeName](#Object/Common/NodeName) of an [interface method parameter](#Object/InterfaceMethodParam).
type InterfaceMethodParamName = NodeName[*InterfaceMethodParam]

func newNodeName[T Node](pos token.Pos, name string, field T) *NodeName[T] {
	return &NodeName[T]{pos: pos, name: name, node: field}
}

// @api(Object/Common/NodeName.Node) represents the [node](#Object/Common/Node) that contains the name.
func (n *NodeName[T]) Node() T { return n.node }

// @api(Object/Common/NodeName.String) represents the string representation of the node name.
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

func newCommonNode[Self Node](
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

func (d *commonNode[Self]) resolve(c *Compiler, file *File, scope Scope) {
	d.annotations = c.resolveAnnotationGroup(file, d.self, d.unresolved.annotations)
}

// Doc implements Node interface.
func (d *commonNode[Self]) Doc() *Doc {
	return d.doc
}

// Annotations implements Node interface.
func (d *commonNode[Self]) Annotations() Annotations {
	return d.annotations
}

// iotaValue represents the iota value of an enum member.
type iotaValue struct {
	value int
	found bool
}

// @api(Object/Value) represents a constant value for a const declaration or an enum member.
type Value struct {
	namePos    token.Pos
	name       string
	typ        *PrimitiveType
	val        constant.Value
	unresolved struct {
		value ast.Expr
	}
	file *File
	enum struct {
		typ   *Enum     // parent enum
		index int       // index in the enum type. start from 0
		iota  iotaValue // iota value
	}
}

func newValue(c *Compiler, file *File, name string, namePos token.Pos, src ast.Expr) *Value {
	v := &Value{
		namePos: namePos,
		name:    name,
		file:    file,
	}
	if src != nil {
		file.addObject(c, src, v)
	}
	v.unresolved.value = src
	return v
}

func (v *Value) File() *File {
	return v.file
}

func (v *Value) resolve(c *Compiler, file *File, scope Scope) {
	v.val = v.resolveValue(c, file, scope, make([]*Value, 0, 16))
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
		c.addErrorf(v.namePos, "unexpected constant type: %T", v.Any())
	}
}

func (v *Value) resolveValue(c *Compiler, file *File, scope Scope, refs []*Value) constant.Value {
	// If value already resolved, return it
	if v.val != nil {
		return v.val
	}

	// If enum type is nil, resolve constant value expression in which iota is not allowed
	if v.enum.typ == nil {
		v.val = c.recursiveResolveValue(file, scope, append(refs, v), v.unresolved.value, nil)
		return v.val
	}

	if v.unresolved.value != nil {
		// Resolve value expression
		v.val = c.recursiveResolveValue(file, v.enum.typ, append(refs, v), v.unresolved.value, &v.enum.iota)
	} else if v.enum.index == 0 {
		// First member of enum type has value 0 if not specified
		v.val = constant.MakeInt64(0)
	} else {
		// Resolve previous value
		prev := v.enum.typ.Members.List[v.enum.index-1]
		prev.value.resolveValue(c, file, v.enum.typ, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.value.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.typ.Members.List[v.enum.index-v.enum.iota.value]

		if start.value.val != nil && start.value.enum.iota.found {
			// If start value is specified and it has iota expression, resolve it with the current iota value
			v.val = c.recursiveResolveValue(file, v.enum.typ, append(refs, v), start.value.unresolved.value, &v.enum.iota)
		} else {
			// Otherwise, add 1 to the previous value
			v.val = constant.BinaryOp(prev.value.val, token.ADD, constant.MakeInt64(1))
		}
	}
	return v.val
}

// @api(Object/Value.IsEnum) returns true if the value is an enum member.
func (v *Value) IsEnum() bool {
	return v.enum.typ != nil
}

// @api(Object/Value.Type) represents the [primitive type](#Object/PrimitiveType) of the value.
func (v *Value) Type() *PrimitiveType {
	return v.typ
}

// @api(Object/Value.String) represents the string representation of the value.
func (v *Value) String() string {
	return v.val.String()
}

// @api(Object/Value.Any) represents the underlying value of the constant.
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

// @api(Object/Const) (extends [Decl](#Object/Common/Decl)) represents a const declaration.
type Const struct {
	*commonNode[*Const]

	// value is the constant value.
	value *Value

	// @api(Object/Const.Comment) is the line [comment](#Object/Comment) of the constant declaration.
	Comment *Comment
}

func newConst(c *Compiler, file *File, src *ast.GenDecl[ast.Expr]) *Const {
	x := &Const{
		Comment: newComment(src.Comment),
	}
	file.addObject(c, src, x)
	x.commonNode = newCommonNode(x, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	x.value = newValue(c, file, src.Name.Name, src.Name.NamePos, src.Spec)
	return x
}

func (x *Const) resolve(c *Compiler, file *File, scope Scope) {
	x.commonNode.resolve(c, file, scope)
	x.value.resolve(c, file, scope)
}

// @api(Object/Const.Name) represents the [NodeName](#Object.NodeName) of the constant.
func (x *Const) Name() *NodeName[*Const] {
	return x.name
}

// @api(Object/Const.Type) represents the type of the constant.
func (x *Const) Type() *PrimitiveType {
	return x.value.Type()
}

// @api(Object/Const.Value) represents the [value object](#Object/Value) of the constant.
func (x *Const) Value() *Value {
	return x.value
}

// @api(Object/Common/Fields) represents a list of fields in a declaration.
type Fields[D Node, F Object] struct {
	// @api(Object/Common/Fields.Decl) is the declaration object that contains the fields.
	//
	// Currently, it is one of following types:
	//
	// - [Enum](#Object/Enum)
	// - [Struct](#Object/Struct)
	// - [Interface](#Object/Interface)
	// - [InterfaceMethod](#Object/InterfaceMethod).
	Decl D

	// @api(Object/Common/Fields.List) is the list of fields in the declaration.
	//
	// Currently, the field object is one of following types:
	//
	// - [EnumMember](#Object/EnumMember)
	// - [StructField](#Object/StructField)
	// - [InterfaceMethod](#Object/InterfaceMethod).
	// - [InterfaceMethodParam](#Object/InterfaceMethodParam).
	List []F
}

// @api(Object/EnumMembers) represents the [list](#Object/Common/Fields) of [enum members](#Object/EnumMember).
type EnumMembers = Fields[*Enum, *EnumMember]

// @api(Object/StructFields) represents the [list](#Object/Common/Fields) of [struct fields](#Object/StructField).
type StructFields = Fields[*Struct, *StructField]

// @api(Object/InterfaceMethods) represents the [list](#Object/Common/Fields) of [interface methods](#Object/InterfaceMethod).
type InterfaceMethods = Fields[*Interface, *InterfaceMethod]

// @api(Object/InterfaceMethodParams) represents the [list](#Object/Common/Fields) of [interface method parameters](#Object/InterfaceMethodParam).
type InterfaceMethodParams = Fields[*InterfaceMethod, *InterfaceMethodParam]

// @api(Object/Enum) (extends [Decl](#Object/Common/Decl)) represents an enum declaration.
type Enum struct {
	*commonNode[*Enum]

	// @api(Object/Enum.MemberType) represents the type of the enum members.
	MemberType *PrimitiveType

	// @api(Object/Enum.Type) is the enum type.
	Type *DeclType[*Enum]

	// @api(Object/Enum.Members) is the list of enum members.
	Members *EnumMembers
}

func newEnum(c *Compiler, file *File, src *ast.GenDecl[*ast.EnumType]) *Enum {
	e := &Enum{}
	file.addObject(c, src, e)
	e.commonNode = newCommonNode(e, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	e.Type = newDeclType(file, src.Pos(), KindEnum, src.Name.Name, e)
	e.Members = &EnumMembers{Decl: e}
	for i, m := range src.Spec.Members.List {
		e.Members.List = append(e.Members.List, newEnumMember(c, file, e, m, i))
	}
	return e
}

func (e *Enum) resolve(c *Compiler, file *File, scope Scope) {
	e.commonNode.resolve(c, file, scope)
	for _, m := range e.Members.List {
		m.resolve(c, file, scope)
	}
	if len(e.Members.List) == 0 {
		e.MemberType = primitiveTypes["int"]
	} else if c.errors.Len() == 0 {
		e.MemberType = e.Members.List[0].Value().Type()
		for _, m := range e.Members.List {
			if kind := e.MemberType.kind.Compatible(m.Value().Type().kind); kind.Valid() {
				if e.MemberType.kind != kind {
					e.MemberType = m.value.Type()
				}
			} else {
				c.addErrorf(m.Value().namePos, "incompatible type %s for enum member %s", m.Value().Type().name, m.Name().String())
				return
			}
		}
	}
}

// @api(Object/EnumMember) (extends [Decl](#Object/Common/Decl)) represents an enum member object in an [enum](#Object/Enum) declaration.
type EnumMember struct {
	*commonNode[*EnumMember]

	// value is the enum member value.
	value *Value

	// @api(Object/EnumMember.Decl) represents the [enum](#Object/Enum) that contains the member.
	Decl *Enum

	// @api(Object/EnumMember.Comment) represents the line [comment](#Object/Comment) of the enum member declaration.
	Comment *Comment
}

func newEnumMember(c *Compiler, file *File, e *Enum, src *ast.EnumMember, index int) *EnumMember {
	m := &EnumMember{
		Comment: newComment(src.Comment),
		Decl:    e,
	}
	file.addObject(c, src, m)
	m.commonNode = newCommonNode(m, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	m.value = newValue(c, file, src.Name.Name, src.Name.NamePos, src.Value)
	m.value.enum.typ = e
	m.value.enum.index = index
	return m
}

func (m *EnumMember) resolve(c *Compiler, file *File, scope Scope) {
	m.commonNode.resolve(c, file, scope)
	m.value.resolve(c, file, scope)
}

// @api(Object/EnumMember.Name) represents the [NodeName](#Object/Common/NodeName) of the enum member.
func (m *EnumMember) Name() *NodeName[*EnumMember] {
	return m.name
}

// @api(Object/EnumMember.Value) represents the [value object](#Object/Value) of the enum member.
func (m *EnumMember) Value() *Value {
	return m.value
}

// @api(Object/Value.IsFirst) reports whether the value is the first member of the enum type.
func (m *EnumMember) IsFirst() bool {
	return m.value.enum.typ != nil && m.value.enum.index == 0
}

// @api(Object/Value.IsLast) reports whether the value is the last member of the enum type.
func (m *EnumMember) IsLast() bool {
	return m.value.enum.typ != nil && m.value.enum.index == len(m.value.enum.typ.Members.List)-1
}

// @api(Object/Struct) (extends [Decl](#Object/Common/Decl)) represents a struct declaration.
type Struct struct {
	*commonNode[*Struct]

	// lang is the current language to generate the struct.
	lang string

	// fields is the list of struct fields.
	fields *StructFields

	// @api(Object/Struct.Type) represents the struct type.
	Type *DeclType[*Struct]
}

func newStruct(c *Compiler, file *File, src *ast.GenDecl[*ast.StructType]) *Struct {
	s := &Struct{}
	file.addObject(c, src, s)
	s.commonNode = newCommonNode(s, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	s.Type = newDeclType(file, src.Pos(), KindStruct, src.Name.Name, s)
	s.fields = &StructFields{Decl: s}
	for _, f := range src.Spec.Fields.List {
		s.fields.List = append(s.fields.List, newStructField(c, file, s, f))
	}
	return s
}

func (s *Struct) resolve(c *Compiler, file *File, scope Scope) {
	s.commonNode.resolve(c, file, scope)
	for _, f := range s.fields.List {
		f.resolve(c, file, scope)
	}
}

// @api(Object/Struct.Fields) represents the list of struct fields.
func (s *Struct) Fields() *StructFields {
	return availableFields(s.fields, s.lang)
}

// @api(Object/StructField) (extends [Node](#Object/Common/Node)) represents a struct field declaration.
type StructField struct {
	*commonNode[*StructField]

	unresolved struct {
		typ ast.Type
	}

	// @api(Object/StructField.Decl) represents the struct that contains the field.
	Decl *Struct

	// @api(Object/StructField.Type) represents the [type](#Object/Common/Type) of the struct field.
	Type Type

	// @api(Object/StructField.Comment) represents the line [comment](#Object/Comment) of the struct field declaration.
	Comment *Comment
}

func newStructField(c *Compiler, file *File, s *Struct, src *ast.StructField) *StructField {
	f := &StructField{Decl: s}
	file.addObject(c, src, f)
	f.commonNode = newCommonNode(f, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	f.unresolved.typ = src.Type
	return f
}

func (f *StructField) resolve(c *Compiler, file *File, scope Scope) {
	f.commonNode.resolve(c, file, scope)
	f.Type = c.resolveType(file, f.unresolved.typ, false)
}

// @api(Object/StructField.Name) represents the [NodeName](#Object/Common/NodeName) of the struct field.
func (f *StructField) Name() *NodeName[*StructField] {
	return f.name
}

// @api(Object/Interface) (extends [Decl](#Object/Common/Decl)) represents an interface declaration.
type Interface struct {
	*commonNode[*Interface]

	// lang is the current language to generate the interface.
	lang string

	// methods is the list of interface methods.
	methods *InterfaceMethods

	// @api(Object/Interface.Type) represents the interface type.
	Type *DeclType[*Interface]
}

func newInterface(c *Compiler, file *File, src *ast.GenDecl[*ast.InterfaceType]) *Interface {
	i := &Interface{}
	file.addObject(c, src, i)
	i.commonNode = newCommonNode(i, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	i.Type = newDeclType(file, src.Pos(), KindInterface, src.Name.Name, i)
	i.methods = &InterfaceMethods{Decl: i}
	for _, m := range src.Spec.Methods.List {
		i.methods.List = append(i.methods.List, newInterfaceMethod(c, file, i, m))
	}
	return i
}

func (i *Interface) resolve(c *Compiler, file *File, scope Scope) {
	i.commonNode.resolve(c, file, scope)
	for _, m := range i.methods.List {
		m.resolve(c, file, scope)
	}
}

// @api(Object/Interface.Methods) represents the list of interface methods.
func (i *Interface) Methods() *InterfaceMethods {
	return availableFields(i.methods, i.lang)
}

// @api(Object/InterfaceMethod) (extends [Node](#Object/Common/Node)) represents an interface method declaration.
type InterfaceMethod struct {
	*commonNode[*InterfaceMethod]

	// @api(Object/InterfaceMethod.Decl) represents the interface that contains the method.
	Decl *Interface

	// @api(Object/InterfaceMethod.Params) represents the list of method parameters.
	Params *InterfaceMethodParams

	// @api(Object/InterfaceMethod.Result) represents the return type of the method.
	Result *InterfaceMethodResult

	// @api(Object/InterfaceMethod.Comment) represents the line [comment](#Object/Comment) of the interface method declaration.
	Comment *Comment
}

func newInterfaceMethod(c *Compiler, file *File, i *Interface, src *ast.Method) *InterfaceMethod {
	m := &InterfaceMethod{
		Decl:    i,
		Comment: newComment(src.Comment),
	}
	file.addObject(c, src, m)
	m.commonNode = newCommonNode(m, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	m.Params = &InterfaceMethodParams{Decl: m}
	for _, p := range src.Params.List {
		m.Params.List = append(m.Params.List, newInterfaceMethodParam(c, file, m, p))
	}
	m.Result = newInterfaceMethodResult(c, file, m, src.Result)
	return m
}

func (m *InterfaceMethod) resolve(c *Compiler, file *File, scope Scope) {
	m.commonNode.resolve(c, file, scope)
	for _, p := range m.Params.List {
		p.resolve(c, file, scope)
	}
	if m.Result != nil {
		m.Result.resolve(c, file, scope)
	}
}

// @api(Object/InterfaceMethod.Name) represents the [NodeName](#Object/Common/NodeName) of the interface method.
func (m *InterfaceMethod) Name() *NodeName[*InterfaceMethod] {
	return m.name
}

// @api(Object/InterfaceMethodParam) (extends [Node](#Object/Common/Node)) represents an interface method parameter declaration.
type InterfaceMethodParam struct {
	*commonNode[*InterfaceMethodParam]

	unresolved struct {
		typ ast.Type
	}

	// @api(Object/InterfaceMethodParam.Method) represents the interface method that contains the parameter.
	Method *InterfaceMethod

	// @api(Object/InterfaceMethodParam.Type) represents the [type](#Object/Common/Type) of the parameter.
	Type Type
}

func newInterfaceMethodParam(c *Compiler, file *File, m *InterfaceMethod, src *ast.MethodParam) *InterfaceMethodParam {
	p := &InterfaceMethodParam{
		Method: m,
	}
	file.addObject(c, src, p)
	p.commonNode = newCommonNode(p, file, src.Pos(), src.Name.Name, src.Doc, src.Annotations)
	p.unresolved.typ = src.Type
	return p
}

func (p *InterfaceMethodParam) resolve(c *Compiler, file *File, scope Scope) {
	p.commonNode.resolve(c, file, scope)
	p.Type = c.resolveType(file, p.unresolved.typ, false)
}

// @api(Object/InterfaceMethodParam.Name) represents the [NodeName](#Object/Common/NodeName) of the interface method parameter.
func (p *InterfaceMethodParam) Name() *NodeName[*InterfaceMethodParam] {
	return p.name
}

// @api(Object/InterfaceMethodResult) represents an interface method result.
type InterfaceMethodResult struct {
	unresolved struct {
		typ ast.Type
	}

	// @api(Object/InterfaceMethodResult.Method) represents the interface method that contains the result.
	Method *InterfaceMethod

	// @api(Object/InterfaceMethodResult.Type) represents the underlying type of the result.
	Type Type
}

func newInterfaceMethodResult(_ *Compiler, _ *File, m *InterfaceMethod, src ast.Type) *InterfaceMethodResult {
	t := &InterfaceMethodResult{Method: m}
	t.unresolved.typ = src
	return t
}

func (t *InterfaceMethodResult) resolve(c *Compiler, file *File, scope Scope) {
	if t.unresolved.typ == nil {
		return
	}
	t.Type = c.resolveType(file, t.unresolved.typ, false)
}

// isAvailable reports whether the declaration is available in the current language.
func isAvailable(decl Node, lang string) bool {
	s, ok := decl.Annotations().get("next").get("available").(string)
	// TODO: support logical expression.
	return !ok || iters.Contains(iters.Map(slices.Values(strings.Split(s, "|")), strings.TrimSpace), lang)
}

// available returns the declaration if it is available in the current language.
func available[T Node](obj T, lang string) (T, bool) {
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
	availables := make([]T, 0, len(list))
	for i, d := range list {
		if _, ok := available(d, lang); ok {
			availables = append(availables, list[i])
		}
	}
	return availables
}
