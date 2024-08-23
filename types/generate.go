package types

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/gopherd/core/flags"
	"github.com/gopherd/next/internal/fsutil"
)

const (
	nextDir     = "next.d"
	langDir     = "lang"
	templateDir = "template"
	templateExt = ".nxp"
)

func hidden(dir string) string {
	return "." + dir
}

// Genertate generates files for all specified languages.
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
		tplDir := c.flags.templates[lang]
		if _, err := os.Stat(tplDir); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("template directory for %q not found: %q", lang, tplDir)
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

	// Generate files for each language
	for lang, dir := range c.flags.outputs {
		tempDir := c.flags.templates[lang]
		if err := c.generateForLang(lang, dir, tempDir); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) searchDirs() []string {

	var dirs []string
	dirs = append(dirs, ".")
	homedir, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(homedir, hidden(nextDir)))
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
		path := filepath.Join(dir, lang+".types")
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

func (c *Context) generateForLang(lang, dir, tplDir string) error {
	tplFiles, err := fsutil.AppendFiles(nil, tplDir, templateExt, true)
	if err != nil {
		return fmt.Errorf("failed to list template files in %q: %v", tplDir, err)
	}

	// Generate files for each template
	for _, tplFile := range tplFiles {
		if err := c.generateForTemplateFile(lang, dir, tplFile); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) generateForTemplateFile(lang, dir, tplFile string) error {
	nodeType, err := nodeTypeOfTemplateFile(tplFile)
	if err != nil {
		return err
	}
	tpl, err := loadTemplate(tplFile)
	if err != nil {
		return fmt.Errorf("failed to load template %q: %v", tplFile, err)
	}
	nodeType = strings.ToLower(nodeType)
	switch nodeType {
	case "package":
		for _, pkg := range c.packages {
			if err := c.generateForPackage(lang, dir, tpl, pkg); err != nil {
				return err
			}
		}
	case "file":
		for _, file := range c.files {
			if err := c.generateForFile(lang, dir, tpl, file); err != nil {
				return err
			}
		}
	case "const", "enum", "struct":
		for _, file := range c.files {
			for _, decl := range file.decls {
				if strings.ToLower(decl.Tok.String()) != nodeType {
					continue
				}
				for _, spec := range decl.Specs {
					if err := c.generateForSpec(lang, dir, tpl, spec); err != nil {
						return err
					}
				}
			}
		}
	default:
		panic(fmt.Sprintf("invalid node type %q in template file %q", nodeType, tplFile))
	}
	return nil
}

func (c *Context) generateForPackage(lang, dir string, tpl Template, pkg *Package) error {
	meta, err := c.applyMeta(pkg, tpl.GetMeta())
	if err != nil {
		return err
	}
	return c.gen(lang, dir, tpl, meta, pkg)
}

func (c *Context) generateForFile(lang, dir string, tpl Template, file *File) error {
	meta, err := c.applyMeta(file, tpl.GetMeta())
	if err != nil {
		return err
	}
	return c.gen(lang, dir, tpl, meta, file)
}

func (c *Context) generateForSpec(lang, dir string, tpl Template, spec Spec) error {
	meta, err := c.applyMeta(spec, tpl.GetMeta())
	if err != nil {
		return err
	}
	return c.gen(lang, dir, tpl, meta, spec)
}

func (c *Context) applyMeta(data any, values map[string]string) (*TemplateMeta, error) {
	if values == nil {
		return nil, nil
	}
	var meta TemplateMeta
	for k, v := range values {
		v, err := executeTemplate(k, v, data)
		if err != nil {
			return nil, fmt.Errorf("failed to execute template %q: %v", v, err)
		}
		if err := meta.setValue(k, v); err != nil {
			return nil, fmt.Errorf("failed to set meta key %q: %v", k, err)
		}
	}
	return &meta, nil
}

func (c *Context) gen(lang, dir string, tpl Template, meta *TemplateMeta, data any) error {
	// TODO
	return nil
}
