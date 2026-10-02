package formatter

import (
	"github.com/vega/ts-json-schema-generator-go/internal/schema"
	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// AliasTypeFormatter mirrors src/TypeFormatter/AliasTypeFormatter.ts.
type AliasTypeFormatter struct {
	childTypeFormatter TypeFormatter
}

func NewAliasTypeFormatter(childTypeFormatter TypeFormatter) *AliasTypeFormatter {
	return &AliasTypeFormatter{childTypeFormatter: childTypeFormatter}
}

func (f *AliasTypeFormatter) SupportsType(t types.Type) bool {
	return isType[*types.AliasType](t)
}

func (f *AliasTypeFormatter) GetDefinition(t types.Type) *schema.Definition {
	return f.childTypeFormatter.GetDefinition(t.(*types.AliasType).Type)
}

func (f *AliasTypeFormatter) GetChildren(t types.Type) []types.Type {
	return f.childTypeFormatter.GetChildren(t.(*types.AliasType).Type)
}
