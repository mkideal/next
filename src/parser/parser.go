package parser

import (
	"fmt"
	"strconv"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/scanner"
	"github.com/next/next/src/token"
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

	imports []*ast.ImportDecl // list of imports

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

// A bailout panic is raised to indicate early termination.
type bailout struct{}

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

func unexpectedToken(tok token.Token, lit string) string {
	if tok == token.ILLEGAL {
		if lit == "~" {
			return "illegal character U+007E '~', did you mean '^'?"
		}
		return "illegal " + strconv.Quote(lit)
	}
	return "'" + tok.String() + "'"
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
			msg += ", found " + unexpectedToken(p.tok, p.lit)
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

func (p *parser) atToken(context string, tok, follow token.Token) bool {
	if p.tok == tok {
		return true
	}
	if p.tok != follow {
		msg := "missing '" + tok.String() + "'"
		if p.lit == "\n" {
			msg += " before newline"
		}
		p.error(p.pos, msg+" in "+context)
		return true // continue
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

var declStmtStart = map[token.Token]bool{
	token.IMPORT:    true,
	token.CONST:     true,
	token.ENUM:      true,
	token.STRUCT:    true,
	token.INTERFACE: true,
	token.IDENT:     true,
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

func (p *parser) parseAnnotationGroup() (*ast.CommentGroup, *ast.AnnotationGroup) {
	var doc = p.leadComment
	var annotations []*ast.Annotation
	for p.tok == token.AT {
		annotations = append(annotations, p.parseAnnotation())
	}
	if len(annotations) == 0 {
		return doc, nil
	}
	return doc, &ast.AnnotationGroup{List: annotations}
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
		rparen, params = parseList("AnnotationParams", p, token.RPAREN, token.COMMA, false, parseAnnotationParam)
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

func parseList[T ast.Node](name string, p *parser, rparen, sep token.Token, ending bool, parseElem func(*parser) T) (closing token.Pos, list []T) {
	if p.trace {
		defer un(trace(p, name))
	}
	for p.tok != rparen && p.tok != token.EOF {
		list = append(list, parseElem(p))
		if (ending || p.tok != rparen) && p.tok != token.EOF {
			if !p.atToken(name, sep, rparen) {
				break
			}
			p.next()
		}
	}
	if p.tok == rparen {
		closing = p.pos
	}
	p.expect(rparen)
	return
}

// parseNamedParamList parses a list of named parameters separated by sep (COMMA or SEMICOLON).
func parseAnnotationParam(p *parser) *ast.AnnotationParam {
	if p.trace {
		defer un(trace(p, "AnnotationParam"))
	}
	var param = &ast.AnnotationParam{
		Name: p.parseIdent(),
	}
	if p.tok == token.ASSIGN {
		param.AssignPos = p.pos
		p.next()
		switch p.tok {
		case token.ARRAY, token.MAP, token.VECTOR:
			param.Value = p.parseType()
		default:
			param.Value = p.parseExpr(true)
		}
	}
	return param
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
		n := p.parseExpr(false)
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

func parseMethodParam(p *parser) *ast.MethodParam {
	if p.trace {
		defer un(trace(p, "MethodParam"))
	}
	doc, annotations := p.parseAnnotationGroup()
	typ := p.parseType()
	name := p.parseIdent()
	return &ast.MethodParam{
		Doc:         doc,
		Annotations: annotations,
		Type:        typ,
		Name:        name,
	}
}

func (p *parser) parseMethod() *ast.Method {
	if p.trace {
		defer un(trace(p, "Method"))
	}

	doc, annotations := p.parseAnnotationGroup()
	name := p.parseIdent()
	openging := p.expect(token.LPAREN)
	closing, params := parseList("MethodParamList", p, token.RPAREN, token.COMMA, false, parseMethodParam)
	var returnType ast.Type
	if p.tok != token.SEMICOLON {
		returnType = p.parseType()
	}
	comment := p.expectSemi()
	method := &ast.Method{
		Doc:         doc,
		Annotations: annotations,
		Name:        name,
		Params:      &ast.MethodParamList{Opening: openging, List: params, Closing: closing},
		Result:      returnType,
		Comment:     comment,
	}
	return method
}

func (p *parser) parseField() *ast.StructField {
	if p.trace {
		defer un(trace(p, "Field"))
	}

	doc, annotations := p.parseAnnotationGroup()
	typ := p.parseType()
	name := p.parseIdent()
	comment := p.expectSemi()
	field := &ast.StructField{
		Doc:         doc,
		Annotations: annotations,
		Name:        name,
		Type:        typ,
		Comment:     comment,
	}
	return field
}

func parseInterfaceType(p *parser) *ast.InterfaceType {
	if p.trace {
		defer un(trace(p, "InterfaceType"))
	}

	lbrace := p.expect(token.LBRACE)
	var list []*ast.Method
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseMethod())
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.InterfaceType{
		Methods: &ast.MethodList{
			Opening: lbrace,
			List:    list,
			Closing: rbrace,
		},
	}
}

func parseStructType(p *parser) *ast.StructType {
	if p.trace {
		defer un(trace(p, "StructType"))
	}

	lbrace := p.expect(token.LBRACE)
	var list []*ast.StructField
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseField())
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

func (p *parser) parseMember() *ast.EnumMember {
	if p.trace {
		defer un(trace(p, "Member"))
	}

	doc, annotations := p.parseAnnotationGroup()
	name := p.parseIdent()
	var value ast.Expr
	var assignPos token.Pos
	if p.tok == token.ASSIGN {
		assignPos = p.pos
		p.next()
		value = p.parseExpr(true)
	}
	comment := p.expectSemi()
	member := &ast.EnumMember{
		Doc:         doc,
		Annotations: annotations,
		Name:        name,
		AssignPos:   assignPos,
		Value:       value,
		Comment:     comment,
	}
	return member
}

func parseEnumType(p *parser) *ast.EnumType {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}

	lbrace := p.expect(token.LBRACE)
	var values []*ast.EnumMember
	for p.tok != token.RBRACE && p.tok != token.EOF {
		values = append(values, p.parseMember())
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.EnumType{
		Members: &ast.MemberList{
			Opening: lbrace,
			List:    values,
			Closing: rbrace,
		},
	}
}

// ----------------------------------------------------------------------------
// Expressions

// parseOperand may return an expression or a raw type (incl. array
// types of the form [...]T). Callers must verify the result.
func (p *parser) parseOperand(cmpAllowed bool) ast.Expr {
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
		x := p.parseExpr(cmpAllowed) // types may be parenthesized: (some type)
		p.exprLev--
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}
	}

	// we have an error
	pos := p.pos
	p.errorExpected(pos, "operand")
	p.advance(declStmtStart)
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
		list = append(list, p.parseExpr(true)) // builtins may expect a type: make(some type, ...)
		if !p.atToken("argument list", token.COMMA, token.RPAREN) {
			break
		}
		p.next()
	}
	p.exprLev--
	rparen := p.expectClosing(token.RPAREN, "argument list")

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Rparen: rparen}
}

func parseValue(p *parser) ast.Expr {
	if p.trace {
		defer un(trace(p, "ValueExpr"))
	}

	return p.parseExpr(true)
}

func parseConstSpec(p *parser) ast.Expr {
	if p.trace {
		defer un(trace(p, "ConstValueExpr"))
	}
	p.expect(token.ASSIGN)
	return p.parseExpr(false)
}

func (p *parser) parsePrimaryExpr(x ast.Expr, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "PrimaryExpr"))
	}

	if x == nil {
		x = p.parseOperand(cmpAllowed)
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

func (p *parser) parseUnaryExpr(cmpAllowed bool) ast.Expr {
	defer decNestLev(incNestLev(p))

	if p.trace {
		defer un(trace(p, "UnaryExpr"))
	}

	switch p.tok {
	case token.ADD, token.SUB, token.NOT, token.XOR, token.AND:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr(cmpAllowed)
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: x}
	}

	return p.parsePrimaryExpr(nil, cmpAllowed)
}

