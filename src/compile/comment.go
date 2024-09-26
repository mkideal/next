package compile

import (
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// @api(Object/Comment) represents a line comment or a comment group in Next source code.
// Use this in templates to access and format comments.
type Comment struct {
	pos  token.Pos
	list []string
}

func newComment(cg *ast.CommentGroup) *Comment {
	if cg == nil {
		return nil
	}
	return &Comment{
		pos:  cg.Pos(),
		list: makeCommentList(cg),
	}
}

// @api(Object/Comment.Text) returns the content of the comment without comment delimiters.
//
// Usage in templates:
// ```npl
// {{.Comment.Text}}
// ```
func (c *Comment) Text() string {
	if c == nil || len(c.list) == 0 {
		return ""
	}
	return strings.Join(ast.TrimComments(c.list), "\n")
}

// @api(Object/Comment.String) returns the full original comment text, including delimiters.
//
// Usage in templates:
// ```npl
// {{.Comment.String}}
// ```
func (c *Comment) String() string {
	if c == nil || len(c.list) == 0 {
		return ""
	}
	return formatComments(c.list, false, " // ")
}

// @api(Object/Doc) represents a documentation comment for a declaration in Next source code.
// Use this in templates to access and format documentation comments.
type Doc struct {
	pos  token.Pos
	list []string
}

func newDoc(cg *ast.CommentGroup) *Doc {
	if cg == nil {
		return nil
	}
	return &Doc{
		pos:  cg.Pos(),
		list: makeCommentList(cg),
	}
}

// @api(Object/Doc.Text) returns the content of the documentation comment without comment delimiters.
//
// Usage in templates:
// ```npl
// {{.Doc.Text}}
// ```
func (d *Doc) Text() string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return strings.Join(ast.TrimComments(d.list), "\n")
}

// @api(Object/Doc.String) returns the full original documentation comment, including delimiters.
//
// Usage in templates:
// ```npl
// {{.Doc.String}}
// ```
func (d *Doc) String() string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return formatComments(d.list, true, "// ", "")
}

// @api(Object/Doc.Format) formats the documentation comment for various output styles.
//
// Parameters: (_indent_ string[, _begin_ string[, _end_ string]])
//
// Usage in templates:
//
// ```npl
// {{.Doc.Format "/// "}}
// {{.Doc.Format " * " "/**\n" " */"}}
// ```
//
// Example output:
//
// ```c
//
//	/// This is a documentation comment.
//	/// It can be multiple lines.
//	/**
//	 * This is a documentation comment.
//	 * It can be multiple lines.
//	 */
//
// ```
func (d *Doc) Format(indent string, beginAndEnd ...string) string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return formatComments(d.list, true, indent, beginAndEnd...)
}

// makeCommentList returns a list of comments from the comment group.
func makeCommentList(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	list := make([]string, len(cg.List))
	for i, c := range cg.List {
		list[i] = c.Text
	}
	return list
}

// formatComments formats the comment group with the given prefix, ident, and begin and end strings.
func formatComments(list []string, appendNewline bool, indent string, beginAndEnd ...string) string {
	if len(list) == 0 {
		return ""
	}
	lines := ast.TrimComments(list)
	if len(lines) == 0 {
		return ""
	}
	isLastEmpty := lines[len(lines)-1] == ""
	begin, end := "", ""
	if len(beginAndEnd) > 0 && beginAndEnd[0] != "" {
		begin = beginAndEnd[0]
		if len(beginAndEnd) > 1 && beginAndEnd[1] != "" {
			end = beginAndEnd[1]
		}
	}
	begins := strings.Split(begin, "\n")
	begin = begins[0]
	if len(begins) > 1 {
		lines = append(begins[1:], lines...)
	}
	if end == "" {
		for i, line := range lines {
			lines[i] = indent + begin + line
		}
	} else {
		lines = append([]string{begin + lines[0]}, lines[1:]...)
		for i := 1; i < len(lines); i++ {
			lines[i] = indent + lines[i]
		}
		lines = append(lines, end)
	}
	if !isLastEmpty && appendNewline {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
