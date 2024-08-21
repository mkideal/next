package types

import (
	"strings"

	"github.com/gopherd/next/ast"
)

type Comment struct {
	// Text is the comment text.
	Text string
}

func resolveComment(c *ast.Comment) *Comment {
	return &Comment{
		Text: c.Text,
	}
}

func (c Comment) String() string {
	return strings.TrimSpace(ast.TrimComments([]string{c.Text})[0])
}

func (c Comment) Format(beginAndEnd ...string) string {
	begin, end := "//", ""
	if len(beginAndEnd) > 0 && beginAndEnd[0] != "" {
		begin = beginAndEnd[0]
		if len(beginAndEnd) > 1 && beginAndEnd[1] != "" {
			end = beginAndEnd[1]
		}
	}
	if end == "" {
		if strings.HasPrefix(c.Text, begin) {
			return c.Text
		}
		return begin + " " + c.String()
	}
	return begin + " " + c.String() + " " + end
}

type CommentGroup struct {
	List []string
}

func newCommentGroup(cg *ast.CommentGroup) CommentGroup {
	if cg == nil {
		return CommentGroup{}
	}
	list := make([]string, len(cg.List))
	for i, c := range cg.List {
		list[i] = c.Text
	}
	return CommentGroup{
		List: list,
	}
}

func (cg CommentGroup) String() string {
	return strings.Join(ast.TrimComments(cg.List), "\n")
}

func (cg CommentGroup) Format(beginAndEnd ...string) string {
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
func (cg CommentGroup) FormatIndent(prefix, indent string, beginAndEnd ...string) string {
	if len(cg.List) == 0 {
		return ""
	}
	lines := ast.TrimComments(cg.List)
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
