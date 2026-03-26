---
paths:
  - "docs/adr/**"
---

# ADR Structure

ADRs (Architecture Decision Records) follow this template:

```
# ADR-NNN: {Title}

## Status

Proposed | Accepted | Deprecated | Superseded by [ADR-NNN](NNN-xxx.md)

## Context
## Decision
## Consequences
```

Rules:
- **Filename**: `{NNN}-{kebab-case}.md` (e.g., `001-custom-struct-tag-validator.md`)
- **000-template.md** is reserved for the template itself
- **H1**: Must start with `ADR-NNN:` and the number must match the filename prefix
- **Required H2s**: Status, Context, Decision, Consequences
- **Status value**: Must be one of `Proposed`, `Accepted`, `Deprecated`, or `Superseded by [ADR-NNN](...)`
- Additional H2s (e.g., `## History`) are allowed
- **Language**: English
