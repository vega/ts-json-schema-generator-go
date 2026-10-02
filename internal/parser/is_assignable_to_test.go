package parser

import (
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/types"
)

// The cases port test/unit/isAssignableTo.test.ts of the TypeScript
// implementation.

func prop(name string, t types.Type, required bool) *types.ObjectProperty {
	return types.NewObjectProperty(name, t, required)
}

func object(id string, bases []types.Type, additionalProperties bool, props ...*types.ObjectProperty) *types.ObjectType {
	return types.NewObjectType(id, bases, props, additionalProperties, false)
}

func union(ts ...types.Type) *types.UnionType { return types.NewUnionType(ts) }

func tuple(ts ...types.Type) *types.TupleType { return types.NewTupleType(ts) }

func literal(v any) *types.LiteralType { return &types.LiteralType{Value: v} }

func reference(t types.Type) *types.ReferenceType {
	ref := types.NewReferenceType()
	ref.SetType(t)
	return ref
}

type assignableCase struct {
	name           string
	target, source types.Type
	want           bool
}

func runAssignableCases(t *testing.T, cases []assignableCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsAssignableTo(c.target, c.source, nil); got != c.want {
				t.Errorf("IsAssignableTo(%s, %s) = %v, want %v", c.target.ID(), c.source.ID(), got, c.want)
			}
		})
	}
}

func TestIsAssignableToPrimitives(t *testing.T) {
	runAssignableCases(t, []assignableCase{
		{"boolean to boolean", &types.BooleanType{}, &types.BooleanType{}, true},
		{"null to null", &types.NullType{}, &types.NullType{}, true},
		{"number to number", &types.NumberType{}, &types.NumberType{}, true},
		{"string to string", &types.StringType{}, &types.StringType{}, true},
		{"undefined to undefined", &types.UndefinedType{}, &types.UndefinedType{}, true},
		{"void to void", &types.VoidType{}, &types.VoidType{}, true},

		{"null to boolean", &types.BooleanType{}, &types.NullType{}, false},
		{"number to null", &types.NullType{}, &types.NumberType{}, false},
		{"boolean to number", &types.NumberType{}, &types.BooleanType{}, false},
		{"string to boolean", &types.BooleanType{}, &types.StringType{}, false},
		{"undefined to string", &types.StringType{}, &types.UndefinedType{}, false},
		{"boolean to undefined", &types.UndefinedType{}, &types.BooleanType{}, false},
		{"string to string[]", &types.ArrayType{Item: &types.StringType{}}, &types.StringType{}, false},
	})
}

func TestIsAssignableToArraysAndUnions(t *testing.T) {
	stringOrNumber := union(&types.StringType{}, &types.NumberType{})
	runAssignableCases(t, []assignableCase{
		{"string[] to string[]", &types.ArrayType{Item: &types.StringType{}}, &types.ArrayType{Item: &types.StringType{}}, true},
		{"number[] to string[]", &types.ArrayType{Item: &types.StringType{}}, &types.ArrayType{Item: &types.NumberType{}}, false},
		{"string to union", stringOrNumber, &types.StringType{}, true},
		{"number to union", stringOrNumber, &types.NumberType{}, true},
		{"boolean to union", stringOrNumber, &types.BooleanType{}, false},
	})
}

