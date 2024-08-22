package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

// Decl represents a declaration: import, constant, enum, struct, protocol
type Decl struct {
	pos token.Pos

	Tok   token.Token
	Specs []Spec
}

func newDecl(ctx *Context, file *File, src *ast.GenDecl) *Decl {
	d := &Decl{
		pos: src.Pos(),
		Tok: src.Tok,
	}
	for _, s := range src.Specs {
		d.Specs = append(d.Specs, newSpec(ctx, file, d, s))
	}
	return d
}

func (d *Decl) Pos() token.Pos { return d.pos }

func (d *Decl) resolve(ctx *Context, file *File, scope Scope) {
	for _, s := range d.Specs {
		s.resolve(ctx, file, scope)
	}
}
