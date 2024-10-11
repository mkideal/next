package ast

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/mkideal/next/src/token"
)

// lineAt returns the line number for the given position in the file set.
func lineAt(fset *token.FileSet, pos token.Pos) int {
	return fset.PositionFor(pos, false).Line
}

// importPath extracts and returns the import path from an ImportDecl.
// It returns an empty string if the path cannot be unquoted.
func importPath(s Decl) string {
	t, err := strconv.Unquote(s.(*ImportDecl).Path.Value)
	if err == nil {
		return t
	}
	return ""
}

// importComment returns the comment text associated with an ImportDecl.
// It returns an empty string if there is no comment.
func importComment(s Decl) string {
	c := s.(*ImportDecl).Comment
	if c == nil {
		return ""
	}
	return c.Text()
}

// collapse determines whether the previous import declaration can be removed,
// leaving only the next one. This is possible when they have the same import path
// and the previous declaration has no comment.
func collapse(prev, next Decl) bool {
	return importPath(prev) == importPath(next) && prev.(*ImportDecl).Comment == nil
}

// posSpan represents a span of positions in the source code.
type posSpan struct {
	Start token.Pos
	End   token.Pos
}

// cgPos represents a comment group position relative to an import specification.
type cgPos struct {
	left bool // true if comment is to the left of the spec, false otherwise
	cg   *CommentGroup
}

// sortSpecs sorts and deduplicates import declarations in a file.
// It also adjusts the positions of comments associated with imports.
func sortSpecs(fset *token.FileSet, f *File, decls []Decl) []Decl {
	if len(decls) <= 1 {
		return decls
	}

	// Record positions for declarations
	pos := make([]posSpan, len(decls))
	for i, s := range decls {
		pos[i] = posSpan{s.Pos(), s.End()}
	}

	// Identify comments in the range of import declarations
	begSpecs := pos[0].Start
	endSpecs := pos[len(pos)-1].End
	beg := fset.File(begSpecs).LineStart(lineAt(fset, begSpecs))
	endLine := lineAt(fset, endSpecs)
	endFile := fset.File(endSpecs)
	var end token.Pos
	if endLine == endFile.LineCount() {
		end = endSpecs
	} else {
		end = endFile.LineStart(endLine + 1) // beginning of next line
	}

	// Find relevant comments
	first, last := len(f.Comments), -1
	for i, g := range f.Comments {
		if g.End() >= end {
			break
		}
		if beg <= g.Pos() {
			if i < first {
				first = i
			}
			if i > last {
				last = i
			}
		}
	}

	var comments []*CommentGroup
	if last >= 0 {
		comments = f.Comments[first : last+1]
	}

	// Assign comments to import specs
	importComments := make(map[*ImportDecl][]cgPos)
	specIndex := 0
	for _, g := range comments {
		for specIndex+1 < len(decls) && pos[specIndex+1].Start <= g.Pos() {
			specIndex++
		}
		var left bool
		if specIndex == 0 && pos[specIndex].Start > g.Pos() {
			left = true
		} else if specIndex+1 < len(decls) &&
			lineAt(fset, pos[specIndex].Start)+1 == lineAt(fset, g.Pos()) {
			specIndex++
			left = true
		}
		s := decls[specIndex].(*ImportDecl)
		importComments[s] = append(importComments[s], cgPos{left: left, cg: g})
	}

	// Sort and deduplicate import declarations
	slices.SortFunc(decls, func(a, b Decl) int {
		ipath, jpath := importPath(a), importPath(b)
		if r := cmp.Compare(ipath, jpath); r != 0 {
			return r
		}
		return cmp.Compare(importComment(a), importComment(b))
	})

	deduped := decls[:0]
	for i, s := range decls {
		if i == len(decls)-1 || !collapse(s, decls[i+1]) {
			deduped = append(deduped, s)
		} else {
			p := s.Pos()
			fset.File(p).MergeLine(lineAt(fset, p))
		}
	}
	decls = deduped

	// Adjust comment positions
	for i, s := range decls {
		s := s.(*ImportDecl)
		s.Path.ValuePos = pos[i].Start
		s.EndPos = pos[i].End
		for _, g := range importComments[s] {
			for _, c := range g.cg.List {
				if g.left {
					c.Slash = pos[i].Start - 1
				} else {
					c.Slash = pos[i].End
				}
			}
		}
	}

	// Sort comments by position
	slices.SortFunc(comments, func(a, b *CommentGroup) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})

	return decls
}
