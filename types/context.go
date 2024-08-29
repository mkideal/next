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
		verbose    int
		importDirs flags.Slice
		macros     flags.Map
		outputs    flags.Map
		templates  flags.MapSlice
		types      flags.Map
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

	// stack is the current position for call expression
	stack []token.Pos

	// searchDirs is a list of search directories
	searchDirs []string
}

func NewContext() *Context {
	c := &Context{
		fset:       token.NewFileSet(),
		files:      make(map[string]*File),
		symbols:    make(map[string]Symbol),
		searchDirs: createSearchDirs(),
	}
	c.flags.macros = make(flags.Map)
	c.flags.outputs = make(flags.Map)
	c.flags.templates = make(flags.MapSlice)
	c.flags.types = make(flags.Map)

	return c
}

func (c *Context) SetupCommandFlags(fs *flag.FlagSet, u flags.UsageFunc) {
	fs.IntVar(&c.flags.verbose, "v", 0, u("Set `verbose` level for debugging: 0=off, 1=info, 2=debug, 3=trace"))
	fs.Var(&c.flags.importDirs, "I", u("Add import directories as `dir[,dir2,...]`, e.g. -I dir1 -I dir2 or -I dir1,dir2"))
	fs.Var(&c.flags.macros, "D", u("Define macro variables as `name[=value]`, e.g. -D A=\"hello next\" -D X=hello -D Y=1 -D Z"))
	fs.Var(&c.flags.outputs, "O", u("Specify output directories as `lang=dir`, e.g. -O go=gen/go -O ts=gen/ts"))
	fs.Var(&c.flags.templates, "T", u("Provide template directories or files as `lang=dir|file`, e.g. -T go=tmpl/go -T ts=tmpl/ts.npl"))
	fs.Var(&c.flags.types, "M", u("Set type mappings as `lang.type=value`, e.g. -M cpp.int=int64_t -M cpp.map<%K%,%V%>=std::map<%K%,%V%>"))
}

// FileSet returns the file set used to track file positions
func (c *Context) FileSet() *token.FileSet {
	return c.fset
}

// AddFile adds a file to the context
func (c *Context) AddFile(f *ast.File) error {
	filename := c.fset.Position(f.Pos()).Filename
	if _, ok := c.files[filename]; ok {
		c.addErrorf(f.Pos(), "file %s already exists", filename)
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
		if pkg.name == f.Name.Name {
			file.pkg = pkg
			pkg.files = append(pkg.files, file)
			return nil
		}
	}
	pkg := &Package{
		name:  f.Name.Name,
		files: []*File{file},
	}
	file.pkg = pkg
	c.packages = append(c.packages, pkg)
	return nil
}

// Output returns the output writer for logging
func (c *Context) Output() io.Writer {
	return os.Stderr
}

// IsDebugEnabled returns true if debug logging is enabled
func (c *Context) IsDebugEnabled() bool {
	return c.flags.verbose >= 2
}

func (c *Context) log(msg string) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	pos := c.Position()
	pos.Column = 0 // ignore column
	if pos.Line > 0 {
		fmt.Fprintf(c.Output(), "%s: %s", pos, msg)
	} else {
		fmt.Fprint(c.Output(), msg)
	}
}

// Trace logs a message if trace logging is enabled
func (c *Context) Trace(args ...any) {
	if c.flags.verbose >= 3 {
		c.log(fmt.Sprint(args...))
	}
}

// Tracef logs a formatted message if trace logging is enabled
func (c *Context) Tracef(format string, args ...any) {
	if c.flags.verbose >= 3 {
		c.log(fmt.Sprintf(format, args...))
	}
}

// Print logs a message if debug logging is enabled
func (c *Context) Print(args ...any) {
	if c.flags.verbose >= 2 {
		c.log(fmt.Sprint(args...))
	}
}

// Printf logs a formatted message if debug logging is enabled
func (c *Context) Printf(format string, args ...any) {
	if c.flags.verbose >= 2 {
		c.log(fmt.Sprintf(format, args...))
	}
}

// Info logs an info message if verbose logging is enabled
func (c *Context) Info(args ...any) {
	if c.flags.verbose < 1 {
		return
	}
	c.log(fmt.Sprint(args...))
}

// Infof logs a formatted info message if verbose logging is enabled
func (c *Context) Infof(format string, args ...any) {
	if c.flags.verbose < 1 {
		return
	}
	c.log(fmt.Sprintf(format, args...))
}

