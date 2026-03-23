---
title: Basic Settings
---

```yaml
spec:
  description: "My awesome project"
  homepage: "https://example.com"
  visibility: public          # public | private | internal
  archived: false

  topics:
    - go
    - cli
    - github
```

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Repository description |
| `homepage` | string | URL displayed on the repo page |
| `visibility` | string | `public`, `private`, or `internal` (GitHub Enterprise) |
| `archived` | bool | `true` to archive (read-only). Reversible — set to `false` to unarchive |
| `topics` | list | GitHub topics for discoverability |
