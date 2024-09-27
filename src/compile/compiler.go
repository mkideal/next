package compile

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/term"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/scanner"
	"github.com/next/next/src/token"
)

const verboseDebug = 1
const verboseTrace = 2

// Compiler represents a compiler of a Next program
type Compiler struct {
	// command line flags
	flags struct {
		verbose   int
		envs      flags.Map
		outputs   flags.Map
		templates flags.MapSlice
		mappings  flags.Map
		solvers   flags.MapSlice
	}
	platform Platform
	// builtin builtin
	builtin FileSystem

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
	c.flags.envs = make(flags.Map)
	c.flags.outputs = make(flags.Map)
	c.flags.templates = make(flags.MapSlice)
	c.flags.mappings = make(flags.Map)
	c.flags.solvers = make(flags.MapSlice)

	return c
}

func (c *Compiler) SetupCommandFlags(flagSet *flag.FlagSet, u flags.UsageFunc) {
	isAnsiSupported := term.IsTerminal(flagSet.Output()) && term.IsSupportsAnsi()
	grey := func(s string) string {
		if isAnsiSupported {
			return term.Gray.Format(s)
		}
		return s
	}
	b := func(s string) string {
		if isAnsiSupported {
			return term.Bold.Format(s)
		}
		return s
	}

	// @api(CommandLine/Flag/-v) represents the verbosity level of the compiler.
	// The default value is **0**, which only shows error messages.
	// The value **1** shows debug messages, and **2** shows trace messages.
	// Usually, the trace message is used for debugging the compiler itself.
	// Levels **1** (debug) and above enable execution of:
	// - `print` and `printf` in Next source files (.next).
	// - `debug` in Next template files (.npl).
	//
	// Example:
	//
	//	```sh
	//	next -v 1 ...
	//	```
	flagSet.IntVar(&c.flags.verbose, "v", 0, u(""+
		"Control verbosity of compiler output and debugging information.\n"+
		"`VERBOSE` levels: "+b("0")+"=error,  "+b("1")+"=debug, "+b("2")+"=trace.\n"+
		"Usually, the trace message used for debugging the compiler itself.\n"+
		"Levels "+b("1")+" (debug) and above enable execution of:\n"+
		" -"+b("print")+" and "+b("printf")+" in Next source filesa (.next).\n"+
		" -"+b("debug")+" in Next template files (.npl).\n",
	))

	// @api(CommandLine/Flag/-D) represents the custom environment variables for code generation.
	// The value is a map of environment variable names and their optional values.
	//
	// Example:
	//
	//	```sh
	//	next -D VERSION=2.1 -D DEBUG -D NAME=myapp ...
	//	```
	//
	//	```npl
	//	{{env.NAME}}
	//	{{env.VERSION}}
	//	```
	//
	// Output:
	//
	//	```
	//	myapp
	//	2.1
	//	```
	flagSet.Var(&c.flags.envs, "D", u(""+
		"Define custom environment variables for use in code generation templates.\n"+
		"`NAME"+grey("[=VALUE]")+"` represents the variable name and its optional value.\n"+
		"Example:\n"+
		"  next -D VERSION=2.1 -D DEBUG -D NAME=myapp\n"+
		"And then, use the variables in templates like this: {{env.NAME}}, {{env.VERSION}}\n",
	))

	// @api(CommandLine/Flag/-O) represents the output directories for generated code of each target language.
	//
	// Example:
	//
	//	```sh
	//	next -O go=./output/go -O ts=./output/ts ...
	//	```
	//
	// :::tip
	//
	// The `{{meta.path}}` is relative to the output directory.
	//
	// :::
	flagSet.Var(&c.flags.outputs, "O", u(""+
		"Set output directories for generated code, organized by target language.\n"+
		"`LANG=DIR` specifies the target language and its output directory.\n"+
		"Example:\n"+
		"  next -O go=./output/go -O ts=./output/ts\n",
	))

	// @api(CommandLine/Flag/-T) represents the custom template directories or files for each target language.
	// You can specify multiple templates for a single language.
	//
	// Example:
	//
	//	```sh
	//	next -T go=./templates/go \
	//	     -T go=./templates/go_extra.npl \
	//	     -T python=./templates/python.npl \
	//	     ...
	//	```
	flagSet.Var(&c.flags.templates, "T", u(""+
		"Specify custom template directories or files for each target language.\n"+
		"`LANG=PATH` defines the target language and its template directory or file.\n"+
		"Multiple templates can be specified for a single language.\n"+
		"Example:\n"+
		"  next \\\n"+
		"    -T go=./templates/go \\\n"+
		"    -T go=./templates/go_extra.npl \\\n"+
		"    -T python=./templates/python.npl\n",
	))

	// @api(CommandLine/Flag/-M) represents the language-specific type mappings and features.
	//
	// Example:
	//
	//	```sh
	//	next -M cpp.vector="std::vector<%T%>" \
	//	     -M java.array="ArrayList<%T%>" \
	//	     -M go.map="map[%K%]%V%" \
	//	     -M python.ext=.py \
	//	     -M ruby.comment="# %T%" \
	//	     ...
	//	```
	flagSet.Var(&c.flags.mappings, "M", u(""+
		"Configure language-specific type mappings and features.\n"+
		"`LANG.KEY=VALUE` specifies the mappings for a given language and type/feature.\n"+
		"Type mappings: Map Next types to language-specific types.\n"+
		"  Primitive types: int, int8, int16, int32, int64, bool, string, any, byte, bytes\n"+
		"  Generic types: vector, array, map\n"+
		"    "+b("%T%")+", "+b("%N%")+", "+b("%K%")+", "+b("%V%")+" are placeholders replaced with actual types or values.\n"+
		"Feature mappings: Set language-specific properties like file extensions or comment styles.\n"+
		"Example:\n"+
		"  next \\\n"+
		"    -M cpp.vector=\"std::vector<%T%>\" \\\n"+
		"    -M java.array=\"ArrayList<%T%>\" \\\n"+
		"    -M go.map=\"map[%K%]%V%\" \\\n"+
		"    -M python.ext=.py \\\n"+
		"    -M ruby.comment=\"# %T%\"\n",
	))

	// @api(CommandLine/Flag/-X) represents the custom annotation solver programs for code generation.
	// Annotation solvers are executed in a separate process to solve annotations.
	// All annotations are passed to the solver program via stdin and stdout.
	// The built-in annotation `next` is reserved for the Next compiler.
	//
	// Example:
	//
	//	```sh
	//	next -X message="message-type-allocator message-types.json" ...
	//	```
	//
	// :::tip
	//
	// In the example above, the `message-type-allocator` is a custom annotation solver program that
	// reads the message types from the `message-types.json` file and rewrite the message types to the
	// `message-types.json` file.
	//
	// :::
	flagSet.Var(&c.flags.solvers, "X", u(""+
		"Specify custom annotation solver programs for code generation.\n"+
		"`ANNOTATION=PROGRAM` defines the target annotation and its solver program.\n"+
		"Annotation solvers are executed in a separate process to solve annotations.\n"+
		"All annotations are passed to the solver program via stdin and stdout.\n"+
		b("NOTE")+": built-in annotation 'next' is reserved for the Next compiler.\n"+
		"Example:\n"+
		"  next -X message=\"message-type-allocator message-types.json\"\n",
	))
}