func (p *parser) tokPrec() (token.Token, int) {
	tok := p.tok
	return tok, tok.Precedence()
}

// parseBinaryExpr parses a (possibly) binary expression.
// If x is non-nil, it is used as the left operand.
func (p *parser) parseBinaryExpr(x ast.Expr, prec1 int, cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "BinaryExpr"))
	}

	if x == nil {
		x = p.parseUnaryExpr(cmpAllowed)
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
		if (op == token.LSS || op == token.GTR) && !cmpAllowed {
			return x
		}
		pos := p.expect(op)
		y := p.parseBinaryExpr(nil, oprec+1, cmpAllowed)
		x = &ast.BinaryExpr{X: x, OpPos: pos, Op: op, Y: y}
	}
}

// The result may be a type or even a raw type ([...]int).
func (p *parser) parseExpr(cmpAllowed bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Expression"))
	}

	return p.parseBinaryExpr(nil, token.LowestPrec+1, cmpAllowed)
}

func (p *parser) parseCallExpr(callType string) *ast.CallExpr {
	x := p.parseExpr(true) // could be a conversion: (some type)(x)
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
	expr := p.parseExpr(true)
	p.expectSemi()
	return &ast.ExprStmt{X: expr}
}

// ----------------------------------------------------------------------------
// Declarations

