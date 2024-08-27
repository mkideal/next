package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

// Decl represents a declaration: import, constant, enum, struct
type Decl struct {
	pos        token.Pos
	tok        token.Token
	file       *File
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Doc         *CommentGroup
	Annotations *AnnotationGroup
	Specs       []Spec
}

func newDecl(ctx *Context, file *File, src *ast.GenDecl) *Decl {
	d := &Decl{
		file: file,
		pos:  src.Pos(),
		tok:  src.Tok,
		Doc:  newCommentGroup(src.Doc),
	}
	d.unresolved.annotations = src.Annotations
	for _, s := range src.Specs {
		d.Specs = append(d.Specs, newSpec(ctx, file, d, s))
	}
	return d
}

func (d *Decl) resolve(ctx *Context, file *File, scope Scope) {
	d.Annotations = ctx.resolveAnnotationGroup(file, d.unresolved.annotations)
	for _, s := range d.Specs {
		s.resolve(ctx, file, scope)
	}
}