func TestIsAssignableToDerefs(t *testing.T) {
	runAssignableCases(t, []assignableCase{
		{"string to string ref", reference(&types.StringType{}), &types.StringType{}, true},
		{"number to string ref", reference(&types.StringType{}), &types.NumberType{}, false},
		{"string ref to string", &types.StringType{}, reference(&types.StringType{}), true},
		{"string ref to number", &types.NumberType{}, reference(&types.StringType{}), false},
		{"string ref to string ref", reference(&types.StringType{}), reference(&types.StringType{}), true},
		{"string ref to number ref", reference(&types.NumberType{}), reference(&types.StringType{}), false},

		{"string to string alias", types.NewAliasType("a", &types.StringType{}), &types.StringType{}, true},
		{"number to string alias", types.NewAliasType("a", &types.StringType{}), &types.NumberType{}, false},
		{"string alias to string", &types.StringType{}, types.NewAliasType("a", &types.StringType{}), true},
		{"string alias to number", &types.NumberType{}, types.NewAliasType("a", &types.StringType{}), false},
		{"string alias to string alias", types.NewAliasType("a", &types.StringType{}), types.NewAliasType("b", &types.StringType{}), true},
		{"string alias to number alias", types.NewAliasType("c", &types.NumberType{}), types.NewAliasType("a", &types.StringType{}), false},

		{"string to annotated string", &types.AnnotatedType{Type: &types.StringType{}, Annotations: types.Annotations{}}, &types.StringType{}, true},
		{"number to annotated string", &types.AnnotatedType{Type: &types.StringType{}, Annotations: types.Annotations{}}, &types.NumberType{}, false},
		{"annotated string to string", &types.StringType{}, &types.AnnotatedType{Type: &types.StringType{}, Annotations: types.Annotations{}}, true},
		{"annotated string to number", &types.NumberType{}, &types.AnnotatedType{Type: &types.StringType{}, Annotations: types.Annotations{}}, false},

		{"string to string definition", types.NewDefinitionType("a", &types.StringType{}), &types.StringType{}, true},
		{"number to string definition", types.NewDefinitionType("a", &types.StringType{}), &types.NumberType{}, false},
		{"string definition to string", &types.StringType{}, types.NewDefinitionType("a", &types.StringType{}), true},
		{"string definition to number", &types.NumberType{}, types.NewDefinitionType("a", &types.StringType{}), false},
		{"string definition to string definition", types.NewDefinitionType("a", &types.StringType{}), types.NewDefinitionType("b", &types.StringType{}), true},
		{"string definition to number definition", types.NewDefinitionType("c", &types.NumberType{}), types.NewDefinitionType("a", &types.StringType{}), false},
	})
}

func TestIsAssignableToTopAndBottom(t *testing.T) {
	obj := object("obj", nil, true, prop("foo", &types.StringType{}, true))
	others := []struct {
		name string
		typ  types.Type
	}{
		{"number[]", &types.ArrayType{Item: &types.NumberType{}}},
		{"string & null", types.NewIntersectionType([]types.Type{&types.StringType{}, &types.NullType{}})},
		{"literal", literal("literal")},
		{"null", &types.NullType{}},
		{"object", obj},
		{"boolean", &types.BooleanType{}},
		{"number", &types.NumberType{}},
		{"string", &types.StringType{}},
		{"tuple", tuple(&types.StringType{}, &types.NumberType{})},
		{"undefined", &types.UndefinedType{}},
	}

	var cases []assignableCase
	for _, other := range others {
		cases = append(cases,
			assignableCase{"any to " + other.name, other.typ, &types.AnyType{}, true},
			assignableCase{"never to " + other.name, other.typ, &types.NeverType{}, true},
			assignableCase{other.name + " to any", &types.AnyType{}, other.typ, true},
			assignableCase{other.name + " to unknown", &types.UnknownType{}, other.typ, true},
			assignableCase{"unknown to " + other.name, other.typ, &types.UnknownType{}, false},
		)
	}
	cases = append(cases,
		assignableCase{"any to never", &types.NeverType{}, &types.AnyType{}, false},
		assignableCase{"never to never", &types.NeverType{}, &types.NeverType{}, true},
		assignableCase{"unknown to never", &types.NeverType{}, &types.UnknownType{}, false},
		assignableCase{"unknown to any", &types.AnyType{}, &types.UnknownType{}, true},
		assignableCase{"unknown to unknown", &types.UnknownType{}, &types.UnknownType{}, true},
		assignableCase{"any to void", &types.VoidType{}, &types.AnyType{}, true},
		assignableCase{"never to void", &types.VoidType{}, &types.NeverType{}, true},
		assignableCase{"null to void", &types.VoidType{}, &types.NullType{}, true},
		assignableCase{"undefined to void", &types.VoidType{}, &types.UndefinedType{}, true},
		assignableCase{"unknown to void", &types.VoidType{}, &types.UnknownType{}, false},
	)
	runAssignableCases(t, cases)
}