type parseSpecFunction[T ast.Node] func(p *parser) T

func (p *parser) parseImportDecl() *ast.ImportDecl {
	if p.trace {
		defer un(trace(p, "ImportDecl"))
	}
	importPos := p.pos
	p.expect(token.IMPORT)

	pos := p.pos
	doc := p.leadComment
	var end token.Pos
	var path string
	if p.tok == token.STRING {
		path = p.lit
		end = p.pos
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
	spec := &ast.ImportDecl{
		Doc:       doc,
		ImportPos: importPos,
		Path:      &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment:   comment,
		EndPos:    end,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func parseGenDecl[T ast.Node](p *parser, doc *ast.CommentGroup, annotations *ast.AnnotationGroup, f parseSpecFunction[T]) *ast.GenDecl[T] {
	tok := p.tok
	if p.trace {
		defer un(trace(p, "GenDecl("+tok.String()+")"))
	}

	pos := p.pos
	p.next()
	name := p.parseIdent()
	spec := f(p)
	var comment *ast.CommentGroup
	if tok == token.CONST {
		comment = p.expectSemi()
	}

	return &ast.GenDecl[T]{
		Doc:         doc,
		Annotations: annotations,
		Tok:         tok,
		TokPos:      pos,
		Name:        name,
		Spec:        spec,
		Comment:     comment,
	}
}

func (p *parser) parseDeclStmt() ast.Node {
	if p.trace {
		defer un(trace(p, "DeclarationStmt"))
	}

	doc, annotations := p.parseAnnotationGroup()

	switch p.tok {
	case token.IMPORT:
		return p.parseImportDecl()

	case token.CONST:
		return parseGenDecl(p, doc, annotations, parseConstSpec)

	case token.ENUM:
		return parseGenDecl(p, doc, annotations, parseEnumType)

	case token.STRUCT:
		return parseGenDecl(p, doc, annotations, parseStructType)

	case token.INTERFACE:
		return parseGenDecl(p, doc, annotations, parseInterfaceType)

	default:
		return p.parseStmt()
	}
}

// ----------------------------------------------------------------------------
// Source files

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}

	// Don't bother parsing the rest if we had errors scanning the first token.
	// Likely not a Next source file at all.
	if p.errors.Len() != 0 {
		return nil
	}

	// package clause
	doc, annotations := p.parseAnnotationGroup()
	pos := p.expect(token.PACKAGE)
	ident := p.parseIdent()
	if ident.Name == "_" {
		p.error(p.pos, "invalid package name _")
	}
	p.expectSemi()

	// Don't bother parsing the rest if we had errors parsing the package clause.
	// Likely not a Next source file at all.
	if p.errors.Len() != 0 {
		return nil
	}

	var decls []ast.Decl
	var stmts []ast.Stmt
	var imports []*ast.ImportDecl
	// import decls
	for p.tok == token.IMPORT {
		imports = append(imports, p.parseImportDecl())
	}

	// rest of package body
	prev := token.IMPORT
	for p.tok != token.EOF {
		// Continue to accept import declarations for error tolerance, but complain.
		if p.tok == token.IMPORT && prev != token.IMPORT {
			p.error(p.pos, "imports must appear before other declarations")
		}
		prev = p.tok

		node := p.parseDeclStmt()
		switch node := node.(type) {
		case ast.Decl:
			if _, isImport := node.(*ast.ImportDecl); isImport {
				p.error(node.Pos(), "imports must appear before other declarations")
			}
			decls = append(decls, node)
		case ast.Stmt:
			stmts = append(stmts, node)
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
