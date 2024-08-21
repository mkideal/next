package parser

import (
	"fmt"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/token"
)

// The parser structure holds the parser's internal state.
type parser struct {
	file    *token.File
	errors  scanner.ErrorList
	scanner scanner.Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode&Trace != 0)
	indent int  // indentation used for tracing output

	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment
	top         bool              // in top of file (before package clause)

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	// Error recovery
	// (used to limit the number of calls to parser.advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos token.Pos // last synchronization position
	syncCnt int       // number of parser.advance calls without progress

	// Non-syntactic parser control
	exprLev int // < 0: in control clause, >= 0: in expression

	imports []*ast.ImportSpec // list of imports

	// nestLev is used to track and limit the recursion depth
	// during parsing.
	nestLev int
}

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(p.file, src, eh, scanner.ScanComments)

	p.top = true
	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)
	p.next()
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *parser) printTrace(a ...any) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.file.Position(p.pos)
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

// maxNestLev is the deepest we're willing to recurse during parsing
const maxNestLev int = 1e5

func incNestLev(p *parser) *parser {
	p.nestLev++
	if p.nestLev > maxNestLev {
		p.error(p.pos, "exceeded max nesting depth")
		panic(bailout{})
	}
	return p
}

// decNestLev is used to track nesting depth during parsing to prevent stack exhaustion.
// It is used along with incNestLev in a similar fashion to how un and trace are used.
func decNestLev(p *parser) {
	p.nestLev--
}

