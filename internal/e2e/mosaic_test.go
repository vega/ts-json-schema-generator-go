package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/config"
	"github.com/vega/ts-json-schema-generator-go/internal/factory"
)

// TestMosaic generates the Mosaic spec schema from the vendored
// @uwdata/mosaic-spec sources (see test/mosaic/README.md) using the same
// invocation mosaic's build uses:
//
//	ts-json-schema-generator -p src/spec/Spec.ts -t Spec \
//	    --no-type-check --no-ref-encode --functions hide
//
// and compares it with the schema published in the npm package.
//
// CSSStyles mirrors the DOM lib's CSSStyleDeclaration, so its property set
// tracks the compiler's lib.dom version: TypeScript 7 knows more CSS
// properties than the TypeScript 5.9 that generated the published schema.
// The comparison therefore requires our CSSStyles properties to be a
// superset of the published ones and everything else to match exactly.
func TestMosaic(t *testing.T) {
	root := repoRoot(t)

	cfg := config.Default()
	cfg.Path = filepath.Join(root, "test", "mosaic", "spec", "Spec.ts")
	cfg.Types = []string{"Spec"}
	cfg.EncodeRefs = false
	cfg.SkipTypeCheck = true
	cfg.Functions = config.FunctionsHide

	generator, release, err := factory.CreateGenerator(cfg)
	if err != nil {
		t.Fatalf("CreateGenerator: %v", err)
	}
	defer release()

	generated, err := generator.CreateSchema(cfg.Types)
	if err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	actual := roundTrip(t, generated)
	goldenBytes, err := os.ReadFile(filepath.Join(root, "test", "mosaic", "schema.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var expected any
	if err := json.Unmarshal(goldenBytes, &expected); err != nil {
		t.Fatalf("parse golden: %v", err)
	}

	actualCSS := extractCSSStyleProperties(t, actual)
	expectedCSS := extractCSSStyleProperties(t, expected)

	for name := range expectedCSS {
		if _, ok := actualCSS[name]; !ok {
			t.Errorf("CSSStyles is missing property %q from the published schema", name)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		var diffs []string
		diffJSON("$", expected, actual, &diffs)
		limit := min(len(diffs), 40)
		for _, line := range diffs[:limit] {
			t.Error(line)
		}
		t.Fatalf("generated mosaic schema differs from test/mosaic/schema.json (%d differences)", len(diffs))
	}
}

func roundTrip(t *testing.T, schema any) any {
	t.Helper()
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

// extractCSSStyleProperties removes and returns definitions.CSSStyles.properties
// so the rest of the schema can be compared exactly.
func extractCSSStyleProperties(t *testing.T, schema any) map[string]any {
	t.Helper()
	root, ok := schema.(map[string]any)
	if !ok {
		t.Fatal("schema is not an object")
	}
	definitions, ok := root["definitions"].(map[string]any)
	if !ok {
		t.Fatal("schema has no definitions")
	}
	styles, ok := definitions["CSSStyles"].(map[string]any)
	if !ok {
		t.Fatal("schema has no CSSStyles definition")
	}
	properties, ok := styles["properties"].(map[string]any)
	if !ok {
		t.Fatal(fmt.Sprint("CSSStyles has no properties: ", styles))
	}
	delete(styles, "properties")
	return properties
}
