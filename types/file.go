package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/token"
)

// File represents a Next source file.
type File struct {
	pos        token.Pos
	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	Context     *Context
	Pkg         string // package name
	Path        string // file path
	Doc         CommentGroup
	Annotations AnnotationGroup
	Decls       []*Decl
	Stmts       []Stmt

	imports []*ImportSpec

	// all symbols used in current file:
	// - values: constant, enum member
	// - types: struct, protocol, enum
	symbols map[string]Symbol
}

func newFile(ctx *Context, src *ast.File) *File {
	file := &File{
		pos:     src.Pos(),
		Pkg:     src.Name.Name,
		Doc:     newCommentGroup(src.Doc),
		symbols: make(map[string]Symbol),
	}
	file.unresolved.annotations = src.Annotations
	for _, d := range src.Decls {
		file.Decls = append(file.Decls, newDecl(ctx, d.(*ast.GenDecl)))
	}
	for _, s := range src.Stmts {
		file.Stmts = append(file.Stmts, newStmt(ctx, s))
	}
	file.createSymbols()
	return file
}

func (f *File) Pos() token.Pos { return f.pos }

func (f *File) lookupSymbol(name string) Symbol {
	if s := f.symbols[name]; s != nil {
		return s
	}
	var files []*File
	pkg, name := splitSymbolName(name)
	for i := range f.imports {
		if f.imports[i].file.Pkg == pkg {
			files = append(files, f.imports[i].file)
		}
	}
	for _, file := range files {
		if s := file.symbols[name]; s != nil {
			return s
		}
	}
	return nil
}

func (f *File) expectTypeSymbol(name string, s Symbol) (Type, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if t, ok := s.(Type); ok {
		return t, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "type", Got: s.symbolType()}
}

func (f *File) expectValueSymbol(name string, s Symbol) (*ValueSpec, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*ValueSpec); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "value", Got: s.symbolType()}
}

// LookupLocalType looks up a type by name in the file's symbol table.
// If the type is not found, it returns an error. If the symbol
// is found but it is not a type, it returns an error.
func (f *File) LookupLocalType(name string) (Type, error) {
	return f.expectTypeSymbol(name, f.symbols[name])
}

// LookupType looks up a type by name in the file's symbol table and
// its imported files. If the type is not found, it returns an error.
func (f *File) LookupType(name string) (Type, error) {
	return f.expectTypeSymbol(name, f.lookupSymbol(name))
}

// LookupLocalValue looks up a value by name in the file's symbol table.
// If the value is not found, it returns an error. If the symbol
// is found but it is not a value, it returns an error.
func (f *File) LookupLocalValue(name string) (*ValueSpec, error) {
	return f.expectValueSymbol(name, f.symbols[name])
}

// LookupValue looks up a value by name in the file's symbol table and
// its imported files. If the value is not found, it returns an error.
func (f *File) LookupValue(name string) (*ValueSpec, error) {
	return f.expectValueSymbol(name, f.lookupSymbol(name))
}

func (f *File) addSymbol(name string, s Symbol) error {
	if prev, dup := f.symbols[name]; dup {
		return &SymbolRedefinedError{Name: name, Prev: prev}
	}
	f.symbols[name] = s
	return nil
}

func (f *File) createSymbols() error {
	for _, d := range f.Decls {
		for _, s := range d.Specs {
			switch s := s.(type) {
			case *ImportSpec:
				f.imports = append(f.imports, s)
			case *ValueSpec:
				if err := f.addSymbol(s.Name, s); err != nil {
					return err
				}
			case *EnumType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return err
				}
				for _, m := range s.Members {
					if err := f.addSymbol(joinSymbolName(s.Name, m.Name), m); err != nil {
						return err
					}
				}
			case *StructType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return err
				}
			case *ProtocolType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (f *File) resolve(ctx *Context) {
	f.Annotations = ctx.resolveAnnotationGroup(f, f.unresolved.annotations)
	for _, d := range f.Decls {
		d.resolve(ctx, f)
	}
}
