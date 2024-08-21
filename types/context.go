package types

import (
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/token"
)

// Context represents a context of a Next program
type Context struct {
	// fset is the file set used to track file positions
	fset  *token.FileSet
	debug bool

	// files maps file name to file
	files       map[string]*File
	sortedFiles []*File

	// packages is a list of files sorted by their position
	packages map[string][]*File

	// symbols maps symbol name to symbol
	symbols map[string]Symbol

	// errors is a list of errors
	errors scanner.ErrorList

	// pos is the current position for call expression
	pos token.Pos
}

func NewContext(debug bool) *Context {
	return &Context{
		fset:     token.NewFileSet(),
		files:    make(map[string]*File),
		packages: make(map[string][]*File),
		symbols:  make(map[string]Symbol),
		debug:    debug,
	}
}

func (c *Context) FileSet() *token.FileSet {
	return c.fset
}

func (c *Context) AddFile(f *ast.File) error {
	filename := c.fset.Position(f.Pos()).Filename
	if _, ok := c.files[filename]; ok {
		c.errorf(f.Pos(), "file %s already exists", filename)
		return c.errors.Err()
	}
	file := newFile(c, f)
	file.Path = filename
	c.files[filename] = file
	c.packages[file.Pkg] = append(c.packages[file.Pkg], file)
	return nil
}

func (c *Context) Output() io.Writer {
	return os.Stderr
}

func (c *Context) Debug() bool {
	return c.debug
}

func (c *Context) Position() token.Position {
	return c.fset.Position(c.pos)
}

func (c *Context) error(pos token.Pos, msg string) {
	c.errors.Add(c.fset.Position(pos), msg)
}

func (c *Context) errorf(pos token.Pos, format string, args ...any) {
	c.errors.Add(c.fset.Position(pos), fmt.Sprintf(format, args...))
}

func (c *Context) lookupSymbol(name string) Symbol {
	return c.symbols[name]
}

func (c *Context) getFileByPos(pos token.Pos) *File {
	return c.files[c.fset.Position(pos).Filename]
}

func (c *Context) lookupFile(fullPath, relativePath string) *File {
	if relativePath == "" {
		return nil
	}
	if relativePath[0] == '.' {
		relativePath = path.Clean(path.Join(path.Dir(fullPath), relativePath))
	}
	return c.files[relativePath]
}

func (c *Context) call(pos token.Pos, name string, args ...constant.Value) (constant.Value, error) {
	c.pos = pos
	return constant.Call(c, name, args)
}

func (c *Context) Resolve() error {
	// sort files by package name and position
	files := make([]*File, 0, len(c.files))
	for i := range c.files {
		files = append(files, c.files[i])
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Pkg != files[j].Pkg {
			return files[i].Pkg < files[j].Pkg
		}
		return files[i].Pos() < files[j].Pos()
	})
	c.sortedFiles = files

	// resolve all imports
	for _, file := range files {
		for i := range file.imports {
			file.imports[i].file = c.lookupFile(file.Path, file.imports[i].Path)
			if file.imports[i].file == nil {
				c.errorf(file.imports[i].Pos(), "import file not found: %s", file.imports[i].Path)
			}
		}
	}
	if c.errors.Len() > 0 {
		return c.errors
	}

	// create all symbols
	for _, file := range files {
		for name, symbol := range file.symbols {
			symbolName := joinSymbolName(file.Pkg, name)
			if prev, ok := c.symbols[symbolName]; ok {
				c.errorf(symbol.Pos(), "symbol %s redeclared: previous declaration at %s", symbolName, c.fset.Position(prev.Pos()))
			} else {
				c.symbols[symbolName] = symbol
			}
		}
	}
	if c.errors.Len() > 0 {
		return c.errors
	}

	// resolve all files
	for _, file := range files {
		file.resolve(c)
	}

	// finally resolve all statements
	for _, file := range files {
		for _, stmt := range file.Stmts {
			stmt.resolve(c, file)
		}
	}

	return c.errors.Err()
}

