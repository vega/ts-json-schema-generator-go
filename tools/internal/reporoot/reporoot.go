// Package reporoot locates the repository root for the tools under tools/,
// so they work from any directory inside the repository.
package reporoot

import (
	"fmt"
	"os"
	"path/filepath"
)

// Find returns the nearest directory at or above the working directory that
// contains test/valid-data.
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "test", "valid-data")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot locate repository root (test/valid-data) above the working directory")
		}
		dir = parent
	}
}
