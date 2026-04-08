# CLAUDE.md

## Package Responsibilities

| Package | Role |
|---|---|
| `cmd` | Cobra CLI command definitions and flag handling |
| `internal/infra` | Orchestrator — drives plan, apply, import workflows |
| `internal/fileset` | File-set resolution, templating, patching |
| `internal/importer` | Import logic: match existing repos, reverse-template, write YAML |
| `internal/repository` | Per-repository state, diff, plan, apply, export |
| `internal/manifest` | YAML manifest parsing, validation, type definitions (shared data model) |
| `internal/gh` | GitHub CLI (`gh`) runner abstraction |
| `internal/ui` | Terminal UI: printer, diff viewer, progress, confirm |
| `internal/parallel` | Generic concurrent map utility |
| `internal/logger` | slog-based structured logger |
| `internal/yamledit` | In-place YAML node editor (preserves formatting) |

## Package Dependency Rules

The internal packages form a clean DAG (no circular dependencies). Respect the following layering:

```
Layer 0 (leaves):  logger, parallel, yamledit   — no internal imports allowed
Layer 1:           gh (→ logger), ui (→ logger)
Layer 2:           manifest (→ gh)
Layer 3:           fileset (→ manifest, gh, parallel), repository (→ manifest, gh, logger, parallel)
Layer 4:           importer (→ fileset, repository, manifest, gh, yamledit)
Layer 5:           infra (→ orchestrator, may import all above + ui)
Layer 6:           cmd (→ infra, ui, manifest, gh, logger, fileset)
```

Key constraints:
- `internal/ui` must NOT depend on domain packages (manifest, repository, infra, etc.) — it is a pure presentation layer
- `internal/manifest` must NOT depend on upper layers — it is the shared data-model layer
- `internal/infra` is the orchestrator; other internal packages must NOT depend on it (except cmd)
- Leaf packages (logger, parallel, yamledit) must NOT import any internal package
- `cmd` should access domain logic through `infra`, not by importing lower layers directly (fileset is an exception for CLI flag types)
