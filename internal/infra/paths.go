package infra

import (
	"os"
	"path/filepath"
	"strings"
)

// tildePath returns path with the user's home directory prefix replaced by "~".
// When the home directory cannot be determined or path is not under it, the
// original path is returned unchanged. Callers should use this only for
// display; filesystem operations must use the original path.
func tildePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	if abs == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(abs, prefix) {
		return "~" + string(filepath.Separator) + abs[len(prefix):]
	}
	return path
}
