// Package ast declares the types used to represent syntax trees for Next source code.
package ast

import (
	"strings"

	"github.com/next/next/src/token"
)

// Node represents any node in the abstract syntax tree.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// Expr represents any expression node in the abstract syntax tree.
type Expr interface {
	Node
	exprNode()
}

// Stmt represents any statement node in the abstract syntax tree.
type Stmt interface {
	Node
	stmtNode()
}

// Type represents any type node in the abstract syntax tree.
type Type interface {
	Node
	typeNode()
}

// Comment represents a single //-style or /*-style comment.
type Comment struct {
	Slash token.Pos // position of "/" starting the comment
	Text  string    // comment text (excluding '\n' for //-style comments)
}

// Pos returns the position of the first character in the comment.
func (c *Comment) Pos() token.Pos { return c.Slash }

// End returns the position of the character immediately after the comment.
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

// CommentGroup represents a sequence of comments with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

// Pos returns the position of the first character in the first comment in the group.
func (g *CommentGroup) Pos() token.Pos { return g.List[0].Pos() }

// End returns the position of the character immediately after the last comment in the group.
func (g *CommentGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment group.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed.
// Multiple empty lines are reduced to one, and trailing space on lines is trimmed.
// Unless the result is empty, it is newline-terminated.
func (g *CommentGroup) Text() string {
	return strings.Join(g.TrimComments(), "\n")
}

// TrimComments returns the trimmed text of the comments in the group.
func (g *CommentGroup) TrimComments() []string {
	if g == nil || len(g.List) == 0 {
		return nil
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}
	comments = TrimComments(comments)
	if len(comments) > 0 && comments[len(comments)-1] != "" {
		comments = append(comments, "")
	}
	return comments
}

// TrimComments removes comment markers and trims whitespace from a slice of comment strings.
func TrimComments(comments []string) []string {
	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		switch c[1] {
		case '/':
			c = c[2:]
			if len(c) == 0 {
				break
			}
			if c[0] == ' ' {
				c = c[1:]
				break
			}
		case '*':
			c = c[2 : len(c)-2]
		}

		// Split on newlines and strip trailing whitespace.
		cl := strings.Split(c, "\n")
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

type AnnotationParam struct {
	Name      *Ident    // name of parameter
	AssignPos token.Pos // position of "=" if any
	Value     Node      // parameter value (Expr or Type), or nil
}

// Pos returns the position of the first character in the annotation parameter.
func (p *AnnotationParam) Pos() token.Pos {
	return p.Name.Pos()
}

// End returns the position of the character immediately after the annotation parameter.
func (p *AnnotationParam) End() token.Pos {
	if p.Value != nil {
		return p.Value.End()
	}
	return p.Name.End()
}

// Annotation represents an annotation in the source code.
type Annotation struct {
	At     token.Pos          // position of "@"
	Name   *Ident             // annotation name
	Lparen token.Pos          // position of "(" if any
	Params []*AnnotationParam // annotation parameters; or nil
	Rparen token.Pos          // position of ")" if any
}

// Pos returns the position of the first character in the annotation.
func (a *Annotation) Pos() token.Pos {
	return a.At
}

// End returns the position of the character immediately after the annotation.
func (a *Annotation) End() token.Pos {
	if a.Rparen.IsValid() {
		return a.Rparen + 1
	}
	if n := len(a.Params); n > 0 {
		return a.Params[n-1].End()
	}
	return a.Name.End()
}

// AnnotationGroup represents a group of annotations.
type AnnotationGroup struct {
	List []*Annotation // len(List) > 0
}

// Pos returns the position of the first character in the first annotation of the group.
func (g *AnnotationGroup) Pos() token.Pos { return g.List[0].Pos() }

// End returns the position of the character immediately after the last annotation in the group.
func (g *AnnotationGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

// List represents a generic list of nodes.
type List[T Node] struct {
	Opening token.Pos // position of opening parenthesis/brace/bracket, if any
	List    []T       // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace/bracket, if any
}

// Pos returns the position of the first character in the list.
func (l *List[T]) Pos() token.Pos {
	if l.Opening.IsValid() {
		return l.Opening
	}
	if len(l.List) > 0 {
		return l.List[0].Pos()
	}
	return token.NoPos
}

// End returns the position of the character immediately after the list.
func (l *List[T]) End() token.Pos {
	if l.Closing.IsValid() {
		return l.Closing + 1
	}
	if n := len(l.List); n > 0 {
		return l.List[n-1].End()
	}
	return token.NoPos
}

// NumFields returns the number of fields in the list.
func (l *List[T]) NumFields() int {
	if l != nil {
		return len(l.List)
	}
	return 0
}

// EnumMember represents an enum member declaration.
type EnumMember struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Name        *Ident           // value name
	AssignPos   token.Pos        // position of "=" if any
	Value       Expr             // value; or nil
	Comment     *CommentGroup    // line comments; or nil
}

// Pos returns the position of the first character in the enum member.
func (m *EnumMember) Pos() token.Pos {
	return m.Name.Pos()
}

// End returns the position of the character immediately after the enum member.
func (m *EnumMember) End() token.Pos {
	if m.Value != nil {
		return m.Value.End()
	}
	return m.Name.End()
}

// MemberList represents a list of enum members.
type MemberList = List[*EnumMember]

// StructField represents a field in a struct type
type StructField struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Type        Type             // field/method/parameter type; or nil
	Name        *Ident           // field/method/(type) parameter names; or nil
	Comment     *CommentGroup    // line comments; or nil
}

// Pos returns the position of the first character in the struct field.
func (f *StructField) Pos() token.Pos {
	return f.Type.Pos()
}

// End returns the position of the character immediately after the struct field.
func (f *StructField) End() token.Pos {
	return f.Name.Pos()
}

// FieldList represents a list of fields, enclosed by parentheses, curly braces, or square brackets.
type FieldList = List[*StructField]

// Method represents a method declaration.
type Method struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Name        *Ident           // method name
	Params      *MethodParamList // method parameters
	Return      Type             // method return type; or nil
	Comment     *CommentGroup    // line comments; or nil
}