func TestIsAssignableToUnionSources(t *testing.T) {
	typeA := object("a", nil, true, prop("a", &types.StringType{}, true))
	typeB := object("b", nil, true, prop("b", &types.StringType{}, true))
	typeC := object("c", nil, true, prop("c", &types.StringType{}, true))
	typeAB := object("ab", []types.Type{typeA, typeB}, true)
	typeAorB := union(typeA, typeB)
	runAssignableCases(t, []assignableCase{
		{"a|a to ab", typeAB, union(typeA, typeA), false},
		{"b|b to ab", typeAB, union(typeB, typeB), false},
		{"a|b to ab", typeAB, union(typeA, typeB), false},
		{"b|a to ab", typeAB, union(typeB, typeA), false},
		{"b|a|c to ab", typeAB, union(typeB, typeA, typeC), false},
		{"b|a to a|b", typeAorB, union(typeB, typeA), true},
		{"a|b to a|b", typeAorB, union(typeA, typeB), true},
		{"ab|b|c to a|b", typeAorB, union(typeAB, typeB, typeC), false},
	})
}

func TestIsAssignableToTuples(t *testing.T) {
	str := &types.StringType{}
	num := &types.NumberType{}
	fixedLengthArrayLike := object("fixedLengthArrayLike", nil, false, prop("length", literal(2.0), true))
	nonFixedLengthArrayLike := object("nonFixedLengthArrayLike", nil, false, prop("length", num, true))
	optionalLengthArrayLike := object("optionalLengthArrayLike", nil, false, prop("length", num, false))
	nonArrayLike := object("nonArrayLike", nil, false, prop("foo", num, true))
	arrayType := &types.ArrayType{Item: str}
	tupleType := tuple(str, num)
	runAssignableCases(t, []assignableCase{
		{"[string, string] to string[]", &types.ArrayType{Item: str}, tuple(str, str), true},
		{"[string, string] to number[]", &types.ArrayType{Item: num}, tuple(str, str), false},
		{"[string, number] to string[]", &types.ArrayType{Item: str}, tuple(str, num), false},

		{"array to fixed length array-like", fixedLengthArrayLike, arrayType, false},
		{"array to array-like", nonFixedLengthArrayLike, arrayType, true},
		{"array to optional length array-like", optionalLengthArrayLike, arrayType, false},
		{"array to non array-like", nonArrayLike, arrayType, false},
		{"tuple to fixed length array-like", fixedLengthArrayLike, tupleType, true},
		{"tuple to array-like", nonFixedLengthArrayLike, tupleType, false},
		{"tuple to optional length array-like", optionalLengthArrayLike, tupleType, false},
		{"tuple to non array-like", nonArrayLike, tupleType, false},

		{"string[] to tuple", tuple(str, str), &types.ArrayType{Item: str}, false},
		{"string to tuple", tuple(str, str), str, false},
		{"[string, number] to [string, string]", tuple(str, str), tuple(str, num), false},
		{"[string, string] to [string, string]", tuple(str, str), tuple(str, str), true},
		{"[string] to [string, string?]", tuple(str, &types.OptionalType{Type: str}), tuple(str), true},
		{"[string, string] to [string, string?]", tuple(str, &types.OptionalType{Type: str}), tuple(str, str), true},
		{"[string, number, string] to [string, infer T]", tuple(str, types.NewInferType("T")), tuple(str, num, str), false},
		{"[string, number] to [string, infer T]", tuple(str, types.NewInferType("T")), tuple(str, num), true},
		{"[string] to [string, infer T]", tuple(str, types.NewInferType("T")), tuple(str), false},
		{"[string] to [string, ...infer T]", tuple(str, &types.RestType{Type: types.NewInferType("T")}), tuple(str), true},
		{"[string, number, string] to [string, ...infer T]", tuple(str, &types.RestType{Type: types.NewInferType("T")}), tuple(str, num, str), true},
	})
}

