package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vega/ts-json-schema-generator-go/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.ChdirRepoRoot()
	os.Exit(m.Run())
}

// fixturePath returns the glob the e2e tests use for a test/valid-data fixture.
func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testutil.RepoRoot(t), "test", "valid-data", name, "*.ts")
}

// TestOutdirMatchesSingleTypeRuns is the property that makes --outdir worth
// having: one parse reused across types must produce exactly the bytes a
// fresh generator per type produces. The generator carries caches keyed by
// node and context, so reuse could in principle leak state between types.
func TestOutdirMatchesSingleTypeRuns(t *testing.T) {
	// MyObject and MySubObject share a definition (MyObject references
	// MySubObject), so the second generation runs against a warm cache.
	const fixture = "interface-multi"
	both := []string{"MyObject", "MySubObject"}

	cases := []struct {
		name  string
		types []string
		extra []string
	}{
		{name: "defaults", types: both},
		{name: "minify-no-top-ref-id", types: both, extra: []string{"--minify", "--no-top-ref", "--id", "https://example.com/s.json"}},
		// --expose none leaves the root type undecorated, so the top-level
		// definition takes its name from TopRefNodeParser rather than from an
		// ExposeNodeParser definition. That name has to follow the type being
		// generated, not the whole --type list.
		{name: "expose-none", types: both, extra: []string{"--expose", "none"}},
		// A single --type is the other naming branch: the parser chain is
		// built already knowing the name.
		{name: "single-type", types: []string{"MyObject"}},
		{name: "single-type-expose-none", types: []string{"MyObject"}, extra: []string{"--expose", "none"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outdir := filepath.Join(t.TempDir(), "nested", "schemas")

			args := []string{"--path", fixturePath(t, fixture), "--no-type-check", "--outdir", outdir}
			for _, typeName := range tc.types {
				args = append(args, "--type", typeName)
			}
			if err := run(append(args, tc.extra...)); err != nil {
				t.Fatalf("run with --outdir: %v", err)
			}

			for _, typeName := range tc.types {
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
	want := readJSON(t, filepath.Join(testutil.RepoRoot(t), "test", "valid-data", "interface-multi", "schema.json"))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("generated schema does not match the golden fixture\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOutdirValidation(t *testing.T) {
	dir := t.TempDir()
	path := fixturePath(t, "interface-multi")

	// An existing regular file where a directory is wanted.
	existingFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(existingFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		// path overrides --path; empty means the valid fixture glob.
		path    string
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
		{
			name:    "empty outdir value",
			args:    []string{"--outdir", "", "--type", "MyObject"},
			wantErr: "--outdir needs a directory path",
		},
		{
			name:    "outdir is an existing file",
			args:    []string{"--outdir", existingFile, "--type", "MyObject"},
			wantErr: "exists and is not a directory",
		},
		{
			// The whole point of validating up front is that a bad flag
			// combination costs no parse. With an unusable --path, reaching
			// the compiler at all would surface a different error.
			name:    "validation precedes parsing",
			path:    filepath.Join(t.TempDir(), "no", "such", "dir", "*.ts"),
			args:    []string{"--outdir", dir, "--out", filepath.Join(dir, "s.json"), "--type", "MyObject"},
			wantErr: "--out and --outdir are mutually exclusive",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argPath := tc.path
			if argPath == "" {
				argPath = path
			}
			err := run(append([]string{"--path", argPath, "--no-type-check"}, tc.args...))
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

// TestFlagErrorExitCodes checks main's contract for flag handling: -h and
// --help exit 0 after the usage text, and a bad flag exits 2 with its error
// printed exactly once (by the flag package, not again by main).
func TestFlagErrorExitCodes(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantOnce string
	}{
		{"short help", []string{"-h"}, 0, "Usage of ts-json-schema-generator"},
		{"long help", []string{"--help"}, 0, "Usage of ts-json-schema-generator"},
		{"unknown flag", []string{"--bogus"}, 2, "flag provided but not defined: -bogus"},
		{"bad value", []string{"--minify=maybe"}, 2, `invalid boolean value "maybe" for -minify`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var code int
			stderr := captureStderr(t, func() {
				code = exitCode(run(tc.args), os.Stderr)
			})
			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d", code, tc.wantCode)
			}
			if n := strings.Count(stderr, tc.wantOnce); n != 1 {
				t.Errorf("stderr contains %q %d times, want once:\n%s", tc.wantOnce, n, stderr)
			}
			if strings.Contains(stderr, "Error:") {
				t.Errorf("stderr has an extra error line:\n%s", stderr)
			}
		})
	}
}

func TestExitCodeReportsOtherErrors(t *testing.T) {
	var stderr strings.Builder
	if code := exitCode(errors.New("boom"), &stderr); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got := stderr.String(); got != "Error: boom\n" {
		t.Errorf("stderr = %q, want %q", got, "Error: boom\n")
	}
}

// captureStderr returns what fn writes to os.Stderr, which the flag package
// uses when no output is set.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	original := os.Stderr
	os.Stderr = file
	defer func() { os.Stderr = original }()
	fn()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
