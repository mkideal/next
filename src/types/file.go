package types

import (
	"path/filepath"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// File represents a Next source file.
// @api(object/File)
type File struct {
	pos token.Pos // position of the file
	pkg *Package  // package containing the file

	imports *Imports          // import declarations
	decls   *Decls            // top-level declarations: const, enum, struct, interface
	stmts   []Stmt            // top-level statements
	symbols map[string]Symbol // symbol table: name -> Symbol(Type|Value)

	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	// Path is the file full path.
	// @api(object/File/Path)
	Path string

	// Doc is the [documentation](#Object.Doc) for the file.
	// @api(object/File/Doc)
	Doc *Doc

	// Annotations is the [annotations](#Object.Annotations) for the file.
	// @api(object/File/Annotations)
	Annotations AnnotationGroup
}

func newFile(ctx *Context, src *ast.File) *File {
	f := &File{
		pos:     src.Pos(),
		Doc:     newDoc(src.Doc),
		imports: &Imports{},
		decls:   &Decls{},
		symbols: make(map[string]Symbol),
	}
	f.unresolved.annotations = src.Annotations

	for _, i := range src.Imports {
		f.imports.List = append(f.imports.List, newImport(ctx, f, i))
	}

	for _, d := range src.Decls {
		switch d := d.(type) {
		case *ast.ConstDecl:
			f.decls.consts = append(f.decls.consts, newConst(ctx, f, d))
		case *ast.EnumDecl:
			f.decls.enums = append(f.decls.enums, newEnum(ctx, f, d))
		case *ast.StructDecl:
			f.decls.structs = append(f.decls.structs, newStruct(ctx, f, d))
		case *ast.InterfaceDecl:
			f.decls.interfaces = append(f.decls.interfaces, newInterface(ctx, f, d))
		default:
			ctx.addErrorf(d.Pos(), "unsupported declaration: %T", d)
		}
	}

	for _, s := range src.Stmts {
		f.stmts = append(f.stmts, newStmt(ctx, f, s))
	}
	if pos, err := f.createSymbols(); err != nil {
		ctx.errors.Add(ctx.fset.Position(pos), err.Error())
	}
	return f
}

// Name returns the file name without the ".next" extension.
// @api(object/File/Name)
func (x *File) Name() string { return strings.TrimSuffix(filepath.Base(x.Path), ".next") }

// File returns the file itself. It is used to implement the Decl interface.
// @api(object/File/File)
func (x *File) File() *File { return x }

// Package returns the package containing the file.
// @api(object/File/Package)
func (x *File) Package() *Package {
	if x == nil {
		return nil
	}
	return x.pkg
}

// Imports returns the file's import declarations.
// @api(object/File/Imports)
func (f *File) Imports() *Imports { return f.imports }

// Decls returns the file's top-level declarations.
// @api(object/File/Decls)
func (f *File) Decls() *Decls {
	if f == nil {
		return nil
	}
	return f.decls
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
func (f *File) LookupLocalValue(name string) (*Value, error) {
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
	if f.decls == nil {
		return token.NoPos, nil
	}
	for _, d := range f.decls.consts {
		if err := f.addSymbol(d.value.name, d.value); err != nil {
			return d.pos, err
		}
	}
	for _, d := range f.decls.enums {
		if err := f.addSymbol(string(d.name), d.Type); err != nil {
			return d.pos, err
		}
		for _, m := range d.Members.List {
			if err := f.addSymbol(joinSymbolName(string(d.name), string(m.name)), m.value); err != nil {
				return m.pos, err
			}
		}
	}
	for _, d := range f.decls.structs {
		if err := f.addSymbol(string(d.name), d.Type); err != nil {
			return d.pos, err
		}
	}
	return token.NoPos, nil
}

func (f *File) resolve(ctx *Context) {
	f.Annotations = ctx.resolveAnnotationGroup(f, f, f.unresolved.annotations)
	f.decls.resolve(ctx, f)
}
