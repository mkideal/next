package compile

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/text/document"

	"github.com/mkideal/next/src/internal/fsutil"
)

const (
	nextDir       = "next.d" // next directory for configuration files
	hiddenNextDir = ".next"  // hidden next directory for configuration files in user's home directory
	templateExt   = ".npl"   // next template file extension
	langMapExt    = ".map"   // next lang map file extension
)

func createSearchDirs(platform Platform) []string {
	var dirs []string

	// Add the current directory for JavaScript runtime
	if runtime.GOOS == "js" {
		return append(dirs, ".")
	}

	// Add the NEXT_PATH
	if nextPath := platform.Getenv(NEXT_PATH); nextPath != "" {
		dirs = addSearchDirs(dirs, filepath.SplitList(nextPath)...)
	}

	// Add the hidden next directory ".next" in the user's home directory
	homedir, err := platform.UserHomeDir()
	if err == nil {
		dirs = addSearchDirs(dirs, filepath.Join(homedir, hiddenNextDir))
	}

	if runtime.GOOS == "windows" {
		// Add the APPDATA directory for Windows
		appData := platform.Getenv("APPDATA")
		if appData != "" {
			dirs = addSearchDirs(dirs, filepath.Join(appData, nextDir))
		}
	} else {
		if runtime.GOOS == "darwin" {
			// Add the Library/Application Support directory for macOS
			dirs = addSearchDirs(dirs, filepath.Join("/Library/Application Support", nextDir))
			if homedir != "" {
				dirs = addSearchDirs(dirs, filepath.Join(homedir, "Library/Application Support", nextDir))
			}
		}
		// Add the system etc directories
		dirs = addSearchDirs(dirs,
			filepath.Join("/usr/local/etc", nextDir),
			filepath.Join("/etc", nextDir),
		)
	}
	return dirs
}

func addSearchDirs(dirs []string, paths ...string) []string {
	for _, path := range paths {
		if path != "" {
			var found bool
			for _, dir := range dirs {
				if dir == path {
					found = true
					break
				}
			}
			if !found {
				dirs = append(dirs, path)
			}
		}
	}
	return dirs
}

