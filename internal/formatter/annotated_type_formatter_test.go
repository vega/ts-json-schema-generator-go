package formatter

import (
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/schema"
)

func TestApplyAnnotation(t *testing.T) {
	sub := &schema.Definition{Type: "string"}
	extra := func(key string, value any) map[string]any { return map[string]any{key: value} }

	cases := []struct {
		name  string
		start schema.Definition
		key   string
		value any
		want  schema.Definition
	}{
		{"$id string", schema.Definition{}, "$id", "x", schema.Definition{ID: "x"}},
		{"$id raw", schema.Definition{ID: "old"}, "$id", 1.0, schema.Definition{Extra: extra("$id", 1.0)}},
		{"$schema string", schema.Definition{}, "$schema", "x", schema.Definition{Schema: "x"}},
		{"$schema raw", schema.Definition{Schema: "old"}, "$schema", true, schema.Definition{Extra: extra("$schema", true)}},
		{"$ref string", schema.Definition{}, "$ref", "#/x", schema.Definition{Ref: "#/x"}},
		{"$ref raw", schema.Definition{Ref: "old"}, "$ref", 1.0, schema.Definition{Extra: extra("$ref", 1.0)}},
		{"$comment string", schema.Definition{}, "$comment", "c", schema.Definition{Comment: "c"}},
		{"$comment raw", schema.Definition{Comment: "old"}, "$comment", 1.0, schema.Definition{Extra: extra("$comment", 1.0)}},
		{"title string", schema.Definition{}, "title", "T", schema.Definition{Title: "T"}},
		{"title raw", schema.Definition{Title: "old"}, "title", 1.0, schema.Definition{Extra: extra("title", 1.0)}},
		{"format string", schema.Definition{}, "format", "date", schema.Definition{Format: "date"}},
		{"format raw", schema.Definition{Format: "old"}, "format", 1.0, schema.Definition{Extra: extra("format", 1.0)}},
		{"type", schema.Definition{Type: "number"}, "type", "string", schema.Definition{Type: "string"}},
		{"enum list", schema.Definition{}, "enum", []any{"a", 1.0}, schema.Definition{Enum: []any{"a", 1.0}}},
		{"enum raw", schema.Definition{Enum: []any{"a"}}, "enum", "a", schema.Definition{Extra: extra("enum", "a")}},
		{"const", schema.Definition{}, "const", "c", schema.Definition{Const: schema.Ptr("c")}},
		{"items", schema.Definition{}, "items", sub, schema.Definition{Items: sub}},
		{"additionalItems", schema.Definition{}, "additionalItems", false, schema.Definition{AdditionalItems: false}},
		{"additionalProperties", schema.Definition{}, "additionalProperties", true, schema.Definition{AdditionalProperties: true}},
		{"minItems number", schema.Definition{}, "minItems", 2.0, schema.Definition{MinItems: schema.IntPtr(2)}},
		{"minItems raw", schema.Definition{MinItems: schema.IntPtr(1)}, "minItems", "2", schema.Definition{Extra: extra("minItems", "2")}},
		{"maxItems number", schema.Definition{}, "maxItems", 3.0, schema.Definition{MaxItems: schema.IntPtr(3)}},
		{"maxItems raw", schema.Definition{MaxItems: schema.IntPtr(1)}, "maxItems", "3", schema.Definition{Extra: extra("maxItems", "3")}},
		{"required list", schema.Definition{}, "required", []any{"a", "b"}, schema.Definition{Required: []string{"a", "b"}}},
		{"required raw", schema.Definition{Required: []string{"a"}}, "required", []any{"a", 1.0}, schema.Definition{Extra: extra("required", []any{"a", 1.0})}},
		{"not", schema.Definition{Not: sub}, "not", "x", schema.Definition{Extra: extra("not", "x")}},
		{"allOf", schema.Definition{AllOf: []*schema.Definition{sub}}, "allOf", "x", schema.Definition{Extra: extra("allOf", "x")}},
		{"anyOf", schema.Definition{AnyOf: []*schema.Definition{sub}}, "anyOf", "x", schema.Definition{Extra: extra("anyOf", "x")}},
		{"oneOf", schema.Definition{OneOf: []*schema.Definition{sub}}, "oneOf", "x", schema.Definition{Extra: extra("oneOf", "x")}},
		{"if", schema.Definition{If: sub}, "if", "x", schema.Definition{Extra: extra("if", "x")}},
		{"then", schema.Definition{Then: sub}, "then", "x", schema.Definition{Extra: extra("then", "x")}},
		{"else", schema.Definition{Else: sub}, "else", "x", schema.Definition{Extra: extra("else", "x")}},
		{"properties", schema.Definition{Properties: schema.NewProperties()}, "properties", "x", schema.Definition{Extra: extra("properties", "x")}},
		{"patternProperties", schema.Definition{PatternProperties: map[string]*schema.Definition{"a": sub}}, "patternProperties", "x", schema.Definition{Extra: extra("patternProperties", "x")}},
		{"propertyNames", schema.Definition{PropertyNames: sub}, "propertyNames", "x", schema.Definition{Extra: extra("propertyNames", "x")}},
		{"discriminator", schema.Definition{Discriminator: "kind"}, "discriminator", "x", schema.Definition{Extra: extra("discriminator", "x")}},
		{"other keyword", schema.Definition{Type: "string"}, "description", "d", schema.Definition{Type: "string", Extra: extra("description", "d")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := tc.start
			applyAnnotation(&def, tc.key, tc.value)
			if !reflect.DeepEqual(def, tc.want) {
				t.Errorf("applyAnnotation(%q, %#v):\n got %#v\nwant %#v", tc.key, tc.value, def, tc.want)
			}
		})
	}
}
