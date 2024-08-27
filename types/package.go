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

	Doc         *CommentGroup
	Annotations *AnnotationGroup
}

func (p *Package) resolve(c *Context) error {
	slices.SortFunc(p.files, func(a, b *File) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, file := range p.files {
		for _, decl := range file.decls {
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *EnumType:
					p.types = append(p.types, spec)
				case *StructType:
					p.types = append(p.types, spec)
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

func (p *Package) Decls() []*Decl {
	var decls []*Decl
	for _, file := range p.files {
		decls = append(decls, file.Decls()...)
	}
	return decls
}

func (p *Package) Types() []Type {
	return p.types
}

func (p *Package) Consts() []*ValueSpec {
	var consts []*ValueSpec
	for _, file := range p.files {
		consts = append(consts, file.Consts()...)
	}
	return consts
}

func (p *Package) Enums() []*EnumType {
	var enums []*EnumType
	for _, file := range p.files {
		enums = append(enums, file.Enums()...)
	}
	return enums
}

func (p *Package) Structs() []*StructType {
	var structs []*StructType
	for _, file := range p.files {
		structs = append(structs, file.Structs()...)
	}
	return structs
}