// Genertate generates files for each language specified in the flags.outputs.
func Generate(c *Compiler) error {
	if len(c.options.Output) == 0 {
		return nil
	}
	c.Trace("flags.envs: ", c.options.Env)
	c.Trace("flags.outputs: ", c.options.Output)
	c.Trace("flags.templates: ", c.options.Templates)
	c.Trace("flags.mappings: ", c.options.Mapping)

	if c.options.Output.Get("next") != "" {
		return fmt.Errorf("output language 'next' is not supported")
	}

	// Check whether the template directory or file exists for each language
	for lang := range c.options.Output {
		for _, tmplPath := range c.options.Templates[lang] {
			if c.platform.IsNotExist(tmplPath) {
				return fmt.Errorf("template path %q not found: %q", lang, tmplPath)
			}
		}
	}

	// Load all mappings from all map files
	m := make(flags.Map)
	for lang := range c.options.Output {
		if err := loadMap(c, m, lang); err != nil {
			return err
		}
	}
	for k, v := range c.options.Mapping {
		m[k] = v
	}
	c.options.Mapping = m
	if c.options.Verbose >= verboseTrace {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			c.Trace("map[%q] = %q", k, m[k])
		}
	}

	// Generate files for each language
	langs := make([]string, 0, len(c.options.Output))
	for lang := range c.options.Output {
		langs = append(langs, lang)
	}
	slices.Sort(langs)
	for _, lang := range langs {
		dir := c.options.Output[lang]
		ext := op.Or(c.options.Mapping[lang+".ext"], "."+lang)
		tempPaths := c.options.Templates[lang]
		if len(tempPaths) == 0 {
			return fmt.Errorf("no template directory specified for %q", lang)
		}
		for _, tempPath := range tempPaths {
			if err := generateForTemplatePath(c, lang, ext, dir, tempPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// @api(Map)
//
//	import Tabs from "@theme/Tabs";
//	import TabItem from "@theme/TabItem";
//
// `Map` represents a language map file. Usually, it is a text file that contains
// key-value pairs separated by a equal sign. The key is the name of the language feature.
// The built-in keys are:
//
//   - **ext**: the file extension for the language
//   - **comment**: the line comment pattern for the language
//   - **int**: the integer type for the language
//   - **int8**: the 8-bit integer type for the language
//   - **int16**: the 16-bit integer type for the language
//   - **int32**: the 32-bit integer type for the language
//   - **int64**: the 64-bit integer type for the language
//   - **float32**: the 32-bit floating-point type for the language
//   - **float64**: the 64-bit floating-point type for the language
//   - **bool**: the boolean type for the language
//   - **string**: the string type for the language
//   - **byte**: the byte type for the language
//   - **bytes**: the byte slice type for the language
//   - **any**: the any type for the language
//   - **array**: the fixed-sized array type for the language
//   - **vector**: the slice type for the language
//   - **map**: the map type for the language
//   - **box**: the box replacement for the language (for java, e.g., box(int) = Integer)
//
// Here are some examples of language map files:
//
//	<Tabs
//	  defaultValue="java"
//	  values={[
//	    { label: 'java.map', value: 'java' },
//	    { label: 'cpp.map', value: 'cpp' },
//	    { label: 'protobuf.map', value: 'protobuf' },
//	  ]}>
//
//	<TabItem value="java">
//	```map title="java.map"
//	ext=.java
//	comment=// %T%
//
//	# primitive types
//	int=int
//	int8=byte
//	int16=short
//	int32=int
//	int64=long
//	float32=float
//	float64=double
//	bool=boolean
//	string=String
//	byte=byte
//	bytes=byte[]
//	any=Object
//	map=Map<box(%K%),box(%V%)>
//	vector=List<box(%T%)>
//	array=%T%[]
//
//	# box types
//	box(int)=Integer
//	box(byte)=Byte
//	box(short)=Short
//	box(long)=Long
//	box(float)=Float
//	box(double)=Double
//	box(boolean)=Boolean
//	```
//	</TabItem>
//
//	<TabItem value="cpp">
//	```map title="cpp.map"
//	ext=.h
//	comment=// %T%
//
//	int=int
//	int8=int8_t
//	int16=int16_t
//	int32=int32_t
//	int64=int64_t
//	float32=float
//	float64=double
//	bool=bool
//	string=std::string
//	byte=unsigned char
//	bytes=std::vector<unsigned char>
//	any=std::any
//	map=std::unordered_map<%K%, %V%>
//	vector=std::vector<%T%>
//	array=std::array<%T%, %N%>
//	```
//	</TabItem>
//
//	<TabItem value="protobuf">
//	```map title="protobuf.map"
//	ext=.proto
//	comment=// %T%
//
//	int=int32
//	int8=int32
//	int16=int32
//	int32=int32
//	int64=int64
//	float32=float
//	float64=double
//	bool=bool
//	string=string
//	byte=int8
//	bytes=bytes
//	any=google.protobuf.Any
//	map=map<%K%, %V%>
//	vector=repeated %T.E%
//	array=repeated %T.E%
//	```
//	</TabItem>
//
//	</Tabs>
//
// :::tip
//
// - The `%T%`, `%K%`, `%V%` and `%N%` are placeholders for the type, key type, value type and number of array elements.
// - `%T.E%` is the final element type of `%T%`. It's useful if you need to get the final element type of a vector or array.
// - Line comments are started with the `#` character and it must be the first character of the line (leading spaces are allowed).
//
//	```plain title="java.map"
//	# This will error
//	ext=.java # This is an invalid comment
//	# Comment must be the first character of the line
//
//		# This is a valid comment (leading spaces are allowed)
//	```
//
// See [Builtin Map Files](https://github.com/mkideal/next/tree/main/builtin) for more information.
//
// :::
func loadMap(c *Compiler, m flags.Map, lang string) error {
	f, err := c.builtin.Open("builtin/" + lang + langMapExt)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to open builtin %q: %v", lang+langMapExt, err)
		}
	}
	if f != nil {
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read builtin %q: %v", lang+langMapExt, err)
		}
		if err := parseLangMap(m, lang, content); err != nil {
			return fmt.Errorf("failed to parse builtin %q: %v", lang+langMapExt, err)
		}
	}

	// Load the language map from the search directories in reverse order
	for i := len(c.searchDirs) - 1; i >= 0; i-- {
		dir := c.searchDirs[i]
		path := filepath.Join(dir, lang+langMapExt)
		if _, err := loadMapFromFile(c, m, lang, path); err != nil {
			return fmt.Errorf("failed to load %q: %v", path, err)
		}
	}
	return nil
}

func loadMapFromFile(c *Compiler, m flags.Map, lang, path string) (bool, error) {
	content, err := c.platform.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, parseLangMap(m, lang, content)
}

func generateForTemplatePath(c *Compiler, lang, ext, dir, tmplPath string) error {
	tmplFiles, err := fsutil.AppendFiles(nil, tmplPath, templateExt, false)
	if err != nil {
		return fmt.Errorf("failed to list template files in %q: %v", tmplPath, err)
	}

	// Generate files for each template
	for _, tmplFile := range tmplFiles {
		if err := generateForTemplateFile(c, lang, ext, dir, tmplFile); err != nil {
			return err
		}
	}
	return nil
}

