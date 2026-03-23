---
title: validate
---

Check YAML syntax and schema without contacting GitHub.

```bash
gh infra validate [path]
```

## Examples

```bash
# Validate all files in a directory
gh infra validate ./repos/

# Validate a single file
gh infra validate ./repos/gomi.yaml
```

Exits with code 0 if all files are valid, non-zero otherwise.