// Pos returns the position of the first character in the method declaration.
func (m *Method) Pos() token.Pos {
	return m.Name.Pos()
}

// End returns the position of the character immediately after the method declaration.
func (m *Method) End() token.Pos {
	if m.Return != nil {
		return m.Return.End()
	}
	return m.Params.End()
}

// MethodList represents a list of method declarations.
type MethodList = List[*Method]

// MethodParam represents a method parameter declaration.
type MethodParam struct {
	Annotations *AnnotationGroup // associated annotations; or nil
	Type        Type             // parameter type
	Name        *Ident           // parameter name
}

// Pos returns the position of the first character in the method parameter.
func (p *MethodParam) Pos() token.Pos {
	return p.Type.Pos()
}

// End returns the position of the character immediately after the method parameter.
func (p *MethodParam) End() token.Pos {
	return p.Name.End()
}

// MethodParamList represents a list of function parameters.
type MethodParamList = List[*MethodParam]

// BadExpr represents an incorrect or unparseable expression.
type BadExpr struct {
	From, To token.Pos // position range of bad expression
}

// Ident represents an identifier.
type Ident struct {
	NamePos token.Pos // identifier position
	Name    string    // identifier name
}

// BasicLit represents a literal of basic type.
type BasicLit struct {
	ValuePos token.Pos   // literal position
	Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
	Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
}

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	Lparen token.Pos // position of "("
	X      Expr      // parenthesized expression
	Rparen token.Pos // position of ")"
}

// SelectorExpr represents an expression followed by a selector.
type SelectorExpr struct {
	X   Expr   // expression
	Sel *Ident // field selector
}

// CallExpr represents an expression followed by an argument list.
type CallExpr struct {
	Fun    Expr      // function expression
	Lparen token.Pos // position of "("
	Args   []Expr    // function arguments; or nil
	Rparen token.Pos // position of ")"
}

// UnaryExpr represents a unary expression.
// Unary "*" expressions are represented via StarExpr nodes.
type UnaryExpr struct {
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	X     Expr        // operand
}

// BinaryExpr represents a binary expression.
type BinaryExpr struct {
	X     Expr        // left operand
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	Y     Expr        // right operand
}

