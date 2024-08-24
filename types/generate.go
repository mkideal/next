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
	"github.com/gopherd/next/internal/fsutil"
)

const (
	nextDir       = "next.d"    // next directory for configuration files
	hiddenNextDir = ".next"     // hidden next directory for configuration files in user's home directory
	typesDir      = "types"     // types directory for types files in next directory
	templatesDir  = "templates" // templates directory for template files in next directory
	templateExt   = ".nxp"      // template file extension
	typesExt      = ".types"    // types file extension
)

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
	// Check whether the template directory exists for each language
	for lang := range c.flags.outputs {
		tmplDir := c.flags.templates[lang]
		if _, err := os.Stat(tmplDir); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("template directory for %q not found: %q", lang, tmplDir)
			}
			return fmt.Errorf("failed to check template directory for %q: %v", lang, err)
		}
	}

	// Load all types from all types files
	searchDirs := c.searchDirs()
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
			c.Printf("types[%q] = %q", k, m[k])
		}
	}

	// Generate files for each language
	for lang, dir := range c.flags.outputs {
		tempDir := c.flags.templates[lang]
		if err := c.generateForTemplateDir(lang, dir, tempDir); err != nil {
			return err
		}
	}
	return nil
}

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
func (c *Context) searchDirs() []string {
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

func (c *Context) loadTypes(m flags.Map, dirs []string, lang string) error {
	for _, dir := range dirs {
		path := filepath.Join(dir, typesDir, lang+typesExt)
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

func (c *Context) generateForTemplateDir(lang, dir, tmplDir string) error {
	tmplFiles, err := fsutil.AppendFiles(nil, tmplDir, templateExt, true)
	if err != nil {
		return fmt.Errorf("failed to list template files in %q: %v", tmplDir, err)
	}

	// Generate files for each template
	for _, tmplFile := range tmplFiles {
		if err := c.generateForTemplateFile(lang, dir, tmplFile); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) generateForTemplateFile(lang, dir, tmplFile string) error {
	ext, objType, err := parseTemplateFilename(tmplFile)
	if err != nil {
		return err
	}
	tmpl, meta, err := loadTemplate(tmplFile)
	if err != nil {
		return fmt.Errorf("failed to load template %q: %v", tmplFile, err)
	}
	objType = strings.ToLower(objType)
	switch objType {
	case "package":
		for _, pkg := range c.packages {
			if err := gen(c, lang, dir, tmpl, meta, ext, pkg); err != nil {
				return err
			}
		}
	case "file":
		for _, file := range c.files {
			if err := gen(c, lang, dir, tmpl, meta, ext, file); err != nil {
				return err
			}
		}
	case "const", "enum", "struct":
		for _, file := range c.files {
			for _, decl := range file.decls {
				if strings.ToLower(decl.Tok.String()) != objType {
					continue
				}
				for _, spec := range decl.Specs {
					var err error
					switch spec := spec.(type) {
					case *ValueSpec:
						err = gen(c, lang, dir, tmpl, meta, ext, spec)
					case *EnumType:
						err = gen(c, lang, dir, tmpl, meta, ext, spec)
					case *StructType:
						err = gen(c, lang, dir, tmpl, meta, ext, spec)
					default:
						err = fmt.Errorf("invalid object type %q in template file %q", objType, tmplFile)
					}
					if err != nil {
						return err
					}
				}
			}
		}
	default:
		panic(fmt.Sprintf("invalid object type %q in template file %q", objType, tmplFile))
	}
	return nil
}

func (c *Context) parseMeta(meta TemplateMeta, data any) (TemplateMeta, error) {
	if meta == nil {
		return nil, nil
	}
	m := make(TemplateMeta)
	for k, v := range meta {
		v, err := executeTemplate(k, v, data)
		if err != nil {
			return nil, fmt.Errorf("failed to execute template %q: %v", v, err)
		}
		m[k] = v
	}
	return m, nil
}

// gen generates a file using the given template, meta data, and object which may be a
// package, file, const, enum or struct.
func gen[T Object](c *Context, lang, dir string, tmpl *template.Template, meta TemplateMeta, ext string, obj T) error {
	meta, err := c.parseMeta(meta, obj)
	if err != nil {
		return err
	}
	if meta == nil {
		meta = make(TemplateMeta)
	}
	data := &TemplateData[T]{
		T:       obj,
		Lang:    lang,
		Meta:    meta,
		context: c,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// write the generated content to the output file
	file := meta.Get("file")
	if file == "" {
		file = obj.Name() + ext
	}
	path := filepath.Join(dir, file)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %v", path, err)
	}

	return nil
}
