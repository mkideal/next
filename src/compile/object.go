package compile

import (
	"cmp"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/parser"
	"github.com/next/next/src/token"
)

// @api(Object/Imports) holds a slice of [Import](#Object/Import) declarations and the declaration that contains the imports.
type Imports[T Decl] struct {
	// @api(Object/Imports.Decl) represents the declaration [Node](#Object/Common/Node) that contains the imports.
	// Currently, it is one of following types:
	//
	// - [File](#Object/File)
	// - [Package](#Object/Package)
	//
	// Example:
	//
	//	```next title="file1.next"
	//	package a;
	//
	//	import "path/to/file2.next";
	//	```
	//
	//	```next title="file2.next"
	//	package a;
	//	```
	//
	//	```npl title="file.npl"
	//	{{- define "meta/this" -}}file{{- end -}}
	//
	//	{{range this.Imports.List}}
	//	{{.Decl.Name}}
	//	{{end}}
	//	```
	//
	//	Output (for file1.next):
	//
	//	```
	//	file1
	//	```
	//
	//	```npl title="package.npl"
	//	{{- define "meta/this" -}}package{{- end -}}
	//
	//	{{range this.Imports.List}}
	//	{{.Decl.Name}}
	//	{{end}}
	//	```
	//
	//	Output:
	//
	//	```
	//	a
	//	```
	Decl Node

	// @api(Object/Imports.List) represents a slice of [Import](#Object/Import) declarations: **\[[Import](#Object/Import)\]**.
	//
	// Example: see [Object/Imports.Decl](#Object/Imports.Decl).
	List []*Import
}

func (i *Imports[T]) resolve(c *Compiler, file *File) {
	for _, spec := range i.List {
		spec.resolve(c, file, file)
	}
}

func (i *Imports[T]) add(x *Import) {
	for _, spec := range i.List {
		if spec.FullPath == x.FullPath {
			return
		}
	}
	i.List = append(i.List, x)
}

