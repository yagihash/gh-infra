package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/babarot/gh-infra/internal/fileset"
	"github.com/babarot/gh-infra/internal/gh"
	"github.com/babarot/gh-infra/internal/manifest"
)

// DiffFiles computes file-level import changes for all matched FileSets.
// It fetches current content from GitHub and compares against local content.
func DiffFiles(ctx context.Context, runner gh.Runner, fileSets []*manifest.FileDocument, filterRepo string, sourceRefCount map[string]int) ([]Change, error) {
	var changes []Change

	for _, doc := range fileSets {
		fs := doc.Resource
		for repoIdx, repo := range fs.Spec.Repositories {
			fullName := fs.Metadata.Owner + "/" + repo.Name
			if filterRepo != "" && fullName != filterRepo {
				continue
			}

			repoCount := len(fs.Spec.Repositories)

			// Resolve files with overrides
			files := fileset.ResolveFiles(fs, repo)

			for _, file := range files {
				shared := file.OriginalSource != "" && sourceRefCount[file.OriginalSource] > 1
				change := planImportEntry(ctx, runner, fullName, file, doc, repoIdx, repo, repoCount, shared)
				changes = append(changes, change)
			}
		}
	}

	return changes, nil
}

// buildSourceRefCount counts how many FileEntries across all FileSets reference
// each local source file (by OriginalSource path). This is used to detect shared
// templates that should not be overwritten during import.
func buildSourceRefCount(fileSets []*manifest.FileDocument) map[string]int {
	counts := make(map[string]int)
	for _, doc := range fileSets {
		for _, file := range doc.Files {
			if file.OriginalSource != "" {
				counts[file.OriginalSource]++
			}
		}
	}
	return counts
}

// planImportEntry determines the WriteMode and computes the diff for a single file entry.
// shared indicates the source template is referenced by multiple file entries.
func planImportEntry(ctx context.Context, runner gh.Runner, fullName string, file manifest.FileEntry, doc *manifest.FileDocument, repoIdx int, repo manifest.FileSetRepository, repoCount int, shared bool) Change {
	change := Change{
		Target: fullName,
		Path:   file.Path,
		Type:   fileset.ChangeNoOp,
	}

	// Populate local path info upfront so it's available even for skipped entries.
	if file.OriginalSource != "" && !strings.HasPrefix(file.Source, "github://") {
		change.LocalTarget = file.OriginalSource
	} else if file.Source == "" || !strings.HasPrefix(file.Source, "github://") {
		change.ManifestPath = doc.SourcePath
		change.DocIndex = doc.DocIndex
		baseIdx := findFileIndex(doc.Resource.Spec.Files, file.Path)
		if baseIdx >= 0 {
			change.YAMLPath = fmt.Sprintf("$.spec.files[%d].content", baseIdx)
		}
	}

	// Templates: skip (reverse transformation is impossible)
	if fileset.HasTemplate(file.Content, file.Vars) {
		change.WriteMode = WriteSkip
		change.Reason = "uses templates"
		return change
	}

	// Patches: generate/update patches instead of skipping
	if len(file.Patches) > 0 {
		return planPatchEntry(ctx, runner, fullName, file, doc, repoIdx, repo, repoCount)
	}

	// Determine write mode based on source
	// (create_only is allowed — importing updates the local master template)
	if file.OriginalSource != "" {
		if strings.HasPrefix(file.Source, "github://") {
			change.WriteMode = WriteSkip
			change.Reason = "remote source (github://)"
			return change
		}
		if shared || len(file.Patches) > 0 {
			// Shared template or already has patches: generate patches to avoid
			// overwriting the template and affecting other repositories.
			return planPatchEntry(ctx, runner, fullName, file, doc, repoIdx, repo, repoCount)
		}
		change.WriteMode = WriteSource
	} else if file.Source != "" && strings.HasPrefix(file.Source, "github://") {
		change.WriteMode = WriteSkip
		change.Reason = "remote source (github://)"
		return change
	} else {
		change.WriteMode = WriteInline
	}

	// Fetch current content from GitHub
	githubContent, err := fetchFileContent(ctx, runner, fullName, file.Path)
	if err != nil {
		// File doesn't exist on GitHub — nothing to import
		change.Type = fileset.ChangeNoOp
		return change
	}

	change.Desired = githubContent

	// Compare with local content
	localContent := file.Content
	change.Current = localContent

	currentNorm := strings.TrimRight(localContent, "\n")
	desiredNorm := strings.TrimRight(githubContent, "\n")

	if currentNorm == desiredNorm {
		change.Type = fileset.ChangeNoOp
		return change
	}

	change.Type = fileset.ChangeUpdate
	return change
}

