package types

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gopherd/core/flags"
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/token"
)

// Context represents a context of a Next program
type Context struct {
	// command line flags
	flags struct {
		verbose    int         // verbose logging level: 0=off, 1=info, 2=debug
		importDirs flags.Slice // import directories
		macros     flags.Map   // macro variables, e.g. -E X=1 -E Y
		outputs    flags.Map   // output directories for each language, e.g. -O go=gen/go -O ts=gen/ts
		templates  flags.Map   // template dir or files for each language, e.g. -T go=templates/go
		types      flags.Map   // type mapping for each language, e.g. -t cpp.map=map -t cpp.float32=float32_t -T cpp.float64=float64_t
	}

	// fset is the file set used to track file positions
	fset *token.FileSet

	// files maps file name to file
	files       map[string]*File
	sortedFiles []*File

	// packages is a list of packages sorted by package name
	packages []*Package

	// symbols maps symbol name to symbol
	symbols map[string]Symbol

	// errors is a list of errors
	errors scanner.ErrorList

	// pos is the current position for call expression
	pos token.Pos
}

func NewContext() *Context {
	c := &Context{
		fset:    token.NewFileSet(),
		files:   make(map[string]*File),
		symbols: make(map[string]Symbol),
	}
	c.flags.macros = make(flags.Map)
	c.flags.outputs = make(flags.Map)
	c.flags.templates = make(flags.Map)
	c.flags.types = make(flags.Map)
	return c
}

func (c *Context) SetupCommandFlags(fs *flag.FlagSet, u flags.UsageFunc) {
	fs.IntVar(&c.flags.verbose, "v", 0, u("Set verbose logging `level`: 0=off, 1=info, 2=debug"))
	fs.Var(&c.flags.importDirs, "I", u("Add import directories as `dir[,dir2,...]`, e.g. -I dir1 -I dir2 or -I dir1,dir2"))
	fs.Var(&c.flags.macros, "D", u("Define macro variables as `name[=value]`, e.g. -D A=\"hello next\" X=hello -D Y=1 -D Z"))
	fs.Var(&c.flags.outputs, "O", u("Specify output directories for each language as `lang=dir`, e.g. -O go=gen/go -O ts=gen/ts"))
	fs.Var(&c.flags.templates, "T", u("Provide template directories or files for each language as `lang=dir|file`, e.g. -T go=templates/go"))
	fs.Var(&c.flags.types, "M", u("Set type mappings for each language as `lang.type=value`, e.g. -M cpp.int=int64_t -M cpp.float64=float64_t"))
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
	path, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of %s: %w", filename, err)
	}
	file := newFile(c, f)
	file.Path = path
	c.files[path] = file
	for _, pkg := range c.packages {
		if pkg.Name == f.Name.Name {
			file.Package = pkg.Name
			pkg.files = append(pkg.files, file)
			return nil
		}
	}
	pkg := &Package{
		Name:  f.Name.Name,
		files: []*File{file},
	}
	file.Package = pkg.Name
	c.packages = append(c.packages, pkg)
	return nil
}

func (c *Context) Output() io.Writer {
	return os.Stderr
}

func (c *Context) IsDebugEnabled() bool {
	return c.flags.verbose >= 2
}

func fileLine(pos token.Position) string {
	if pos.IsValid() {
		return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
	}
	return ""
}

func (c *Context) log(msg string) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	if c.Position().Line > 0 {
		fmt.Fprintf(c.Output(), "%s: %s", fileLine(c.Position()), msg)
	} else {
		fmt.Fprint(c.Output(), msg)
	}
}

func (c *Context) Print(args ...any) {
	if c.IsDebugEnabled() {
		c.log(fmt.Sprint(args...))
	}
}

func (c *Context) Printf(format string, args ...any) {
	if c.IsDebugEnabled() {
		c.log(fmt.Sprintf(format, args...))
	}
}

func (c *Context) Info(args ...any) {
	if c.flags.verbose < 1 {
		return
	}
	c.log(fmt.Sprint(args...))
}

func (c *Context) Infof(format string, args ...any) {
	if c.flags.verbose < 1 {
		return
	}
	c.log(fmt.Sprintf(format, args...))
}

func (c *Context) Error(args ...any) {
	c.log(fmt.Sprint(args...))
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
	path, err := filepath.Abs(c.fset.Position(pos).Filename)
	if err != nil {
		c.errorf(pos, "failed to get absolute path of %s: %v", c.fset.Position(pos).Filename, err)
		return nil
	}
	return c.files[path]
}

