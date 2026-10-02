package parser

import (
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

func TestAnnotationsFromTags(t *testing.T) {
	reader := NewBasicAnnotationsReader(map[string]bool{"custom": true})
	tests := []struct {
		name string
		tags []jsDocTagInfo
		want types.Annotations
	}{
		{"no tags", nil, nil},
		{"unknown tags", []jsDocTagInfo{{"param", "x desc"}, {"see", "{a: 1}"}}, nil},
		{"text tag", []jsDocTagInfo{{"title", "T"}}, types.Annotations{"title": "T"}},
		{"empty text tag", []jsDocTagInfo{{"description", ""}}, types.Annotations{"description": ""}},
		{"dollar tag", []jsDocTagInfo{{"id", "foo"}}, types.Annotations{"$id": "foo"}},
		{"json tag", []jsDocTagInfo{{"minimum", "5"}}, types.Annotations{"minimum": 5.0}},
		{"json tag without value", []jsDocTagInfo{{"deprecated", ""}}, types.Annotations{"deprecated": true}},
		{"json tag with invalid JSON", []jsDocTagInfo{{"default", "not json"}}, types.Annotations{"default": "not json"}},
		{"extra tag", []jsDocTagInfo{{"custom", "{a: 1}"}}, types.Annotations{"custom": map[string]any{"a": 1.0}}},
		{"last tag wins", []jsDocTagInfo{{"minimum", "1"}, {"minimum", "2"}}, types.Annotations{"minimum": 2.0}},
	}
	for _, tt := range tests {
		if got := reader.annotationsFromTags(tt.tags); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: annotationsFromTags() = %#v, want %#v", tt.name, got, tt.want)
		}
	}
}
