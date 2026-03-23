---
title: File
sidebar:
  label: Overview
  order: 0
---

`File` manages **files** in a **single** repository — CODEOWNERS, LICENSE, CI workflows, security policies, and any other files you want to keep in a declared state.

:::tip[Example]
```yaml
apiVersion: gh-infra/v1
kind: File
metadata:
  owner: babarot
  name: gomi

spec:
  files:
    # Inline content
    - path: .github/CODEOWNERS
      content: |
        * @babarot

    # Inline content with templating (<% %> expanded per repo)
    - path: go.mod
      content: |
        module github.com/<% .Repo.FullName %>
        go 1.24.0

    # From local file
    - path: LICENSE
      source: ./templates/LICENSE

    # From GitHub repository
    - path: .github/workflows/ci.yml
      source: github://babarot/shared-config/workflows/ci.yml

  on_drift: warn                          # warn | overwrite | skip
  strategy: direct                        # direct | pull_request
  commit_message: "ci: sync managed files"
```
:::

## Metadata

```yaml
metadata:
  owner: babarot    # GitHub owner or organization
  name: gomi        # Repository name
```

The combination of `owner` and `name` identifies the target repository (`babarot/gomi`).

## Spec

| Section | What you can configure |
|---|---|
| [File Sources](./sources/) | Inline content, local files, directories, and `github://` references |
| [Templating](./templating/) | `<% %>` syntax, built-in variables, custom vars |
| [Drift Handling](./drift/) | How to handle manual edits: `warn`, `overwrite`, or `skip` |
| [Apply Strategy](./strategy/) | Commit directly or open a pull request |

## When to Use

Use `File` when you want to manage files in one repo's YAML file. This is the simplest file management resource and the starting point for most users.

It works best when:

- **Each repo has its own distinct files** — different CODEOWNERS, different CI workflows, different licenses.
- **You want per-repo change tracking** — each file maps to one repo, so `git blame` tells you exactly who changed what and when.
- **You're managing files in a small number of repos** — for 1–5 repos, separate files are easy to maintain.

If you find yourself copying the same files across many repos, consider [FileSet](../fileset/) instead.
