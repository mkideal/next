package types

import (
	"cmp"
	"slices"
)

// @api(template/object) Package
type Package struct {
	name  string
	files []*File
	types []Type

	Doc         *Doc
	Annotations AnnotationGroup
}

func (p *Package) Name() string { return p.name }

func (p *Package) In(pkg *Package) bool { return p == nil || p == pkg }

func (p *Package) resolve(c *Context) error {
	slices.SortFunc(p.files, func(a, b *File) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, file := range p.files {
		for _, d := range file.decls.enums {
			p.types = append(p.types, d.Type)
		}
		for _, d := range file.decls.structs {
			p.types = append(p.types, d.Type)
		}
		for _, d := range file.decls.interfaces {
			p.types = append(p.types, d.Type)
		}
	}
	for _, file := range p.files {
		if file.Doc != nil {
			p.Doc = file.Doc
			break
		}
	}
	p.Annotations = make(AnnotationGroup)
	for _, file := range p.files {
		if file.Annotations != nil {
			for name, group := range file.Annotations {
				p.Annotations[name] = group
			}
		}
	}
	return nil
}

func (p *Package) Types() []Type {
	return p.types
}