func (c *Context) lookupFile(fullPath, relativePath string) *File {
	if relativePath == "" {
		return nil
	}
	if relativePath[0] == '.' {
		var err error
		relativePath, err = filepath.Abs(filepath.Join(filepath.Dir(fullPath), relativePath))
		if err != nil {
			c.errorf(token.NoPos, "failed to get absolute path of %s: %v", relativePath, err)
			return nil
		}
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
		if files[i].Package != files[j].Package {
			return files[i].Package < files[j].Package
		}
		return files[i].Pos() < files[j].Pos()
	})
	c.sortedFiles = files

	// resolve all imports
	for _, file := range files {
		for i := range file.imports {
			file.imports[i].importedFile = c.lookupFile(file.Path, file.imports[i].Path)
			if file.imports[i].importedFile == nil {
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
			symbolName := joinSymbolName(file.Package, name)
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
	if c.errors.Len() > 0 {
		return c.errors
	}

	// finally resolve all statements
	for _, file := range files {
		for _, stmt := range file.stmts {
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
	return c.recursiveResolveValue(file, file, make([]*ValueSpec, 0, 16), expr, iota)
}

func (c *Context) resolveSymbolValue(file *File, scope Scope, refs []*ValueSpec, v *ValueSpec) constant.Value {
	if v.value != nil {
		// value already resolved
		return v.value
	}
	if index := slices.Index(refs, v); index >= 0 {
		var sb strings.Builder
		for i := index; i < len(refs); i++ {
			fmt.Fprintf(&sb, "\n%s: %s ↓", c.fset.Position(refs[i].Pos()), refs[i].Name)
		}
		c.errorf(v.pos, "cyclic references: %s\n%s: %s", sb.String(), c.fset.Position(v.Pos()), v.Name)
		return constant.MakeUnknown()
	}
	v.resolveValue(c, file, scope, refs)
	return v.value
}

func (c *Context) recursiveResolveValue(file *File, scope Scope, refs []*ValueSpec, expr ast.Expr, iota *iotaValue) (result constant.Value) {
	defer func() {
		if r := recover(); r != nil {
			c.errorf(expr.Pos(), "%v", r)
			result = constant.MakeUnknown()
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
		v, err := expectValueSymbol(expr.Name, scope.LookupLocalSymbol(expr.Name))
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
		return c.resolveSymbolValue(file, scope, refs, v)

	case *ast.SelectorExpr:
		name := joinSymbolName(c.resolveSelectorExprChain(expr)...)
		v, err := lookupValue(scope, name)
		if err != nil {
			c.errorf(expr.Pos(), "%s is not defined", name)
			return constant.MakeUnknown()
		}
		file = c.getFileByPos(v.Pos())
		if file == nil {
			c.errorf(expr.Pos(), "%s is not defined (file %q not found)", name, c.fset.Position(v.Pos()).Filename)
			return constant.MakeUnknown()
		}
		return c.resolveSymbolValue(file, scope, refs, v)

	case *ast.BinaryExpr:
		x := c.recursiveResolveValue(file, scope, refs, expr.X, iota)
		y := c.recursiveResolveValue(file, scope, refs, expr.Y, iota)
		switch expr.Op {
		case token.SHL, token.SHR:
			return constant.Shift(x, expr.Op, uint(c.recursiveResolveUint64(file, scope, expr.Y, refs, iota)))
		case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
			return constant.MakeBool(constant.Compare(x, expr.Op, y))
		default:
			return constant.BinaryOp(x, expr.Op, y)
		}

	case *ast.UnaryExpr:
		x := c.recursiveResolveValue(file, scope, refs, expr.X, iota)
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
			args[i] = c.recursiveResolveValue(file, scope, refs, arg, iota)
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
	return c.recursiveResolveUint64(file, file, expr, make([]*ValueSpec, 0, 16), nil)
}

func (c *Context) recursiveResolveUint64(file *File, scope Scope, expr ast.Expr, refs []*ValueSpec, iota *iotaValue) uint64 {
	val := c.recursiveResolveValue(file, scope, refs, expr, iota)
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

func (c *Context) resolveIdentType(file *File, i *ast.Ident) Type {
	if t, ok := basicTypes[i.Name]; ok {
		return t
	}
	t, err := file.LookupLocalType(i.Name)
	if err != nil {
		c.errorf(i.Pos(), "failed to lookup type %s: %s", i.Name, err)
		return nil
	}
	return t
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
	typ, err := lookupType(file, fullName)
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