// Advance to the next token.
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is token.ILLEGAL), so don't print it.
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator(), p.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	for {
		p.pos, p.tok, p.lit = p.scanner.Scan()
		if p.tok == token.COMMENT {
			if p.mode&ParseComments == 0 {
				continue
			}
		} else {
			// Found a non-comment; top of file is over.
			p.top = false
		}
		break
	}
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{Slash: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos
	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == token.SEMICOLON || p.tok == token.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

// A bailout panic is raised to indicate early termination. pos and msg are
// only populated when bailing out of object resolution.
type bailout struct {
	pos token.Pos
	msg string
}

func (p *parser) error(pos token.Pos, msg string) {
	if p.trace {
		defer un(trace(p, "error: "+msg))
	}

	epos := p.file.Position(pos)

	// If AllErrors is not set, discard errors reported on the same line
	// as the last recorded error and stop parsing if there are more than
	// 10 errors.
	if p.mode&AllErrors == 0 {
		n := len(p.errors)
		if n > 0 && p.errors[n-1].Pos.Line == epos.Line {
			return // discard - likely a spurious error
		}
		if n > 10 {
			panic(bailout{})
		}
	}

	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok == token.SEMICOLON && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

// expect2 is like expect, but it returns an invalid position
// if the expected token is not found.
func (p *parser) expect2(tok token.Token) (pos token.Pos) {
	if p.tok == tok {
		pos = p.pos
	} else {
		p.errorExpected(p.pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return
}

// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
func (p *parser) expectClosing(tok token.Token, context string) token.Pos {
	if p.tok != tok && p.tok == token.SEMICOLON && p.lit == "\n" {
		p.error(p.pos, "missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(tok)
}

// expectSemi consumes a semicolon and returns the applicable line comment.
func (p *parser) expectSemi() (comment *ast.CommentGroup) {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		switch p.tok {
		case token.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.SEMICOLON:
			if p.lit == ";" {
				// explicit semicolon
				p.next()
				comment = p.lineComment // use following comments
			} else {
				// artificial semicolon
				comment = p.lineComment // use preceding comments
				p.next()
			}
			return comment
		default:
			p.errorExpected(p.pos, "';'")
		}
	}
	return nil
}

func (p *parser) atComma(context string, follow token.Token) bool {
	if p.tok == token.COMMA {
		return true
	}
	if p.tok != follow {
		msg := "missing ','"
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += " before newline"
		}
		p.error(p.pos, msg+" in "+context)
		return true // "insert" comma and continue
	}
	return false
}

func assert(cond bool, msg string) {
	if !cond {
		panic("go/parser internal error: " + msg)
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or token.EOF. For error recovery.
func (p *parser) advance(to map[token.Token]bool) {
	for ; p.tok != token.EOF; p.next() {
		if to[p.tok] {
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 advance calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop (it is possible that
			// both parseOperand and parseStmt call advance and
			// correctly do not advance, thus the need for the
			// invocation limit p.syncCnt).
			if p.pos == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if p.pos > p.syncPos {
				p.syncPos = p.pos
				p.syncCnt = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		}
	}
}

var declStart = map[token.Token]bool{
	token.IMPORT:   true,
	token.CONST:    true,
	token.STRUCT:   true,
	token.PROTOCOL: true,
	token.ENUM:     true,
}

var exprEnd = map[token.Token]bool{
	token.COMMA:     true,
	token.SEMICOLON: true,
	token.RPAREN:    true,
	token.RBRACK:    true,
	token.RBRACE:    true,
}

// safePos returns a valid file position for a given position: If pos
// is valid to begin with, safePos returns pos. If pos is out-of-range,
// safePos returns the EOF position.
//
// This is hack to work around "artificial" end positions in the AST which
// are computed by adding 1 to (presumably valid) token positions. If the
// token positions are invalid due to parse errors, the resulting end position
// may be past the file's EOF position, which would lead to panics if used
// later on.
func (p *parser) safePos(pos token.Pos) (res token.Pos) {
	defer func() {
		if recover() != nil {
			res = token.Pos(p.file.Base() + p.file.Size()) // EOF position
		}
	}()
	_ = p.file.Offset(pos) // trigger a panic if position is out-of-range
	return pos
}

// ----------------------------------------------------------------------------
// Annotation parsing

func (p *parser) parseAnnotationGroup() *ast.AnnotationGroup {
	var annotations []*ast.Annotation
	for p.tok == token.AT {
		annotations = append(annotations, p.parseAnnotation())
	}
	if len(annotations) == 0 {
		return nil
	}
	return &ast.AnnotationGroup{List: annotations}
}

func (p *parser) parseAnnotation() *ast.Annotation {
	pos := p.expect(token.AT)
	name := p.parseIdent()
	var params []*ast.AnnotationParam
	var lparen token.Pos
	var rparen token.Pos
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		if p.tok != token.RPAREN {
			for _, expr := range p.parseExprList(true, true) {
				expr = ast.Unparen(expr)
				switch x := expr.(type) {
				case *ast.Ident:
					params = append(params, &ast.AnnotationParam{Name: x})
				case *ast.BinaryExpr:
					if x.Op == token.ASSIGN {
						if ident, ok := x.X.(*ast.Ident); ok {
							params = append(params, &ast.AnnotationParam{
								Name:      ident,
								AssignPos: x.OpPos,
								Value:     x.Y,
							})
						} else {
							p.error(x.Pos(), "expected identifier")
						}
					} else {
						params = append(params, &ast.AnnotationParam{Value: x})
					}
				default:
					params = append(params, &ast.AnnotationParam{Value: x})
				}
			}
		}
		rparen = p.expect(token.RPAREN)
	}
	return &ast.Annotation{
		At:     pos,
		Name:   name,
		Lparen: lparen,
		Params: params,
		Rparen: rparen,
	}
}

// ----------------------------------------------------------------------------
// Identifiers

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

// ----------------------------------------------------------------------------
// Common productions

// If lhs is set, result list elements which are identifiers are not resolved.
func (p *parser) parseExprList(assignAllowed, cmpAllowed bool) (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ExpressionList"))
	}

	list = append(list, p.parseExpr(assignAllowed, cmpAllowed))
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseExpr(assignAllowed, cmpAllowed))
	}

	return
}

// ----------------------------------------------------------------------------
// Types

func (p *parser) parseType() ast.Type {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	var typ ast.Type
	pos := p.pos
	switch p.tok {
	case token.IDENT:
		typ = p.parseTypeName(nil)
	case token.MAP:
		p.next()
		p.expect(token.LSS)
		k := p.parseType()
		p.expect(token.COMMA)
		v := p.parseType()
		gt := p.pos
		p.expect(token.GTR)
		typ = &ast.MapType{
			Map: pos,
			K:   k,
			V:   v,
			GT:  gt,
		}
	case token.ARRAY:
		p.next()
		p.expect(token.LSS)
		t := p.parseType()
		p.expect(token.COMMA)
		n := p.parseExpr(false, false)
		gt := p.pos
		p.expect(token.GTR)
		typ = &ast.ArrayType{
			Array: pos,
			T:     t,
			N:     n,
			GT:    gt,
		}
	case token.VECTOR:
		p.next()
		p.expect(token.LSS)
		t := p.parseType()
		gt := p.pos
		p.expect(token.GTR)
		typ = &ast.VectorType{
			Vector: pos,
			T:      t,
			GT:     gt,
		}
	default:
		p.errorExpected(pos, "type")
		p.advance(exprEnd)
		typ = &ast.BadExpr{From: pos, To: p.pos}
	}

	return typ
}

// If the result is an identifier, it is not resolved.
func (p *parser) parseTypeName(ident *ast.Ident) ast.Type {
	if p.trace {
		defer un(trace(p, "TypeName"))
	}

	if ident == nil {
		ident = p.parseIdent()
	}

	if p.tok == token.PERIOD {
		// ident is a package name
		p.next()
		sel := p.parseIdent()
		return &ast.SelectorExpr{X: ident, Sel: sel}
	}

	return ident
}

func (p *parser) parseFieldDecl() *ast.Field {
	if p.trace {
		defer un(trace(p, "FieldDecl"))
	}

	doc := p.leadComment
	annotations := p.parseAnnotationGroup()
	typ := p.parseType()
	name := p.parseIdent()
	comment := p.expectSemi()
	field := &ast.Field{
		Doc:         doc,
		Annotations: annotations,
		Name:        name,
		Type:        typ,
		Comment:     comment,
	}
	return field
}

func (p *parser) parseStructType() *ast.StructType {
	if p.trace {
		defer un(trace(p, "StructType"))
	}

	lbrace := p.expect(token.LBRACE)
	var list []*ast.Field
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseFieldDecl())
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.StructType{
		Fields: &ast.FieldList{
			Opening: lbrace,
			List:    list,
			Closing: rbrace,
		},
	}
}

func (p *parser) parseEnumType() *ast.EnumType {
	if p.trace {
		defer un(trace(p, "StructType"))
	}

	lbrace := p.expect(token.LBRACE)
	var values []*ast.ValueSpec
	for iota := 0; p.tok != token.RBRACE && p.tok != token.EOF; iota++ {
		annotations := p.parseAnnotationGroup()
		values = append(values, p.parseValueSpec(p.leadComment, annotations, token.ENUM, iota).(*ast.ValueSpec))
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.EnumType{
		Opening: lbrace,
		Values:  values,
		Closing: rbrace,
	}
}

// ----------------------------------------------------------------------------
// Expressions

// parseOperand may return an expression or a raw type (incl. array
// types of the form [...]T). Callers must verify the result.
func (p *parser) parseOperand(assignAllowed, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Operand"))
	}

	switch p.tok {
	case token.IDENT:
		x := p.parseIdent()
		return x

	case token.INT, token.FLOAT, token.CHAR, token.STRING:
		x := &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
		p.next()
		return x

	case token.LPAREN:
		lparen := p.pos
		p.next()
		p.exprLev++
		x := p.parseExpr(assignAllowed, cmpAllowed) // types may be parenthesized: (some type)
		p.exprLev--
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}
	}

	// we have an error
	pos := p.pos
	p.errorExpected(pos, "operand")
	return &ast.BadExpr{From: pos, To: p.pos}
}

func (p *parser) parseSelector(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "Selector"))
	}

	sel := p.parseIdent()

	return &ast.SelectorExpr{X: x, Sel: sel}
}