// planPatchEntry handles files with patches: generates a new patch from template vs GitHub content.
func planPatchEntry(ctx context.Context, runner gh.Runner, fullName string, file manifest.FileEntry, doc *manifest.FileDocument, repoIdx int, repo manifest.FileSetRepository, repoCount int) Change {
	change := Change{
		Target:       fullName,
		Path:         file.Path,
		Type:         fileset.ChangeNoOp,
		ManifestPath: doc.SourcePath,
		DocIndex:     doc.DocIndex,
		WriteMode:    WritePatch,
	}

	// Resolve YAMLPath: check if this file comes from an override or the base spec.
	// We target the entire FileEntry (not just .patches) so we can add the patches
	// field even when it doesn't exist yet.
	overrideIdx := findFileIndex(repo.Overrides, file.Path)
	if overrideIdx >= 0 {
		change.YAMLPath = fmt.Sprintf("$.spec.repositories[%d].overrides[%d]", repoIdx, overrideIdx)
	} else {
		baseIdx := findFileIndex(doc.Resource.Spec.Files, file.Path)
		if baseIdx < 0 {
			// File comes from directory expansion — cannot write patches back
			change.WriteMode = WriteSkip
			change.Reason = "expanded from directory source"
			return change
		}
		change.YAMLPath = fmt.Sprintf("$.spec.files[%d]", baseIdx)
	}

	// Shared manifest warning
	if repoCount > 1 && overrideIdx < 0 {
		change.Warnings = append(change.Warnings,
			fmt.Sprintf("shared manifest: affects %d repositories", repoCount))
	}

	// Compute current patched content (template + existing patches)
	patchedContent, err := fileset.ApplyPatches(fileset.EnsureTrailingNewline(file.Content), file.Patches)
	if err != nil {
		change.WriteMode = WriteSkip
		change.Reason = fmt.Sprintf("cannot apply existing patches: %v", err)
		return change
	}

	// Fetch current content from GitHub
	githubContent, err := fetchFileContent(ctx, runner, fullName, file.Path)
	if err != nil {
		change.Type = fileset.ChangeNoOp
		return change
	}

	change.Current = patchedContent
	change.Desired = githubContent

	// If GitHub matches the already-patched content, nothing to do
	if strings.TrimRight(patchedContent, "\n") == strings.TrimRight(githubContent, "\n") {
		change.Type = fileset.ChangeNoOp
		return change
	}

	// Keep a copy of the original entry for write-back (needed to reconstruct with patches).
	entryCopy := file
	change.PatchEntry = &entryCopy

	// If GitHub matches the raw template, patches should be removed
	if strings.TrimRight(file.Content, "\n") == strings.TrimRight(githubContent, "\n") {
		change.Type = fileset.ChangeUpdate
		change.PatchContent = "" // sentinel: remove patches
		return change
	}

	// Generate new patch: template → GitHub content
	patch, err := fileset.GeneratePatch(file.Content, githubContent, file.Path)
	if err != nil {
		change.WriteMode = WriteSkip
		change.Reason = fmt.Sprintf("patch generation failed: %v", err)
		return change
	}

	change.Type = fileset.ChangeUpdate
	change.PatchContent = patch
	return change
}

// findFileIndex returns the index of the first FileEntry with the given path, or -1 if not found.
func findFileIndex(files []manifest.FileEntry, path string) int {
	for i, f := range files {
		if f.Path == path {
			return i
		}
	}
	return -1
}

// fetchFileContent fetches a file's content from GitHub via the Contents API.
func fetchFileContent(ctx context.Context, runner gh.Runner, repo, path string) (string, error) {
	if runner == nil {
		return "", fmt.Errorf("no runner available")
	}
	out, err := runner.Run(ctx, "api", fmt.Sprintf("repos/%s/contents/%s", repo, path))
	if err != nil {
		return "", err
	}

	var raw struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return "", err
	}

	content := raw.Content
	if raw.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
		if err != nil {
			return "", fmt.Errorf("decode base64 for %s: %w", path, err)
		}
		content = string(decoded)
	}

	return content, nil
}