// Error logs an error message
func (c *Context) Error(args ...any) {
	c.log(fmt.Sprint(args...))
}

// Position returns the current position for call expression
func (c *Context) Position() token.Position {
	if len(c.stack) == 0 {
		return token.Position{}
	}
	return c.fset.Position(c.stack[len(c.stack)-1])
}

// addError adds an error message with position to the error list
func (c *Context) addError(pos token.Pos, msg string) {
	c.errors.Add(c.fset.Position(pos), msg)
}

// addErrorf adds a formatted error message with position to the error list
func (c *Context) addErrorf(pos token.Pos, format string, args ...any) {
	c.errors.Add(c.fset.Position(pos), fmt.Sprintf(format, args...))
}

// getFileByPos returns the file by position
func (c *Context) getFileByPos(pos token.Pos) *File {
	path, err := filepath.Abs(c.fset.Position(pos).Filename)
	if err != nil {
		c.addErrorf(pos, "failed to get absolute path of %s: %v", c.fset.Position(pos).Filename, err)
		return nil
	}
	return c.files[path]
}

// lookupFile returns the file by full path or relative path
func (c *Context) lookupFile(fullPath, relativePath string) *File {
	if relativePath == "" {
		return nil
	}
	if relativePath[0] == '.' {
		var err error
		relativePath, err = filepath.Abs(filepath.Join(filepath.Dir(fullPath), relativePath))
		if err != nil {
			c.addErrorf(token.NoPos, "failed to get absolute path of %s: %v", relativePath, err)
			return nil
		}
	}
	return c.files[relativePath]
}

// pushTrace pushes a position to the trace
func (c *Context) pushTrace(pos token.Pos) {
	c.stack = append(c.stack, pos)
}

// popTrace pops a position from the trace
func (c *Context) popTrace() {
	c.stack = c.stack[:len(c.stack)-1]
}

// call calls a function with arguments
func (c *Context) call(pos token.Pos, name string, args ...constant.Value) (constant.Value, error) {
	c.pushTrace(pos)
	defer c.popTrace()
	return constant.Call(c, name, args)
}

