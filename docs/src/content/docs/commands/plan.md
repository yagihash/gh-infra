---
title: plan
---

Show diff between YAML and current GitHub state. No mutations are made.

```bash
gh infra plan [path]
```

## Flags

| Flag | Description |
|------|-------------|
| `-r, --repo <owner/repo>` | Target a specific repository |
| `--ci` | Exit with code 1 if changes detected (useful for CI drift detection) |

## Examples

```bash
# Plan all YAML files in a directory
gh infra plan ./repos/

# Plan a single file
gh infra plan ./repos/gomi.yaml

# Plan a specific repository
gh infra plan ./repos/ --repo babarot/gomi

# CI drift detection
gh infra plan ./repos/ --ci
```
