package parser

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// InferTypeNodeParser handles `infer T` placeholders which are resolved later
// by the conditional type parser (src/NodeParser/InferTypeNodeParser.ts).
type InferTypeNodeParser struct{}

func NewInferTypeNodeParser() *InferTypeNodeParser {
	return &InferTypeNodeParser{}
}

func (p *InferTypeNodeParser) SupportsNode(node *ast.Node) bool {
	return node.Kind == ast.KindInferType
}

func (p *InferTypeNodeParser) CreateType(node *ast.Node, _ *Context, _ *types.ReferenceType) types.Type {
	return types.NewInferType(node.AsInferTypeNode().TypeParameter.Name().Text())
}
