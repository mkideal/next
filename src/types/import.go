package types

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// Imports holds a list of imports.
// @api(object/Imports)
type Imports struct {
	// List is the list of imports.
	// @api(object/Imports/List)
	List []*Import
}

func (i *Imports) resolve(ctx *Context, file *File) {
	for _, spec := range i.List {
		spec.target = ctx.lookupFile(file.Path, spec.Path)
		if spec.target == nil {
			ctx.addErrorf(spec.pos, "import file not found: %s", spec.Path)
		}
	}
}

// TrimmedList returns a list of unique imports sorted by package name.
// @api(object/Imports/TrimmedList)
func (i *Imports) TrimmedList() []*Import {
	var seen = make(map[string]bool)
	var pkgs []*Import
	for _, spec := range i.List {
		if seen[spec.target.pkg.name] {
			continue
		}
		seen[spec.target.pkg.name] = true
		pkgs = append(pkgs, spec)
	}
	slices.SortFunc(pkgs, func(a, b *Import) int {
		return cmp.Compare(a.target.pkg.name, b.target.pkg.name)
	})
	return pkgs
}

// Import represents a file import.
// @api(object/Import)
type Import struct {
	pos    token.Pos // position of the import declaration
	target *File     // imported file
	file   *File     // file containing the import

	// Doc is the [documentation](#Object.Doc) for the import declaration.
	// @api(object/Import/Doc)
	Doc *Doc

	// Comment is the [line comment](#Object.Comment) of the import declaration.
	// @api(object/import/Comment)
	Comment *Comment

	// Path is the import path.
	// @api(object/import/Path)
	Path string
}

func newImport(ctx *Context, file *File, src *ast.ImportDecl) *Import {
	path, err := strconv.Unquote(src.Path.Value)
	if err != nil {
		ctx.addErrorf(src.Path.Pos(), "invalid import path %v: %v", src.Path.Value, err)
		path = "!BAD-IMPORT-PATH!"
	}
	i := &Import{
		pos:     src.Pos(),
		file:    file,
		Doc:     newDoc(src.Doc),
		Comment: newComment(src.Comment),
		Path:    path,
	}
	return i
}

// Target returns the imported file.
// @api(object/Import/Target
func (i *Import) Target() *File { return i.target }

// File returns the file containing the import.
// @api(object/Import/File)
func (i *Import) File() *File { return i.target }

func (i *Import) resolve(ctx *Context, file *File, _ Scope) {}
