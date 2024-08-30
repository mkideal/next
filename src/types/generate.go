package types

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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

// searchDirs returns ordered a list of directories to search for types files.
// The order is from the most specific to the least specific.
// The most specific directory is the user's home directory.
// The least specific directories are the system directories.
// The system directories are:
//   - /etc/next.d
//   - /usr/local/etc/next.d
//   - /Library/Application Support/next.d
//   - ~/Library/Application Support/next.d
//   - %APPDATA%/next.d
func createSearchDirs() []string {
	var dirs []string
	dirs = append(dirs, ".")
	homedir, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(homedir, hiddenNextDir))
	}
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
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
func (c *Context) Generate() error {
	if len(c.flags.outputs) == 0 {
		return nil
	}
	c.Print("flags.importDirs: ", c.flags.importDirs)
	c.Print("flags.macros: ", c.flags.macros)
	c.Print("flags.outputs: ", c.flags.outputs)
	c.Print("flags.templates: ", c.flags.templates)
	c.Print("flags.types: ", c.flags.types)

	if c.flags.outputs.Get("next") != "" {
		return fmt.Errorf("output language 'next' is not supported")
	}

	// Check whether the template directory or file exists for each language
	for lang := range c.flags.outputs {
		for _, tmplPath := range c.flags.templates[lang] {
			if _, err := os.Stat(tmplPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("template path %q not found: %q", lang, tmplPath)
				}
				return fmt.Errorf("failed to check template path %q: %v", lang, err)
			}
		}
	}

	// Load all types from all types files
	searchDirs := c.searchDirs
	m := make(flags.Map)
	for lang := range c.flags.outputs {
		if err := c.loadTypes(m, searchDirs, lang); err != nil {
			return err
		}
	}
	for k, v := range c.flags.types {
		m[k] = v
	}
	c.flags.types = m
	if c.IsDebugEnabled() {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			c.Tracef("types[%q] = %q", k, m[k])
		}
	}

	// Generate files for each language
	for _, lang := range slices.Sorted(maps.Keys(c.flags.outputs)) {
		dir := c.flags.outputs[lang]
		ext := op.Or(c.flags.types[lang+".ext"], "."+lang)
		tempPaths := c.flags.templates[lang]
		if len(tempPaths) == 0 {
			return fmt.Errorf("no template directory specified for %q", lang)
		}
		for _, tempPath := range tempPaths {
			if err := c.generateForTemplatePath(lang, ext, dir, tempPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Context) loadTypes(m flags.Map, dirs []string, lang string) error {
	for _, dir := range dirs {
		path := filepath.Join(dir, lang+langMapExt)
		if err := c.loadTypesFromFile(m, lang, path); err != nil {
			return fmt.Errorf("failed to load types from %q: %v", path, err)
		}
	}
	return nil
}

func (c *Context) loadTypesFromFile(m flags.Map, lang, path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	return parseLangTypes(m, lang, f)
}

func (c *Context) generateForTemplatePath(lang, ext, dir, tmplPath string) error {
	tmplFiles, err := fsutil.AppendFiles(nil, tmplPath, templateExt, true)
	if err != nil {
		return fmt.Errorf("failed to list template files in %q: %v", tmplPath, err)
	}

	// Generate files for each template
	for _, tmplFile := range tmplFiles {
		if err := c.generateForTemplateFile(lang, ext, dir, tmplFile); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) generateForTemplateFile(lang, ext, dir, tmplFile string) error {
	tmplContent, err := os.ReadFile(tmplFile)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", tmplFile, err)
	}
	meta, content, err := parseMeta(string(tmplContent))
	if err != nil {
		return err
	}

	objType := "file"
	this := meta.Get("this")
	if this != nil {
		objType = this.content
	}
	delete(meta, "this")

	info := templateContextInfo{
		context: c,
		lang:    lang,
		dir:     dir,
		ext:     ext,
	}

	switch strings.ToLower(objType) {
	case "file":
		return generateForFile(newTemplateContext[*File](info), tmplFile, string(content), meta)

	case "const":
		return generateForSpec(newTemplateContext[*ValueSpec](info), tmplFile, string(content), meta)

	case "enum":
		return generateForSpec(newTemplateContext[*EnumSpec](info), tmplFile, string(content), meta)

	case "struct":
		return generateForSpec(newTemplateContext[*StructSpec](info), tmplFile, string(content), meta)

	default:
		return fmt.Errorf(`unknown value for 'this': %q, expected "file", "const", "enum" or "struct"`, objType)
	}
}

func generateForFile(tc *templateContext[*File], file, content string, meta templateMeta[string]) error {
	t, mt, err := createTemplates(file, content, meta, tc.withThis())
	if err != nil {
		return err
	}
	for _, f := range tc.context.files {
		tc.reset(f)
		if err := gen(tc, t, mt); err != nil {
			return err
		}
	}
	return nil
}

func generateForSpec[T Node](tc *templateContext[T], file, content string, meta templateMeta[string]) error {
	t, mt, err := createTemplates(file, content, meta, tc.withThis())
	if err != nil {
		return err
	}

	for _, file := range tc.context.files {
		for _, decl := range file.decls {
			for _, spec := range decl.Specs {
				if spec, ok := spec.(T); ok {
					tc.reset(spec)
					if err := gen(tc, t, mt); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// gen generates a file using the given template, meta data, and object which may be a
// file, const, enum or struct.
func gen[T Node](tc *templateContext[T], t *template.Template, mt templateMeta[*template.Template]) error {
	tc.entrypoint = t
	if err := tc.init(); err != nil {
		return err
	}

	meta, err := resolveMeta(mt, tc)
	if err != nil {
		return err
	}
	if meta.Get("skip").value() == "true" {
		return nil
	}
	tc.buf.Reset()
	if err := t.Execute(&tc.buf, tc); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// write the generated content to the output file
	path := op.Or(meta.Get("path").value(), tc.obj.Name()+tc.ext)
	if !filepath.IsAbs(path) {
		path = filepath.Join(tc.dir, path)
	}
	if op.Or(meta.Get("overwrite").value(), "true") != "true" {
		if _, err := os.Stat(path); err == nil {
			tc.context.Printf("file %q already exists, and will not be overwritten", path)
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %v", tc.dir, err)
	}
	if err := os.WriteFile(path, []byte(tc.buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %v", path, err)
	}

	return nil
}
