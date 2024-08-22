package types

type Package struct {
	Name string

	files []*File
}

func (p *Package) Files() []*File {
	return p.files
}

func (p *Package) Decls() []*Decl {
	var decls []*Decl
	for _, file := range p.files {
		decls = append(decls, file.decls...)
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

func (p *Package) Protocols() []*ProtocolType {
	var protocols []*ProtocolType
	for _, file := range p.files {
		protocols = append(protocols, file.Protocols()...)
	}
	return protocols
}
