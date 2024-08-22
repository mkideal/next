package parser

import (
	"testing"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

func TestParseSingle(t *testing.T) {
	const filename = "./testdata/single/demo.next"
	const src = `package demo;

enum Day {
	Weekday = 1,
	Weekend = 2,
}

const Complex3 = (~Day.Weekend);
`
	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, ParseComments|AllErrors|Trace)
	if err != nil {
		t.Fatal(err)
	}
	ast.Print(fset, f)
}
