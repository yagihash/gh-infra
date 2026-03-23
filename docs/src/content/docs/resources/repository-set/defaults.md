---
title: Defaults & Overrides
---

The `defaults` block defines shared settings applied to all repositories in the set. Individual repositories can override any default value.

```yaml
defaults:
  spec:
    visibility: public
    features:
      wiki: false
    merge_strategy:
      allow_squash_merge: true

repositories:
  - name: my-repo
    spec:
      features:
        wiki: true    # overrides the default (false → true)
```

## Merge Behavior

- **Scalar values** (strings, booleans, numbers): the per-repo value replaces the default.
- **Lists** (topics, branch_protection): the per-repo list replaces the default list entirely — lists are not merged.
- **Maps** (features, merge_strategy): per-repo keys are merged into the defaults. Only specified keys are overridden; unspecified keys retain the default value.
