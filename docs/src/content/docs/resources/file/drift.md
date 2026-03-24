---
title: Drift Handling
sidebar:
  order: 3
---

## What is Drift?

Drift occurs when someone manually edits a file that gh-infra manages. For example, a developer pushes a change to `.github/CODEOWNERS` directly, making it different from what your manifest declares.

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

## Interaction with `sync_mode: mirror`

`on_drift` and `sync_mode: mirror` cannot be used together. If any file entry has `sync_mode: mirror`, specifying `on_drift` explicitly is a validation error:

```yaml
# ✗ Error: on_drift cannot be set when sync_mode "mirror" is used
spec:
  on_drift: warn
  files:
    - path: .github/workflows
      source: ./templates/workflows/
      sync_mode: mirror
```

This is because mirror means "make the directory exactly match the source" — content drift is always resolved by overwriting, which contradicts `warn` or `skip`.

If you omit `on_drift` (let it default), mirror files silently use `overwrite` while non-mirror files use the default `warn`:

```yaml
# ✓ Valid: on_drift not specified, defaults apply
spec:
  files:
    - path: .github/workflows
      source: ./templates/workflows/
      sync_mode: mirror    # always overwrites

    - path: LICENSE
      source: ./templates/LICENSE
      # uses default on_drift: warn
```
