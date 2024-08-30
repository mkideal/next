package types

import (
	"testing"
)

func TestFormatComments(t *testing.T) {
	list := []string{
		"// This is a comment.",
		"// This is another comment.",
	}
	const appendNewline = true
	if got, want := formatComments(list, appendNewline, "", ""), "// This is a comment.\n// This is another comment.\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := formatComments(list, appendNewline, "", " *", "/*", "*/"), "/* This is a comment.\n * This is another comment.\n */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := formatComments(list, appendNewline, "  ", " *", "/*", "*/"), "  /* This is a comment.\n   * This is another comment.\n   */"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := formatComments(list, appendNewline, "", "", "<!--\n", "-->"), "<!--\n This is a comment.\n This is another comment.\n -->"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
