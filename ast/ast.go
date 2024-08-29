// Package ast declares the types used to represent syntax trees for Go
// packages.
package ast

import (
	"strings"

	"github.com/gopherd/next/token"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Go spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

// All node types implement the Node interface.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All type nodes implement the Type interface.
type Type interface {
	Node
	typeNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single //-style or /*-style comment.
//
// The Text field contains the comment text without carriage returns (\r) that
// may have been present in the source. Because a comment's end position is
// computed using len(Text), the position reported by [Comment.End] does not match the
// true source end position for comments containing carriage returns.
type Comment struct {
	Slash token.Pos // position of "/" starting the comment
	Text  string    // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Pos() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() token.Pos { return g.List[0].Pos() }
func (g *CommentGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed.
// Multiple empty lines are reduced to one, and trailing space on lines is trimmed.
// Unless the result is empty, it is newline-terminated.
func (g *CommentGroup) Text() string {
	return strings.Join(g.TrimComments(), "\n")
}

func (g *CommentGroup) TrimComments() []string {
	if g == nil || len(g.List) == 0 {
		return nil
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}
	comments = TrimComments(comments)
	// Add final "" entry to get trailing newline from Join.
	if len(comments) > 0 && comments[len(comments)-1] != "" {
		comments = append(comments, "")
	}
	return comments
}

func TrimComments(comments []string) []string {
	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '/':
			//-style comment (no newline at the end)
			c = c[2:]
			if len(c) == 0 {
				// empty line
				break
			}
			if c[0] == ' ' {
				// strip first space - required for Example tests
				c = c[1:]
				break
			}
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	return lines[0:n]
}

type NamedParam struct {
	Name      *Ident    // name of parameter
	AssignPos token.Pos // position of "=" if any
	Value     Expr      // parameter value, or nil
}

func (p *NamedParam) Pos() token.Pos {
	if p.Name != nil {
		return p.Name.Pos()
	}
	return p.Value.Pos()
}

func (p *NamedParam) End() token.Pos {
	if p.Value != nil {
		return p.Value.End()
	}
	return p.Name.End()
}

// An Annotation represents an annotation.
type Annotation struct {
	At     token.Pos     // position of "@"
	Name   *Ident        // annotation name
	Lparen token.Pos     // position of "(" if any
	Params []*NamedParam // annotation parameters; or nil
	Rparen token.Pos     // position of ")" if any
}

func (a *Annotation) Pos() token.Pos {
	return a.At
}

func (a *Annotation) End() token.Pos {
	if a.Rparen.IsValid() {
		return a.Rparen + 1
	}
	if n := len(a.Params); n > 0 {
		return a.Params[n-1].End()
	}
	return a.Name.End()
}

type AnnotationGroup struct {
	List []*Annotation // len(List) > 0
}

func (g *AnnotationGroup) Pos() token.Pos { return g.List[0].Pos() }
func (g *AnnotationGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

// ----------------------------------------------------------------------------
// Expressions and types

type List[T Node] struct {
	Opening token.Pos // position of opening parenthesis/brace/bracket, if any
	List    []T       // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace/bracket, if any
}

func (l *List[T]) Pos() token.Pos {
	if l.Opening.IsValid() {
		return l.Opening
	}
	if len(l.List) > 0 {
		return l.List[0].Pos()
	}
	return token.NoPos
}

func (l *List[T]) End() token.Pos {
	if l.Closing.IsValid() {
		return l.Closing + 1
	}
	if n := len(l.List); n > 0 {
		return l.List[n-1].End()
	}
	return token.NoPos
}

// NumFields returns the number of parameters or struct fields represented by a [List].
func (l *List[T]) NumFields() int {
	if l != nil {
		return len(l.List)
	}
	return 0
}

// A Field represents a Field declaration list in a struct type,
// a method list in an interface type, or a parameter/result declaration
// in a signature.
// [Field.Names] is nil for unnamed parameters (parameter lists which only contain types)
// and embedded struct fields. In the latter case, the field name is the type name.
type Field struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Type        Type             // field/method/parameter type; or nil
	Name        *Ident           // field/method/(type) parameter names; or nil
	Comment     *CommentGroup    // line comments; or nil
}

func (f *Field) Pos() token.Pos {
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	return f.Name.Pos()
}

// A FieldList represents a list of Fields, enclosed by parentheses,
// curly braces, or square brackets.
type FieldList = List[*Field]

// A ValueList represents a list of ValueSpecs, enclosed by parentheses,
type ValueList = List[*ValueSpec]

// Method represents a method declaration.
//
// Example:
//
// MethodName(Type1 param1, Type2 param2) Type3
type Method struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Name        *Ident           // method name
	Type        *FuncType        // method type
	Comment     *CommentGroup    // line comments; or nil
}

func (m *Method) Pos() token.Pos {
	return m.Name.Pos()
}

func (m *Method) End() token.Pos {
	return m.Type.End()
}

// A MethodList represents a list of Method declarations.
type MethodList = List[*Method]

// FuncType represents a function type.
type FuncType struct {
	Params     *MethodParamList // method parameters
	ReturnType Type             // method return type; or nil
}

func (ft *FuncType) Pos() token.Pos {
	return ft.Params.Pos()
}

func (ft *FuncType) End() token.Pos {
	if ft.ReturnType != nil {
		return ft.ReturnType.End()
	}
	return ft.Params.End()
}

// A ValueSpec node represents a constant or variable declaration
type MethodParam struct {
	Type Type   // parameter type
	Name *Ident // parameter name
}

func (p *MethodParam) Pos() token.Pos {
	return p.Type.Pos()
}

func (p *MethodParam) End() token.Pos {
	return p.Name.End()
}

// MethodParamList represents a list of function parameters.
type MethodParamList = List[*MethodParam]

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
type (
	// A BadExpr node is a placeholder for an expression containing
	// syntax errors for which a correct expression node cannot be
	// created.
	//
	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
	}

	// A BasicLit node represents a literal of basic type.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen token.Pos // position of "("
		X      Expr      // parenthesized expression
		Rparen token.Pos // position of ")"
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X   Expr   // expression
		Sel *Ident // field selector
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun    Expr      // function expression
		Lparen token.Pos // position of "("
		Args   []Expr    // function arguments; or nil
		Rparen token.Pos // position of ")"
	}

	// A UnaryExpr node represents a unary expression.
	// Unary "*" expressions are represented via StarExpr nodes.
	//
	UnaryExpr struct {
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		X     Expr        // operand
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr        // left operand
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		Y     Expr        // right operand
	}
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
type (
	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Array token.Pos // position of token.ARRAY
		T     Type      // element type
		N     Expr      // size of the array
		GT    token.Pos // position of ">"
	}

	// A VectorType node represents a vector type.
	VectorType struct {
		Vector token.Pos // position of token.VECTOR
		T      Type      // element type
		GT     token.Pos // position of ">"
	}

	// A MapType node represents a map type.
	MapType struct {
		Map token.Pos // position of token.MAP
		K   Type      // key type
		V   Type      // value type
		GT  token.Pos // position of ">"
	}

	// An InterfaceType node represents an interface type.
	InterfaceType struct {
		Methods *MethodList // list of methods
	}

	// A StructType node represents a struct type.
	StructType struct {
		Fields *FieldList // list of field declarations
	}

	// A EnumType node represents an enum type.
	EnumType struct {
		Members *ValueList // list of enum members
	}
)

// Pos and End implementations for expression/type nodes.

func (x *BadExpr) Pos() token.Pos       { return x.From }
func (x *Ident) Pos() token.Pos         { return x.NamePos }
func (x *BasicLit) Pos() token.Pos      { return x.ValuePos }
func (x *ParenExpr) Pos() token.Pos     { return x.Lparen }
func (x *SelectorExpr) Pos() token.Pos  { return x.X.Pos() }
func (x *CallExpr) Pos() token.Pos      { return x.Fun.Pos() }
func (x *UnaryExpr) Pos() token.Pos     { return x.OpPos }
func (x *BinaryExpr) Pos() token.Pos    { return x.X.Pos() }
func (x *ArrayType) Pos() token.Pos     { return x.Array }
func (x *VectorType) Pos() token.Pos    { return x.Vector }
func (x *MapType) Pos() token.Pos       { return x.Map }
func (x *EnumType) Pos() token.Pos      { return x.Members.Pos() }
func (x *StructType) Pos() token.Pos    { return x.Fields.Pos() }
func (x *InterfaceType) Pos() token.Pos { return x.Methods.Pos() }

func (x *BadExpr) End() token.Pos       { return x.To }
func (x *Ident) End() token.Pos         { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() token.Pos      { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *ParenExpr) End() token.Pos     { return x.Rparen + 1 }
func (x *SelectorExpr) End() token.Pos  { return x.Sel.End() }
func (x *CallExpr) End() token.Pos      { return x.Rparen + 1 }
func (x *UnaryExpr) End() token.Pos     { return x.X.End() }
func (x *BinaryExpr) End() token.Pos    { return x.Y.End() }
func (x *ArrayType) End() token.Pos     { return x.GT }
func (x *VectorType) End() token.Pos    { return x.GT }
func (x *MapType) End() token.Pos       { return x.GT }
func (x *EnumType) End() token.Pos      { return x.Members.End() }
func (x *StructType) End() token.Pos    { return x.Fields.End() }
func (x *InterfaceType) End() token.Pos { return x.Methods.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an Expr.
func (*BadExpr) exprNode()      {}
func (*Ident) exprNode()        {}
func (*BasicLit) exprNode()     {}
func (*ParenExpr) exprNode()    {}
func (*SelectorExpr) exprNode() {}
func (*CallExpr) exprNode()     {}
func (*UnaryExpr) exprNode()    {}
func (*BinaryExpr) exprNode()   {}

func (*ArrayType) exprNode()     {}
func (*VectorType) exprNode()    {}
func (*MapType) exprNode()       {}
func (*EnumType) exprNode()      {}
func (*StructType) exprNode()    {}
func (*InterfaceType) exprNode() {}

func (*BadExpr) typeNode()       {}
func (*SelectorExpr) typeNode()  {}
func (*Ident) typeNode()         {}
func (*ArrayType) typeNode()     {}
func (*VectorType) typeNode()    {}
func (*MapType) typeNode()       {}
func (*EnumType) typeNode()      {}
func (*StructType) typeNode()    {}
func (*InterfaceType) typeNode() {}

// ----------------------------------------------------------------------------
// Convenience functions for Idents

func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// ----------------------------------------------------------------------------
// Statements

type (
	BadStmt struct {
		From, To token.Pos // position range of bad statement
	}

	// ExprStmt represents a call declaration.
	ExprStmt struct {
		X Expr // expression
	}
)

func (s *BadStmt) Pos() token.Pos  { return s.From }
func (s *ExprStmt) Pos() token.Pos { return s.X.Pos() }

func (s *BadStmt) End() token.Pos  { return s.To }
func (s *ExprStmt) End() token.Pos { return s.X.End() }

// stmtNode() ensures that only statement nodes can be
// assigned to a Stmt.
func (*BadStmt) stmtNode()  {}
func (*ExprStmt) stmtNode() {}

// ----------------------------------------------------------------------------
// Declarations

// A Spec node represents a single (non-parenthesized) import,
// constant, type (enum, struct) declaration.
type (
	// The Spec type stands for any of *ImportSpec, *ValueSpec, and *TypeSpec.
	Spec interface {
		Node
		specNode()
	}

	// An ImportSpec node represents a single package import.
	ImportSpec struct {
		Doc         *CommentGroup    // associated documentation; or nil
		Annotations *AnnotationGroup // associated annotations; or nil
		Path        *BasicLit        // import path
		Comment     *CommentGroup    // line comments; or nil
		EndPos      token.Pos        // end of spec (overrides Path.Pos if nonzero)
	}

	// A ValueSpec node represents a constant or variable declaration
	// (ConstSpec or VarSpec production).
	//
	ValueSpec struct {
		Doc         *CommentGroup    // associated documentation; or nil
		Annotations *AnnotationGroup // associated annotations; or nil
		Name        *Ident           // value name
		Value       Expr             // value; or nil
		Comment     *CommentGroup    // line comments; or nil
	}

	// A TypeSpec node represents a type declaration (TypeSpec production).
	TypeSpec struct {
		Doc         *CommentGroup    // associated documentation; or nil
		Annotations *AnnotationGroup // associated annotations; or nil
		Keyword     token.Token      // ENUM, STRUCT
		Name        *Ident           // type name
		Type        Type             // any of the *XxxTypes
		Comment     *CommentGroup    // line comments; or nil
	}
)

// Pos and End implementations for spec nodes.

func (s *ImportSpec) Pos() token.Pos {
	return s.Path.Pos()
}
func (s *ValueSpec) Pos() token.Pos { return s.Name.Pos() }
func (s *TypeSpec) Pos() token.Pos  { return s.Name.Pos() }

func (s *ImportSpec) End() token.Pos {
	if s.EndPos != 0 {
		return s.EndPos
	}
	return s.Path.End()
}

func (s *ValueSpec) End() token.Pos {
	if s.Value != nil {
		return s.Value.End()
	}
	return s.Name.End()
}
func (s *TypeSpec) End() token.Pos { return s.Type.End() }

// specNode() ensures that only spec nodes can be
// assigned to a Spec.
func (*ImportSpec) specNode() {}
func (*ValueSpec) specNode()  {}
func (*TypeSpec) specNode()   {}

// A declaration is represented by one of the following declaration nodes.
type (
	// A BadDecl node is a placeholder for a declaration containing
	// syntax errors for which a correct declaration node cannot be
	// created.
	//
	BadDecl struct {
		From, To token.Pos // position range of bad declaration
	}

	// A GenDecl node (generic declaration node) represents an import,
	// constant, type or variable declaration. A valid Lparen position
	// (Lparen.IsValid()) indicates a parenthesized declaration.
	//
	// Relationship between Tok value and Specs element type:
	//
	//	token.IMPORT         *ImportSpec
	//	token.CONST          *ValueSpec
	//	token.ENUM/STRUCT    *TypeSpec
	//
	GenDecl struct {
		Doc         *CommentGroup    // associated documentation; or nil
		Annotations *AnnotationGroup // associated annotations; or nil
		TokPos      token.Pos        // position of Tok
		Tok         token.Token      // IMPORT, CONST, ENUM, STRUCT
		Lparen      token.Pos        // position of '(', if any
		Specs       []Spec
		Rparen      token.Pos // position of ')', if any
	}
)

// Pos and End implementations for declaration nodes.

func (d *BadDecl) Pos() token.Pos { return d.From }
func (d *GenDecl) Pos() token.Pos { return d.TokPos }

func (d *BadDecl) End() token.Pos { return d.To }
func (d *GenDecl) End() token.Pos {
	if d.Rparen.IsValid() {
		return d.Rparen + 1
	}
	return d.Specs[0].End()
}

// declNode() ensures that only declaration nodes can be
// assigned to a Decl.
func (*BadDecl) declNode() {}
func (*GenDecl) declNode() {}

// ----------------------------------------------------------------------------
// Files and packages

// A File node represents a Go source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.
//
// For correct printing of source code containing comments (using packages
// go/format and go/printer), special care must be taken to update comments
// when a File's syntax tree is modified: For printing, comments are interspersed
// between tokens based on their position. If syntax tree nodes are
// removed or moved, relevant comments in their vicinity must also be removed
// (from the [File.Comments] list) or moved accordingly (by updating their
// positions). A [CommentMap] may be used to facilitate some of these operations.
//
// Whether and how a comment is associated with a node depends on the
// interpretation of the syntax tree by the manipulating program: except for Doc
// and [Comment] comments directly associated with nodes, the remaining comments
// are "free-floating"
type File struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Package     token.Pos        // position of "package" keyword
	Name        *Ident           // package name
	Decls       []Decl           // top-level declarations; or nil
	Stmts       []Stmt           // top-level statements; or nil

	FileStart, FileEnd token.Pos       // start and end of entire file
	Imports            []*ImportSpec   // imports in this file
	Comments           []*CommentGroup // list of all comments in the source file
}

// Pos returns the position of the package declaration.
// (Use FileStart for the start of the entire file.)
func (f *File) Pos() token.Pos { return f.Package }

// End returns the end of the last declaration in the file.
// (Use FileEnd for the end of the entire file.)
func (f *File) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.Name.End()
}

// IsGenerated reports whether the file was generated by a program,
// not handwritten, by detecting the special comment described
// at https://go.dev/s/generatedcode.
//
// The syntax tree must have been parsed with the [parser.ParseComments] flag.
// Example:
//
//	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments|parser.PackageClauseOnly)
//	if err != nil { ... }
//	gen := ast.IsGenerated(f)
func IsGenerated(file *File) bool {
	_, ok := generator(file)
	return ok
}

func generator(file *File) (string, bool) {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if comment.Pos() > file.Package {
				break // after package declaration
			}
			// opt: check Contains first to avoid unnecessary array allocation in Split.
			const prefix = "// Code generated "
			if strings.Contains(comment.Text, prefix) {
				for _, line := range strings.Split(comment.Text, "\n") {
					if rest, ok := strings.CutPrefix(line, prefix); ok {
						if gen, ok := strings.CutSuffix(rest, " DO NOT EDIT."); ok {
							return gen, true
						}
					}
				}
			}
		}
	}
	return "", false
}

// Unparen returns the expression with any enclosing parentheses removed.
func Unparen(e Expr) Expr {
	for {
		paren, ok := e.(*ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}