func (p *parser) parseCall(fun ast.Expr) *ast.CallExpr {
	if p.trace {
		defer un(trace(p, "Call"))
	}

	lparen := p.expect(token.LPAREN)
	p.exprLev++
	var list []ast.Expr
	var ellipsis token.Pos
	for p.tok != token.RPAREN && p.tok != token.EOF && !ellipsis.IsValid() {
		list = append(list, p.parseExpr(true, true)) // builtins may expect a type: make(some type, ...)
		if !p.atComma("argument list", token.RPAREN) {
			break
		}
		p.next()
	}
	p.exprLev--
	rparen := p.expectClosing(token.RPAREN, "argument list")

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Rparen: rparen}
}

func (p *parser) parseValue(assignAllowed, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Element"))
	}

	return p.parseExpr(assignAllowed, cmpAllowed)
}

func (p *parser) parsePrimaryExpr(x ast.Expr, assignAllowed, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "PrimaryExpr"))
	}

	if x == nil {
		x = p.parseOperand(assignAllowed, cmpAllowed)
	}
	// We track the nesting here rather than at the entry for the function,
	// since it can iteratively produce a nested output, and we want to
	// limit how deep a structure we generate.
	var n int
	defer func() { p.nestLev -= n }()
	for n = 1; ; n++ {
		incNestLev(p)
		switch p.tok {
		case token.PERIOD:
			p.next()
			switch p.tok {
			case token.IDENT:
				x = p.parseSelector(x)
			default:
				pos := p.pos
				p.errorExpected(pos, "selector or type assertion")
				if p.tok != token.RBRACE {
					p.next() // make progress
				}
				sel := &ast.Ident{NamePos: pos, Name: "_"}
				x = &ast.SelectorExpr{X: x, Sel: sel}
			}
		case token.LPAREN:
			x = p.parseCall(x)
		default:
			return x
		}
	}
}

