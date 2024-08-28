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
	Annotations *AnnotationGroup
}

func (p *Package) Name() string { return p.name }

func (p *Package) In(pkg *Package) bool { return p == nil || p == pkg }

func (p *Package) resolve(c *Context) error {
	slices.SortFunc(p.files, func(a, b *File) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, file := range p.files {
		for _, decl := range file.decls {
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *EnumSpec:
					p.types = append(p.types, spec.Type)
				case *StructSpec:
					p.types = append(p.types, spec.Type)
				}
			}
		}
	}
	for _, file := range p.files {
		if file.Doc != nil {
			p.Doc = file.Doc
			break
		}
	}
	var annotations []Annotation
	for _, file := range p.files {
		if file.Annotations != nil {
			annotations = append(annotations, file.Annotations.list...)
		}
	}
	if len(annotations) > 0 {
		p.Annotations = &AnnotationGroup{list: annotations}
	}
	return nil
}

func (p *Package) Types() []Type {
	return p.types
}
