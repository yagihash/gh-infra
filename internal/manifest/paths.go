package manifest

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolvePaths normalises the given paths to absolute form and checks for
// duplicates or containment (e.g. "." and "./repos/"). It returns an error
// when overlapping paths are detected.
func ResolvePaths(args []string) ([]string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}

	resolved := make([]string, 0, len(args))
	for _, a := range args {
		abs, err := filepath.Abs(a)
		if err != nil {
			return nil, fmt.Errorf("resolve path %s: %w", a, err)
		}
		resolved = append(resolved, abs)
	}

	// Dedup exact matches.
	seen := make(map[string]bool, len(resolved))
	unique := make([]string, 0, len(resolved))
	for _, p := range resolved {
		if seen[p] {
			continue
		}
		seen[p] = true
		unique = append(unique, p)
	}

	// Check containment: if path A is a prefix of path B, they overlap.
	for i := 0; i < len(unique); i++ {
		for j := i + 1; j < len(unique); j++ {
			if isSubpath(unique[i], unique[j]) || isSubpath(unique[j], unique[i]) {
				return nil, fmt.Errorf("overlapping paths: %s and %s", unique[i], unique[j])
			}
		}
	}

	return unique, nil
}

// isSubpath reports whether child is under parent (parent is a prefix of child).
func isSubpath(parent, child string) bool {
	if parent == child {
		return false // exact match is handled by dedup
	}
	prefix := parent
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(child, prefix)
}
