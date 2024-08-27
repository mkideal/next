package types

// @api(template/object) Package
type Package struct {
	name  string
	files []*File
}

func (p *Package) Decls() []*Decl {
	var decls []*Decl
	for _, file := range p.files {
		decls = append(decls, file.Decls()...)
	}
	return decls
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
