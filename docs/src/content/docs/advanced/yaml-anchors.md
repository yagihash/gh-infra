---
title: YAML Anchors
---

Use YAML anchors to keep file content DRY within a single file:

```yaml
_templates:
  codeowners: &codeowners |
    * @babarot

  license: &license |
    MIT License
    Copyright (c) 2025 babarot

spec:
  files:
    - path: .github/CODEOWNERS
      content: *codeowners
    - path: LICENSE
      content: *license
```

:::note
YAML anchors work within a single file only. They cannot cross file boundaries — this is a YAML spec limitation. If you need to share content across files, use `source` to reference a local file instead.
:::
