---
title: Overview
sidebar:
  order: 0
---

gh-infra manages GitHub infrastructure through three resource kinds:

| Kind | Scope | Description |
|------|-------|-------------|
| [Repository](/reference/repository/) | 1 repo | Settings, features, branch protection, secrets, variables |
| [RepositorySet](/reference/repository-set/) | N repos | Shared defaults across multiple repositories |
| [FileSet](/reference/fileset/) | N repos | Distribute files (CODEOWNERS, LICENSE, workflows, etc.) |

## Common Structure

All resources share the same top-level structure:

```yaml
apiVersion: gh-infra/v1
kind: <Repository | RepositorySet | FileSet>
metadata:
  name: <resource-name>
  owner: <github-owner>

spec:
  # Resource-specific fields
```

## File Organization

You can organize YAML files however you like. gh-infra accepts a file or directory path:

```bash
gh infra plan ./repos/           # All YAML files in the directory
gh infra plan ./repos/gomi.yaml  # A single file
```

Multiple resource kinds can coexist in the same directory. gh-infra processes each file based on its `kind`.
