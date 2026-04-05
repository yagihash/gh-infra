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

// planImportEntry determines the suggested write mode, selectable actions, and
// diff contents for a single file entry. shared indicates the source template
// is referenced by multiple file entries.
func planImportEntry(ctx context.Context, runner gh.Runner, fullName string, file manifest.FileEntry, doc *manifest.FileDocument, repoIdx int, repo manifest.FileSetRepository, repoCount int, shared bool) Change {
	change := Change{
		Target: fullName,
		Path:   file.Path,
		Type:   fileset.ChangeNoOp,
	}

	sourceBacked := file.OriginalSource != "" && !strings.HasPrefix(file.Source, "github://")
	inlineBacked := file.OriginalSource == "" && (file.Source == "" || !strings.HasPrefix(file.Source, "github://"))

	// Populate write target info upfront so it's available even for skipped entries.
	if sourceBacked {
		change.LocalTarget = file.OriginalSource
	} else if inlineBacked {
		change.ManifestPath = doc.SourcePath
		change.DocIndex = doc.DocIndex
		overrideIdx := findFileIndex(repo.Overrides, file.Path)
		if overrideIdx >= 0 {
			change.YAMLPath = fmt.Sprintf("$.spec.repositories[%d].overrides[%d].content", repoIdx, overrideIdx)
		} else {
			baseIdx := findFileIndex(doc.Resource.Spec.Files, file.Path)
			if baseIdx >= 0 {
				change.YAMLPath = fmt.Sprintf("$.spec.files[%d].content", baseIdx)
			}
		}
	}

	// Templates: skip (reverse transformation is impossible)
	if fileset.HasTemplate(file.Content, file.Vars) {
		setActionMetadata(&change, WriteSkip, ActionSkip)
		change.Reason = "uses templates"
		return change
	}

	if file.Source != "" && strings.HasPrefix(file.Source, "github://") {
		setActionMetadata(&change, WriteSkip, ActionSkip)
		change.Reason = "remote source (github://)"
		return change
	}

	patchSupported := false
	if len(file.Patches) > 0 || sourceBacked {
		patchSupported = configurePatchTarget(&change, file, doc, repoIdx, repo, repoCount)
	}
	if len(file.Patches) > 0 && !patchSupported {
		setActionMetadata(&change, WriteSkip, ActionSkip)
		change.Reason = "expanded from directory source"
		return change
	}
	patchPreferred := len(file.Patches) > 0 || shared
	allowedActions := []ImportAction{ActionWrite, ActionSkip}
	suggestedMode := WriteInline
	if sourceBacked {
		suggestedMode = WriteSource
		if patchSupported {
			allowedActions = []ImportAction{ActionWrite, ActionPatch, ActionSkip}
			if patchPreferred {
				suggestedMode = WritePatch
			}
		}
	} else if patchSupported {
		allowedActions = []ImportAction{ActionWrite, ActionPatch, ActionSkip}
		suggestedMode = WritePatch
	}
	setActionMetadata(&change, suggestedMode, allowedActions...)

	// Fetch current content from GitHub
	githubContent, err := fetchFileContent(ctx, runner, fullName, file.Path)
	if err != nil {
		// File doesn't exist on GitHub — nothing to import
		change.Type = fileset.ChangeNoOp
		return change
	}

	change.Desired = githubContent

	change.WriteCurrent = file.Content
	if change.SelectedAction == ActionWrite {
		change.Current = change.WriteCurrent
	}

	if len(file.Patches) > 0 {
		patchedContent, err := fileset.ApplyPatches(fileset.EnsureTrailingNewline(file.Content), file.Patches)
		if err != nil {
			setActionMetadata(&change, WriteSkip, ActionSkip)
			change.Type = fileset.ChangeNoOp
			change.Reason = fmt.Sprintf("cannot apply existing patches: %v", err)
			return change
		}
		change.PatchCurrent = patchedContent
		if change.SelectedAction == ActionPatch {
			change.Current = change.PatchCurrent
		}
	} else if patchSupported {
		change.PatchCurrent = file.Content
		if change.SelectedAction == ActionPatch {
			change.Current = change.PatchCurrent
		}
	}

	change.UpdateTypeForAction()
	if change.Type == fileset.ChangeNoOp {
		return change
	}

	if patchSupported {
		if strings.TrimRight(change.WriteCurrent, "\n") == strings.TrimRight(githubContent, "\n") {
			change.PatchContent = ""
		} else {
			patch, err := fileset.GeneratePatch(file.Content, githubContent, file.Path)
			if err != nil {
				setActionMetadata(&change, WriteSkip, ActionSkip)
				change.Type = fileset.ChangeNoOp
				change.Reason = fmt.Sprintf("patch generation failed: %v", err)
				return change
			}
			change.PatchContent = patch
		}
	}

	return change
}

func configurePatchTarget(change *Change, file manifest.FileEntry, doc *manifest.FileDocument, repoIdx int, repo manifest.FileSetRepository, repoCount int) bool {
	overrideIdx := findFileIndex(repo.Overrides, file.Path)
	if overrideIdx >= 0 {
		change.ManifestPath = doc.SourcePath
		change.DocIndex = doc.DocIndex
		change.PatchYAMLPath = fmt.Sprintf("$.spec.repositories[%d].overrides[%d]", repoIdx, overrideIdx)
		if file.OriginalSource == "" {
			change.YAMLPath = change.PatchYAMLPath + ".content"
		}
	} else {
		baseIdx := findFileIndex(doc.Resource.Spec.Files, file.Path)
		if baseIdx < 0 {
			return false
		}
		change.ManifestPath = doc.SourcePath
		change.DocIndex = doc.DocIndex
		change.PatchYAMLPath = fmt.Sprintf("$.spec.files[%d]", baseIdx)
		if file.OriginalSource == "" {
			change.YAMLPath = change.PatchYAMLPath + ".content"
		}
		if repoCount > 1 {
			change.Warnings = append(change.Warnings,
				fmt.Sprintf("shared manifest: affects %d repositories", repoCount))
		}
	}

	entryCopy := file
	change.PatchEntry = &entryCopy
	return true
}

func setActionMetadata(change *Change, suggested WriteMode, allowed ...ImportAction) {
	change.WriteMode = suggested
	change.SuggestedWriteMode = suggested
	change.AllowedActions = append([]ImportAction(nil), allowed...)
	change.SelectedAction = DefaultAction(suggested)
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
