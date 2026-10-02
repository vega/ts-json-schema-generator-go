package parser

import "testing"

func TestJsChangeFirstCase(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		mapCase func(string) string
		want    string
	}{
		{"capitalize", "foo", jsToUpperCase, "Foo"},
		{"capitalize keeps the rest", "fOO bar", jsToUpperCase, "FOO bar"},
		{"capitalize an upper-case string", "Foo", jsToUpperCase, "Foo"},
		{"uncapitalize", "FOO", jsToLowerCase, "fOO"},
		{"capitalize a non-letter", "1a", jsToUpperCase, "1a"},
		{"full case mapping", "ßa", jsToUpperCase, "SSa"},
		{"non-ASCII letter", "élan", jsToUpperCase, "Élan"},
		{"astral first character is a lone surrogate", "𝒶b", jsToUpperCase, "𝒶b"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := jsChangeFirstCase(c.input, c.mapCase); got != c.want {
				t.Errorf("jsChangeFirstCase(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestJsChangeFirstCasePanicsOnEmptyString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("jsChangeFirstCase(\"\") did not panic")
		}
	}()
	jsChangeFirstCase("", jsToUpperCase)
}
