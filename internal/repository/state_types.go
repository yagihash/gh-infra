package repository

// CurrentState represents the current state of a GitHub repository.
type CurrentState struct {
	Owner         string
	Name          string
	IsNew         bool // true if the repository does not exist yet
	Description   string
	Archived      bool
	Homepage      string
	Visibility    string
	Topics        []string
	Features      CurrentFeatures
	MergeStrategy CurrentMergeStrategy

	BranchProtection map[string]*CurrentBranchProtection // pattern → protection
	Rulesets         map[string]*CurrentRuleset          // name → ruleset
	Secrets          []string                            // names only (values are opaque)
	Variables        map[string]string                   // name → value
}

func (r *CurrentState) FullName() string {
	return r.Owner + "/" + r.Name
}

type CurrentFeatures struct {
	Issues      bool
	Projects    bool
	Wiki        bool
	Discussions bool
}

type CurrentMergeStrategy struct {
	AllowMergeCommit         bool
	AllowSquashMerge         bool
	AllowRebaseMerge         bool
	AutoDeleteHeadBranches   bool
	MergeCommitTitle         string
	MergeCommitMessage       string
	SquashMergeCommitTitle   string
	SquashMergeCommitMessage string
}

type CurrentBranchProtection struct {
	Pattern                 string
	RequiredReviews         int
	DismissStaleReviews     bool
	RequireCodeOwnerReviews bool
	RequireStatusChecks     *CurrentStatusChecks
	EnforceAdmins           bool
	AllowForcePushes        bool
	AllowDeletions          bool
}

type CurrentStatusChecks struct {
	Strict   bool
	Contexts []string
}

// CurrentRuleset represents the current state of a GitHub repository ruleset.
type CurrentRuleset struct {
	ID           int
	Name         string
	Target       string
	Enforcement  string
	BypassActors []CurrentRulesetBypassActor
	Conditions   *CurrentRulesetConditions
	Rules        CurrentRulesetRules
}

type CurrentRulesetBypassActor struct {
	ActorID    int
	ActorType  string
	BypassMode string
}

type CurrentRulesetConditions struct {
	RefName *CurrentRulesetRefCondition
}

type CurrentRulesetRefCondition struct {
	Include []string
	Exclude []string
}

type CurrentRulesetRules struct {
	PullRequest           *CurrentRulesetPullRequest
	RequiredStatusChecks  *CurrentRulesetStatusChecks
	NonFastForward        bool
	Deletion              bool
	Creation              bool
	RequiredLinearHistory bool
	RequiredSignatures    bool
}

type CurrentRulesetPullRequest struct {
	RequiredApprovingReviewCount   int
	DismissStaleReviewsOnPush      bool
	RequireCodeOwnerReview         bool
	RequireLastPushApproval        bool
	RequiredReviewThreadResolution bool
}

type CurrentRulesetStatusChecks struct {
	StrictRequiredStatusChecksPolicy bool
	Contexts                         []CurrentRulesetStatusCheck
}

type CurrentRulesetStatusCheck struct {
	Context       string
	IntegrationID int
}
