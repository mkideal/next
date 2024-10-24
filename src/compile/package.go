package compile

import (
	"cmp"
	"fmt"
	"slices"
)

// @api(Object/Packages) represents a list of Next packages.
type Packages []*Package

func (Packages) Doc() *Doc                { return nil }
func (Packages) Annotations() Annotations { return nil }
func (Packages) File() *File              { return nil }
func (Packages) Package() *Package        { return nil }
func (Packages) Name() string             { return "" }
func (Packages) NamePos() Position        { return Position{} }
func (Packages) Pos() Position            { return Position{} }

// @api(Object/Package) (extends [Decl](#Object/Common/Decl)) represents a Next package.
type Package struct {
	name    string
	files   []*File
	decls   *Decls[*Package]
	types   []Type
	imports *Imports[*Package]

	doc         *Doc
	annotations Annotations
}

func newPackage(c *Compiler, name string) *Package {
	p := &Package{
		name: name,
	}
	p.decls = &Decls[*Package]{Decl: p, compiler: c}
	p.imports = &Imports[*Package]{Decl: p}
	return p
}

// @api(Object/Package.Decls) represents the top-level declarations in the package.
func (p *Package) Decls() *Decls[*Package] {
	if p == nil {
		return nil
	}
	return p.decls
}

func (p *Package) Pos() Position { return Position{} }

// @api(Object/Package.Name) represents the package name string.
func (p *Package) Name() string { return p.name }

func (p *Package) Doc() *Doc { return p.doc }

func (p *Package) Annotations() Annotations { return p.annotations }

func (p *Package) Package() *Package { return p }

func (p *Package) addFile(f *File) {
	f.pkg = p
	p.files = append(p.files, f)
	p.decls.consts = append(p.decls.consts, f.decls.consts...)
	p.decls.enums = append(p.decls.enums, f.decls.enums...)
	p.decls.structs = append(p.decls.structs, f.decls.structs...)
	p.decls.interfaces = append(p.decls.interfaces, f.decls.interfaces...)
	for _, x := range f.imports.List {
		p.imports.add(x)
	}
}

func (p *Package) File() *File {
	if len(p.files) == 0 {
		return nil
	}
	return p.files[0]
}

// @api(Object/Package.Files) represents the all declared file objects in the package.
func (p *Package) Files() []*File {
	return p.files
}

// @api(Object/Package.Types) represents the all declared types in the package.
func (p *Package) Types() []Type {
	return p.types
}

// @api(Object/Package.Imports) represents the package's import declarations.
func (p *Package) Imports() *Imports[*Package] { return p.imports }

// @api(Object/Package.Has) reports whether the package contains the given [Type](#Object/Common/Type) or [Symbol](#Object/Common/Symbol).
// If the current package is nil, it always returns true.
//
// Example:
//
//	```npl
//	{{- define "next/go/used.type" -}}
//	{{if not (.File.Package.Has .Type) -}}
//	{{.Type.Decl.File.Package.Name -}}.
//	{{- end -}}
//	{{next .Type}}
//	{{- end}}
//	```
func (p *Package) Has(obj Object) (bool, error) {
	var p2 *Package
	switch node := obj.(type) {
	case Type:
		p2 = node.Decl().File().Package()
	case Symbol:
		p2 = node.File().Package()
	default:
		return false, fmt.Errorf("package.Has: unexpected type %T, want Type or Symbol", node)
	}
	return p2 == nil || p == p2, nil
}

func (p *Package) resolve(c *Compiler) error {
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
		if file.doc != nil {
			p.doc = file.doc
			break
		}
	}
	p.annotations = make(Annotations)
	for _, file := range p.files {
		if file.annotations != nil {
			for name, group := range file.annotations {
				p.annotations[name] = group
			}
		}
	}
	return nil
}

// @api(Object/Package.LookupLocalType) looks up a type by name in the package's symbol table.
func (p *Package) LookupLocalType(name string) (Type, error) {
	for _, f := range p.files {
		if s, ok := f.symbols[name]; ok {
			return expectTypeSymbol(name, s)
		}
	}
	return nil, nil
}

// @api(Object/Package.LookupLocalValue) looks up a value by name in the package's symbol table.
func (p *Package) LookupLocalValue(name string) (*Value, error) {
	for _, f := range p.files {
		if s, ok := f.symbols[name]; ok {
			return expectValueSymbol(name, s)
		}
	}
	return nil, nil
}
