package formatter

import (
	"github.com/vega/ts-json-schema-generator-go/internal/schema"
	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// NumberTypeFormatter mirrors src/TypeFormatter/NumberTypeFormatter.ts.
type NumberTypeFormatter struct{}

func NewNumberTypeFormatter() *NumberTypeFormatter { return &NumberTypeFormatter{} }

func (f *NumberTypeFormatter) SupportsType(t types.Type) bool {
	return isType[*types.NumberType](t)
}

func (f *NumberTypeFormatter) GetDefinition(t types.Type) *schema.Definition {
	return &schema.Definition{Type: "number"}
}

func (f *NumberTypeFormatter) GetChildren(t types.Type) []types.Type { return nil }