func (p *parser) parseUnaryExpr(assignAllowed, cmpAllowed bool) ast.Expr {
	defer decNestLev(incNestLev(p))

	if p.trace {
		defer un(trace(p, "UnaryExpr"))
	}

	switch p.tok {
	case token.ADD, token.SUB, token.NOT, token.XOR, token.AND:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr(assignAllowed, cmpAllowed)
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: x}
	}

	return p.parsePrimaryExpr(nil, assignAllowed, cmpAllowed)
}

func (p *parser) tokPrec() (token.Token, int) {
	tok := p.tok
	return tok, tok.Precedence()
}

// parseBinaryExpr parses a (possibly) binary expression.
// If x is non-nil, it is used as the left operand.
//
// TODO(rfindley): parseBinaryExpr has become overloaded. Consider refactoring.
func (p *parser) parseBinaryExpr(x ast.Expr, prec1 int, assignAllowed, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "BinaryExpr"))
	}

	if x == nil {
		x = p.parseUnaryExpr(assignAllowed, cmpAllowed)
	}
	// We track the nesting here rather than at the entry for the function,
	// since it can iteratively produce a nested output, and we want to
	// limit how deep a structure we generate.
	var n int
	defer func() { p.nestLev -= n }()
	for n = 1; ; n++ {
		incNestLev(p)
		op, oprec := p.tokPrec()
		if oprec < prec1 {
			return x
		}
		if op == token.ASSIGN {
			if !assignAllowed {
				return x
			}
			if _, ok := x.(*ast.Ident); !ok {
				p.error(x.Pos(), "assignment to non-assignable")
			}

		}
		if (op == token.LSS || op == token.GTR) && !cmpAllowed {
			return x
		}
		pos := p.expect(op)
		y := p.parseBinaryExpr(nil, oprec+1, assignAllowed, cmpAllowed)
		x = &ast.BinaryExpr{X: x, OpPos: pos, Op: op, Y: y}
	}
}

