---
title: Branch Protection (Classic)
---

:::caution
This configures **classic** branch protection rules via the [Branch Protection API](https://docs.github.com/en/rest/branches/branch-protection). GitHub now recommends [Repository Rulesets](../rulesets/) as the successor, which offer richer controls such as enforcement modes (dry-run), granular bypass actors, tag protection, and audit history.

Classic branch protection still works and is fully supported by gh-infra, but for new setups consider using [`rulesets`](../rulesets/) instead.
:::

## Example

```yaml
spec:
  branch_protection:
    - pattern: main
      required_reviews: 1
      dismiss_stale_reviews: true
      require_code_owner_reviews: false
      require_status_checks:
        strict: true
        contexts:
          - "ci / test"
          - "ci / lint"
      enforce_admins: false
      allow_force_pushes: false
      allow_deletions: false

    - pattern: "release/*"
      required_reviews: 2
      allow_force_pushes: false
```

Multiple patterns can be defined. Each entry creates a separate branch protection rule.

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `pattern` | string | Branch name or pattern (e.g., `main`, `release/*`) |
| `required_reviews` | int | Number of required approving reviews |
| `dismiss_stale_reviews` | bool | Dismiss approvals when new commits are pushed |
| `require_code_owner_reviews` | bool | Require review from code owners |
| `require_status_checks.strict` | bool | Require branch to be up to date before merging |
| `require_status_checks.contexts` | list | Required status check names |
| `enforce_admins` | bool | Apply rules to admins too |
| `allow_force_pushes` | bool | Allow force pushes to matching branches |
| `allow_deletions` | bool | Allow deleting matching branches |

## When to Use Classic vs Rulesets

Use **classic `branch_protection`** when:
- Your GitHub plan or GHES version does not support rulesets
- You have existing classic rules and don't need to migrate yet

Use **[`rulesets`](../rulesets/)** when:
- You want `evaluate` mode to dry-run rules before enforcing
- You need bypass controls per role, team, or GitHub App
- You want to protect tags (not just branches)
- You want stacking (multiple rulesets applied to the same branch)
