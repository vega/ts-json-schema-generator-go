package parser

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// TupleNodeParser parses tuple type nodes (src/NodeParser/TupleNodeParser.ts).
type TupleNodeParser struct {
	typeChecker     *checker.Checker
	childNodeParser NodeParser
}

func NewTupleNodeParser(typeChecker *checker.Checker, childNodeParser NodeParser) *TupleNodeParser {
	return &TupleNodeParser{typeChecker: typeChecker, childNodeParser: childNodeParser}
}

func (p *TupleNodeParser) SupportsNode(node *ast.Node) bool {
	return node.Kind == ast.KindTupleType
}

func (p *TupleNodeParser) CreateType(node *ast.Node, ctx *Context, _ *types.ReferenceType) types.Type {
	return types.NewTupleType(createTypes(p.childNodeParser, node.AsTupleTypeNode().Elements.Nodes, ctx))
}