// The result may be a type or even a raw type ([...]int).
func (p *parser) parseExpr(assignAllowed, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Expression"))
	}

	return p.parseBinaryExpr(nil, token.LowestPrec+1, assignAllowed, cmpAllowed)
}

func (p *parser) parseCallExpr(callType string) *ast.CallExpr {
	x := p.parseExpr(false, true) // could be a conversion: (some type)(x)
	if t := ast.Unparen(x); t != x {
		p.error(x.Pos(), fmt.Sprintf("expression in %s must not be parenthesized", callType))
		x = t
	}
	if call, isCall := x.(*ast.CallExpr); isCall {
		return call
	}
	if _, isBad := x.(*ast.BadExpr); !isBad {
		// only report error if it's a new one
		p.error(p.safePos(x.End()), fmt.Sprintf("expression in %s must be function call", callType))
	}
	return nil
}

// ----------------------------------------------------------------------------
// Statements

func (p *parser) parseStmt() ast.Stmt {
	return p.parseExprStmt()
}

func (p *parser) parseExprStmt() *ast.ExprStmt {
	if p.trace {
		defer un(trace(p, "ExpressionStatement"))
	}
	expr := p.parseExpr(true, true)
	p.expectSemi()
	return &ast.ExprStmt{X: expr}
}

// ----------------------------------------------------------------------------
// Declarations

type parseSpecFunction func(doc *ast.CommentGroup, annotations *ast.AnnotationGroup, keyword token.Token, iota int) ast.Spec

func (p *parser) parseImportSpec(doc *ast.CommentGroup, annotations *ast.AnnotationGroup, _ token.Token, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "ImportSpec"))
	}

	pos := p.pos
	var path string
	if p.tok == token.STRING {
		path = p.lit
		p.next()
	} else if p.tok.IsLiteral() {
		p.error(pos, "import path must be a string")
		p.next()
	} else {
		p.error(pos, "missing import path")
		p.advance(exprEnd)
	}
	comment := p.expectSemi()

	// collect imports
	spec := &ast.ImportSpec{
		Doc:         doc,
		Annotations: annotations,
		Path:        &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment:     comment,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func (p *parser) parseValueSpec(doc *ast.CommentGroup, annotations *ast.AnnotationGroup, keyword token.Token, iota int) ast.Spec {
	if p.trace {
		defer un(trace(p, keyword.String()+"Spec"))
	}

	ident := p.parseIdent()
	var value ast.Expr
	var comment *ast.CommentGroup
	switch keyword {
	case token.CONST:
		p.expect(token.ASSIGN)
		value = p.parseExpr(false, true)
		comment = p.expectSemi()
	case token.ENUM:
		if p.tok == token.ASSIGN {
			p.next()
			value = p.parseExpr(false, true)
		}
		p.expect(token.COMMA)
		comment = p.lineComment
	default:
		panic("unreachable")
	}

	return &ast.ValueSpec{
		Doc:         doc,
		Annotations: annotations,
		Name:        ident,
		Value:       value,
		Comment:     comment,
	}
}

func (p *parser) parseTypeSpec(doc *ast.CommentGroup, annotations *ast.AnnotationGroup, keyword token.Token, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}
	name := p.parseIdent()

	var typ ast.Type
	switch keyword {
	case token.STRUCT, token.PROTOCOL:
		typ = p.parseStructType()
	case token.ENUM:
		typ = p.parseEnumType()
	default:
		p.errorExpected(p.pos, "Enum, Struct or Protocol")
	}

	return &ast.TypeSpec{
		Doc:         doc,
		Annotations: annotations,
		Keyword:     keyword,
		Name:        name,
		Type:        typ,
	}
}

