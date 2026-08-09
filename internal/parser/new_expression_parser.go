package parser

// Port of src/NodeParser/NewExpressionParser.ts.

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

type NewExpressionParser struct {
	typeChecker     *checker.Checker
	childNodeParser NodeParser
	synth           *SynthesizedSymbols
}

func NewNewExpressionParser(typeChecker *checker.Checker, childNodeParser NodeParser, synth *SynthesizedSymbols) *NewExpressionParser {
	return &NewExpressionParser{typeChecker: typeChecker, childNodeParser: childNodeParser, synth: synth}
}

func (p *NewExpressionParser) SupportsNode(node *ast.Node) bool {
	return node.Kind == ast.KindNewExpression
}

func (p *NewExpressionParser) CreateType(node *ast.Node, context *Context, _ *types.ReferenceType) types.Type {
	t := p.typeChecker.GetTypeAtLocation(node)

	decl := expressionDeclaration(p.typeChecker, t, node, p.synth)
	return p.childNodeParser.CreateType(decl, newCallArgumentContext(p.childNodeParser, node, context), nil)
}
