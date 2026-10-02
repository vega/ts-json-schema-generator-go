package formatter

import (
	"github.com/vega/ts-json-schema-generator-go/internal/schema"
	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// BooleanTypeFormatter mirrors src/TypeFormatter/BooleanTypeFormatter.ts.
type BooleanTypeFormatter struct{}

func NewBooleanTypeFormatter() *BooleanTypeFormatter { return &BooleanTypeFormatter{} }

func (f *BooleanTypeFormatter) SupportsType(t types.Type) bool {
	return isType[*types.BooleanType](t)
}

func (f *BooleanTypeFormatter) GetDefinition(t types.Type) *schema.Definition {
	return &schema.Definition{Type: "boolean"}
}

func (f *BooleanTypeFormatter) GetChildren(t types.Type) []types.Type { return nil }
