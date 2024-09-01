// Package parser implements a parser for Next source files.
// Input may be provided in a variety of forms (see the various Parse* functions);
// the output is an abstract syntax tree (AST) representing the Next source.
package parser

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// readSource converts src to a []byte if possible, or reads from the file specified by filename if src is nil.
// It returns the resulting []byte and any error encountered.
func readSource(filename string, src interface{}) ([]byte, error) {
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
func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
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