func TestIsAssignableToObjects(t *testing.T) {
	empty := object("empty", nil, false)
	typeA := object("a", nil, false, prop("a", &types.StringType{}, true))
	typeB := object("b", nil, false, prop("b", &types.StringType{}, true))
	typeC := object("c", nil, false, prop("c", &types.StringType{}, true))
	typeAB := object("ab", []types.Type{typeA, typeB}, false)
	optionalAB := object("a", nil, false, prop("a", &types.StringType{}, false), prop("b", &types.StringType{}, false))
	optionalB := object("b", nil, false, prop("b", &types.StringType{}, false))
	flatAB := object("ab", nil, false, prop("a", &types.StringType{}, true), prop("b", &types.StringType{}, true))
	aAndB := types.NewIntersectionType([]types.Type{typeA, typeB})
	nonPrimitive := types.NewObjectType("obj", nil, nil, true, true)

	runAssignableCases(t, []assignableCase{
		{"any to {}", empty, &types.AnyType{}, true},
		{"number[] to {}", empty, &types.ArrayType{Item: &types.NumberType{}}, true},
		{"string & null to {}", empty, types.NewIntersectionType([]types.Type{&types.StringType{}, &types.NullType{}}), true},
		{"literal to {}", empty, literal("literal"), true},
		{"never to {}", empty, &types.NeverType{}, true},
		{"null to {}", empty, &types.NullType{}, false},
		{"object to {}", empty, object("obj", nil, true, prop("foo", &types.StringType{}, true)), true},
		{"boolean to {}", empty, &types.BooleanType{}, true},
		{"number to {}", empty, &types.NumberType{}, true},
		{"string to {}", empty, &types.StringType{}, true},
		{"tuple to {}", empty, tuple(&types.StringType{}, &types.NumberType{}), true},
		{"undefined to {}", empty, &types.UndefinedType{}, false},

		{"string to a", typeA, &types.StringType{}, false},
		{"ab to a", typeA, typeAB, true},
		{"ab to b", typeB, typeAB, true},
		{"ab to c", typeC, typeAB, false},
		{"a to ab", typeAB, typeA, false},
		{"b to ab", typeAB, typeB, false},

		{"optional with one in common", optionalB, optionalAB, true},
		{"optional with none in common", optionalB, typeA, false},

		{"string & number to string", &types.StringType{}, types.NewIntersectionType([]types.Type{&types.StringType{}, &types.NumberType{}}), true},
		{"string & number to number", &types.NumberType{}, types.NewIntersectionType([]types.Type{&types.StringType{}, &types.NumberType{}}), true},
		{"string & number to boolean", &types.BooleanType{}, types.NewIntersectionType([]types.Type{&types.StringType{}, &types.NumberType{}}), false},

		{"a & b to a", typeA, aAndB, true},
		{"a & b to b", typeB, aAndB, true},
		{"a & b to c", typeC, aAndB, false},
		{"a & b to flat ab", flatAB, aAndB, true},
		{"a to a & b", aAndB, typeA, false},
		{"b to a & b", aAndB, typeB, false},
		{"c to a & b", aAndB, typeC, false},
		{"flat ab to a & b", aAndB, flatAB, true},
		{"a & b to a & b", aAndB, aAndB, true},

		{"number to {} with additional properties", object("obj", nil, true), &types.NumberType{}, true},
		{"number to object", nonPrimitive, &types.NumberType{}, false},
		{"string to object", nonPrimitive, &types.StringType{}, false},
		{"boolean to object", nonPrimitive, &types.BooleanType{}, false},
	})
}

