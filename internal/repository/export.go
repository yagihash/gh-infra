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

	for _, rs := range r.Rulesets {
		mrs := manifest.Ruleset{
			Name:        rs.Name,
			Target:      manifest.Ptr(rs.Target),
			Enforcement: manifest.Ptr(rs.Enforcement),
			Rules: manifest.RulesetRules{
				NonFastForward:        manifest.Ptr(rs.Rules.NonFastForward),
				Deletion:              manifest.Ptr(rs.Rules.Deletion),
				Creation:              manifest.Ptr(rs.Rules.Creation),
				RequiredLinearHistory: manifest.Ptr(rs.Rules.RequiredLinearHistory),
				RequiredSignatures:    manifest.Ptr(rs.Rules.RequiredSignatures),
			},
		}
		for _, ba := range rs.BypassActors {
			mrs.BypassActors = append(mrs.BypassActors, manifest.RulesetBypassActor{
				ActorID:    ba.ActorID,
				ActorType:  ba.ActorType,
				BypassMode: ba.BypassMode,
			})
		}
		if rs.Conditions != nil && rs.Conditions.RefName != nil {
			mrs.Conditions = &manifest.RulesetConditions{
				RefName: &manifest.RulesetRefCondition{
					Include: rs.Conditions.RefName.Include,
					Exclude: rs.Conditions.RefName.Exclude,
				},
			}
		}
		if rs.Rules.PullRequest != nil {
			mrs.Rules.PullRequest = &manifest.RulesetPullRequest{
				RequiredApprovingReviewCount:   manifest.Ptr(rs.Rules.PullRequest.RequiredApprovingReviewCount),
				DismissStaleReviewsOnPush:      manifest.Ptr(rs.Rules.PullRequest.DismissStaleReviewsOnPush),
				RequireCodeOwnerReview:         manifest.Ptr(rs.Rules.PullRequest.RequireCodeOwnerReview),
				RequireLastPushApproval:        manifest.Ptr(rs.Rules.PullRequest.RequireLastPushApproval),
				RequiredReviewThreadResolution: manifest.Ptr(rs.Rules.PullRequest.RequiredReviewThreadResolution),
			}
		}
		if rs.Rules.RequiredStatusChecks != nil {
			sc := &manifest.RulesetStatusChecks{
				StrictRequiredStatusChecksPolicy: manifest.Ptr(rs.Rules.RequiredStatusChecks.StrictRequiredStatusChecksPolicy),
			}
			for _, c := range rs.Rules.RequiredStatusChecks.Contexts {
				check := manifest.RulesetStatusCheck{Context: c.Context}
				if c.IntegrationID != 0 {
					check.IntegrationID = manifest.Ptr(c.IntegrationID)
				}
				sc.Contexts = append(sc.Contexts, check)
			}
			mrs.Rules.RequiredStatusChecks = sc
		}
		repo.Spec.Rulesets = append(repo.Spec.Rulesets, mrs)
	}

	for name, value := range r.Variables {
		repo.Spec.Variables = append(repo.Spec.Variables, manifest.Variable{
			Name:  name,
			Value: value,
		})
	}

	return repo
}