func (c *Context) resolveAnnotationGroup(file *File, annotations *ast.AnnotationGroup) AnnotationGroup {
	if annotations == nil {
		return AnnotationGroup{}
	}
	list := make([]Annotation, len(annotations.List))
	for i, a := range annotations.List {
		params := make([]*AnnotationParam, len(a.Params))
		for j, p := range a.Params {
			var name string
			if p.Name != nil {
				name = p.Name.Name
			}
			var value constant.Value
			if p.Value != nil {
				value = c.resolveValue(file, p.Value, nil)
			}
			params[j] = &AnnotationParam{
				Name:  name,
				Value: value,
			}
		}
		list[i] = Annotation{
			pos:    c.fset.Position(a.Pos()),
			Name:   a.Name.Name,
			Params: params,
		}
	}
	return AnnotationGroup{
		List: list,
	}
}

func (c *Context) resolveValue(file *File, expr ast.Expr, iota *iotaValue) constant.Value {
	return c.recusiveResolveValue(file, make([]string, 0, 16), expr, iota)
}

func (c *Context) resolveSymbolValue(file *File, refs []string, expr ast.Expr, name string, v *ValueSpec) constant.Value {
	if v.Value != nil {
		// value already resolved
		return v.Value
	}
	if index := slices.Index(refs, name); index >= 0 {
		c.errorf(expr.Pos(), "cyclic reference %v->%s", strings.Join(refs[index:], "->"), name)
		return constant.MakeUnknown()
	}
	v.resolveValue(c, file, refs)
	return v.Value
}

func (c *Context) recusiveResolveValue(file *File, refs []string, expr ast.Expr, iota *iotaValue) constant.Value {
	defer func() {
		if r := recover(); r != nil {
			c.errorf(expr.Pos(), "%v", r)
		}
	}()
	expr = ast.Unparen(expr)
	switch expr := expr.(type) {
	case *ast.BasicLit:
		return constant.MakeFromLiteral(expr.Value, expr.Kind, 0)

	case *ast.Ident:
		if expr.Name == "false" {
			return constant.MakeBool(false)
		} else if expr.Name == "true" {
			return constant.MakeBool(true)
		}
		v, err := file.LookupLocalValue(expr.Name)
		if err != nil {
			if expr.Name == "iota" {
				if iota == nil {
					c.errorf(expr.Pos(), "iota is not allowed in this context")
					return constant.MakeUnknown()
				}
				iota.found = true
				return constant.MakeInt64(int64(iota.value))
			}
			c.errorf(expr.Pos(), "%s is not defined", expr.Name)
			return constant.MakeUnknown()
		}
		name := joinSymbolName(file.Pkg, expr.Name)
		return c.resolveSymbolValue(file, refs, expr, name, v)

	case *ast.SelectorExpr:
		name := joinSymbolName(c.resolveSelectorExprChain(expr)...)
		v, err := file.LookupValue(name)
		if err != nil {
			c.errorf(expr.Pos(), "%s is not defined", name)
			return constant.MakeUnknown()
		}
		file = c.getFileByPos(v.Pos())
		if file == nil {
			c.errorf(expr.Pos(), "%s is not defined (file %q not found)", name, c.fset.Position(v.Pos()).Filename)
			return constant.MakeUnknown()
		}
		return c.resolveSymbolValue(file, refs, expr, name, v)

	case *ast.BinaryExpr:
		x := c.recusiveResolveValue(file, refs, expr.X, iota)
		y := c.recusiveResolveValue(file, refs, expr.Y, iota)
		switch expr.Op {
		case token.SHL, token.SHR:
			return constant.Shift(x, expr.Op, uint(c.recusiveResolveUint64(file, expr.Y, refs, iota)))
		case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
			return constant.MakeBool(constant.Compare(x, expr.Op, y))
		default:
			return constant.BinaryOp(x, expr.Op, y)
		}

	case *ast.UnaryExpr:
		x := c.recusiveResolveValue(file, refs, expr.X, iota)
		return constant.UnaryOp(expr.Op, x, 0)

	case *ast.CallExpr:
		fun := ast.Unparen(expr.Fun)
		ident, ok := fun.(*ast.Ident)
		if !ok {
			c.errorf(expr.Pos(), "unexpected function %T", fun)
			return constant.MakeUnknown()
		}
		args := make([]constant.Value, len(expr.Args))
		for i, arg := range expr.Args {
			args[i] = c.recusiveResolveValue(file, refs, arg, iota)
		}
		result, err := c.call(expr.Pos(), ident.Name, args...)
		if err != nil {
			c.errorf(expr.Pos(), err.Error())
			return constant.MakeUnknown()
		}
		return result

	default:
		c.errorf(expr.Pos(), "unexpected expression %T", expr)
	}
	return constant.MakeUnknown()
}

