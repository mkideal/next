package compile

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/gopherd/core/flags"

	"github.com/mkideal/next/src/ast"
	"github.com/mkideal/next/src/constant"
	"github.com/mkideal/next/src/grammar"
	"github.com/mkideal/next/src/internal/stringutil"
	"github.com/mkideal/next/src/parser"
	"github.com/mkideal/next/src/scanner"
	"github.com/mkideal/next/src/token"
)

const verboseDebug = 1
const verboseTrace = 2

// Compiler represents a compiler of a Next program
type Compiler struct {
	// command line options
	options  Options
	platform Platform
	builtin  FileSystem
	grammar  grammar.Grammar

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

	// all annotations
	annotations map[token.Pos]*linkedAnnotation
}

// NewCompiler creates a new compiler with builtin language supports
func NewCompiler(platform Platform, builtin FileSystem) *Compiler {
	c := &Compiler{
		platform:    platform,
		builtin:     builtin,
		fset:        token.NewFileSet(),
		files:       make(map[string]*File),
		symbols:     make(map[string]Symbol),
		searchDirs:  createSearchDirs(platform),
		annotations: make(map[token.Pos]*linkedAnnotation),
	}
	c.options.Env = make(flags.Map)
	c.options.Output = make(flags.Map)
	c.options.Templates = make(flags.MapSlice)
	c.options.Mapping = make(flags.Map)
	c.options.Solvers = make(flags.MapSlice)

	return c
}

func (c *Compiler) SetupCommandFlags(flagSet *flag.FlagSet, u flags.UsageFunc) {
	c.options.SetupCommandFlags(flagSet, u)
}

// FileSet returns the file set used to track file positions
func (c *Compiler) FileSet() *token.FileSet {
	return c.fset
}

// GetFile returns the file by path
func (c *Compiler) GetFile(path string) *File {
	return c.files[path]
}

// Files returns all files in the context
func (c *Compiler) Files() map[string]*File {
	return c.files
}

// AddFile adds a file to the context
func (c *Compiler) AddFile(f *ast.File) (*File, error) {
	filename := c.fset.Position(f.Pos()).Filename
	if file, ok := c.files[filename]; ok {
		c.addErrorf(f.Pos(), "file %s already exists", filename)
		return file, c.errors.Err()
	}
	path, err := absolutePath(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of %s: %w", filename, err)
	}
	file := newFile(c, f, path)
	c.files[path] = file
	for _, pkg := range c.packages {
		if pkg.name == f.Name.Name {
			pkg.addFile(file)
			return file, nil
		}
	}
	pkg := newPackage(c, f.Name.Name)
	pkg.addFile(file)
	c.packages = append(c.packages, pkg)
	return file, nil
}

// Output returns the output writer for logging
func (c *Compiler) Output() io.Writer {
	return c.platform.Stderr()
}

