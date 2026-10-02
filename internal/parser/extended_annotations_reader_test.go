package parser

import (
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

func TestCollapseSingleNewlines(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a\nb", "a b"},
		{"a\nb\nc", "a b c"},
		{"a\n\nb", "a\n\nb"},
		{"a\n* item", "a\n* item"},
		{"a\n- item", "a\n- item"},
		{"\na", "\na"},
		{"a\n", "a\n"},
		{"a\r\nb", "a\r b"},
	}
	for _, tt := range tests {
		if got := collapseSingleNewlines(tt.input); got != tt.want {
			t.Errorf("collapseSingleNewlines(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindLinkNameEnd(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 0},
		{" text", 0},
		{"://example.com", 14},
		{"://example.com|label", 14},
		{"() rest", 2},
		{"<T> rest", 3},
		{"<A<B>>x", 6},
		{"<unclosed", 0},
	}
	for _, tt := range tests {
		if got := findLinkNameEnd(tt.text); got != tt.want {
			t.Errorf("findLinkNameEnd(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestSkipSeparatorFromLinkText(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"", ""},
		{"label", "label"},
		{"|", ""},
		{"|label", "label"},
		{"|  label", "label"},
		{" |label", " |label"},
	}
	for _, tt := range tests {
		if got := skipSeparatorFromLinkText(tt.text); got != tt.want {
			t.Errorf("skipSeparatorFromLinkText(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestTypeAnnotation(t *testing.T) {
	tests := []struct {
		name string
		tags []jsDocTagInfo
		want types.Annotations
	}{
		{"no tags", nil, nil},
		{"no asType", []jsDocTagInfo{{"example", "1"}}, nil},
		{"asType", []jsDocTagInfo{{"example", "1"}, {"asType", "integer"}}, types.Annotations{"type": "integer"}},
		{"first asType wins", []jsDocTagInfo{{"asType", "integer"}, {"asType", "string"}}, types.Annotations{"type": "integer"}},
	}
	for _, tt := range tests {
		if got := typeAnnotation(tt.tags); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: typeAnnotation() = %#v, want %#v", tt.name, got, tt.want)
		}
	}
}

func TestExampleAnnotation(t *testing.T) {
	tests := []struct {
		name string
		tags []jsDocTagInfo
		want types.Annotations
	}{
		{"no tags", nil, nil},
		{"only invalid", []jsDocTagInfo{{"example", "not json"}}, nil},
		{
			"valid examples in order",
			[]jsDocTagInfo{{"example", "1"}, {"example", "not json"}, {"default", "2"}, {"example", "{a: 'b'}"}},
			types.Annotations{"examples": []any{1.0, map[string]any{"a": "b"}}},
		},
	}
	for _, tt := range tests {
		if got := exampleAnnotation(tt.tags); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: exampleAnnotation() = %#v, want %#v", tt.name, got, tt.want)
		}
	}
}
