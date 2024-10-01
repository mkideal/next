package compile

import (
	"path/filepath"
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// @api(Object/File) (extends [Decl](#Object/Common/Decl)) represents a Next source file.
type File struct {
	compiler *Compiler
	pos      token.Pos // position of the file
	pkg      *Package  // package containing the file
	src      *ast.File // the original AST

	imports *Imports[*File]     // import declarations
	decls   *Decls[*File]       // top-level declarations: const, enum, struct, interface
	stmts   []Stmt              // top-level statements
	symbols map[string]Symbol   // symbol table: name -> Symbol(Type|Value)
	objects map[ast.Node]Object // AST node -> Node

	unresolved struct {
		annotations *ast.AnnotationGroup
	}

	doc         *Doc
	annotations Annotations

	// @api(Object/File.Path) represents the file full path.
	Path string
}

func newFile(c *Compiler, src *ast.File, path string) *File {
	f := &File{
		compiler: c,
		src:      src,
		pos:      src.Pos(),
		doc:      newDoc(src.Doc),
		symbols:  make(map[string]Symbol),
		objects:  make(map[ast.Node]Object),
		Path:     path,
	}
	f.decls = &Decls[*File]{Decl: f, compiler: c}
	f.imports = &Imports[*File]{Decl: f}
	f.unresolved.annotations = src.Annotations

	for _, i := range src.Imports {
		f.imports.add(newImport(c, f, i))
	}

	for _, d := range src.Decls {
		switch d := d.(type) {
		case *ast.ConstDecl:
			f.decls.consts = append(f.decls.consts, newConst(c, f, d))
		case *ast.EnumDecl:
			f.decls.enums = append(f.decls.enums, newEnum(c, f, d))
		case *ast.StructDecl:
			f.decls.structs = append(f.decls.structs, newStruct(c, f, d))
		case *ast.InterfaceDecl:
			f.decls.interfaces = append(f.decls.interfaces, newInterface(c, f, d))
		default:
			c.addErrorf(d.Pos(), "unsupported declaration: %T", d)
		}
	}

	for _, s := range src.Stmts {
		f.stmts = append(f.stmts, newStmt(c, f, s))
	}
	if pos, err := f.createSymbols(); err != nil {
		c.errors.Add(c.fset.Position(pos), err.Error())
	}
	return f
}

func (x *File) Package() *Package {
	if x == nil {
		return nil
	}
	return x.pkg
}

// @api(Object/File.Name) represents the file name without the ".next" extension.
func (x *File) Name() string { return strings.TrimSuffix(filepath.Base(x.Path), ".next") }

// @api(Object/File.Imports) represents the file's import declarations.
func (f *File) Imports() *Imports[*File] { return f.imports }

// @api(Object/File.Decls) returns the file's all top-level declarations.
func (f *File) Decls() *Decls[*File] {
	if f == nil {
		return nil
	}
	return f.decls
}

func (f *File) Doc() *Doc { return f.doc }

func (f *File) Annotations() Annotations { return f.annotations }

func (f *File) addObject(c *Compiler, n ast.Node, x Object) {
	if _, dup := f.objects[n]; dup {
		c.addErrorf(n.Pos(), "node already added: %T", n)
		return
	}
	f.objects[n] = x
}

// FileNode represents the original AST node of the file.
func FileNode(f *File) *ast.File { return f.src }

// LookupFileObject looks up an object by its AST node in the file's symbol table.
func LookupFileObject(f *File, n ast.Node) Object {
	if f == nil {
		return nil
	}
	return f.objects[n]
}

// @api(Object/File.LookupLocalType) looks up a type by name in the file's symbol table.
// If the type is not found, it returns an error. If the symbol
// is found but it is not a type, it returns an error.
func (f *File) LookupLocalType(name string) (Type, error) {
	return expectTypeSymbol(name, f.symbols[name])
}

// @api(Object/File.LookupLocalValue) looks up a value by name in the file's symbol table.
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
		if err := f.addSymbol(d.name, d.Type); err != nil {
			return d.pos, err
		}
		for _, m := range d.Members.List {
			if err := f.addSymbol(joinSymbolName(d.name, m.name), m.value); err != nil {
				return m.pos, err
			}
		}
	}
	for _, d := range f.decls.structs {
		if err := f.addSymbol(d.name, d.Type); err != nil {
			return d.pos, err
		}
	}
	for _, d := range f.decls.interfaces {
		if err := f.addSymbol(d.name, d.Type); err != nil {
			return d.pos, err
		}
	}
	return token.NoPos, nil
}

func (f *File) resolve(c *Compiler) {
	f.annotations = resolveAnnotations(c, f, f, f.unresolved.annotations)
	f.decls.resolve(c, f)
}
