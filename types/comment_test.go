package types_test

import (
	"testing"

	"github.com/gopherd/next/types"
)

func TestCommentFormat(t *testing.T) {
	c := types.Comment{Text: "// This is a comment."}
	if got, want := c.Format(), "// This is a comment."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := c.Format("//", ""), "// This is a comment."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := c.Format("//", "//"), "// This is a comment. //"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := c.Format("/*", "*/"), "/* This is a comment. */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCommentGroupFormat(t *testing.T) {
	cg := types.CommentGroup{
		List: []string{
			"// This is a comment.",
			"// This is another comment.",
		},
	}
	if got, want := cg.Format(), "// This is a comment.\n// This is another comment.\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.Format("//", ""), "// This is a comment.\n// This is another comment.\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.Format("//", "//"), "// This is a comment.\n This is another comment.\n //"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.Format("/*", "*/"), "/* This is a comment.\n This is another comment.\n */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCommentGroupFormatIndent(t *testing.T) {
	cg := types.CommentGroup{
		List: []string{
			"// This is a comment.",
			"// This is another comment.",
		},
	}
	if got, want := cg.FormatIndent("", ""), "// This is a comment.\n// This is another comment.\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.FormatIndent("", " *", "/*", "*/"), "/* This is a comment.\n * This is another comment.\n */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.FormatIndent("  ", " *", "/*", "*/"), "  /* This is a comment.\n   * This is another comment.\n   */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := cg.FormatIndent("", "", "<!--\n", "-->"), "<!--\n This is a comment.\n This is another comment.\n -->"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
