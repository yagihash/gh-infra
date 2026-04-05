package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/babarot/gh-infra/internal/fileset"
	"github.com/babarot/gh-infra/internal/yamledit"
)

// Write writes the planned changes to disk.
// It applies manifest edits (repo spec patches) and file changes.
func Write(plan *Result) error {
	// Apply manifest edits (repo/reposet YAML patches).
	for path, data := range plan.ManifestEdits {
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write manifest %s: %w", path, err)
		}
	}

	// Apply file changes.
	for _, c := range plan.FileChanges {
		if c.Type != fileset.ChangeUpdate {
			continue
		}

		switch c.WriteMode {
		case WriteSource:
			if err := os.WriteFile(c.LocalTarget, []byte(c.Desired), 0644); err != nil {
				return fmt.Errorf("write source %s: %w", c.LocalTarget, err)
			}

		case WriteInline:
			// Read current manifest (may already be patched by repo edits).
			data, ok := plan.ManifestEdits[c.ManifestPath]
			if !ok {
				var err error
				data, err = os.ReadFile(c.ManifestPath)
				if err != nil {
					return fmt.Errorf("read manifest for inline edit %s: %w", c.ManifestPath, err)
				}
			}
			updated, err := yamledit.ReplaceContent(data, c.DocIndex, c.YAMLPath, c.Desired)
			if err != nil {
				return fmt.Errorf("inline edit %s in %s: %w", c.YAMLPath, c.ManifestPath, err)
			}
			plan.ManifestEdits[c.ManifestPath] = updated
			if err := os.WriteFile(c.ManifestPath, updated, 0644); err != nil {
				return fmt.Errorf("write manifest %s: %w", c.ManifestPath, err)
			}

		case WritePatch:
			if c.PatchEntry == nil {
				return fmt.Errorf("WritePatch for %s but PatchEntry is nil", c.Path)
			}
			// Read current manifest (may already be patched by repo edits).
			data, ok := plan.ManifestEdits[c.ManifestPath]
			if !ok {
				var err error
				data, err = os.ReadFile(c.ManifestPath)
				if err != nil {
					return fmt.Errorf("read manifest for patch edit %s: %w", c.ManifestPath, err)
				}
			}
			// Step 1: Replace the entire FileEntry with a placeholder patches value
			// to ensure the patches field exists in the YAML.
			entry := writePatchEntry(c)
			updated, err := yamledit.ReplaceNode(data, c.DocIndex, c.YAMLPath, entry)
			if err != nil {
				return fmt.Errorf("patch edit %s in %s: %w", c.YAMLPath, c.ManifestPath, err)
			}
			// Step 2: If there's actual patch content, re-write just the first
			// patch string using ReplaceContent so it gets proper literal block
			// formatting with correct indentation.
			if c.PatchContent != "" {
				patchPath := c.YAMLPath + ".patches[0]"
				updated, err = yamledit.ReplaceContent(updated, c.DocIndex, patchPath, c.PatchContent)
				if err != nil {
					return fmt.Errorf("patch content edit %s in %s: %w", patchPath, c.ManifestPath, err)
				}
			}
			plan.ManifestEdits[c.ManifestPath] = updated
			if err := os.WriteFile(c.ManifestPath, updated, 0644); err != nil {
				return fmt.Errorf("write manifest %s: %w", c.ManifestPath, err)
			}

		case WriteSkip:
			// Nothing to do.
		}
	}

	return nil
}

// writePatchEntry builds a YAML-serializable entry for WritePatch.
// It preserves the original fields (path, source, reconcile, etc.)
// and sets or clears the patches field.
func writePatchEntry(c Change) map[string]any {
	e := c.PatchEntry
	entry := map[string]any{
		"path": e.Path,
	}

	// Recover source path: after resolution, Source is cleared but OriginalSource
	// holds the absolute path. Convert back to a relative path from the manifest dir.
	if e.OriginalSource != "" {
		entry["source"] = recoverSourcePath(e.OriginalSource, c.ManifestPath)
	} else if e.Source != "" {
		entry["source"] = e.Source
	}

	// Preserve the original field name (sync_mode vs reconcile)
	if e.DeprecatedSyncMode != "" {
		entry["sync_mode"] = e.DeprecatedSyncMode
	} else if e.Reconcile != "" {
		entry["reconcile"] = e.Reconcile
	}

	if len(e.Vars) > 0 {
		entry["vars"] = e.Vars
	}

	if c.PatchContent != "" {
		entry["patches"] = []string{c.PatchContent}
	}

	return entry
}

// recoverSourcePath converts an absolute OriginalSource path back to a relative
// path from the manifest directory, prefixed with "./" for consistency.
func recoverSourcePath(originalSource, manifestPath string) string {
	manifestDir := filepath.Dir(manifestPath)
	rel, err := filepath.Rel(manifestDir, originalSource)
	if err != nil {
		return originalSource
	}
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}