func generateForTemplateFile(c *Compiler, lang, ext, dir, tmplFile string) error {
	content, err := c.platform.ReadFile(tmplFile)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", tmplFile, err)
	}

	tc := newTemplateContext(templateContextInfo{
		compiler: c,
		lang:     lang,
		dir:      dir,
		ext:      ext,
	})
	t, err := createTemplate(tmplFile, string(content), tc.funcs)
	if err != nil {
		return err
	}
	this := "file"
	if values, _, err := resolveMeta(tc, t, string(content), "this"); err != nil {
		return err
	} else if m := values.lookup("this"); m.Second {
		this = m.First
	}

	var formatter Formatter
	if f := tc.compiler.options.Formatter.Get(lang); f != "" {
		formatter, err = newFormatter(f)
		if err != nil {
			return fmt.Errorf("failed to create %s formatter: %w", lang, err)
		}
	}

	switch strings.ToLower(this) {
	case "packages":
		return generateForPackages(tc, t, content, formatter)

	case "package":
		return generateForPackage(tc, t, content, formatter)

	case "file":
		return generateForFile(tc, t, content, formatter)

	case "const":
		return generateForConst(tc, t, content, formatter)

	case "enum":
		return generateForEnum(tc, t, content, formatter)

	case "struct":
		return generateForStruct(tc, t, content, formatter)

	case "interface":
		return generateForInterface(tc, t, content, formatter)

	default:
		return fmt.Errorf(`unknown value for 'this': %q, expected "package", "file", "const", "enum", "struct" or "interface"`, this)
	}
}

func generateForPackages(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	return gen(tc, t, Packages(tc.compiler.packages), content, formatter)
}

func generateForPackage(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, pkg := range tc.compiler.packages {
		if err := gen(tc, t, pkg, content, formatter); err != nil {
			return err
		}
	}
	return nil
}

func generateForFile(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, f := range tc.compiler.files {
		if err := gen(tc, t, f, content, formatter); err != nil {
			return err
		}
	}
	return nil
}

func generateForConst(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.consts {
			if err := gen(tc, t, d, content, formatter); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForEnum(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.enums {
			if err := gen(tc, t, d, content, formatter); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForStruct(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.structs {
			if err := gen(tc, t, d, content, formatter); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForInterface(tc *templateContext, t *template.Template, content []byte, formatter Formatter) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.interfaces {
			if err := gen(tc, t, d, content, formatter); err != nil {
				return err
			}
		}
	}
	return nil
}

// gen generates a file using the given template, meta data, and object which may be a
// file, const, enum or struct.
func gen[T Decl](tc *templateContext, t *template.Template, decl T, content []byte, formatter Formatter) error {
	// Skip if the declaration is an alias
	if decl.Annotations().get("next").get(tc.lang+"_alias") != nil {
		return nil
	}

	// Skip if the declaration is not available in the target language
	decl, ok := available(tc.compiler, decl, tc.lang)
	if !ok {
		return nil
	}

	// Reset the template context with the template and the declaration
	if err := tc.reset(t, reflect.ValueOf(decl)); err != nil {
		return err
	}
	tc.push(filepath.Dir(t.ParseName))
	defer tc.pop()

	// Resolve meta data
	meta, positions, err := resolveMeta(tc, t, string(content))
	if err != nil {
		return err
	}
	for k, v := range meta {
		tc.compiler.Trace("%s: meta[%q] = %q", t.Name(), k, v)
	}

	// Skip if the meta data contains 'skip' and its value is true
	if m := meta.lookup("skip").First; m != "" {
		skip, err := strconv.ParseBool(m)
		if err != nil {
			pos := positions["skip"]
			return fmt.Errorf("%s: failed to parse 'skip' meta data: %v", document.FormatPosition(t.ParseName, pos), err)
		}
		if skip {
			return nil
		}
	}

	var parsedPerm os.FileMode
	if m := meta.lookup("perm"); m.Second {
		x, err := strconv.ParseInt(m.First, 8, 32)
		if err != nil {
			pos := positions["perm"]
			return fmt.Errorf("%s: failed to parse 'perm' meta data: %v", document.FormatPosition(t.ParseName, pos), err)
		}
		parsedPerm = os.FileMode(x)
		if parsedPerm|os.ModePerm != os.ModePerm {
			pos := positions["perm"]
			return fmt.Errorf("%s: invalid file permission: %o", document.FormatPosition(t.ParseName, pos), parsedPerm)
		}
		if parsedPerm == 0 {
			pos := positions["perm"]
			return fmt.Errorf("%s: zero file permission", document.FormatPosition(t.ParseName, pos))
		}
	}

	// Execute the template with the template context
	buf := tc.pc().buf
	if err := t.Execute(buf, tc); err != nil {
		return err
	}

	// Do not write the generated content to the output file if the test flag is set
	if tc.compiler.options.test {
		return nil
	}

	// Write the generated content to the output file
	path := op.Or(meta.lookup("path").First, decl.Name()+tc.ext)
	if !filepath.IsAbs(path) {
		path = filepath.Join(tc.dir, path)
	}
	perm := cmp.Or(parsedPerm, op.If(tc.headCalled && tc.compiler.options.Perm > 0, os.FileMode(tc.compiler.options.Perm), 0644))
	if err := tc.compiler.platform.WriteFile(path, buf.Bytes(), perm, formatter); err != nil {
		return fmt.Errorf("%s: failed to write file %s: %v", t.ParseName, path, err)
	}

	return nil
}
