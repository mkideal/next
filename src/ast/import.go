package ast

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/next/next/src/token"
)

func lineAt(fset *token.FileSet, pos token.Pos) int {
	return fset.PositionFor(pos, false).Line
}

func importPath(s Decl) string {
	t, err := strconv.Unquote(s.(*ImportDecl).Path.Value)
	if err == nil {
		return t
	}
	return ""
}

func importComment(s Decl) string {
	c := s.(*ImportDecl).Comment
	if c == nil {
		return ""
	}
	return c.Text()
}

// collapse indicates whether prev may be removed, leaving only next.
func collapse(prev, next Decl) bool {
	if importPath(next) != importPath(prev) {
		return false
	}
	return prev.(*ImportDecl).Comment == nil
}

type posSpan struct {
	Start token.Pos
	End   token.Pos
}

type cgPos struct {
	left bool // true if comment is to the left of the spec, false otherwise.
	cg   *CommentGroup
}

func sortSpecs(fset *token.FileSet, f *File, decls []Decl) []Decl {
	// Can't short-circuit here even if decls are already sorted,
	// since they might yet need deduplication.
	// A lone import, however, may be safely ignored.
	if len(decls) <= 1 {
		return decls
	}

	// Record positions for decls.
	pos := make([]posSpan, len(decls))
	for i, s := range decls {
		pos[i] = posSpan{s.Pos(), s.End()}
	}

	// Identify comments in this range.
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
	first := len(f.Comments)
	last := -1
	for i, g := range f.Comments {
		if g.End() >= end {
			break
		}
		// g.End() < end
		if beg <= g.Pos() {
			// comment is within the range [beg, end[ of import declarations
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

	// Assign each comment to the import spec on the same line.
	importComments := map[*ImportDecl][]cgPos{}
	specIndex := 0
	for _, g := range comments {
		for specIndex+1 < len(decls) && pos[specIndex+1].Start <= g.Pos() {
			specIndex++
		}
		var left bool
		// A block comment can appear before the first import spec.
		if specIndex == 0 && pos[specIndex].Start > g.Pos() {
			left = true
		} else if specIndex+1 < len(decls) && // Or it can appear on the left of an import spec.
			lineAt(fset, pos[specIndex].Start)+1 == lineAt(fset, g.Pos()) {
			specIndex++
			left = true
		}
		s := decls[specIndex].(*ImportDecl)
		importComments[s] = append(importComments[s], cgPos{left: left, cg: g})
	}

	// Sort the import decls by import path.
	// Remove duplicates, when possible without data loss.
	// Reassign the import paths to have the same position sequence.
	// Reassign each comment to the spec on the same line.
	// Sort the comments by new position.
	slices.SortFunc(decls, func(a, b Decl) int {
		ipath := importPath(a)
		jpath := importPath(b)
		r := cmp.Compare(ipath, jpath)
		if r != 0 {
			return r
		}
		return cmp.Compare(importComment(a), importComment(b))
	})

	// Dedup. Thanks to our sorting, we can just consider
	// adjacent pairs of imports.
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

	// Fix up comment positions
	for i, s := range decls {
		s := s.(*ImportDecl)
		s.Path.ValuePos = pos[i].Start
		s.EndPos = pos[i].End
		for _, g := range importComments[s] {
			for _, c := range g.cg.List {
				if g.left {
					c.Slash = pos[i].Start - 1
				} else {
					// An import spec can have both block comment and a line comment
					// to its right. In that case, both of them will have the same pos.
					// But while formatting the AST, the line comment gets moved to
					// after the block comment.
					c.Slash = pos[i].End
				}
			}
		}
	}

	slices.SortFunc(comments, func(a, b *CommentGroup) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})

	return decls
}
