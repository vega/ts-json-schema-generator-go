package schema

import "testing"

func TestEncodeRef(t *testing.T) {
	// Matches encodeURIComponent: A-Za-z0-9 -_.!~*'() unescaped.
	cases := map[string]string{
		"Simple":        "Simple",
		"Generic<A,B>":  "Generic%3CA%2CB%3E",
		"with space":    "with%20space",
		"quote'paren()": "quote'paren()",
		"ümlaut":        "%C3%BCmlaut",
	}
	for in, want := range cases {
		if got := EncodeRef(in); got != want {
			t.Errorf("EncodeRef(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDefinitionMarshal(t *testing.T) {
	props := NewProperties()
	props.Set("b", &Definition{Type: "string"})
	props.Set("a", &Definition{Type: "number"})
	def := &Definition{
		Type:       "object",
		Properties: props,
		Required:   []string{"b"},
	}
	def.SetExtra("description", "d")

	got, err := MarshalStable(def, false, true)
	if err != nil {
		t.Fatal(err)
	}
	// Insertion order preserved without sortProps.
	want := `{"type":"object","properties":{"b":{"type":"string"},"a":{"type":"number"}},"required":["b"],"description":"d"}` + "\n"
	if string(got) != want {
		t.Errorf("unsorted marshal:\n got %s\nwant %s", got, want)
	}

	got, err = MarshalStable(def, true, true)
	if err != nil {
		t.Fatal(err)
	}
	// Alphabetical everywhere with sortProps.
	want = `{"description":"d","properties":{"a":{"type":"number"},"b":{"type":"string"}},"required":["b"],"type":"object"}` + "\n"
	if string(got) != want {
		t.Errorf("sorted marshal:\n got %s\nwant %s", got, want)
	}
}

func TestDefinitionMarshalEmptyProperties(t *testing.T) {
	// The allOf reducer distinguishes "no properties key" from
	// "properties: {}" (src/Utils/allOfDefinition.ts always sets it).
	withEmpty := &Definition{Type: "object", Properties: NewProperties()}
	got, err := MarshalStable(withEmpty, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"type":"object","properties":{}}`+"\n" {
		t.Errorf("empty properties should be emitted: %s", got)
	}

	without := &Definition{Type: "object"}
	got, err = MarshalStable(without, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"type":"object"}`+"\n" {
		t.Errorf("absent properties should be omitted: %s", got)
	}
}
