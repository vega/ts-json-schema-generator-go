package parser

import (
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

func isNotString(t types.Type) bool {
	_, isString := t.(*types.StringType)
	return !isString
}

func isNotStringLiteral(t types.Type) bool {
	literal, isLiteral := t.(*types.LiteralType)
	return !isLiteral || !literal.IsString()
}

func TestNarrowType(t *testing.T) {
	stringDefinition := types.NewDefinitionType("S", &types.StringType{})
	numberDefinition := types.NewDefinitionType("N", &types.NumberType{})
	nestedAlias := types.NewAliasType("nested", union(&types.StringType{}, &types.BooleanType{}))

	cases := []struct {
		name      string
		input     types.Type
		predicate func(types.Type) bool
		want      types.Type
		// same requires the result to be the input itself, which keeps
		// definitions when nothing was narrowed.
		same bool
	}{
		{
			name:      "kept non-union",
			input:     &types.NumberType{},
			predicate: isNotString,
			same:      true,
		},
		{
			name:      "kept definition keeps the definition",
			input:     numberDefinition,
			predicate: isNotString,
			same:      true,
		},
		{
			name:      "removed non-union",
			input:     &types.StringType{},
			predicate: isNotString,
			want:      &types.NeverType{},
		},
		{
			name:      "removed definition",
			input:     stringDefinition,
			predicate: isNotString,
			want:      &types.NeverType{},
		},
		{
			name:      "unchanged union",
			input:     union(&types.NumberType{}, &types.BooleanType{}),
			predicate: isNotString,
			same:      true,
		},
		{
			name:      "union narrowed to one member",
			input:     union(&types.StringType{}, &types.NumberType{}),
			predicate: isNotString,
			want:      &types.NumberType{},
		},
		{
			name:      "union narrowed to several members",
			input:     union(&types.StringType{}, &types.NumberType{}, &types.BooleanType{}),
			predicate: isNotString,
			want:      union(&types.NumberType{}, &types.BooleanType{}),
		},
		{
			name:      "union narrowed to nothing",
			input:     union(&types.StringType{}, stringDefinition),
			predicate: isNotString,
			want:      &types.NeverType{},
		},
		{
			name:      "kept members stay wrapped",
			input:     union(&types.StringType{}, numberDefinition, &types.NullType{}),
			predicate: isNotString,
			want:      union(numberDefinition, &types.NullType{}),
		},
		{
			name:      "nested union in alias",
			input:     union(nestedAlias, &types.NullType{}),
			predicate: isNotString,
			want:      union(&types.BooleanType{}, &types.NullType{}),
		},
		{
			name:      "enum",
			input:     types.NewEnumType("E", []types.EnumValue{"a", 1.0, nil}),
			predicate: isNotStringLiteral,
			want:      union(literal(1.0), &types.NullType{}),
		},
		{
			name:      "unchanged enum",
			input:     types.NewEnumType("E", []types.EnumValue{1.0, 2.0}),
			predicate: isNotStringLiteral,
			same:      true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NarrowType(c.input, c.predicate)
			if c.same {
				if got != c.input {
					t.Errorf("NarrowType(%s) = %s, want the input itself", c.input.ID(), got.ID())
				}
				return
			}
			if got.ID() != c.want.ID() {
				t.Errorf("NarrowType(%s) = %s, want %s", c.input.ID(), got.ID(), c.want.ID())
			}
		})
	}
}