// isTypeElem reports whether x is a (possibly parenthesized) type element expression.
// The result is false if x could be a type element OR an ordinary (value) expression.
func (p *parser) parseGenDecl(annotations *ast.AnnotationGroup, keyword token.Token, f parseSpecFunction) *ast.GenDecl {
	if p.trace {
		defer un(trace(p, "GenDecl("+keyword.String()+")"))
	}

	doc := p.leadComment
	pos := p.expect(keyword)
	var lparen, rparen token.Pos
	var list []ast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			annotations := p.parseAnnotationGroup()
			list = append(list, f(p.leadComment, annotations, keyword, iota))
		}
		rparen = p.expect(token.RPAREN)
	} else {
		list = append(list, f(nil, nil, keyword, 0))
	}

	return &ast.GenDecl{
		Doc:         doc,
		Annotations: annotations,
		TokPos:      pos,
		Tok:         keyword,
		Lparen:      lparen,
		Specs:       list,
		Rparen:      rparen,
	}
}

func (p *parser) parseDeclStmt(_ map[token.Token]bool) ast.Node {
	if p.trace {
		defer un(trace(p, "Declaration"))
	}

	annotations := p.parseAnnotationGroup()

	var f parseSpecFunction
	switch p.tok {
	case token.IMPORT:
		f = p.parseImportSpec

	case token.CONST:
		f = p.parseValueSpec

	case token.ENUM, token.STRUCT, token.PROTOCOL:
		f = p.parseTypeSpec

	default:
		return p.parseStmt()
	}

	return p.parseGenDecl(annotations, p.tok, f)
}

// ----------------------------------------------------------------------------
// Source files

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}

	// Don't bother parsing the rest if we had errors scanning the first token.
	// Likely not a Go source file at all.
	if p.errors.Len() != 0 {
		return nil
	}

	// package clause
	doc := p.leadComment
	annotations := p.parseAnnotationGroup()
	pos := p.expect(token.PACKAGE)
	// Go spec: The package clause is not a declaration;
	// the package name does not appear in any scope.
	ident := p.parseIdent()
	if ident.Name == "_" && p.mode&DeclarationErrors != 0 {
		p.error(p.pos, "invalid package name _")
	}
	p.expectSemi()

	// Don't bother parsing the rest if we had errors parsing the package clause.
	// Likely not a Go source file at all.
	if p.errors.Len() != 0 {
		return nil
	}

	var decls []ast.Decl
	var stmts []ast.Stmt
	if p.mode&PackageClauseOnly == 0 {
		// import decls
		for p.tok == token.IMPORT {
			annotations := p.parseAnnotationGroup()
			decls = append(decls, p.parseGenDecl(annotations, token.IMPORT, p.parseImportSpec))
		}

		if p.mode&ImportsOnly == 0 {
			// rest of package body
			prev := token.IMPORT
			for p.tok != token.EOF {
				// Continue to accept import declarations for error tolerance, but complain.
				if p.tok == token.IMPORT && prev != token.IMPORT {
					p.error(p.pos, "imports must appear before other declarations")
				}
				prev = p.tok

				node := p.parseDeclStmt(declStart)
				switch node := node.(type) {
				case ast.Decl:
					decls = append(decls, node)
				case ast.Stmt:
					stmts = append(stmts, node)
				}
			}
		}
	}

	return &ast.File{
		Doc:         doc,
		Annotations: annotations,
		Package:     pos,
		Name:        ident,
		Decls:       decls,
		Stmts:       stmts,
		FileStart:   token.Pos(p.file.Base()),
		FileEnd:     token.Pos(p.file.Base() + p.file.Size()),
		Imports:     p.imports,
		Comments:    p.comments,
	}
}
