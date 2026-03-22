package repository

import (
	"github.com/babarot/gh-infra/internal/manifest"
)

// ToManifest converts current state to a manifest Repository for export (import command).
func ToManifest(r *CurrentState) *manifest.Repository {
	repo := &manifest.Repository{
		APIVersion: manifest.APIVersion,
		Kind:       manifest.KindRepository,
		Metadata: manifest.RepositoryMetadata{
			Name:  r.Name,
			Owner: r.Owner,
		},
		Spec: manifest.RepositorySpec{
			Description: manifest.Ptr(r.Description),
			Visibility:  manifest.Ptr(r.Visibility),
			Archived:    manifest.Ptr(r.Archived),
			Topics:      r.Topics,
			Features: &manifest.Features{
				Issues:      manifest.Ptr(r.Features.Issues),
				Projects:    manifest.Ptr(r.Features.Projects),
				Wiki:        manifest.Ptr(r.Features.Wiki),
				Discussions: manifest.Ptr(r.Features.Discussions),
			},
			MergeStrategy: &manifest.MergeStrategy{
				AllowMergeCommit:         manifest.Ptr(r.MergeStrategy.AllowMergeCommit),
				AllowSquashMerge:         manifest.Ptr(r.MergeStrategy.AllowSquashMerge),
				AllowRebaseMerge:         manifest.Ptr(r.MergeStrategy.AllowRebaseMerge),
				AutoDeleteHeadBranches:   manifest.Ptr(r.MergeStrategy.AutoDeleteHeadBranches),
				MergeCommitTitle:         manifest.Ptr(r.MergeStrategy.MergeCommitTitle),
				MergeCommitMessage:       manifest.Ptr(r.MergeStrategy.MergeCommitMessage),
				SquashMergeCommitTitle:   manifest.Ptr(r.MergeStrategy.SquashMergeCommitTitle),
				SquashMergeCommitMessage: manifest.Ptr(r.MergeStrategy.SquashMergeCommitMessage),
			},
		},
	}

	if r.Homepage != "" {
		repo.Spec.Homepage = manifest.Ptr(r.Homepage)
	}

	for _, bp := range r.BranchProtection {
		mbp := manifest.BranchProtection{
			Pattern:                 bp.Pattern,
			RequiredReviews:         manifest.Ptr(bp.RequiredReviews),
			DismissStaleReviews:     manifest.Ptr(bp.DismissStaleReviews),
			RequireCodeOwnerReviews: manifest.Ptr(bp.RequireCodeOwnerReviews),
			EnforceAdmins:           manifest.Ptr(bp.EnforceAdmins),
			AllowForcePushes:        manifest.Ptr(bp.AllowForcePushes),
			AllowDeletions:          manifest.Ptr(bp.AllowDeletions),
		}
		if bp.RequireStatusChecks != nil {
			mbp.RequireStatusChecks = &manifest.StatusChecks{
				Strict:   bp.RequireStatusChecks.Strict,
				Contexts: bp.RequireStatusChecks.Contexts,
			}
		}
		repo.Spec.BranchProtection = append(repo.Spec.BranchProtection, mbp)
	}

	for name, value := range r.Variables {
		repo.Spec.Variables = append(repo.Spec.Variables, manifest.Variable{
			Name:  name,
			Value: value,
		})
	}

	return repo
}
