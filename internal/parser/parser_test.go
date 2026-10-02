package parser

import "testing"

func TestNodeKeyFileName(t *testing.T) {
	base := baseDir()
	if base == "" {
		t.Skip("working directory is unknown")
	}
	cases := []struct {
		name     string
		filename string
		want     string
	}{
		{"inside the working directory", base + "/src/main.ts", "src_main.ts"},
		{"unrelated absolute path", "/other/dir/main.ts", "_other_dir_main.ts"},
		{"sibling with the working directory as prefix", base + "x/main.ts", nodeKeyFileName(base) + "x_main.ts"},
		{"relative path", "lib/main.ts", "lib_main.ts"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := nodeKeyFileName(c.filename); got != c.want {
				t.Errorf("nodeKeyFileName(%q) = %q, want %q", c.filename, got, c.want)
			}
		})
	}
}
