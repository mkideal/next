package types

import (
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

	"github.com/next/next/src/fsutil"
)

const (
	nextDir       = "next.d" // next directory for configuration files
	hiddenNextDir = ".next"  // hidden next directory for configuration files in user's home directory
	templateExt   = ".npl"   // next template file extension
	langMapExt    = ".map"   // next lang map file extension
)

// searchDirs returns ordered a list of directories to search for map files.
// The order is from the most specific to the least specific.
// The most specific directory is the user's home directory.
// The least specific directories are the system directories.
// The system directories are:
//   - /etc/next.d
//   - /usr/local/etc/next.d
//   - /Library/Application Support/next.d
//   - ~/Library/Application Support/next.d
//   - %APPDATA%/next.d
func createSearchDirs(platform Platform) []string {
	var dirs []string
	dirs = append(dirs, ".")
	homedir, err := platform.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(homedir, hiddenNextDir))
	}
	if runtime.GOOS == "js" {
		return dirs
	}
	if runtime.GOOS == "windows" {
		appData := platform.Getenv("APPDATA")
		if appData != "" {
			dirs = append(dirs, filepath.Join(appData, nextDir))
		}
	} else {
		if runtime.GOOS == "darwin" {
			dirs = append(dirs, filepath.Join("/Library/Application Support", nextDir))
			if homedir != "" {
				dirs = append(dirs, filepath.Join(homedir, "Library/Application Support", nextDir))
			}
		}
		dirs = append(dirs,
			filepath.Join("/usr/local/etc", nextDir),
			filepath.Join("/etc", nextDir),
		)
	}
	slices.Reverse(dirs)
	return dirs
}

// Genertate generates files for each language specified in the flags.outputs.
func Generate(c *Compiler) error {
	if len(c.flags.outputs) == 0 {
		return nil
	}
	c.Print("flags.envs: ", c.flags.envs)
	c.Print("flags.outputs: ", c.flags.outputs)
	c.Print("flags.templates: ", c.flags.templates)
	c.Print("flags.mappings: ", c.flags.mappings)

	if c.flags.outputs.Get("next") != "" {
		return fmt.Errorf("output language 'next' is not supported")
	}

	// Check whether the template directory or file exists for each language
	for lang := range c.flags.outputs {
		for _, tmplPath := range c.flags.templates[lang] {
			if c.platform.IsNotExist(tmplPath) {
				return fmt.Errorf("template path %q not found: %q", lang, tmplPath)
			}
		}
	}

	// Load all mappings from all map files
	m := make(flags.Map)
	for lang := range c.flags.outputs {
		if err := loadMap(c, m, lang); err != nil {
			return err
		}
	}
	for k, v := range c.flags.mappings {
		m[k] = v
	}
	c.flags.mappings = m
	if c.IsDebugEnabled() {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			c.Tracef("map[%q] = %q", k, m[k])
		}
	}

	// Generate files for each language
	langs := make([]string, 0, len(c.flags.outputs))
	for lang := range c.flags.outputs {
		langs = append(langs, lang)
	}
	slices.Sort(langs)
	for _, lang := range langs {
		dir := c.flags.outputs[lang]
		ext := op.Or(c.flags.mappings[lang+".ext"], "."+lang)
		tempPaths := c.flags.templates[lang]
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
	for _, dir := range c.searchDirs {
		path := filepath.Join(dir, lang+langMapExt)
		if err := loadMapFromFile(c, m, lang, path); err != nil {
			return fmt.Errorf("failed to load %q: %v", path, err)
		}
	}
	return nil
}

func loadMapFromFile(c *Compiler, m flags.Map, lang, path string) error {
	content, err := c.platform.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return parseLangMap(m, lang, content)
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
	if values, err := ResolveMeta(tc, t, "this"); err != nil {
		return err
	} else if m := values.lookup("this"); m.Second {
		this = m.First
	}

	switch strings.ToLower(this) {
	case "package":
		return generateForPackage(tc, t)

	case "file":
		return generateForFile(tc, t)

	case "const":
		return generateForConst(tc, t)

	case "enum":
		return generateForEnum(tc, t)

	case "struct":
		return generateForStruct(tc, t)

	case "interface":
		return generateForInterface(tc, t)

	default:
		return fmt.Errorf(`unknown value for 'this': %q, expected "package", "file", "const", "enum", "struct" or "interface"`, this)
	}
}

func generateForPackage(tc *templateContext, t *template.Template) error {
	for _, pkg := range tc.compiler.packages {
		if err := gen(tc, t, pkg); err != nil {
			return err
		}
	}
	return nil
}

func generateForFile(tc *templateContext, t *template.Template) error {
	for _, f := range tc.compiler.files {
		if err := gen(tc, t, f); err != nil {
			return err
		}
	}
	return nil
}

func generateForConst(tc *templateContext, t *template.Template) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.consts {
			if err := gen(tc, t, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForEnum(tc *templateContext, t *template.Template) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.enums {
			if err := gen(tc, t, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForStruct(tc *templateContext, t *template.Template) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.structs {
			if err := gen(tc, t, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateForInterface(tc *templateContext, t *template.Template) error {
	for _, file := range tc.compiler.files {
		if file.decls == nil {
			continue
		}
		for _, d := range file.decls.interfaces {
			if err := gen(tc, t, d); err != nil {
				return err
			}
		}
	}
	return nil
}

// gen generates a file using the given template, meta data, and object which may be a
// file, const, enum or struct.
func gen[T Decl](tc *templateContext, t *template.Template, decl T) error {
	// skip if the declaration is an alias
	if decl.Annotations().get("next").get(tc.lang+"_alias") != nil {
		return nil
	}

	// skip if the declaration is not available in the target language
	decl, ok := available(tc.compiler, decl, tc.lang)
	if !ok {
		return nil
	}

	// reset the template context with the template and the declaration
	if err := tc.reset(t, reflect.ValueOf(decl)); err != nil {
		return err
	}

	// resolve meta data
	meta, err := ResolveMeta(tc, t)
	if err != nil {
		return err
	}
	for k, v := range meta {
		tc.compiler.Printf("%s: meta[%q] = %q", t.Name(), k, v)
	}

	// skip if the meta data contains 'skip' and its value is true
	if m := meta.lookup("skip").First; m != "" {
		skip, err := strconv.ParseBool(m)
		if err != nil {
			return fmt.Errorf("failed to parse 'skip' meta data: %v", err)
		}
		if skip {
			return nil
		}
	}

	// execute the template with the template context
	tc.pushPwd(filepath.Dir(t.ParseName))
	defer tc.popPwd()
	if err := t.Execute(&tc.buf, tc); err != nil {
		return err
	}

	// write the generated content to the output file
	path := op.Or(meta.lookup("path").First, decl.Name()+tc.ext)
	if !filepath.IsAbs(path) {
		path = filepath.Join(tc.dir, path)
	}
	if err := tc.compiler.platform.WriteFile(path, []byte(tc.buf.String())); err != nil {
		return fmt.Errorf("failed to write file %q: %v", path, err)
	}

	return nil
}
