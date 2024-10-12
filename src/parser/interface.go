// Package parser implements a parser for Next source files.
// Input may be provided in a variety of forms (see the various Parse* functions);
// the output is an abstract syntax tree (AST) representing the Next source.
package parser

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/mkideal/next/src/ast"
	"github.com/mkideal/next/src/token"
)

// readSource converts src to a []byte if possible, or reads from the file specified by filename if src is nil.
// It returns the resulting []byte and any error encountered.
func readSource(filename string, src any) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return io.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return os.ReadFile(filename)
}

// Mode is a set of flags controlling parser behavior.
type Mode uint

const (
	ParseComments Mode = 1 << iota // Parse comments and add them to AST
	Trace                          // Print a trace of parsed productions
	AllErrors                      // Report all errors (not just the first 10 on different lines)
)

// ParseFile parses a single Next source file and returns the corresponding ast.File node.
// It takes the following parameters:
//   - fset: A token.FileSet for position information (must not be nil)
//   - filename: The filename of the source (used for recording positions)
//   - src: The source code (string, []byte, io.Reader, or nil to read from filename)
//   - mode: Parser mode flags
//
// If parsing fails, it returns a nil ast.File and an error. For syntax errors,
// it returns a partial AST and a scanner.ErrorList sorted by source position.
func ParseFile(fset *token.FileSet, filename string, src any, mode Mode) (f *ast.File, err error) {
	if fset == nil {
		panic("parser.ParseFile: no token.FileSet provided (fset == nil)")
	}

	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		if f == nil {
			f = &ast.File{
				Name: new(ast.Ident),
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	p.init(fset, filename, text, mode)
	f = p.parseFile()

	return
}

// ParseExprFrom is a convenience function for parsing an expression.
// The arguments have the same meaning as for [ParseFile], but the source must
// be a valid Next (type or value) expression. Specifically, fset must not
// be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with [ast.Bad]* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by source position.
func ParseExprFrom(fset *token.FileSet, filename string, src any, mode Mode) (expr ast.Expr, err error) {
	if fset == nil {
		panic("parser.ParseExprFrom: no token.FileSet provided (fset == nil)")
	}

	// get source
	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			bail, ok := e.(bailout)
			if !ok {
				panic(e)
			} else if bail.msg != "" {
				p.errors.Add(p.file.Position(bail.pos), bail.msg)
			}
		}
		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse expr
	p.init(fset, filename, text, mode)
	expr = p.parseRhs()

	// If a semicolon was inserted, consume it;
	// report an error if there's more tokens.
	if p.tok == token.SEMICOLON && p.lit == "\n" {
		p.next()
	}
	p.expect(token.EOF)

	return
}

// ParseExpr is a convenience function for obtaining the AST of an expression x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
//
// If syntax errors were found, the result is a partial AST (with [ast.Bad]* nodes
// representing the fragments of erroneous source code). Multiple errors are
// returned via a scanner.ErrorList which is sorted by source position.
func ParseExpr(x string) (ast.Expr, error) {
	return ParseExprFrom(token.NewFileSet(), "", []byte(x), 0)
}
