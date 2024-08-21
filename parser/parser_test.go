package parser

import (
	"testing"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

func TestParseSingle(t *testing.T) {
	const filename = "./testdata/single/demo.next"
	const src = `package demo;

// integer constant
const X = 1;
const V1 = 1000_1000;
const V2 = 100;
const V3 = V1 + v2;

// string constant
const Y = "y";

// constant group
const (
    A = 1;
    B = 2.0;
    C = false;
    D = "hello";
)

// Point
@json(name="Point")
struct Point {
	@json(name="x", omitempty=true)
	int x;
	@json(omitempty=true)
	int y;
}

@type(100)
protocol User {
	@key
	int id;
}

// color enum
enum Color {
    Red = 1,
    Green = 2,
    Blue = 3,
}
`
	fset := token.NewFileSet()
	f, err := ParseFile(fset, filename, nil, ParseComments|AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	ast.Print(fset, f)
}