// @api(Object/Imports.TrimmedList) represents a slice of unique imports sorted by package name.
//
// Example:
//
//	```next
//	package c;
//
//	import "path/to/file1.next"; // package: a
//	import "path/to/file2.next"; // package: a
//	import "path/to/file3.next"; // package: b
//	```
//
//	```npl
//	{{range .Imports.TrimmedList}}
//	{{.Target.Package.Name}}
//	{{end}}
//	```
//
// Output:
//
//	```
//	a
//	b
//	```
func (i *Imports[T]) TrimmedList() []*Import {
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

// @api(Object/Import) represents a import declaration in a [File](#Object/File).
type Import struct {
	pos    token.Pos // position of the import declaration
	target *File     // the imported file
	file   *File     // the file containing the import declaration

	// @api(Object/Import.Doc) represents the [Doc](#Object/Doc) comment of the import.
	//
	// Example:
	//
	//	```next
	//	// This is a documenation comment for the import.
	//	// It can be multiline.
	//	import "path/to/file.next"; // This is a line comment for the import declaration.
	//	```
	//
	//	```npl
	//	{{.Doc.Text}}
	//	```
	//
	// Output:
	//
	//	```
	//	This is a documenation comment for the import.
	//	It can be multiline.
	//	```
	Doc *Doc

	// @api(Object/Import.Comment) represents the line [Comment](#Object/Comment) of the import declaration.
	//
	// Example:
	//
	//	```next
	//	// This is a documenation comment for the import.
	//	// It can be multiline.
	//	import "path/to/file.next"; // This is a line comment for the import.
	//	```
	//
	//	```npl
	//	{{.Comment.Text}}
	//	```
	//
	// Output:
	//
	//	```
	//	This is a line comment for the import.
	//	```
	Comment *Comment

	// @api(Object/Import.Path) represents the import path.
	//
	// Example:
	//
	//	```next
	//	import "path/to/file.next";
	//	```
	//
	//	```npl
	//	{{.Path}}
	//	```
	//
	// Output:
	//
	//	```
	//	path/to/file.next
	//	```
	Path string

	// @api(Object/Import.FullPath) represents the full path of the import.
	//
	// Example:
	//
	//	```next
	//	import "path/to/file.next";
	//	```
	//
	//	```npl
	//	{{.FullPath}}
	//	```
	//
	// Output:
	//
	//	```
	//	/full/path/to/file.next
	//	```
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

// @api(Object/Import.Target) represents the imported [File](#Object/File) object.
//
// Example:
//
//	```next title="file1.next"
//	package a;
//
//	import "file2.next";
//	import "file3.next";
//	```
//
//	```npl
//	{{range .Imports.List}}
//	{{.Target.Path}}
//	{{end}}
//	```
//
//	Output:
//
//	```
//	file2.next
//	file3.next
//	```
func (i *Import) Target() *File { return i.target }

// @api(Object/Import.File) represents the [File](#Object/File) object that contains the import declaration.
//
// Example:
//
//	```next title="file1.next"
//	package a;
//
//	import "file2.next";
//	import "file3.next";
//	```
//
//	```npl
//	{{range .Imports.List}}
//	{{.File.Path}}
//	{{end}}
//	```
//
//	Output:
//
//	```
//	file1.next
//	file1.next
//	```
func (i *Import) File() *File { return i.file }

func (i *Import) resolve(c *Compiler, file *File, _ Scope) {
	i.target = c.lookupFile(file.Path, i.Path)
	if i.target == nil {
		c.addErrorf(i.pos, "import file not found: %s", i.Path)
	}
}

// @api(Object/Common/List)
// `List<T>` represents a slice of objects: **\[T: [Object](#Object)\]**.
type List[T Object] []T

// @api(Object/Common/List.List) represents the slice of [Object](#Object)s. It is used to provide a uniform way to access.
func (l List[T]) List() []T {
	return l
}

// @api(Object/Consts) represents a [List](#Object/Common/List) of [Const](#Object/Const) declarations.
type Consts = List[*Const]

// @api(Object/Enums) represents a [List](#Object/Common/List) of [Enum](#Object/Enum) declarations.
type Enums = List[*Enum]

// @api(Object/Structs) represents a [List](#Object/Common/List) of [Struct](#Object/Struct) declarations.
type Structs = List[*Struct]

// @api(Object/Interfaces) represents a [List](#Object/Common/List) of [Interface](#Object/Interface) declarations.
type Interfaces = List[*Interface]

// @api(Object/Decls) holds all top-level declarations in a file.
type Decls struct {
	compiler   *Compiler
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

// @api(Object/Decls.Consts) represents the [List](#Object/Common/List) of [Const](#Object/Const) declarations.
func (d *Decls) Consts() Consts {
	if d == nil {
		return nil
	}
	return availableList(d.compiler, d.consts, d.lang)
}

// @api(Object/Decls.Enums) represents the [List](#Object/Common/List) of [Enum](#Object/Enum) declarations.
func (d *Decls) Enums() Enums {
	if d == nil {
		return nil
	}
	return availableList(d.compiler, d.enums, d.lang)
}

// @api(Object/Decls.Structs) represents the [List](#Object/Common/List) of [Struct](#Object/Struct) declarations.
func (d *Decls) Structs() Structs {
	if d == nil {
		return nil
	}
	return availableList(d.compiler, d.structs, d.lang)
}

// @api(Object/Decls.Interfaces) represents the [List](#Object/Common/List) of [Interface](#Object/Interface) declarations.
func (d *Decls) Interfaces() Interfaces {
	if d == nil {
		return nil
	}
	return availableList(d.compiler, d.interfaces, d.lang)
}

// commonNode represents a common node.
type commonNode[Self Node] struct {
	self    Self      // self represents the declaration object itself
	pos     token.Pos // position of the declaration
	name    string    // name of the declaration
	namePos token.Pos // position of the name
	file    *File     // file containing the declaration

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
	pos token.Pos, name *ast.Ident,
	doc *ast.CommentGroup, annotations *ast.AnnotationGroup,
) *commonNode[Self] {
	d := &commonNode[Self]{
		self:    self,
		pos:     pos,
		name:    name.Name,
		namePos: name.Pos(),
		file:    file,
		doc:     newDoc(doc),
	}
	d.unresolved.annotations = annotations
	return d
}

func (d *commonNode[Self]) resolve(c *Compiler, file *File, scope Scope) {
	d.annotations = resolveAnnotations(c, file, d.self, d.unresolved.annotations)
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

// @api(Object/Value) represents a value for a const declaration or an enum member.
type Value struct {
	pos        token.Pos
	namePos    token.Pos
	name       string
	typ        *PrimitiveType
	val        constant.Value
	unresolved struct {
		value ast.Expr
	}
	file *File
	enum struct {
		decl  *Enum     // enum declaration that contains the value
		index int       // index in the enum type. start from 0
		iota  iotaValue // iota value
	}
}

func newValue(c *Compiler, file *File, name *ast.Ident, src ast.Expr) *Value {
	v := &Value{
		namePos: name.Pos(),
		name:    name.Name,
		file:    file,
	}
	if src != nil {
		v.pos = src.Pos()
		file.addObject(c, src, v)
	} else {
		v.pos = name.End()
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
	case int32:
		v.typ = primitiveTypes["int32"]
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
	if v.enum.decl == nil {
		v.val = c.recursiveResolveValue(file, scope, append(refs, v), v.unresolved.value, nil)
		return v.val
	}

	if v.unresolved.value != nil {
		// Resolve value expression
		v.val = c.recursiveResolveValue(file, v.enum.decl, append(refs, v), v.unresolved.value, &v.enum.iota)
	} else if v.enum.index == 0 {
		// First member of enum type has value 0 if not specified
		v.val = constant.MakeInt64(0)
	} else {
		// Resolve previous value
		prev := v.enum.decl.Members.List[v.enum.index-1]
		prev.value.resolveValue(c, file, v.enum.decl, append(refs, v))

		// Increment iota value
		v.enum.iota.value = prev.value.enum.iota.value + 1

		// Find the start value of the enum type for iota expression
		start := v.enum.decl.Members.List[v.enum.index-v.enum.iota.value]

		if start.value.val != nil && start.value.enum.iota.found {
			// If start value is specified and it has iota expression, resolve it with the current iota value
			v.val = c.recursiveResolveValue(file, v.enum.decl, append(refs, v), start.value.unresolved.value, &v.enum.iota)
		} else {
			// Otherwise, add 1 to the previous value
			v.val = constant.BinaryOp(prev.value.val, token.ADD, constant.MakeInt64(1))
		}
	}
	return v.val
}

// @api(Object/Value.Enum) represents the enum object that contains the value if it is an enum member.
// Otherwise, it returns nil.
func (v *Value) Enum() *Enum {
	return v.enum.decl
}

// @api(Object/Value.Type) represents the [PrimitiveType](#Object/PrimitiveType) of the value.
func (v *Value) Type() *PrimitiveType {
	return v.typ
}

// @api(Object/Value.String) represents the string representation of the value.
func (v *Value) String() string {
	return v.val.String()
}

// @api(Object/Value.Any) represents the underlying value of the constant.
// It returns one of the following types:
//
// - **int32**
// - **int64**
// - **float32**
// - **float64**
// - **bool**
// - **string**
// - **nil**
func (v *Value) Any() any {
	if v.val == nil {
		return nil
	}
	switch v.val.Kind() {
	case constant.Int:
		if i, exactly := constant.Int64Val(v.val); exactly {
			if i >= math.MinInt32 && i <= math.MaxInt32 {
				return int32(i)
			}
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

	// @api(Object/Const.Comment) is the line [Comment](#Object/Comment) of the constant declaration.
	//
	// Example:
	//
	//	```next
	//	const x = 1; // This is a line comment for the constant.
	//	```
	//
	//	```npl
	//	{{.Comment.Text}}
	//	```
	//
	// Output:
	//
	//	```
	//	This is a line comment for the constant.
	//	```
	Comment *Comment
}

func newConst(c *Compiler, file *File, src *ast.GenDecl[ast.Expr]) *Const {
	x := &Const{
		Comment: newComment(src.Comment),
	}
	file.addObject(c, src, x)
	x.commonNode = newCommonNode(x, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	x.value = newValue(c, file, src.Name, src.Spec)
	return x
}

func (x *Const) resolve(c *Compiler, file *File, scope Scope) {
	x.commonNode.resolve(c, file, scope)
	x.value.resolve(c, file, scope)
}

// @api(Object/Const.Type) represents the type of the constant.
func (x *Const) Type() *PrimitiveType {
	return x.value.Type()
}

// @api(Object/Const.Value) represents the [Value](#Object/Value) object of the constant.
func (x *Const) Value() *Value {
	return x.value
}

// @api(Object/Common/Fields)
// `Fields<D, F>` represents a list of fields of a declaration where `D` is the declaration [Node](#Object/Common/Node) and
// `F` is the field [Object](#Object).
type Fields[D Node, F Object] struct {
	// @api(Object/Common/Fields.Decl) is the declaration [Node](#Object/Common/Node) that contains the fields.
	//
	// Currently, it is one of following types:
	//
	// - [Enum](#Object/Enum)
	// - [Struct](#Object/Struct)
	// - [Interface](#Object/Interface)
	// - [InterfaceMethod](#Object/InterfaceMethod).
	Decl D

	// @api(Object/Common/Fields.List) is the slice of fields: **\[[Object](#Object)\]**.
	//
	// Currently, the field object is one of following types:
	//
	// - [EnumMember](#Object/EnumMember)
	// - [StructField](#Object/StructField)
	// - [InterfaceMethod](#Object/InterfaceMethod).
	// - [InterfaceMethodParam](#Object/InterfaceMethodParam).
	List []F
}

// @api(Object/EnumMembers) represents the [Fields](#Object/Common/Fields) of [EnumEember](#Object/EnumMember)
// in an [Enum](#Object/Enum) declaration.
type EnumMembers = Fields[*Enum, *EnumMember]

// @api(Object/StructFields) represents the [Fields](#Object/Common/Fields) of [StructField](#Object/StructField)
// in a [Struct](#Object/Struct) declaration.
type StructFields = Fields[*Struct, *StructField]

// @api(Object/InterfaceMethods) represents the [Fields](#Object/Common/Fields) of [InterfaceMethod](#Object/InterfaceMethod)
// in an [Interface](#Object/Interface) declaration.
type InterfaceMethods = Fields[*Interface, *InterfaceMethod]

// @api(Object/InterfaceMethodParams) represents the [Fields](#Object/Common/Fields) of [InterfaceMethodParameter](#Object/InterfaceMethodParam)
// in an [InterfaceMethod](#Object/InterfaceMethod) declaration.
type InterfaceMethodParams = Fields[*InterfaceMethod, *InterfaceMethodParam]

// @api(Object/Enum) (extends [Decl](#Object/Common/Decl)) represents an enum declaration.
type Enum struct {
	*commonNode[*Enum]

	// @api(Object/Enum.MemberType) represents the [PrimitiveType](#Object/PrimitiveType) of the enum members.
	MemberType *PrimitiveType

	// @api(Object/Enum.Type) is the [EnumType](#Object/EnumType) of the enum.
	Type *EnumType

	// @api(Object/Enum.Members) is the [Fields](#Object/Common/Fields) of [EnumMember](#Object/EnumMember).
	Members *EnumMembers
}

func newEnum(c *Compiler, file *File, src *ast.GenDecl[*ast.EnumType]) *Enum {
	e := &Enum{}
	file.addObject(c, src, e)
	e.commonNode = newCommonNode(e, file, src.Pos(), src.Name, src.Doc, src.Annotations)
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
				c.addErrorf(m.Value().namePos, "incompatible type %s for enum member %s", m.Value().Type().name, m.Name())
				return
			}
		}
	}
}

// @api(Object/EnumMember) (extends [Decl](#Object/Common/Decl)) represents an enum member object in an [Enum](#Object/Enum) declaration.
type EnumMember struct {
	*commonNode[*EnumMember]

	// value is the enum member value.
	value *Value

	// @api(Object/EnumMember.Decl) represents the [Enum](#Object/Enum) that contains the member.
	Decl *Enum

	// @api(Object/EnumMember.Comment) represents the line [Comment](#Object/Comment) of the enum member declaration.
	Comment *Comment
}

func newEnumMember(c *Compiler, file *File, e *Enum, src *ast.EnumMember, index int) *EnumMember {
	m := &EnumMember{
		Comment: newComment(src.Comment),
		Decl:    e,
	}
	file.addObject(c, src, m)
	m.commonNode = newCommonNode(m, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	m.value = newValue(c, file, src.Name, src.Value)
	m.value.enum.decl = e
	m.value.enum.index = index
	return m
}

func (m *EnumMember) resolve(c *Compiler, file *File, scope Scope) {
	m.commonNode.resolve(c, file, scope)
	m.value.resolve(c, file, scope)
}

// @api(Object/EnumMember.Index) represents the index of the enum member in the enum type.
func (m *EnumMember) Index() int {
	return m.value.enum.index
}

// @api(Object/EnumMember.Value) represents the [Value](#Object/Value) object of the enum member.
func (m *EnumMember) Value() *Value {
	return m.value
}

// @api(Object/Value.IsFirst) reports whether the value is the first member of the enum type.
//
// Example:
//
//	```next
//	enum Color {
//	    Red = 1; // IsFirst is true for Red
//	    Green = 2;
//	    Blue = 3;
//	}
//	```
func (m *EnumMember) IsFirst() bool {
	return m.value.enum.decl != nil && m.value.enum.index == 0
}

// @api(Object/Value.IsLast) reports whether the value is the last member of the enum type.
//
// Example:
//
//	```next
//	enum Color {
//	    Red = 1;
//	    Green = 2;
//	    Blue = 3; // IsLast is true for Blue
//	}
//	```
func (m *EnumMember) IsLast() bool {
	return m.value.enum.decl != nil && m.value.enum.index == len(m.value.enum.decl.Members.List)-1
}

// @api(Object/Struct) (extends [Decl](#Object/Common/Decl)) represents a struct declaration.
type Struct struct {
	*commonNode[*Struct]

	// lang is the current language to generate the struct.
	lang string

	// fields is the list of struct fields.
	fields *StructFields

	// @api(Object/Struct.Type) represents [StructType](#Object/StructType) of the struct.
	Type *StructType
}

func newStruct(c *Compiler, file *File, src *ast.GenDecl[*ast.StructType]) *Struct {
	s := &Struct{}
	file.addObject(c, src, s)
	s.commonNode = newCommonNode(s, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	s.Type = newDeclType(file, src.Pos(), KindStruct, src.Name.Name, s)
	s.fields = &StructFields{Decl: s}
	for i, f := range src.Spec.Fields.List {
		s.fields.List = append(s.fields.List, newStructField(c, file, s, f, i))
	}
	return s
}

func (s *Struct) resolve(c *Compiler, file *File, scope Scope) {
	s.commonNode.resolve(c, file, scope)
	for _, f := range s.fields.List {
		f.resolve(c, file, scope)
	}
}

// @api(Object/Struct.Fields) represents the [Fields](#Object/Common/Fields) of [StructField](#Object/StructField).
//
// Example:
//
//	```next
//	struct Point {
//	    int x;
//	    int y;
//	}
//	```
//
//	```npl
//	{{range .Fields.List}}
//	{{.Name}} {{.Type}}
//	{{end}}
//	```
//
// Output:
//
//	```
//	x int
//	y int
//	```
func (s *Struct) Fields() *StructFields {
	return availableFields(s.file.compiler, s.fields, s.lang)
}

// @api(Object/StructField) (extends [Node](#Object/Common/Node)) represents a struct field declaration.
type StructField struct {
	*commonNode[*StructField]

	unresolved struct {
		typ ast.Type
	}
	index int

	// @api(Object/StructField.Decl) represents the [Struct](#Object/Struct) that contains the field.
	Decl *Struct

	// @api(Object/StructField.Type) represents the [Type](#Object/Common/Type) of the struct field.
	Type Type

	// @api(Object/StructField.Comment) represents the line [Comment](#Object/Comment) of the struct field declaration.
	Comment *Comment
}

func newStructField(c *Compiler, file *File, s *Struct, src *ast.StructField, index int) *StructField {
	f := &StructField{Decl: s, index: index}
	file.addObject(c, src, f)
	f.commonNode = newCommonNode(f, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	f.unresolved.typ = src.Type
	return f
}

func (f *StructField) resolve(c *Compiler, file *File, scope Scope) {
	f.commonNode.resolve(c, file, scope)
	f.Type = c.resolveType(file, f.unresolved.typ, false)
}

// @api(Object/StructField.Index) represents the index of the struct field in the struct type.
//
// Example:
//
//	```next
//	struct Point {
//	    int x; // Index is 0 for x
//	    int y; // Index is 1 for y
//	}
//	```
func (f *StructField) Index() int {
	return f.index
}

// @api(Object/StructField.IsFirst) reports whether the field is the first field in the struct type.
//
// Example:
//
//	```next
//	struct Point {
//	    int x; // IsFirst is true for x
//	    int y;
//	}
//	```
func (f *StructField) IsFirst() bool {
	return f.index == 0
}

// @api(Object/StructField.IsLast) reports whether the field is the last field in the struct type.
//
// Example:
//
//	```next
//	struct Point {
//	    int x;
//	    int y; // IsLast is true for y
//	}
//	```
func (f *StructField) IsLast() bool {
	return f.index == len(f.Decl.Fields().List)-1
}

// @api(Object/Interface) (extends [Decl](#Object/Common/Decl)) represents an interface declaration.
type Interface struct {
	*commonNode[*Interface]

	// lang is the current language to generate the interface.
	lang string

	// methods is the list of interface methods.
	methods *InterfaceMethods

	// @api(Object/Interface.Type) represents [InterfaceType](#Object/InterfaceType) of the interface.
	Type *InterfaceType
}

func newInterface(c *Compiler, file *File, src *ast.GenDecl[*ast.InterfaceType]) *Interface {
	x := &Interface{}
	file.addObject(c, src, x)
	x.commonNode = newCommonNode(x, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	x.Type = newDeclType(file, src.Pos(), KindInterface, src.Name.Name, x)
	x.methods = &InterfaceMethods{Decl: x}
	for i, m := range src.Spec.Methods.List {
		x.methods.List = append(x.methods.List, newInterfaceMethod(c, file, x, m, i))
	}
	return x
}

func (i *Interface) resolve(c *Compiler, file *File, scope Scope) {
	i.commonNode.resolve(c, file, scope)
	for _, m := range i.methods.List {
		m.resolve(c, file, scope)
	}
}

// @api(Object/Interface.Methods) represents the list of interface methods.
func (i *Interface) Methods() *InterfaceMethods {
	return availableFields(i.file.compiler, i.methods, i.lang)
}

// @api(Object/InterfaceMethod) (extends [Node](#Object/Common/Node)) represents an interface method declaration.
type InterfaceMethod struct {
	*commonNode[*InterfaceMethod]
	index int

	// @api(Object/InterfaceMethod.Decl) represents the interface that contains the method.
	Decl *Interface

	// @api(Object/InterfaceMethod.Params) represents the [Fields](#Object/Common/Fields) of [InterfaceMethodParam](#Object/InterfaceMethodParam).
	Params *InterfaceMethodParams

	// @api(Object/InterfaceMethod.Result) represents the [InterfaceMethodResult](#Object/InterfaceMethodResult) of the method.
	Result *InterfaceMethodResult

	// @api(Object/InterfaceMethod.Comment) represents the line [Comment](#Object/Comment) of the interface method declaration.
	Comment *Comment
}

func newInterfaceMethod(c *Compiler, file *File, i *Interface, src *ast.Method, index int) *InterfaceMethod {
	m := &InterfaceMethod{
		index:   index,
		Decl:    i,
		Comment: newComment(src.Comment),
	}
	file.addObject(c, src, m)
	m.commonNode = newCommonNode(m, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	m.Params = &InterfaceMethodParams{Decl: m}
	for i, p := range src.Params.List {
		m.Params.List = append(m.Params.List, newInterfaceMethodParam(c, file, m, p, i))
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

// @api(Object/InterfaceMethod.Index) represents the index of the interface method in the interface.
//
// Example:
//
//	```next
//	interface Shape {
//	    draw(); // Index is 0 for draw
//	    area() float64; // Index is 1 for area
//	}
//	```
func (m *InterfaceMethod) Index() int {
	return m.index
}

// @api(Object/InterfaceMethod.IsFirst) reports whether the method is the first method in the interface.
//
// Example:
//
//	```next
//	interface Shape {
//	    draw(); // IsFirst is true for draw
//	    area() float64;
//	}
//	```
func (m *InterfaceMethod) IsFirst() bool {
	return m.index == 0
}

// @api(Object/InterfaceMethod.IsLast) reports whether the method is the last method in the interface type.
//
// Example:
//
//	```next
//	interface Shape {
//	    draw();
//	    area() float64; // IsLast is true for area
//	}
//	```
func (m *InterfaceMethod) IsLast() bool {
	return m.index == len(m.Decl.Methods().List)-1
}

// @api(Object/InterfaceMethodParam) (extends [Node](#Object/Common/Node)) represents an interface method parameter declaration.
type InterfaceMethodParam struct {
	*commonNode[*InterfaceMethodParam]
	index int

	unresolved struct {
		typ ast.Type
	}

	// @api(Object/InterfaceMethodParam.Method) represents the [InterfaceMethod](#Object/InterfaceMethod) that contains the parameter.
	Method *InterfaceMethod

	// @api(Object/InterfaceMethodParam.Type) represents the [Type](#Object/Common/Type) of the parameter.
	Type Type
}

func newInterfaceMethodParam(c *Compiler, file *File, m *InterfaceMethod, src *ast.MethodParam, index int) *InterfaceMethodParam {
	p := &InterfaceMethodParam{
		Method: m,
		index:  index,
	}
	file.addObject(c, src, p)
	p.commonNode = newCommonNode(p, file, src.Pos(), src.Name, src.Doc, src.Annotations)
	p.unresolved.typ = src.Type
	return p
}

func (p *InterfaceMethodParam) resolve(c *Compiler, file *File, scope Scope) {
	p.commonNode.resolve(c, file, scope)
	p.Type = c.resolveType(file, p.unresolved.typ, false)
}

// @api(Object/InterfaceMethodParam.Index) represents the index of the interface method parameter in the method.
//
// Example:
//
//	```next
//	interface Shape {
//	    draw(int x, int y); // Index is 0 for x, 1 for y
//	}
//	```
func (p *InterfaceMethodParam) Index() int {
	return p.index
}

// @api(Object/InterfaceMethodParam.IsFirst) reports whether the parameter is the first parameter in the method.
//
// Example:
//
//	```next
//	interface Shape {
//	    draw(int x, int y); // IsFirst is true for x
//	}
//	```
func (p *InterfaceMethodParam) IsFirst() bool {
	return p.index == 0
}

// @api(Object/InterfaceMethodParam.IsLast) reports whether the parameter is the last parameter in the method.
// Example:
//
//	```next
//	interface Shape {
//	    draw(int x, int y); // IsLast is true for y
//	}
//	```
func (p *InterfaceMethodParam) IsLast() bool {
	return p.index == len(p.Method.Params.List)-1
}

// @api(Object/InterfaceMethodResult) represents an interface method result.
type InterfaceMethodResult struct {
	unresolved struct {
		typ ast.Type
	}

	// @api(Object/InterfaceMethodResult.Method) represents the [InterfaceMethod](#Object/InterfaceMethod) that contains the result.
	Method *InterfaceMethod

	// @api(Object/InterfaceMethodResult.Type) represents the underlying [Type](#Object/Common/Type) of the result.
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
func isAvailable(c *Compiler, decl Node, lang string) bool {
	a := decl.Annotations().get("next")
	s, ok := a.get("available").(string)
	if !ok {
		return true
	}
	pos := a.ValuePos("available")
	expr, err := parser.ParseExpr(s)
	if err != nil {
		panic(fmt.Sprintf("%s: invalid @next(available) expression", pos))
	}
	return evalAvailableExpr(c, pos.pos, expr, lang)
}

func evalAvailableExpr(c *Compiler, pos token.Pos, expr ast.Expr, lang string) bool {
	expr = ast.Unparen(expr)
	switch expr := expr.(type) {
	case *ast.Ident:
		if expr.Name == "true" {
			return true
		} else if expr.Name == "false" {
			return false
		}
		return expr.Name == lang
	case *ast.UnaryExpr:
		if expr.Op == token.NOT {
			return !evalAvailableExpr(c, pos, expr.X, lang)
		}
		panic(c.fset.Position(pos+expr.OpPos).String() + ": unexpected unary operator: " + expr.Op.String())
	case *ast.BinaryExpr:
		switch expr.Op {
		case token.OR:
			return evalAvailableExpr(c, pos, expr.X, lang) || evalAvailableExpr(c, pos, expr.Y, lang)
		case token.AND:
			return evalAvailableExpr(c, pos, expr.X, lang) && evalAvailableExpr(c, pos, expr.Y, lang)
		case token.EQL:
			return evalAvailableExpr(c, pos, expr.X, lang) == evalAvailableExpr(c, pos, expr.Y, lang)
		case token.NEQ:
			return evalAvailableExpr(c, pos, expr.X, lang) != evalAvailableExpr(c, pos, expr.Y, lang)
		default:
			panic(c.fset.Position(pos+expr.OpPos).String() + ": unexpected binary operator: " + expr.Op.String())
		}
	default:
		panic(c.fset.Position(pos+expr.Pos()).String() + ": unexpected expression")
	}
}

// available returns the declaration if it is available in the current language.
func available[T Node](c *Compiler, obj T, lang string) (T, bool) {
	if !isAvailable(c, obj, lang) {
		return obj, false
	}
	switch decl := any(obj).(type) {
	case *Package:
		decl.decls.lang = lang
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
func availableFields[D, F Node](c *Compiler, fields *Fields[D, F], lang string) *Fields[D, F] {
	for i, f := range fields.List {
		if isAvailable(c, f, lang) {
			continue
		}
		list := make([]F, 0, len(fields.List))
		list = append(list, fields.List[:i]...)
		for j := i + 1; j < len(fields.List); j++ {
			if isAvailable(c, fields.List[j], lang) {
				list = append(list, fields.List[j])
			}
		}
		return &Fields[D, F]{Decl: fields.Decl, List: list}
	}
	return fields
}

// availableList returns the list of declarations that are available in the current language.
func availableList[T Node](c *Compiler, list List[T], lang string) List[T] {
	availables := make([]T, 0, len(list))
	for i, d := range list {
		if _, ok := available(c, d, lang); ok {
			availables = append(availables, list[i])
		}
	}
	return availables
}
