package types

import (
	"strings"

	"github.com/next/next/src/ast"
	"github.com/next/next/src/token"
)

// @template(object/Comment)
// Comment represents a line comment.
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

// @template(object/Comment.Text)
// Text returns the content of the comment in Next source code.
// The content is trimmed by the comment characters.
// For example, the comment "// hello comment" will return "hello comment".
func (c *Comment) Text() string {
	if c == nil || len(c.list) == 0 {
		return ""
	}
	return strings.Join(ast.TrimComments(c.list), "\n")
}

// @template(object/Comment.String)
// String returns the origin content of the comment in Next source code.
func (c *Comment) String() string {
	if c == nil || len(c.list) == 0 {
		return ""
	}
	return formatComments(c.list, false, "", " ")
}

// @template(object/Doc)
// Doc represents a documentation comment for a declaration.
//
// Example:
//
// ```next
//
//	// This is a documentation comment.
//	// It can be multiple lines.
//	struct User {
//		// This is a field documentation comment.
//		// It can be multiple lines.
//		name string
//	}
//
// ```
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

// @template(object/Doc.Text)
// Text returns the content of the documentation comment in Next source code.
// The content is trimmed by the comment characters.
// For example, the comment "// hello comment" will return "hello comment".
func (d *Doc) Text() string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return strings.Join(ast.TrimComments(d.list), "\n")
}

// @template(object/Doc.String)
// String returns the origin content of the documentation comment in Next source code.
func (d *Doc) String() string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return formatComments(d.list, true, "", "")
}

// @template(object/Doc.Format)
// Format formats the documentation comment with the given prefix, ident, and begin and end strings.
//
// Example:
//
// ```next
//
//	// This is a documentation comment.
//	// It can be multiple lines.
//	struct User {
//		// This is a field documentation comment.
//		// It can be multiple lines.
//		name string
//	}
//
// ```
//
// ```npl
// {{- define "next/c/doc" -}}
// {{.Format "" " * " "/**\n" " */" | align}}
// {{- end}}
// ```
//
// Output:
//
// ```c
//
//	/**
//	 * This is a documentation comment.
//	 * It can be multiple lines.
//	 */
//	typedef struct {
//		/**
//		 * This is a field documentation comment.
//		 * It can be multiple lines.
//		 */
//		char *name;
//	} User;
//
// ```
func (d *Doc) Format(prefix, indent string, beginAndEnd ...string) string {
	if d == nil || len(d.list) == 0 {
		return ""
	}
	return formatComments(d.list, true, prefix, indent, beginAndEnd...)
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
func formatComments(list []string, appendNewline bool, prefix, indent string, beginAndEnd ...string) string {
	if len(list) == 0 {
		return ""
	}
	lines := ast.TrimComments(list)
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
	begins := strings.Split(begin, "\n")
	begin = begins[0]
	if len(begins) > 1 {
		lines = append(begins[1:], lines...)
	}
	if end == "" {
		for i, line := range lines {
			lines[i] = prefix + indent + begin + " " + line
		}
	} else {
		lines = append([]string{prefix + begin + lines[0]}, lines[1:]...)
		for i := 1; i < len(lines); i++ {
			lines[i] = prefix + indent + lines[i]
		}
		lines = append(lines, prefix+end)
	}
	if !isLastEmpty && appendNewline {
		lines = append(lines, prefix)
	}
	return strings.Join(lines, "\n")
}