// ArrayType represents an array or slice type.
type ArrayType struct {
	Array token.Pos // position of token.ARRAY
	T     Type      // element type
	N     Expr      // size of the array
	GT    token.Pos // position of ">"
}

// VectorType represents a vector type.
type VectorType struct {
	Vector token.Pos // position of token.VECTOR
	T      Type      // element type
	GT     token.Pos // position of ">"
}

// MapType represents a map type.
type MapType struct {
	Map token.Pos // position of token.MAP
	K   Type      // key type
	V   Type      // value type
	GT  token.Pos // position of ">"
}

// InterfaceType represents an interface type.
type InterfaceType struct {
	Methods *MethodList // list of methods
}

// StructType represents a struct type.
type StructType struct {
	Fields *FieldList // list of field declarations
}

// EnumType represents an enum type.
type EnumType struct {
	Members *MemberList // list of enum members
}

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

// String returns the identifier name.
func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// BadStmt represents an invalid statement.
type BadStmt struct {
	From, To token.Pos // position range of bad statement
}

// ExprStmt represents a standalone expression in a statement list.
type ExprStmt struct {
	X Expr // expression
}

func (s *BadStmt) Pos() token.Pos  { return s.From }
func (s *ExprStmt) Pos() token.Pos { return s.X.Pos() }

func (s *BadStmt) End() token.Pos  { return s.To }
func (s *ExprStmt) End() token.Pos { return s.X.End() }

// stmtNode() ensures that only statement nodes can be
// assigned to a Stmt.
func (*BadStmt) stmtNode()  {}
func (*ExprStmt) stmtNode() {}

// Decl represents a declaration in the Next source code.
type Decl interface {
	Node
	Token() token.Token
	declNode()
}

// ImportDecl represents a single package import.
type ImportDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Path    *BasicLit     // import path
	Comment *CommentGroup // line comments; or nil
	EndPos  token.Pos     // end of spec (overrides Path.Pos if nonzero)
}

// GenDecl represents a general declaration node.
type GenDecl[T Node] struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Tok         token.Token      // CONST, ENUM, STRUCT or INTERFACE
	TokPos      token.Pos        // position of Tok
	Name        *Ident           // value name
	Spec        T                // spec of the declaration
	Comment     *CommentGroup    // line comments const declarations; or nil
}

// ConstDecl represents a constant declaration.
type ConstDecl = GenDecl[Expr]

// EnumDecl represents an enum declaration.
type EnumDecl = GenDecl[*EnumType]

// StructDecl represents a struct declaration.
type StructDecl = GenDecl[*StructType]

// InterfaceDecl represents an interface declaration.
type InterfaceDecl = GenDecl[*InterfaceType]

// Pos and End implementations for decl nodes.
func (s *ImportDecl) Pos() token.Pos { return s.Path.Pos() }
func (s *GenDecl[T]) Pos() token.Pos { return s.Name.Pos() }

func (s *ImportDecl) End() token.Pos {
	if s.EndPos != 0 {
		return s.EndPos
	}
	return s.Path.End()
}

func (s *GenDecl[T]) End() token.Pos {
	return s.Spec.End()
}

func (x *ImportDecl) Token() token.Token { return token.IMPORT }
func (x *GenDecl[T]) Token() token.Token { return x.Tok }

func (*ImportDecl) declNode() {}
func (*GenDecl[T]) declNode() {}

// File represents a Next source file.
type File struct {
	Doc         *CommentGroup    // associated documentation; or nil
	Annotations *AnnotationGroup // associated annotations; or nil
	Package     token.Pos        // position of "package" keyword
	Name        *Ident           // package name
	Decls       []Decl           // top-level declarations except imports
	Stmts       []Stmt           // top-level statements; or nil

	FileStart, FileEnd token.Pos       // start and end of entire file
	Imports            []*ImportDecl   // imports in this file
	Comments           []*CommentGroup // list of all comments in the source file
}

// Pos returns the position of the package declaration.
func (f *File) Pos() token.Pos { return f.Package }

// End returns the end of the last declaration in the file.
func (f *File) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.Name.End()
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
