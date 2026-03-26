# ADR-001: Custom struct tag validator over go-playground/validator

## Status

Accepted

## Context

Manifest validation logic was scattered across three locations:

1. **types.go** -- struct field definitions (yaml tags only)
2. **validation.go** -- hand-written `Validate()` methods with required/oneof/duplicate checks
3. **types.go UnmarshalYAML** -- deprecated field migration (3 nearly-identical if/else blocks)

Adding a new field with validation or deprecation required touching 2-3 files. The question was whether to adopt [go-playground/validator](https://github.com/go-playground/validator) (the de facto standard for Go struct tag validation) or build a lightweight custom validator to achieve co-location.

## Decision

Built a custom `tagvalidator.go` (~250 lines) instead of adopting go-playground/validator.

### Evaluation of go-playground/validator

We cloned the library and read the actual source code (`baked_in.go`, `validator.go`, `errors.go`). Findings:

| Feature | go-playground/validator | Impact |
|---------|----------------------|--------|
| `*string` + `omitempty,oneof` | Works correctly (nil = skip, non-nil = validate) | Initial concern was unfounded |
| `unique=Name` on `[]Struct` | Works, but uses **Go field names only** (`FieldByName(param)`) | Cannot use yaml field names in tags |
| `excluded_with` (mutual exclusion) | Works, but Go field names in error messages | yaml name inconsistency |
| Error messages | Default: `Key: 'Repository.Metadata.Name' Error:Field validation for 'Name' failed on the 'required' tag` | Requires translation layer (~40 lines) to produce `metadata.name is required` |
| `unique` error detail | Returns pass/fail only -- **cannot report which value was duplicated** | Cannot produce `duplicate name "dup" in spec.rulesets` |
| `dive` (auto array traversal) | Eliminates manual for-loops over slice elements | Only clear advantage |
| `deprecated` field migration | Not supported | Must keep custom code regardless |
| Struct-level validation | `RegisterStructValidation` -- works but registered separately from struct | Breaks co-location goal |
| yaml names in error paths | `RegisterTagNameFunc` works for error paths, but NOT for tag parameters (`unique=Name`, `excluded_with=Source`) | Split convention: yaml names in errors, Go names in tags |

### Why custom

1. **`unique` cannot report the duplicated value** -- The library's `isUnique` checks membership in a `seen` map and returns `false` immediately, with no way to surface which value collided. Our custom `checkUnique` returns the exact value: `duplicate name "dup" in spec.rulesets`.

2. **yaml name consistency** -- We use yaml field names everywhere: tag parameters (`unique=name`, `exclusive=source`, `deprecated:"via:..."`), error messages, and structural paths. go-playground/validator forces Go field names in tag parameters while supporting yaml names only in error paths via `RegisterTagNameFunc`.

3. **Three-layer complexity** -- Adopting go-playground/validator would not eliminate custom code. The result would be: library config (~40 lines) + custom validators (unique reimplementation, bypass_actors) + custom deprecated handler. Three layers instead of one.

4. **Error message control** -- The translation function is not complex (~30 lines), but it is an indirection layer that must be maintained in sync with every new validation tag.

5. **`deprecated` tag** -- No library supports this. `MigrateDeprecated` would remain custom regardless. Co-locating it with validation in the same tag system keeps things unified.

### What the custom validator provides

```go
// All rules are declared on the struct field -- single source of truth
type FileEntry struct {
    Path      string `yaml:"path"      validate:"required"`
    Content   string `yaml:"content"   validate:"exclusive=source"`
    Source    string  `yaml:"source"`
    Reconcile string `yaml:"reconcile" validate:"omitempty,oneof=patch mirror create_only"`

    DeprecatedSyncMode string `yaml:"sync_mode" deprecated:"reconcile:use \"reconcile\" instead"`
}
```

Supported tags:
- `validate:"required"` -- non-empty string, non-nil slice/pointer
- `validate:"oneof=a b c"` -- enum check
- `validate:"omitempty,..."` -- skip if zero/nil
- `validate:"unique=name"` -- slice elements have unique values (reports which value is duplicated)
- `validate:"exclusive=source"` -- mutual exclusion with another field
- `deprecated:"target:message"` -- migrate value to target field, collect warning

Recursive struct walking builds full yaml paths automatically (`metadata.name`, `spec.merge_strategy.squash_merge_commit_title`).

### What remains hand-written in validation.go

- Slice element iteration with contextual prefix (`my-repo: spec.rulesets[0]`)
- bypass_actors exactly-one-of check (5 fields, mixed types)
- Default value assignment (`FileSet.Spec.Via = ViaPush`)
- `commit_strategy`/`on_apply` conflict (two deprecated aliases for the same target)

## Consequences

- Adding a new validated field requires only a struct tag -- no changes to validation.go unless cross-field logic is needed
- Adding a new deprecated field requires only a struct tag + the standard UnmarshalYAML boilerplate
- The ~250-line tagvalidator.go must be maintained in-house, but it has no external dependencies and is covered by unit tests
- If go-playground/validator adds duplicate-value reporting and yaml-name tag parameters in the future, this decision could be revisited
