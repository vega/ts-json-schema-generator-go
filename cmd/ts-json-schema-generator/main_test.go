package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// TestMain switches the working directory to the repo root, like internal/e2e:
// node-key hashes embed cwd-relative filenames, so the golden schemas only
// reproduce when generating from the root.
func TestMain(m *testing.M) {
	if _, file, _, ok := runtime.Caller(0); ok {
		if err := os.Chdir(rootOf(file)); err != nil {
			fmt.Fprintln(os.Stderr, "cannot chdir to repo root:", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func rootOf(file string) string {
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine caller file")
	}
	return rootOf(file)
}

// fixturePath returns the glob the e2e tests use for a test/valid-data fixture.
func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "test", "valid-data", name, "*.ts")
}

// TestOutdirMatchesSingleTypeRuns is the property that makes --outdir worth
// having: one parse reused across types must produce exactly the bytes a
// fresh generator per type produces. The generator carries caches keyed by
// node and context, so reuse could in principle leak state between types.
func TestOutdirMatchesSingleTypeRuns(t *testing.T) {
	// MyObject and MySubObject share a definition (MyObject references
	// MySubObject), so the second generation runs against a warm cache.
	const fixture = "interface-multi"
	types := []string{"MyObject", "MySubObject"}

	cases := []struct {
		name  string
		extra []string
	}{
		{name: "defaults"},
		{name: "minify-no-top-ref-id", extra: []string{"--minify", "--no-top-ref", "--id", "https://example.com/s.json"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outdir := filepath.Join(t.TempDir(), "nested", "schemas")

			args := []string{"--path", fixturePath(t, fixture), "--no-type-check", "--outdir", outdir}
			for _, typeName := range types {
				args = append(args, "--type", typeName)
			}
			if err := run(append(args, tc.extra...)); err != nil {
				t.Fatalf("run with --outdir: %v", err)
			}

			for _, typeName := range types {
				reference := filepath.Join(t.TempDir(), "reference.json")
				single := []string{
					"--path", fixturePath(t, fixture), "--no-type-check",
					"--type", typeName, "--out", reference,
				}
				if err := run(append(single, tc.extra...)); err != nil {
					t.Fatalf("run with --out for %s: %v", typeName, err)
				}

				want, err := os.ReadFile(reference)
				if err != nil {
					t.Fatalf("read reference schema for %s: %v", typeName, err)
				}
				got, err := os.ReadFile(filepath.Join(outdir, typeName+".schema.json"))
				if err != nil {
					t.Fatalf("read --outdir schema for %s: %v", typeName, err)
				}
				if string(got) != string(want) {
					t.Errorf("%s.schema.json differs from a fresh single-type run\n got: %s\nwant: %s",
						typeName, truncate(got), truncate(want))
				}
			}
		})
	}
}

// TestOutdirWritesOnlyRequestedFiles guards against the loop writing extra or
// misnamed files.
func TestOutdirWritesOnlyRequestedFiles(t *testing.T) {
	outdir := t.TempDir()
	err := run([]string{
		"--path", fixturePath(t, "interface-multi"), "--no-type-check",
		"--outdir", outdir, "--type", "MyObject", "--type", "MySubObject",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	entries, err := os.ReadDir(outdir)
	if err != nil {
		t.Fatalf("read outdir: %v", err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	want := []string{"MyObject.schema.json", "MySubObject.schema.json"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("outdir contains %v, want %v", names, want)
	}
}

// TestOutdirMatchesGoldenSchema checks the per-type file against the fixture's
// checked-in schema, i.e. that it really is the ordinary single-type schema.
func TestOutdirMatchesGoldenSchema(t *testing.T) {
	outdir := t.TempDir()
	err := run([]string{
		"--path", fixturePath(t, "interface-multi"), "--no-type-check",
		"--outdir", outdir, "--type", "MyObject",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	got := readJSON(t, filepath.Join(outdir, "MyObject.schema.json"))
	want := readJSON(t, filepath.Join(repoRoot(t), "test", "valid-data", "interface-multi", "schema.json"))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("generated schema does not match the golden fixture\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOutdirValidation(t *testing.T) {
	dir := t.TempDir()
	path := fixturePath(t, "interface-multi")

	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "with --out",
			args:    []string{"--outdir", dir, "--out", filepath.Join(dir, "s.json"), "--type", "MyObject"},
			wantErr: "--out and --outdir are mutually exclusive",
		},
		{
			name:    "no types",
			args:    []string{"--outdir", dir},
			wantErr: "--outdir requires at least one --type",
		},
		{
			name:    "star type",
			args:    []string{"--outdir", dir, "--type", "*"},
			wantErr: "--outdir cannot be used with --type '*'",
		},
		{
			name:    "duplicate type",
			args:    []string{"--outdir", dir, "--type", "MyObject", "--type", "MyObject"},
			wantErr: `duplicate --type "MyObject"`,
		},
		{
			name:    "path separator in type name",
			args:    []string{"--outdir", dir, "--type", "some/Type"},
			wantErr: `cannot be used as a file name`,
		},
		{
			name:    "empty type name",
			args:    []string{"--outdir", dir, "--type", ""},
			wantErr: `cannot be used as a file name`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := run(append([]string{"--path", path, "--no-type-check"}, tc.args...))
			if err == nil {
				t.Fatalf("expected an error, got none")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
			if entries, readErr := os.ReadDir(dir); readErr == nil && len(entries) > 0 {
				t.Errorf("validation failure still wrote %d file(s) to the output directory", len(entries))
			}
		})
	}
}

// TestOutdirUnknownTypeFails names the offending type so a typo in a long
// --type list is findable.
func TestOutdirUnknownTypeFails(t *testing.T) {
	outdir := t.TempDir()
	err := run([]string{
		"--path", fixturePath(t, "interface-multi"), "--no-type-check",
		"--outdir", outdir, "--type", "MyObject", "--type", "NoSuchType",
	})
	if err == nil {
		t.Fatal("expected an error for an unknown type")
	}
	if !strings.Contains(err.Error(), "NoSuchType") {
		t.Errorf("error %q does not name the missing type", err.Error())
	}
}

func readJSON(t *testing.T, path string) any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return value
}

func truncate(data []byte) string {
	if len(data) > 200 {
		return string(data[:200]) + "..."
	}
	return string(data)
}
