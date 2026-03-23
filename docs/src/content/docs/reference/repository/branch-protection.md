---
title: Branch Protection
---

Configure branch protection rules:

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
