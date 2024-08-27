package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	src := `package main

// Comment 1
// Comment 2
const X = 1

// Comment 3
// Comment 4
const (
    // Comment 5
    // Comment 6
    Y = 2
)
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				if x.Doc != nil {
					fmt.Printf("Const declaration doc: %s\n", x.Doc.Text())
				}
				for _, spec := range x.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						if vs.Doc != nil {
							fmt.Printf("Const value doc: %s\n", vs.Doc.Text())
						}
						fmt.Printf("Const name: %s\n", vs.Names[0].Name)
					}
				}
			}
		}
		return true
	})
}