func (c *Context) resolveUint64(file *File, expr ast.Expr) uint64 {
	return c.recusiveResolveUint64(file, expr, make([]string, 0, 16), nil)
}

func (c *Context) recusiveResolveUint64(file *File, expr ast.Expr, refs []string, iota *iotaValue) uint64 {
	val := c.recusiveResolveValue(file, refs, expr, iota)
	switch val.Kind() {
	case constant.Int:
		n, ok := constant.Uint64Val(val)
		if ok {
			return n
		}
		c.errorf(expr.Pos(), "constant %s overflows uint64", val)
	case constant.Float:
		f, ok := constant.Float64Val(val)
		if ok && f == float64(uint64(f)) {
			return uint64(f)
		}
		c.errorf(expr.Pos(), "constant %s overflows uint64", val)
	default:
		c.errorf(expr.Pos(), "constant %s is not an integer", val)
	}
	return 0
}

func (c *Context) resolveType(file *File, t ast.Type) Type {
	switch t := t.(type) {
	case *ast.Ident:
		return c.resolveIdentType(file, t)
	case *ast.SelectorExpr:
		return c.resolveSelectorExprType(file, t)
	case *ast.ArrayType:
		return c.resolveArrayType(file, t)
	case *ast.VectorType:
		return c.resolveVectorType(file, t)
	case *ast.MapType:
		return c.resolveMapType(file, t)
	default:
		c.errorf(t.Pos(), "unexpected type %T", t)
		return nil
	}
}

func (c *Context) resolveIdentType(file *File, t *ast.Ident) Type {
	switch t.Name {
	case "bool":
		return &BasicType{kind: Bool}
	case "int":
		return &BasicType{kind: Int}
	case "int8":
		return &BasicType{kind: Int8}
	case "int16":
		return &BasicType{kind: Int16}
	case "int32":
		return &BasicType{kind: Int32}
	case "int64":
		return &BasicType{kind: Int64}
	case "float32":
		return &BasicType{kind: Float32}
	case "float64":
		return &BasicType{kind: Float64}
	case "byte":
		return &BasicType{kind: Byte}
	case "bytes":
		return &BasicType{kind: Bytes}
	case "string":
		return &BasicType{kind: String}
	default:
		typ, err := file.LookupLocalType(t.Name)
		if err != nil {
			c.errorf(t.Pos(), "failed to lookup type %s: %s", t.Name, err)
			return nil
		}
		return typ
	}
}

func (c *Context) resolveSelectorExprChain(t *ast.SelectorExpr) []string {
	var names []string
	for t != nil {
		names = append(names, t.Sel.Name)
		switch x := ast.Unparen(t.X).(type) {
		case *ast.Ident:
			names = append(names, x.Name)
			t = nil
		case *ast.SelectorExpr:
			t = x
		default:
			c.errorf(t.Pos(), "unexpected selector expression %T", t)
			return nil
		}
	}

	slices.Reverse(names)
	return names
}

func (c *Context) resolveSelectorExprType(file *File, t *ast.SelectorExpr) Type {
	names := c.resolveSelectorExprChain(t)
	if len(names) < 2 {
		return nil
	}
	if len(names) != 2 {
		c.errorf(t.Pos(), "unexpected selector expression %s", names)
		return nil
	}
	fullName := joinSymbolName(names...)
	typ, err := file.LookupType(fullName)
	if err != nil {
		c.error(t.Pos(), err.Error())
		return nil
	}
	return typ
}

func (c *Context) resolveArrayType(file *File, t *ast.ArrayType) Type {
	return &ArrayType{
		typ:      typ{pos: t.Pos()},
		ElemType: c.resolveType(file, t.T),
		N:        c.resolveUint64(file, t.N),
	}
}

func (c *Context) resolveVectorType(file *File, t *ast.VectorType) Type {
	return &VectorType{
		typ:      typ{pos: t.Pos()},
		ElemType: c.resolveType(file, t.T),
	}
}

func (c *Context) resolveMapType(file *File, t *ast.MapType) Type {
	return &MapType{
		typ:      typ{pos: t.Pos()},
		KeyType:  c.resolveType(file, t.K),
		ElemType: c.resolveType(file, t.V),
	}
}