// FileSet returns the file set used to track file positions
func (c *Compiler) FileSet() *token.FileSet {
	return c.fset
}

// GetFile returns the file by path
func (c *Compiler) GetFile(path string) *File {
	return c.files[path]
}

// AddFile adds a file to the context
func (c *Compiler) AddFile(f *ast.File) (*File, error) {
	filename := c.fset.Position(f.Pos()).Filename
	if file, ok := c.files[filename]; ok {
		c.addErrorf(f.Pos(), "file %s already exists", filename)
		return file, c.errors.Err()
	}
	path, err := filepath.Abs(filename)
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

// IsDebugEnabled returns true if debug logging is enabled
func (c *Compiler) IsDebugEnabled() bool {
	return c.flags.verbose >= 2
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
	if c.flags.verbose >= verboseTrace {
		c.log(msg, args...)
	}
}

// Debug logs a message if debug logging is enabled
func (c *Compiler) Debug(msg string, args ...any) {
	if c.flags.verbose >= verboseDebug {
		c.log(msg, args...)
	}
}

// Error logs an error message
func (c *Compiler) Error(args ...any) {
	c.log(fmt.Sprint(args...))
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
	path, err := filepath.Abs(c.fset.Position(pos).Filename)
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
func (c *Compiler) pushTrace(pos token.Pos) {
	c.stack = append(c.stack, pos)
}

// popTrace pops a position from the trace
func (c *Compiler) popTrace() {
	c.stack = c.stack[:len(c.stack)-1]
}

// call calls a function with arguments
func (c *Compiler) call(pos token.Pos, name string, args ...constant.Value) (constant.Value, error) {
	c.pushTrace(pos)
	defer c.popTrace()
	return constant.Call(c, name, args)
}

// Resolve resolves all files in the context.
func (c *Compiler) Resolve() error {
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
func (c *Compiler) resolveType(file *File, t ast.Type, ignoreError bool) Type {
	var result Type
	switch t := t.(type) {
	case *ast.Ident:
		result = c.resolveIdentType(file, t, ignoreError)
	case *ast.SelectorExpr:
		result = c.resolveSelectorExprType(file, t, ignoreError)
	case *ast.ArrayType:
		result = c.resolveArrayType(file, t, ignoreError)
	case *ast.VectorType:
		result = c.resolveVectorType(file, t, ignoreError)
	case *ast.MapType:
		result = c.resolveMapType(file, t, ignoreError)
	default:
		if ignoreError {
			return nil
		}
		c.addErrorf(t.Pos(), "unexpected type %T", t)
		return nil
	}
	if result != nil {
		result = Use(result, file, t)
		file.addObject(c, t, result)
	}
	return result
}

func (c *Compiler) resolveIdentType(file *File, i *ast.Ident, ignoreError bool) Type {
	if t, ok := primitiveTypes[i.Name]; ok {
		return t
	}
	t, err := file.LookupLocalType(i.Name)
	if err != nil {
		if !ignoreError {
			c.addErrorf(i.Pos(), "failed to lookup type %s: %s", i.Name, err)
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

func (c *Compiler) resolveSelectorExprType(file *File, t *ast.SelectorExpr, ignoreError bool) Type {
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
	typ, err := LookupType(file, fullName)
	if err != nil {
		if !ignoreError {
			c.addError(t.Pos(), err.Error())
		}
		return nil
	}
	return typ
}

func (c *Compiler) resolveArrayType(file *File, t *ast.ArrayType, ignoreError bool) Type {
	typ := c.resolveType(file, t.T, ignoreError)
	if typ == nil {
		return nil
	}
	return &ArrayType{
		pos:      t.Pos(),
		ElemType: typ,
		N:        c.resolveInt64(file, t.N),
	}
}

func (c *Compiler) resolveVectorType(file *File, t *ast.VectorType, ignoreError bool) Type {
	typ := c.resolveType(file, t.T, ignoreError)
	if typ == nil {
		return nil
	}
	return &VectorType{
		pos:      t.Pos(),
		ElemType: typ,
	}
}

func (c *Compiler) resolveMapType(file *File, t *ast.MapType, ignoreError bool) Type {
	k := c.resolveType(file, t.K, ignoreError)
	if k == nil {
		return nil
	}
	v := c.resolveType(file, t.V, ignoreError)
	if v == nil {
		return nil
	}
	return &MapType{
		pos:      t.Pos(),
		KeyType:  k,
		ElemType: v,
	}
}