func TestIsAssignableToFirstMemberOfAName(t *testing.T) {
	// A property shadowing a base property of the same name wins, like
	// Array#find over getObjectProperties upstream.
	base := object("base", nil, false, prop("a", &types.NumberType{}, true))
	derived := object("derived", []types.Type{base}, false, prop("a", &types.StringType{}, true))
	runAssignableCases(t, []assignableCase{
		{"string member to derived", derived, object("s", nil, false, prop("a", &types.StringType{}, true)), true},
		{"number member to derived", derived, object("n", nil, false, prop("a", &types.NumberType{}, true)), false},
		// Every source member is checked, including the shadowed one.
		{"derived to string member", object("s", nil, false, prop("a", &types.StringType{}, true)), derived, false},
	})
}

func TestIsAssignableToCircular(t *testing.T) {
	selfReferencing := func(id, property string) *types.ObjectType {
		ref := types.NewReferenceType()
		obj := object(id, nil, false, prop(property, ref, false))
		ref.SetType(obj)
		return obj
	}
	nodeA := selfReferencing("a", "parent")
	nodeB := selfReferencing("b", "parent")
	nodeC := selfReferencing("c", "child")
	runAssignableCases(t, []assignableCase{
		{"a to a", nodeA, nodeA, true},
		{"b to a", nodeA, nodeB, true},
		{"a to b", nodeB, nodeA, true},
		{"a to c", nodeC, nodeA, false},
		{"b to c", nodeC, nodeB, false},
		{"c to a", nodeA, nodeC, false},
		{"c to b", nodeB, nodeC, false},
	})
}

func TestIsAssignableToDeepUnion(t *testing.T) {
	objectType := object("interface-src/test.ts-0-53-src/test.ts-0-317", nil, false, prop("a", &types.StringType{}, true))
	innerUnion := union(&types.NumberType{}, types.NewDefinitionType("NumericValueRef", objectType))
	alias := types.NewAliasType("alias-src/test.ts-53-106-src/test.ts-0-317", innerUnion)
	outerUnion := union(types.NewDefinitionType("NumberValue", alias), &types.UndefinedType{})
	runAssignableCases(t, []assignableCase{
		{"definition to deep union", outerUnion, types.NewDefinitionType("NumericValueRef", objectType), true},
	})
}

func TestIsAssignableToLiterals(t *testing.T) {
	str := &types.StringType{}
	num := &types.NumberType{}
	boolean := &types.BooleanType{}
	runAssignableCases(t, []assignableCase{
		{`"foo" to string`, str, literal("foo"), true},
		{`"foo" to number`, num, literal("foo"), false},
		{`"foo" to boolean`, boolean, literal("foo"), false},
		{"1 to string", str, literal(1.0), false},
		{"1 to number", num, literal(1.0), true},
		{"1 to boolean", boolean, literal(1.0), false},
		{"true to string", str, literal(true), false},
		{"true to number", num, literal(true), false},
		{"true to boolean", boolean, literal(true), true},

		{`string to "foo"`, literal("foo"), str, false},
		{"number to 1", literal(1.0), num, false},
		{"boolean to true", literal(true), boolean, false},

		{`"bar" to "foo"`, literal("foo"), literal("bar"), false},
		{"2 to 1", literal(1.0), literal(2.0), false},
		{"false to true", literal(true), literal(false), false},

		{`"foo" to "foo"`, literal("foo"), literal("foo"), true},
		{"1 to 1", literal(1.0), literal(1.0), true},
		{"true to true", literal(true), literal(true), true},
	})
}

func TestIsAssignableToInfersTypes(t *testing.T) {
	inferMap := NewInferMap()
	target := tuple(&types.StringType{}, types.NewInferType("T"))
	if !IsAssignableTo(target, tuple(&types.StringType{}, &types.NumberType{}), inferMap) {
		t.Fatal("IsAssignableTo([string, infer T], [string, number]) = false, want true")
	}
	inferred, ok := inferMap.Get("T")
	if !ok {
		t.Fatal("T was not inferred")
	}
	if _, isNumber := inferred.(*types.NumberType); !isNumber {
		t.Errorf("T inferred as %s, want number", inferred.ID())
	}
}
