package parser

import "testing"

func TestGetTextWithoutStars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/** foo */", "foo"},
		{"/**\n * a\n *  b\n *c\n */", "a\n b\nc"},
		{"/**\r\n * a\r\n */", "a"},
		{"/**\n plain\n */", " plain"},
		{"/**\n *\n */", ""},
	}
	for _, tt := range tests {
		if got := getTextWithoutStars(tt.input); got != tt.want {
			t.Errorf("getTextWithoutStars(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
