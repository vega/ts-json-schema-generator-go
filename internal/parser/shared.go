package parser

// Helpers that several node parsers use. The TypeScript implementation
// repeats them as private methods in each of the corresponding files.

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"

	"github.com/vega/ts-json-schema-generator-go/internal/tsutils"
)

// newTypeArgumentContext builds the sub context for a node's explicit type
// arguments, each parsed in the parent context.
func newTypeArgumentContext(child NodeParser, node *ast.Node, parentContext *Context) *Context {
	subContext := NewContext(node)
	for _, typeArg := range node.TypeArguments() {
		subContext.PushArgument(child.CreateType(typeArg, parentContext, nil))
	}
	return subContext
}

// newCallArgumentContext builds the sub context for the value arguments of a
// call or new expression, each parsed in the parent context.
func newCallArgumentContext(child NodeParser, node *ast.Node, parentContext *Context) *Context {
	subContext := NewContext(node)
	for _, arg := range node.Arguments() {
		subContext.PushArgument(child.CreateType(arg, parentContext, nil))
	}
	return subContext
}

// expressionDeclaration resolves the declaration standing behind the type of a
// call or new expression. For generic signatures such as <T>(type: T) => T
// there is no reference to the original type, so the checker's synthesized
// type node is preferred over the symbol's declaration.
func expressionDeclaration(typeChecker *checker.Checker, t *checker.Type, node *ast.Node, synth *SynthesizedSymbols) *ast.Node {
	symbol := t.Symbol()
	if symbol == nil && t.Alias() != nil {
		symbol = t.Alias().Symbol()
	}

	decl := typeChecker.TypeToTypeNode(t, node, nodeBuilderFlagsIgnoreErrors, synth.Map())
	if decl == nil && symbol != nil {
		decl = symbol.ValueDeclaration
		if decl == nil && len(symbol.Declarations) > 0 {
			decl = symbol.Declarations[0]
		}
	}

	if decl == nil {
		panic(NewUnknownNodeError(node))
	}
	return decl
}

// memberAdditionalProperties resolves the additionalProperties value implied
// by a member list's index signature, falling back to the parser's default.
func memberAdditionalProperties(child NodeParser, node *ast.Node, context *Context, fallback any) any {
	for _, member := range node.Members() {
		if ast.IsIndexSignatureDeclaration(member) {
			if t := child.CreateType(member.Type(), context, nil); t != nil {
				return t
			}
			return fallback
		}
	}
	return fallback
}

// memberPropertyName renders a member's name, resolving computed names through
// the checker. nodeText falls back to the text property for synthesized nodes,
// mirroring the try/catch with escapedText/text upstream.
func memberPropertyName(typeChecker *checker.Checker, propertyName *ast.Node) string {
	if propertyName.Kind == ast.KindComputedPropertyName {
		if symbol := tsutils.GetSymbolAtLocation(typeChecker, propertyName); symbol != nil {
			return symbol.Name
		}
	}
	return nodeText(propertyName)
}
