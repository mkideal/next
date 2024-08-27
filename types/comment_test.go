package types

import (
	"testing"
)

func TestCommentGroupFormat(t *testing.T) {
	cg := CommentGroup{
		list: []string{
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
	cg := CommentGroup{
		list: []string{
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
