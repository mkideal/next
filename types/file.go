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
		file.Decls = append(file.Decls, newDecl(ctx, file, d.(*ast.GenDecl)))
	}
	for _, s := range src.Stmts {
		file.Stmts = append(file.Stmts, newStmt(ctx, file, s))
	}
	if pos, err := file.createSymbols(); err != nil {
		ctx.errors.Add(ctx.fset.Position(pos), err.Error())
	}
	return file
}

func (f *File) Pos() token.Pos { return f.pos }

func (f *File) ParentScope() Scope { return &fileParentScope{f} }

func (f *File) LookupLocalSymbol(name string) Symbol {
	return f.symbols[name]
}

type fileParentScope struct {
	f *File
}

func (s *fileParentScope) ParentScope() Scope {
	return nil
}

func (s *fileParentScope) LookupLocalSymbol(name string) Symbol {
	var files []*File
	pkg, name := splitSymbolName(name)
	for i := range s.f.imports {
		if s.f.imports[i].file.Pkg == pkg {
			files = append(files, s.f.imports[i].file)
		}
	}
	for _, file := range files {
		if s := file.symbols[name]; s != nil {
			return s
		}
	}
	return nil
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
	for _, d := range f.Decls {
		for _, s := range d.Specs {
			switch s := s.(type) {
			case *ImportSpec:
				f.imports = append(f.imports, s)
			case *ValueSpec:
				if err := f.addSymbol(s.Name, s); err != nil {
					return s.pos, err
				}
			case *EnumType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return s.pos, err
				}
				for _, m := range s.Members {
					if err := f.addSymbol(joinSymbolName(s.Name, m.Name), m); err != nil {
						return m.pos, err
					}
				}
			case *StructType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return s.pos, err
				}
			case *ProtocolType:
				if err := f.addSymbol(s.Name, s); err != nil {
					return s.pos, err
				}
			}
		}
	}
	return token.NoPos, nil
}

func (f *File) resolve(ctx *Context) {
	f.Annotations = ctx.resolveAnnotationGroup(f, f.unresolved.annotations)
	for _, d := range f.Decls {
		d.resolve(ctx, f, f)
	}
}
