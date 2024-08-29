package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

// File represents a Next source file.
type File struct {
	pos        token.Pos
	pkg        *Package
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Path        string
	Doc         *Doc
	Annotations AnnotationGroup

	decls   []*Decl
	stmts   []Stmt
	imports *Imports
	specs   Specs

	// all symbols used in current file:
	// - values: constant, enum member
	// - types: struct, enum
	symbols map[string]Symbol
}

func newFile(ctx *Context, src *ast.File) *File {
	f := &File{
		pos:     src.Pos(),
		Doc:     newDoc(src.Doc),
		imports: &Imports{},
		symbols: make(map[string]Symbol),
	}
	f.unresolved.annotations = src.Annotations
	for _, d := range src.Decls {
		f.decls = append(f.decls, newDecl(ctx, f, d.(*ast.GenDecl)))
	}
	f.specs.Consts = f.getConsts()
	f.specs.Enums = f.getEnums()
	f.specs.Structs = f.getStructs()
	for _, s := range src.Stmts {
		f.stmts = append(f.stmts, newStmt(ctx, f, s))
	}
	if pos, err := f.createSymbols(); err != nil {
		ctx.errors.Add(ctx.fset.Position(pos), err.Error())
	}
	return f
}

func (f *File) Decls() []*Decl {
	var hasImport bool
	for _, d := range f.decls {
		if d.tok == token.IMPORT {
			hasImport = true
			break
		}
	}
	if !hasImport {
		return f.decls
	}
	var decls = make([]*Decl, 0, len(f.decls))
	for _, d := range f.decls {
		if d.tok != token.IMPORT {
			decls = append(decls, d)
		}
	}
	return decls
}

func (f *File) Imports() *Imports { return f.imports }

func (f *File) Specs() *Specs {
	return &f.specs
}

func (f *File) getConsts() *ConstSpecs {
	var consts = &ConstSpecs{File: f}
	for _, d := range f.decls {
		if d.tok == token.CONST {
			for _, s := range d.Specs {
				consts.List = append(consts.List, s.(*ValueSpec))
			}
		}
	}
	return consts
}

func (f *File) getEnums() *EnumSpecs {
	var enums = &EnumSpecs{File: f}
	for _, d := range f.decls {
		if d.tok == token.ENUM {
			for _, s := range d.Specs {
				enums.List = append(enums.List, s.(*EnumSpec))
			}
		}
	}
	return enums
}

func (f *File) getStructs() *StructSpecs {
	var structs = &StructSpecs{File: f}
	for _, d := range f.decls {
		if d.tok == token.STRUCT {
			for _, s := range d.Specs {
				structs.List = append(structs.List, s.(*StructSpec))
			}
		}
	}
	return structs
}

// LookupLocalType looks up a type by name in the file's symbol table.
// If the type is not found, it returns an error. If the symbol
// is found but it is not a type, it returns an error.
func (f *File) LookupLocalType(name string) (Type, error) {
	return expectTypeSymbol(name, f.symbols[name])
}

// LookupLocalValue looks up a value by name in the file's symbol table.
// If the value is not found, it returns an error. If the symbol
// is found but it is not a value, it returns an error.
func (f *File) LookupLocalValue(name string) (*ValueSpec, error) {
	return expectValueSymbol(name, f.symbols[name])
}

func (f *File) addSymbol(name string, s Symbol) error {
	if prev, dup := f.symbols[name]; dup {
		return &SymbolRedefinedError{Name: name, Prev: prev}
	}
	f.symbols[name] = s
	return nil
}

func (f *File) createSymbols() (token.Pos, error) {
	for _, d := range f.decls {
		for _, s := range d.Specs {
			switch s := s.(type) {
			case *ImportSpec:
				f.imports.List = append(f.imports.List, s)
			case *ValueSpec:
				if err := f.addSymbol(s.name, s); err != nil {
					return s.pos, err
				}
			case *EnumSpec:
				if err := f.addSymbol(s.Type.name, s.Type); err != nil {
					return s.Type.pos, err
				}
				for _, m := range s.Members.List {
					if err := f.addSymbol(joinSymbolName(s.Type.name, m.name), m); err != nil {
						return m.pos, err
					}
				}
			case *StructSpec:
				if err := f.addSymbol(s.Type.name, s.Type); err != nil {
					return s.Type.pos, err
				}
			}
		}
	}
	return token.NoPos, nil
}

func (f *File) resolve(ctx *Context) {
	f.Annotations = ctx.resolveAnnotationGroup(f, f.unresolved.annotations)
	for _, d := range f.decls {
		d.resolve(ctx, f, f)
	}
}
