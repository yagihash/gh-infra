---
title: Features & Merge Strategy
---

## Features

Toggle repository features on or off:

```yaml
spec:
  features:
    issues: true
    projects: false
    wiki: false
    discussions: false
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `issues` | bool | `true` | Enable Issues tab |
| `projects` | bool | `true` | Enable Projects tab |
| `wiki` | bool | `true` | Enable Wiki tab |
| `discussions` | bool | `false` | Enable Discussions tab |

## Merge Strategy

Control how pull requests can be merged:

```yaml
spec:
  merge_strategy:
    allow_merge_commit: false
    allow_squash_merge: true
    allow_rebase_merge: false
    auto_delete_head_branches: true
    squash_merge_commit_title: PR_TITLE        # PR_TITLE | COMMIT_OR_PR_TITLE
    squash_merge_commit_message: COMMIT_MESSAGES # COMMIT_MESSAGES | PR_BODY | BLANK
    merge_commit_title: MERGE_MESSAGE           # MERGE_MESSAGE | PR_TITLE
    merge_commit_message: PR_TITLE              # PR_TITLE | PR_BODY | BLANK
```

| Field | Type | Description |
|-------|------|-------------|
| `allow_merge_commit` | bool | Allow merge commits |
| `allow_squash_merge` | bool | Allow squash merging |
| `allow_rebase_merge` | bool | Allow rebase merging |
| `auto_delete_head_branches` | bool | Automatically delete head branches after merge |
| `squash_merge_commit_title` | string | Title format for squash merges |
| `squash_merge_commit_message` | string | Message format for squash merges |
| `merge_commit_title` | string | Title format for merge commits |
| `merge_commit_message` | string | Message format for merge commits |
