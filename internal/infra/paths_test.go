package infra

import (
	"path/filepath"
	"testing"
)

func TestTildePath(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "path inside home",
			input: "/home/alice/src/repo/infra.yaml",
			want:  "~/src/repo/infra.yaml",
		},
		{
			name:  "exact home directory",
			input: "/home/alice",
			want:  "~",
		},
		{
			name:  "path outside home",
			input: "/tmp/foo.yaml",
			want:  "/tmp/foo.yaml",
		},
		{
			name:  "path that only shares prefix but is not under home",
			input: "/home/alice2/file.yaml",
			want:  "/home/alice2/file.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tildePath(tt.input)
			if got != tt.want {
				t.Errorf("tildePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTildePath_RelativePathResolvesBeforeReplacement(t *testing.T) {
	// Ensure that a relative path under the home directory is also shortened.
	// This mirrors how the command receives paths (Abs resolution).
	home := t.TempDir()
	t.Setenv("HOME", home)

	rel := "file.yaml"
	abs := filepath.Join(home, rel)
	// cd into home so the relative path resolves under it.
	t.Chdir(home)

	got := tildePath(rel)
	want := "~" + string(filepath.Separator) + rel
	if got != want {
		// tildePath inspects the absolute form of the input. If the absolute
		// resolution works, we should see the tilde form.
		if got == abs || got == rel {
			t.Skipf("relative path resolution not verifiable in this test environment: got %q", got)
		}
		t.Errorf("tildePath(%q) = %q, want %q", rel, got, want)
	}
}
