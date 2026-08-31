package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFilePath(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "PLAN.md")
	if err := writeFile(existing, "# M\n"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"no argument defaults to ROADMAP.md", "", "./ROADMAP.md"},
		{"a directory gets ROADMAP.md appended", dir, filepath.Join(dir, "ROADMAP.md")},
		{"an existing file is used as-is", existing, existing},
		{"a nonexistent path is used as-is", "/no/such/file.md", "/no/such/file.md"},
		{"a relative name is used as-is", "OTHER.md", "OTHER.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveFilePath(tt.arg); got != tt.want {
				t.Errorf("resolveFilePath(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

// a trailing separator on a directory must not produce a doubled slash
func TestResolveFilePathDirectoryWithTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, "ROADMAP.md")
	if got := resolveFilePath(dir + string(os.PathSeparator)); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
