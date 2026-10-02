package parser

import (
	"slices"
	"testing"
)

func TestExpandTemplateMatrix(t *testing.T) {
	cases := []struct {
		name   string
		matrix [][]string
		want   []string
	}{
		{"single row", [][]string{{"a", "b"}}, []string{"a", "b"}},
		{"leftmost row varies slowest", [][]string{{"a", "b"}, {"1", "2"}}, []string{"a1", "a2", "b1", "b2"}},
		{"three rows", [][]string{{"a"}, {"1", "2"}, {"x", "y"}}, []string{"a1x", "a1y", "a2x", "a2y"}},
		{"empty row", [][]string{{"a", "b"}, {}}, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := expandTemplateMatrix(c.matrix); !slices.Equal(got, c.want) {
				t.Errorf("expandTemplateMatrix(%q) = %q, want %q", c.matrix, got, c.want)
			}
		})
	}
}
