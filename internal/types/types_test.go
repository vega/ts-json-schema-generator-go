package types

import (
	"math"
	"testing"
)

func TestNumberToString(t *testing.T) {
	cases := map[float64]string{
		0:                  "0",
		1:                  "1",
		-1:                 "-1",
		1.5:                "1.5",
		1e21:               "1e+21",
		1e-7:               "1e-7",
		0.000001:           "0.000001",
		123456789012345680: "123456789012345680",
	}
	for in, want := range cases {
		if got := NumberToString(in); got != want {
			t.Errorf("NumberToString(%v) = %q, want %q", in, got, want)
		}
	}
	// String(-0) is "0", so -0 and 0 share a literal type ID.
	if got := NumberToString(math.Copysign(0, -1)); got != "0" {
		t.Errorf("NumberToString(-0) = %q, want 0", got)
	}
}

func TestStableStringify(t *testing.T) {
	got := StableStringify(map[string]any{"b": 1.0, "a": []any{"x", true, nil}})
	want := `{"a":["x",true,null],"b":1}`
	if got != want {
		t.Errorf("StableStringify = %q, want %q", got, want)
	}
	// Non-finite numbers stringify as null, like safe-stable-stringify.
	got = StableStringify([]any{math.Inf(1), math.Inf(-1), math.NaN()})
	if want := `[null,null,null]`; got != want {
		t.Errorf("StableStringify(non-finite) = %q, want %q", got, want)
	}
}

func TestHash(t *testing.T) {
	// Short strings pass through; long ones use Java's String.hashCode.
	if got := Hash("short"); got != "short" {
		t.Errorf("Hash(short) = %q", got)
	}
	// "this is a longer string!" hashed with the JS reference
	// implementation from src/Utils/nodeKey.ts.
	if got := Hash("this is a longer string!"); got != "1848086156" {
		t.Errorf("Hash(long) = %q, want 1848086156", got)
	}
	if got := Hash(42.0); got != "42" {
		t.Errorf("Hash(42) = %q", got)
	}
	// String#length counts UTF-16 code units: 8 CJK characters are 24 bytes
	// but 8 units, so they pass through.
	const cjk = "七転八倒四字熟語"
	if got := Hash(cjk); got != cjk {
		t.Errorf("Hash(cjk) = %q, want %q", got, cjk)
	}
	// This string hashes to exactly -2^31, where Math.abs gives 2147483648.
	if got := Hash("@aaajaaa5aaa=aaaaaa9"); got != "2147483648" {
		t.Errorf("Hash(min int32) = %q, want 2147483648", got)
	}
}

func TestStripQuotes(t *testing.T) {
	cases := map[string]string{
		`"quoted"`: "quoted",
		`'quoted'`: "quoted",
		`unquoted`: "unquoted",
		`"mis'ed"`: `"mis'ed"`,
		`""`:       "",
	}
	for in, want := range cases {
		if got := StripQuotes(in); got != want {
			t.Errorf("StripQuotes(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnionFlattening(t *testing.T) {
	u := NewUnionType([]Type{
		&StringType{},
		NewUnionType([]Type{&NumberType{}, &StringType{}}),
		&NeverType{},
		&HiddenType{},
	})
	if got := u.ID(); got != "(string|number)" {
		t.Errorf("union ID = %q, want (string|number)", got)
	}
}
