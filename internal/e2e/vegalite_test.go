package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/config"
	"github.com/vega/ts-json-schema-generator-go/internal/factory"
	"github.com/vega/ts-json-schema-generator-go/internal/testutil"
)

// TestVegaLite generates the vega-lite schema from the installed
// node_modules/vega-lite sources and compares it with the golden file,
// mirroring test/vega-lite/vega-lite.test.ts.
func TestVegaLite(t *testing.T) {
	root := testutil.RepoRoot(t)
	entry := filepath.Join(root, "node_modules", "vega-lite", "src", "index.ts")
	requireNpmSources(t, entry)

	cfg := config.Default()
	cfg.Path = entry
	cfg.Types = []string{"TopLevelSpec"}
	cfg.EncodeRefs = false
	cfg.SkipTypeCheck = true

	generator, release, err := factory.CreateGenerator(cfg)
	if err != nil {
		t.Fatalf("CreateGenerator: %v", err)
	}
	defer release()

	schema, err := generator.CreateSchema(cfg.Types)
	if err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	generated, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var actual any
	if err := json.Unmarshal(generated, &actual); err != nil {
		t.Fatalf("unmarshal generated: %v", err)
	}

	goldenBytes, err := os.ReadFile(filepath.Join(root, "test", "vega-lite", "schema.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var expected any
	if err := json.Unmarshal(goldenBytes, &expected); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		var diffs []string
		diffJSON("$", expected, actual, &diffs)
		limit := min(len(diffs), 40)
		for _, line := range diffs[:limit] {
			t.Error(line)
		}
		t.Fatalf("generated vega-lite schema differs from test/vega-lite/schema.json (%d differences)", len(diffs))
	}
}
