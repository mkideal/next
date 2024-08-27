package types

import (
	"strings"

	"github.com/gopherd/next/ast"
)

type CommentGroup struct {
	list []string
}

func newCommentGroup(cg *ast.CommentGroup) *CommentGroup {
	if cg == nil {
		return nil
	}
	list := make([]string, len(cg.List))
	for i, c := range cg.List {
		list[i] = c.Text
	}
	return &CommentGroup{
		list: list,
	}
}

func (cg *CommentGroup) Text() string {
	if cg == nil || len(cg.list) == 0 {
		return ""
	}
	return cg.Format()
}

func (cg *CommentGroup) String() string {
	if cg == nil || len(cg.list) == 0 {
		return ""
	}
	return strings.Join(ast.TrimComments(cg.list), "\n")
}

func (cg *CommentGroup) Format(beginAndEnd ...string) string {
	if cg == nil {
		return ""
	}
	return cg.FormatIndent("", "", beginAndEnd...)
}

// FormatIndent formats the comment group with the given prefix, ident, and begin and end strings.
//
// Example:
//
//	cg.FormatIndent("", "")
//
//	// comment1
//	// comment2
//
//	cg.FormatIndent("", " *", "/*", "*/")
//
//	/* comment1
//	 * comment2
//	 */
//
//	cg.FormatIndent("  ", " *", "/*", "*/")
//
//	  /* comment1
//	   * comment2
//	   */
//
//	cg.FormatIndent("", "", "<!--\n", "-->")
//
//	<!--
//	 comment1
//	 comment2
//	 -->
func (cg *CommentGroup) FormatIndent(prefix, indent string, beginAndEnd ...string) string {
	if cg == nil || len(cg.list) == 0 {
		return ""
	}
	lines := ast.TrimComments(cg.list)
	if len(lines) == 0 {
		return ""
	}
	isLastEmpty := lines[len(lines)-1] == ""
	begin, end := "//", ""
	if len(beginAndEnd) > 0 && beginAndEnd[0] != "" {
		begin = beginAndEnd[0]
		if len(beginAndEnd) > 1 && beginAndEnd[1] != "" {
			end = beginAndEnd[1]
		}
	}
	if end == "" {
		for i, line := range lines {
			lines[i] = prefix + indent + begin + " " + line
		}
		if !isLastEmpty {
			lines = append(lines, prefix)
		}
		return strings.Join(lines, "\n")
	}
	lines = append([]string{prefix + begin + " " + lines[0]}, lines[1:]...)
	for i := 1; i < len(lines); i++ {
		lines[i] = prefix + indent + " " + lines[i]
	}
	lines = append(lines, prefix+" "+end)
	return strings.Join(lines, "\n")
}
