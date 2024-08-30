package types

import (
	"cmp"
	"slices"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// Imports represents a list of import specs
type Imports struct {
	List []*ImportSpec
}

func (i *Imports) resolve(ctx *Context, file *File) {
	for _, spec := range i.List {
		spec.importedFile = ctx.lookupFile(file.Path, spec.Path)
		if spec.importedFile == nil {
			ctx.addErrorf(spec.pos, "import file not found: %s", spec.Path)
		}
	}
}

func (i *Imports) ListForPackage() []*ImportSpec {
	var seen = make(map[string]bool)
	var pkgs []*ImportSpec
	for _, spec := range i.List {
		if seen[spec.importedFile.pkg.name] {
			continue
		}
		seen[spec.importedFile.pkg.name] = true
		pkgs = append(pkgs, spec)
	}
	slices.SortFunc(pkgs, func(a, b *ImportSpec) int {
		return cmp.Compare(a.importedFile.pkg.name, b.importedFile.pkg.name)
	})
	return pkgs
}

// Decl represents a declaration: import, constant, enum, struct
type Decl struct {
	pos        token.Pos
	tok        token.Token
	file       *File
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Doc         *Doc
	Annotations AnnotationGroup
	Specs       []Spec
}

func newDecl(ctx *Context, file *File, src *ast.GenDecl) *Decl {
	d := &Decl{
		file: file,
		pos:  src.Pos(),
		tok:  src.Tok,
		Doc:  newDoc(src.Doc),
	}
	d.unresolved.annotations = src.Annotations
	for _, s := range src.Specs {
		d.Specs = append(d.Specs, newSpec(ctx, file, d, s))
	}
	return d
}

func (d *Decl) File() *File {
	return d.file
}

func (d *Decl) resolve(ctx *Context, file *File, scope Scope) {
	d.Annotations = ctx.resolveAnnotationGroup(file, d.unresolved.annotations)
	for _, s := range d.Specs {
		s.resolve(ctx, file, scope)
	}
}
