package formatter

import (
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/schema"
)

func TestMakeNullable(t *testing.T) {
	str := &schema.Definition{Type: "string"}
	null := &schema.Definition{Type: "null"}

	cases := []struct {
		name string
		def  schema.Definition
		want schema.Definition
	}{
		{"oneOf without null", schema.Definition{OneOf: []*schema.Definition{str}}, schema.Definition{OneOf: []*schema.Definition{str, null}}},
		{"anyOf without null", schema.Definition{AnyOf: []*schema.Definition{str}}, schema.Definition{AnyOf: []*schema.Definition{str, null}}},
		{"anyOf with null", schema.Definition{AnyOf: []*schema.Definition{str, null}}, schema.Definition{AnyOf: []*schema.Definition{str, null}}},
		{"string type", schema.Definition{Type: "string"}, schema.Definition{Type: []string{"string", "null"}}},
		{"null type", schema.Definition{Type: "null"}, schema.Definition{Type: "null"}},
		{"[]string type", schema.Definition{Type: []string{"string"}}, schema.Definition{Type: []string{"string", "null"}}},
		{"[]string type with null", schema.Definition{Type: []string{"string", "null"}}, schema.Definition{Type: []string{"string", "null"}}},
		{"[]any type", schema.Definition{Type: []any{"string"}}, schema.Definition{Type: []any{"string", "null"}}},
		{"[]any type with null", schema.Definition{Type: []any{"null", 1.0}}, schema.Definition{Type: []any{"null", 1.0}}},
		{"enum", schema.Definition{Type: "string", Enum: []any{"a"}}, schema.Definition{Type: []string{"string", "null"}, Enum: []any{"a", nil}}},
		{"enum with null", schema.Definition{Type: []string{"string", "null"}, Enum: []any{"a", nil}}, schema.Definition{Type: []string{"string", "null"}, Enum: []any{"a", nil}}},
		{
			"object type",
			schema.Definition{Type: "object", Title: "T", Extra: map[string]any{"description": "d", "x": 1.0}},
			schema.Definition{Title: "T", Extra: map[string]any{"description": "d"}, AnyOf: []*schema.Definition{
				{Type: "object", Extra: map[string]any{"x": 1.0}}, null,
			}},
		},
		{"ref", schema.Definition{Ref: "#/definitions/A"}, schema.Definition{AnyOf: []*schema.Definition{{Ref: "#/definitions/A"}, null}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := tc.def
			got := makeNullable(&def)
			if !reflect.DeepEqual(*got, tc.want) {
				t.Errorf("makeNullable:\n got %v\nwant %v", got, &tc.want)
			}
		})
	}
}