func (c *Compiler) log(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
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
func (c *Compiler) Trace(msg string, args ...any) {
	if c.options.Verbose >= verboseTrace {
		c.log(msg, args...)
	}
}

// Debug logs a message if debug logging is enabled
func (c *Compiler) Debug(msg string, args ...any) {
	if c.options.Verbose >= verboseDebug {
		c.log(msg, args...)
	}
}

// Error logs an error message
func (c *Compiler) Error(args ...any) {
	c.log(fmt.Sprint(args...))
}

// Getenv returns the value of an environment variable
func (c *Compiler) Getenv(key string) (string, bool) {
	v, ok := c.options.Env[key]
	return v, ok
}

// Position returns the current position for call expression
func (c *Compiler) Position() token.Position {
	if len(c.stack) == 0 {
		return token.Position{}
	}
	return c.fset.Position(c.stack[len(c.stack)-1])
}

// addError adds an error message with position to the error list
func (c *Compiler) addError(pos token.Pos, msg string) {
	c.errors.Add(c.fset.Position(pos), msg)
}

// addErrorf adds a formatted error message with position to the error list
func (c *Compiler) addErrorf(pos token.Pos, format string, args ...any) {
	c.errors.Add(c.fset.Position(pos), fmt.Sprintf(format, args...))
}

// getFileByPos returns the file by position
func (c *Compiler) getFileByPos(pos token.Pos) *File {
	path, err := absolutePath(c.fset.Position(pos).Filename)
	if err != nil {
		c.addErrorf(pos, "failed to get absolute path of %s: %v", c.fset.Position(pos).Filename, err)
		return nil
	}
	return c.files[path]
}

// lookupFile returns the file by full path or relative path
func (c *Compiler) lookupFile(fullPath, relativePath string) *File {
	if relativePath == "" {
		return nil
	}
	if !filepath.IsAbs(relativePath) {
		var err error
		relativePath, err = absolutePath(filepath.Join(filepath.Dir(fullPath), relativePath))
		if err != nil {
			c.addErrorf(token.NoPos, "failed to get absolute path of %s: %v", relativePath, err)
			return nil
		}
	}
	relativePath = filepath.ToSlash(relativePath)
	return c.files[relativePath]
}

// pushTrace pushes a position to the trace
func (c *Compiler) pushTrace(pos token.Pos) {
	c.stack = append(c.stack, pos)
}

// popTrace pops a position from the trace
func (c *Compiler) popTrace() {
	c.stack = c.stack[:len(c.stack)-1]
}

// call calls a function with arguments
func (c *Compiler) call(pos token.Pos, fun ast.Expr, args ...constant.Value) (constant.Value, error) {
	fun = ast.Unparen(fun)
	var name string
	switch fun := fun.(type) {
	case *ast.Ident:
		name = fun.Name
	default:
		return constant.MakeUnknown(), fmt.Errorf("unexpected function %T", fun)
	}
	c.pushTrace(pos)
	defer c.popTrace()
	return constant.Call(c, name, args)
}

// Resolve resolves all files in the context.
func (c *Compiler) Resolve() error {
	// imports all files
	seen := make(map[string]bool)
	var imports []string
	for _, pkg := range c.packages {
		for _, x := range pkg.imports.List {
			if _, ok := c.files[x.FullPath]; ok {
				seen[x.FullPath] = true
				continue
			}
			if seen[x.FullPath] {
				continue
			}
			seen[x.FullPath] = true
			imports = append(imports, x.FullPath)
		}
	}
	for len(imports) > 0 {
		path := imports[len(imports)-1]
		imports = imports[:len(imports)-1]
		content, err := c.platform.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		f, err := parser.ParseFile(c.fset, path, content, parser.ParseComments)
		if err != nil {
			return err
		}
		file, err := c.AddFile(f)
		if err != nil {
			return err
		}
		file.importedOnly = true
		for _, x := range file.imports.List {
			if seen[x.FullPath] {
				continue
			}
			seen[x.FullPath] = true
			imports = append(imports, x.FullPath)
		}
	}

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

	// create all symbols
	for _, file := range files {
		for name, symbol := range file.symbols {
			symbolName := joinSymbolName(file.pkg.name, name)
			if prev, ok := c.symbols[symbolName]; ok {
				c.addErrorf(symbol.Pos().pos, "symbol %s redeclared: previous declaration at %s", symbolName, prev.Pos())
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

	// resolve all packages
	for _, pkg := range c.packages {
		pkg.resolve(c)
	}

	// resolve all statements
	for _, file := range files {
		for _, stmt := range file.stmts {
			stmt.resolve(c, file)
		}
	}

	// return if there are errors
	if c.errors.Len() > 0 {
		return c.errors
	}

	// solve all annotations by external programs
	if err := c.solveAnnotations(); err != nil {
		return err
	}

	return c.errors.Err()
}

// resolveValue resolves a value of an expression
func (c *Compiler) resolveValue(file *File, expr ast.Expr, iota *iotaValue) constant.Value {
	return c.recursiveResolveValue(file, file, make([]*Value, 0, 16), expr, iota)
}

// resolveSymbolValue resolves a value of a symbol
func (c *Compiler) resolveSymbolValue(file *File, scope Scope, refs []*Value, v *Value) constant.Value {
	if v.val != nil {
		// value already resolved
		return v.val
	}
	if index := slices.Index(refs, v); index >= 0 {
		var sb strings.Builder
		for i := index; i < len(refs); i++ {
			fmt.Fprintf(&sb, "\n%s: %s ↓", refs[i].Pos(), refs[i].name)
		}
		c.addErrorf(v.namePos, "cyclic references: %s\n%s: %s", sb.String(), v.Pos(), v.name)
		return constant.MakeUnknown()
	}
	v.resolveValue(c, file, scope, refs)
	return v.val
}

func (c *Compiler) recursiveResolveValue(file *File, scope Scope, refs []*Value, expr ast.Expr, iota *iotaValue) (result constant.Value) {
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
			c.addError(expr.Pos(), err.Error())
			return constant.MakeUnknown()
		} else if v == nil {
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
		name := joinSymbolName(c.resolveSelectorExprChain(expr, false)...)
		v, err := LookupValue(scope, name)
		if err != nil {
			c.addErrorf(expr.Pos(), err.Error())
			return constant.MakeUnknown()
		} else if v == nil {
			c.addErrorf(expr.Pos(), "%s is not defined", name)
			return constant.MakeUnknown()
		}
		file = c.getFileByPos(v.namePos)
		if file == nil {
			c.addErrorf(expr.Pos(), "%s is not defined (file %q not found)", name, v.Pos().Filename)
			return constant.MakeUnknown()
		}
		return c.resolveSymbolValue(file, scope, refs, v)

	case *ast.BinaryExpr:
		x := c.recursiveResolveValue(file, scope, refs, expr.X, iota)
		y := c.recursiveResolveValue(file, scope, refs, expr.Y, iota)
		switch expr.Op {
		case token.SHL, token.SHR:
			return constant.Shift(x, expr.Op, uint(c.recursiveResolveInt64(file, scope, expr.Y, refs, iota)))
		case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
			return constant.MakeBool(constant.Compare(x, expr.Op, y))
		default:
			return constant.BinaryOp(x, expr.Op, y)
		}

	case *ast.UnaryExpr:
		x := c.recursiveResolveValue(file, scope, refs, expr.X, iota)
		return constant.UnaryOp(expr.Op, x, 0)

	case *ast.CallExpr:
		args := make([]constant.Value, len(expr.Args))
		for i, arg := range expr.Args {
			args[i] = c.recursiveResolveValue(file, scope, refs, arg, iota)
		}
		result, err := c.call(expr.Pos(), expr.Fun, args...)
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
func (c *Compiler) resolveInt64(file *File, expr ast.Expr) int64 {
	return c.recursiveResolveInt64(file, file, expr, make([]*Value, 0, 16), nil)
}

func (c *Compiler) recursiveResolveInt64(file *File, scope Scope, expr ast.Expr, refs []*Value, iota *iotaValue) int64 {
	val := c.recursiveResolveValue(file, scope, refs, expr, iota)
	switch val.Kind() {
	case constant.Int:
		n, ok := constant.Int64Val(val)
		if ok {
			return n
		}
		c.addErrorf(expr.Pos(), "constant %s overflows uint64", val)
	case constant.Float:
		f, ok := constant.Float64Val(val)
		if ok && f == float64(int64(f)) {
			return int64(f)
		}
		c.addErrorf(expr.Pos(), "constant %s overflows uint64", val)
	default:
		c.addErrorf(expr.Pos(), "constant %s is not an integer", val)
	}
	return 0
}

// resolveType resolves a type
func (c *Compiler) resolveType(depth int, node Node, src ast.Type, ignoreError bool) Type {
	var result Type
	switch t := src.(type) {
	case *ast.Ident:
		result = c.resolveIdentType(node, t, ignoreError)
	case *ast.SelectorExpr:
		result = c.resolveSelectorExprType(node, t, ignoreError)
	case *ast.ArrayType:
		result = c.resolveArrayType(depth, node, t, ignoreError)
	case *ast.VectorType:
		result = c.resolveVectorType(depth, node, t, ignoreError)
	case *ast.MapType:
		result = c.resolveMapType(depth, node, t, ignoreError)
	default:
		if ignoreError {
			return nil
		}
		c.addErrorf(t.Pos(), "unexpected type %T", t)
		return nil
	}
	if result != nil {
		result = Use(depth, src, node, result)
		node.File().addObject(c, src, result)
	}
	return result
}

func (c *Compiler) resolveIdentType(node Node, i *ast.Ident, ignoreError bool) Type {
	if t, ok := primitiveTypes[i.Name]; ok {
		return t
	}
	t, err := node.File().LookupLocalType(i.Name)
	if err != nil {
		if !ignoreError {
			c.addErrorf(i.Pos(), "failed to lookup type %s: %s", i.Name, err)
		}
		return nil
	} else if t == nil {
		if !ignoreError {
			c.addErrorf(i.Pos(), "type %s is not defined", i.Name)
		}
		return nil
	}
	return t
}

func (c *Compiler) resolveSelectorExprChain(t *ast.SelectorExpr, ignoreError bool) []string {
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
			if !ignoreError {
				c.addErrorf(t.Pos(), "unexpected selector expression %T", t)
			}
			return nil
		}
	}

	slices.Reverse(names)
	return names
}

func (c *Compiler) resolveSelectorExprType(node Node, t *ast.SelectorExpr, ignoreError bool) Type {
	names := c.resolveSelectorExprChain(t, ignoreError)
	if len(names) < 2 {
		return nil
	}
	if len(names) != 2 {
		if !ignoreError {
			c.addErrorf(t.Pos(), "unexpected selector expression %s", names)
		}
		return nil
	}
	fullName := joinSymbolName(names...)
	typ, err := LookupType(node.File(), fullName)
	if err != nil {
		if !ignoreError {
			c.addError(t.Pos(), err.Error())
		}
		return nil
	} else if typ == nil {
		if !ignoreError {
			c.addErrorf(t.Pos(), "type %s is not defined", fullName)
		}
		return nil
	}
	return typ
}

func (c *Compiler) resolveArrayType(depth int, node Node, t *ast.ArrayType, ignoreError bool) Type {
	typ := c.resolveType(depth+1, node, t.T, ignoreError)
	if typ == nil {
		return nil
	}
	return &ArrayType{
		pos:      t.Pos(),
		ElemType: typ,
		N:        c.resolveInt64(node.File(), t.N),
		lenExpr:  t.N,
	}
}

func (c *Compiler) resolveVectorType(depth int, node Node, t *ast.VectorType, ignoreError bool) Type {
	typ := c.resolveType(depth+1, node, t.T, ignoreError)
	if typ == nil {
		return nil
	}
	return &VectorType{
		pos:      t.Pos(),
		ElemType: typ,
	}
}

func (c *Compiler) resolveMapType(depth int, node Node, t *ast.MapType, ignoreError bool) Type {
	k := c.resolveType(depth+1, node, t.K, ignoreError)
	if k == nil {
		return nil
	}
	v := c.resolveType(depth+1, node, t.V, ignoreError)
	if v == nil {
		return nil
	}
	return &MapType{
		pos:      t.Pos(),
		KeyType:  k,
		ElemType: v,
	}
}

// ValidateGrammar validates the grammar of the context
func (c *Compiler) ValidateGrammar() error {
	for _, pkg := range c.packages {
		// Validate package
		if err := c.validatePackageGrammar(pkg); err != nil {
			return err
		}
		// Validate imports
		if c.grammar.Import.Disabled && len(pkg.imports.List) > 0 {
			c.addErrorf(pkg.imports.List[0].pos, "import declaration is not allowed")
		}
		// Validate constants
		if c.grammar.Const.Disabled && len(pkg.decls.consts) > 0 {
			c.addErrorf(pkg.decls.consts[0].pos, "const declaration is not allowed")
		} else {
			for _, x := range pkg.decls.consts {
				if err := c.validateConstGrammar(x); err != nil {
					return err
				}
			}
		}
		// Validate enums
		if c.grammar.Enum.Disabled && len(pkg.decls.enums) > 0 {
			c.addErrorf(pkg.decls.enums[0].pos, "enum declaration is not allowed")
		} else {
			for _, x := range pkg.decls.enums {
				if err := c.validateEnumGrammar(x); err != nil {
					return err
				}
			}
		}
		// Validate structs
		if c.grammar.Struct.Disabled && len(pkg.decls.structs) > 0 {
			c.addErrorf(pkg.decls.structs[0].pos, "struct declaration is not allowed")
		} else {
			for _, x := range pkg.decls.structs {
				if err := c.validateStructGrammar(x); err != nil {
					return err
				}
			}
		}
		// Validate interfaces
		if c.grammar.Interface.Disabled && len(pkg.decls.interfaces) > 0 {
			c.addErrorf(pkg.decls.interfaces[0].pos, "interface declaration is not allowed")
		} else {
			for _, x := range pkg.decls.interfaces {
				if err := c.validateInterfaceGrammar(x); err != nil {
					return err
				}
			}
		}
	}
	return c.errors.Err()
}

func (c *Compiler) validatePackageGrammar(x *Package) error {
	if err := c.executeValidators(x, c.grammar.Package.Validators); err != nil {
		return err
	}
	for _, file := range x.files {
		if err := c.validateAnnotations(file, c.grammar.Package.Annotations); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) validateConstGrammar(x *Const) error {
	if err := c.executeValidators(x, c.grammar.Const.Validators); err != nil {
		return err
	}
	if err := c.validateAnnotations(x, c.grammar.Const.Annotations); err != nil {
		return err
	}
	if len(c.grammar.Const.Types) > 0 {
		if !slices.Contains(c.grammar.Const.Types, x.value.typ.kind.grammarType()) {
			c.addErrorf(x.namePos, "constant type %s is not allowed: expected one of %v", x.value.typ.kind, c.grammar.Const.Types)
		}
	}
	return nil
}

func (c *Compiler) validateEnumGrammar(x *Enum) error {
	if err := c.executeValidators(x, c.grammar.Enum.Validators); err != nil {
		return err
	}
	if err := c.validateAnnotations(x, c.grammar.Enum.Annotations); err != nil {
		return err
	}
	if len(c.grammar.Enum.Member.Types) > 0 {
		if !slices.Contains(c.grammar.Enum.Member.Types, x.MemberType.kind.grammarType()) {
			c.addErrorf(x.namePos, "enum member type %s is not allowed: expected one of %v", x.MemberType.kind, c.grammar.Enum.Member.Types)
		}
	}
	if c.grammar.Enum.Member.ValueRequired {
		for _, m := range x.Members.List {
			if m.value.unresolved.value == nil {
				c.addErrorf(m.namePos, "enum member %s: value is required", m.Name())
			}
		}
	}
	if c.grammar.Enum.Member.ZeroRequired {
		if len(x.Members.List) == 0 {
			c.addErrorf(x.namePos, "enum should have at least one member")
		}
		hasZero := false
		for _, m := range x.Members.List {
			if rv := reflect.ValueOf(m.value.Actual()); (rv.CanInt() && rv.Int() == 0) || (rv.CanUint() && rv.Uint() == 0) {
				hasZero = true
				break
			}
		}
		if !hasZero {
			c.addErrorf(x.namePos, "enum should have a zero value member")
		}
	}
	for _, m := range x.Members.List {
		if err := c.executeValidators(m, c.grammar.Enum.Member.Validators); err != nil {
			return err
		}
		if err := c.validateAnnotations(m, c.grammar.Enum.Member.Annotations); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) validateStructGrammar(x *Struct) error {
	if err := c.executeValidators(x, c.grammar.Struct.Validators); err != nil {
		return err
	}
	if err := c.validateAnnotations(x, c.grammar.Struct.Annotations); err != nil {
		return err
	}
	for _, f := range x.fields.List {
		if err := c.executeValidators(f, c.grammar.Struct.Field.Validators); err != nil {
			return err
		}
		if err := c.validateAnnotations(f, c.grammar.Struct.Field.Annotations); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) validateInterfaceGrammar(x *Interface) error {
	if err := c.executeValidators(x, c.grammar.Interface.Validators); err != nil {
		return err
	}
	if err := c.validateAnnotations(x, c.grammar.Interface.Annotations); err != nil {
		return err
	}
	for _, m := range x.methods.List {
		if err := c.executeValidators(m, c.grammar.Interface.Method.Validators); err != nil {
			return err
		}
		if err := c.validateAnnotations(m, c.grammar.Interface.Method.Annotations); err != nil {
			return err
		}
		for _, p := range m.Params.List {
			if err := c.executeValidators(p, c.grammar.Interface.Method.Parameter.Validators); err != nil {
				return err
			}
			if err := c.validateAnnotations(p, c.grammar.Interface.Method.Parameter.Annotations); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiler) executeValidators(node Node, validators grammar.Validators) error {
	for _, v := range validators {
		if ok, err := v.Value().Validate(node); err != nil {
			return fmt.Errorf("%s: failed to validate %s %s: %w", node.NamePos(), node.Typeof(), node.Name(), err)
		} else if !ok {
			c.addErrorf(node.NamePos().pos, "%s %s: %s", node.Typeof(), node.Name(), v.Value().Message)
		}
	}
	return nil
}

func (c *Compiler) validateAnnotations(node Node, annotations grammar.Annotations) error {
	for name, annotation := range node.Annotations() {
		ga := grammar.LookupAnnotation(annotations, name)
		if ga == nil {
			if !c.options.Strict {
				continue
			}
			s := stringutil.FindBestMatchFunc(slices.All(annotations), name, stringutil.DefaultSimilarityThreshold, func(i int, a grammar.Options[grammar.Annotation]) string {
				return a.Value().Name
			})
			if s != "" {
				c.addErrorf(annotation.Pos().pos, "unknown annotation %s, did you mean %s?", name, s)
			} else {
				c.addErrorf(annotation.Pos().pos, "unknown annotation %s", name)
			}
			continue
		}
		for _, av := range ga.Validators {
			if ok, err := av.Value().Validate(annotation); err != nil {
				return fmt.Errorf("%s: failed to validate annotation %s(%s): %w", annotation.Pos(), name, annotation.Node().Name(), err)
			} else if !ok {
				c.addErrorf(annotation.Pos().pos, "annotation %s: %s", name, av.Value().Message)
			}
		}
		for p, v := range annotation {
			if strings.HasPrefix(p, "_") {
				// Skip private parameters added by the compiler
				continue
			}
			gp := grammar.LookupAnnotationParameter(ga.Parameters, p)
			if gp == nil {
				if !c.options.Strict && name != "next" {
					continue
				}
				s := stringutil.FindBestMatchFunc(slices.All(ga.Parameters), p, stringutil.DefaultSimilarityThreshold, func(i int, a grammar.Options[grammar.AnnotationParameter]) string {
					name := a.Value().Name
					if strings.HasPrefix(name, ".+_") {
						if index := strings.Index(p, "_"); index > 0 {
							return p[:index] + strings.TrimPrefix(name, ".+")
						}
					} else if strings.HasSuffix(name, "_.+") {
						if index := strings.LastIndex(p, "_"); index > 0 {
							return strings.TrimSuffix(name, "_.+") + p[index:]
						}
					}
					return name
				})
				if s != "" {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: unknown parameter %s, did you mean %s?", name, p, s)
				} else {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: unknown parameter %s", name, p)
				}
				continue
			}
			if slices.Contains(gp.Types, grammar.Any) {
				continue
			}
			rv := reflect.ValueOf(v)
			switch {
			case rv.Kind() == reflect.String:
				if !slices.Contains(gp.Types, grammar.String) {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s should be string", name, p)
				}
			case rv.CanInt() || rv.CanUint():
				if !slices.Contains(gp.Types, grammar.Int) {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s should be integer", name, p)
				}
			case rv.Kind() == reflect.Bool:
				if !slices.Contains(gp.Types, grammar.Bool) {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s should be boolean", name, p)
				}
			case rv.CanFloat():
				if !slices.Contains(gp.Types, grammar.Float) {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s should be float", name, p)
				}
			default:
				if !slices.Contains(gp.Types, grammar.Type) {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s: unexpected value type %T, expected one of %v", name, p, v, gp.Types)
				}
				if t, ok := v.(Type); !ok || t == nil {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s should be a type", name, p)
				}
			}
			for _, pv := range gp.Validators {
				if ok, err := pv.Value().Validate(v); err != nil {
					return fmt.Errorf("%s: failed to validate annotation %s parameter %s: %w", annotation.NamePos(p), name, p, err)
				} else if !ok {
					c.addErrorf(annotation.NamePos(p).pos, "annotation %s: parameter %s: %s", name, p, pv.Value().Message)
				}
			}
			// Validate available expression
			if name == "next" && p == "available" {
				if pos, err := validateAvailable(c, annotation, v); err != nil {
					if !pos.IsValid() {
						pos = annotation.NamePos(name).pos
					}
					c.addError(pos, err.Error())
				}
			}
		}
		for _, gp := range ga.Parameters {
			if !gp.Value().Required {
				continue
			}
			if _, ok := annotation[gp.Value().Name]; !ok {
				c.addErrorf(annotation.Pos().pos, "annotation %s: missing required parameter %s", name, gp.Value().Name)
			}
		}
	}
	return nil
}
