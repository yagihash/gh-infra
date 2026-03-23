---
title: apply
---

Apply changes to GitHub. By default, requires interactive confirmation before proceeding.

```bash
gh infra apply [path]
```

## Flags

| Flag | Description |
|------|-------------|
| `-r, --repo <owner/repo>` | Target a specific repository |
| `--auto-approve` | Skip confirmation prompt |
| `--force-secrets` | Re-set all secrets (even existing ones) |

## Examples

```bash
# Apply all changes
gh infra apply ./repos/

# Apply without confirmation (for CI)
gh infra apply ./repos/ --auto-approve

# Force re-set secrets
gh infra apply ./repos/ --force-secrets

# Apply to a specific repository
gh infra apply ./repos/ --repo babarot/gomi
```