// Resolve resolves all files in the context
func (c *Context) Resolve() error {
	// sort files by package name and position
	files := make([]*File, 0, len(c.files))
	for i := range c.files {
		files = append(files, c.files[i])
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].pkg.name != files[j].pkg.name {
			return files[i].pkg.name < files[j].pkg.name
		}
		return files[i].pos < files[j].pos
	})
	c.sortedFiles = files

	// resolve all imports
	for _, file := range files {
		file.imports.resolve(c, file)
	}
	if c.errors.Len() > 0 {
		return c.errors
	}

	// create all symbols
	for _, file := range files {
		for name, symbol := range file.symbols {
			symbolName := joinSymbolName(file.pkg.name, name)
			if prev, ok := c.symbols[symbolName]; ok {
				c.addErrorf(symbol.symbolPos(), "symbol %s redeclared: previous declaration at %s", symbolName, c.fset.Position(prev.symbolPos()))
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

	// resolve all packages
	for _, pkg := range c.packages {
		pkg.resolve(c)
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

// resolveAnnotationGroup resolves an annotation group
func (c *Context) resolveAnnotationGroup(file *File, annotations *ast.AnnotationGroup) AnnotationGroup {
	if annotations == nil {
		return nil
	}
	result := make(AnnotationGroup)
	for _, a := range annotations.List {
		if _, dup := result[a.Name.Name]; dup {
			c.addErrorf(a.Pos(), "annotation %s redeclared", a.Name.Name)
			continue
		}
		params := make(Annotation)
		for _, p := range a.Params {
			name := p.Name.Name
			if _, dup := params[name]; dup {
				c.addErrorf(p.Pos(), "named parameter %s redefined", name)
				continue
			}
			var value constant.Value
			if p.Value != nil {
				value = c.resolveValue(file, p.Value, nil)
			} else {
				value = constant.MakeBool(true)
			}
			params[name] = &AnnotationParam{
				pos:   p.Pos(),
				name:  name,
				value: value,
			}
		}
		result[a.Name.Name] = params
	}
	return result
}

// resolveValue resolves a value of an expression
func (c *Context) resolveValue(file *File, expr ast.Expr, iota *iotaValue) constant.Value {
	return c.recursiveResolveValue(file, file, make([]*ValueSpec, 0, 16), expr, iota)
}

// resolveSymbolValue resolves a value of a symbol
func (c *Context) resolveSymbolValue(file *File, scope Scope, refs []*ValueSpec, v *ValueSpec) constant.Value {
	if v.value != nil {
		// value already resolved
		return v.value
	}
	if index := slices.Index(refs, v); index >= 0 {
		var sb strings.Builder
		for i := index; i < len(refs); i++ {
			fmt.Fprintf(&sb, "\n%s: %s ↓", c.fset.Position(refs[i].symbolPos()), refs[i].name)
		}
		c.addErrorf(v.pos, "cyclic references: %s\n%s: %s", sb.String(), c.fset.Position(v.symbolPos()), v.name)
		return constant.MakeUnknown()
	}
	v.resolveValue(c, file, scope, refs)
	return v.value
}

func (c *Context) recursiveResolveValue(file *File, scope Scope, refs []*ValueSpec, expr ast.Expr, iota *iotaValue) (result constant.Value) {
	defer func() {
		if r := recover(); r != nil {
			c.addErrorf(expr.Pos(), "%v", r)
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
					c.addErrorf(expr.Pos(), "iota is not allowed in this context")
					return constant.MakeUnknown()
				}
				iota.found = true
				return constant.MakeInt64(int64(iota.value))
			}
			c.addErrorf(expr.Pos(), "%s is not defined", expr.Name)
			return constant.MakeUnknown()
		}
		return c.resolveSymbolValue(file, scope, refs, v)

	case *ast.SelectorExpr:
		name := joinSymbolName(c.resolveSelectorExprChain(expr)...)
		v, err := lookupValue(scope, name)
		if err != nil {
			c.addErrorf(expr.Pos(), "%s is not defined", name)
			return constant.MakeUnknown()
		}
		file = c.getFileByPos(v.symbolPos())
		if file == nil {
			c.addErrorf(expr.Pos(), "%s is not defined (file %q not found)", name, c.fset.Position(v.symbolPos()).Filename)
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
			c.addErrorf(expr.Pos(), "unexpected function %T", fun)
			return constant.MakeUnknown()
		}
		args := make([]constant.Value, len(expr.Args))
		for i, arg := range expr.Args {
			args[i] = c.recursiveResolveValue(file, scope, refs, arg, iota)
		}
		result, err := c.call(expr.Pos(), ident.Name, args...)
		if err != nil {
			c.addErrorf(expr.Pos(), err.Error())
			return constant.MakeUnknown()
		}
		return result

	default:
		c.addErrorf(expr.Pos(), "unexpected expression %T", expr)
	}
	return constant.MakeUnknown()
}

// resolveUint64 resolves an unsigned integer value of an expression
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
		c.addErrorf(expr.Pos(), "constant %s overflows uint64", val)
	case constant.Float:
		f, ok := constant.Float64Val(val)
		if ok && f == float64(uint64(f)) {
			return uint64(f)
		}
		c.addErrorf(expr.Pos(), "constant %s overflows uint64", val)
	default:
		c.addErrorf(expr.Pos(), "constant %s is not an integer", val)
	}
	return 0
}

// resolveType resolves a type
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
		c.addErrorf(t.Pos(), "unexpected type %T", t)
		return nil
	}
}

func (c *Context) resolveIdentType(file *File, i *ast.Ident) Type {
	if t, ok := basicTypes[i.Name]; ok {
		return t
	}
	t, err := file.LookupLocalType(i.Name)
	if err != nil {
		c.addErrorf(i.Pos(), "failed to lookup type %s: %s", i.Name, err)
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
			c.addErrorf(t.Pos(), "unexpected selector expression %T", t)
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
		c.addErrorf(t.Pos(), "unexpected selector expression %s", names)
		return nil
	}
	fullName := joinSymbolName(names...)
	typ, err := lookupType(file, fullName)
	if err != nil {
		c.addError(t.Pos(), err.Error())
		return nil
	}
	return typ
}

func (c *Context) resolveArrayType(file *File, t *ast.ArrayType) Type {
	return &ArrayType{
		pos:      t.Pos(),
		ElemType: c.resolveType(file, t.T),
		N:        c.resolveUint64(file, t.N),
	}
}

func (c *Context) resolveVectorType(file *File, t *ast.VectorType) Type {
	return &VectorType{
		pos:      t.Pos(),
		ElemType: c.resolveType(file, t.T),
	}
}

func (c *Context) resolveMapType(file *File, t *ast.MapType) Type {
	return &MapType{
		pos:      t.Pos(),
		KeyType:  c.resolveType(file, t.K),
		ElemType: c.resolveType(file, t.V),
	}
}
