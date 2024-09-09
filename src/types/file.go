package types

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// Package represents a Next package.
// @api(template/object) Package
type Package struct {
	name  string
	files []*File
	types []Type

	// Doc is the [documentation](#Object.Doc) for the package.
	// @api(object/Package/Doc)
	Doc *Doc

	// Annotations is the [annotations](#Object.Annotations) for the package.
	// @api(object/Package/Annotations)
	Annotations AnnotationGroup
}

// Name returns the package name.
// @api(object/Package/Name)
func (p *Package) Name() string { return p.name }

// In reports whether the package is the same as the given package.
// @api(object/Package/In)
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

// Types returns the all declared types in the package.
// @api(object/Package/Types)
func (p *Package) Types() []Type {
	return p.types
}

// File represents a Next source file.
// @api(object/File)
type File struct {
	pos token.Pos // position of the file
	pkg *Package  // package containing the file
	src *ast.File // the original AST

	imports *Imports            // import declarations
	decls   *Decls              // top-level declarations: const, enum, struct, interface
	stmts   []Stmt              // top-level statements
	symbols map[string]Symbol   // symbol table: name -> Symbol(Type|Value)
	nodes   map[ast.Node]Object // AST node -> Node

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

func newFile(ctx *Context, src *ast.File, path string) *File {
	f := &File{
		src:     src,
		pos:     src.Pos(),
		Path:    path,
		Doc:     newDoc(src.Doc),
		decls:   &Decls{},
		symbols: make(map[string]Symbol),
		nodes:   make(map[ast.Node]Object),
	}
	f.imports = &Imports{File: f}
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

// Node returns the original AST node for the file.
func (f *File) Node() *ast.File { return f.src }

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

func (f *File) addNode(ctx *Context, n ast.Node, x Object) {
	if _, dup := f.nodes[n]; dup {
		ctx.addErrorf(n.Pos(), "node already added: %T", n)
		return
	}
	f.nodes[n] = x
}

// GetNode returns the Node for the given AST node.
func (f *File) GetNode(n ast.Node) Object {
	return f.nodes[n]
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
		if err := f.addSymbol(d.name.name, d.Type); err != nil {
			return d.pos, err
		}
		for _, m := range d.Members.List {
			if err := f.addSymbol(joinSymbolName(d.name.name, m.name.name), m.value); err != nil {
				return m.pos, err
			}
		}
	}
	for _, d := range f.decls.structs {
		if err := f.addSymbol(d.name.name, d.Type); err != nil {
			return d.pos, err
		}
	}
	for _, d := range f.decls.interfaces {
		if err := f.addSymbol(d.name.name, d.Type); err != nil {
			return d.pos, err
		}
	}
	return token.NoPos, nil
}

func (f *File) resolve(ctx *Context) {
	f.Annotations = ctx.resolveAnnotationGroup(f, f, f.unresolved.annotations)
	f.decls.resolve(ctx, f)
}
