// Package testutil holds scaffolding shared by the test packages.
package testutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot determine caller file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file))), nil
}

// RepoRoot returns the absolute path of the repository root.
func RepoRoot(t testing.TB) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// ChdirRepoRoot switches the working directory to the repository root and
// exits the process if it cannot. Call it from TestMain: node-key hashes
// embed cwd-relative file names (process.cwd() in the TypeScript
// implementation), so the golden schemas only reproduce from the root.
func ChdirRepoRoot() {
	root, err := repoRoot()
	if err == nil {
		err = os.Chdir(root)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot chdir to repo root:", err)
		os.Exit(1)
	}
}
