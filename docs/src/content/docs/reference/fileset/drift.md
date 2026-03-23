---
title: Drift Handling
---

## What is Drift?

Drift occurs when someone manually edits a file that gh-infra manages. For example, a developer pushes a change to `.github/CODEOWNERS` directly, making it different from what your FileSet declares.

On the next `plan`, gh-infra detects this difference. The `on_drift` field controls what happens next.

## Options

```yaml
spec:
  on_drift: warn    # warn (default) | overwrite | skip
```

### `warn` (default)

Shows a warning during `plan` and **skips the file** during `apply`. This is the safest option — it tells you drift exists without overwriting someone's intentional changes.

Use this when you're not sure whether the manual edit was intentional.

### `overwrite`

Shows the diff during `plan` and **overwrites the file** during `apply`, restoring it to the declared state.

Use this when you want strict enforcement — the YAML is the source of truth, and manual edits should be reverted.

### `skip`

Ignores the file entirely — no warning, no action. gh-infra pretends the file doesn't exist.

Use this when you've intentionally allowed a repo to diverge and don't want noise in the plan output.

## Summary

| Value | plan | apply |
|-------|------|-------|
| `warn` | Shows drift warning | Skips the file |
| `overwrite` | Shows diff | Overwrites with declared content |
| `skip` | No output | No action |
