---
paths:
  - "cmd/**/*.go"
  - "internal/infra/**/*.go"
  - "internal/importer/**/*.go"
  - "internal/ui/**/*.go"
---

# Command UX Symmetry

gh-infra ships four user-facing commands: `plan`, `apply`, `import`, and `import --into`. A user moving between them should see the same shape of output — the same vocabulary for "reading", "fetching", "spinner progress", "error reporting", and "summary". When adding or modifying one of these commands (or the code that powers their output), preserve this symmetry. Deviating from it silently is how drift, noisy output, and spinner-corruption bugs creep back in.

## Phase skeleton

Every command follows this skeleton (adapt, don't rearrange):

1. **Parse manifests / CLI args.**
2. **`p.Phase("Reading desired state from <path> ...")`** — for commands that read YAML files (`plan`, `apply`, `import --into`). Display the path through `internal/infra.tildePath(...)` so `/Users/<user>/…` becomes `~/…`. Omit this phase for commands that do not read manifests (`import`).
3. **`p.Phase("Fetching current state from GitHub API ...")`** — standard for all commands that call the GitHub API.
4. **`p.BlankLine()`** — separates phases from the spinner region.
5. **`ui.RunRefresh(tasks)`** — start the spinner and hold on to the `*ui.RefreshTracker`.
6. **Per-target work (parallel).** Report per-target progress through the tracker (see below).
7. **`tracker.Wait()`** — drain the spinner.
8. **`tracker.PrintErrors()`** — renders per-target error blocks. See the table below for the one exception.
9. **`p.Separator()`** — before body output.
10. **Body output** — plan table, apply results, YAML dump, or import plan.
11. **Summary / Message** — `p.Summary(...)` on success; `p.Message(...)` for no-op terminal states. Distinguish "no changes" from "no changes because some targets were skipped" (see below).

## Error reporting discipline

### Never write to stderr while the spinner is running

The bubbletea spinner owns the terminal between `ui.RunRefresh` and `tracker.Wait()`. Do not call `p.Warning`, `p.Error`, `p.ErrorMessage`, or `fmt.Fprintln(p.ErrWriter(), ...)` inside that window — the incremental redraws corrupt each other. This was the plan-time "spinner freezes mid-frame" bug that originally motivated this rule.

Route every per-target error through the tracker. The tracker renders a truncated single-line form on the task row during the spinner and buffers the full detail for `tracker.PrintErrors` to render after the spinner finishes.

### `tracker.Error` vs `tracker.Fail`

| Method | When to use |
|---|---|
| `tracker.Error(name, err)` | Default. Per-target failure that should appear inline on the spinner row and again in the post-spinner error report. |
| `tracker.Fail(name)` | Rare. Only when the error is already being surfaced through a different channel (e.g., the function returns the error and cobra prints it, or the command's summary explicitly names the target). Do not reach for `Fail` just to "skip the target". |

### `tracker.PrintErrors()` placement

| Command | Calls `PrintErrors`? | Reason |
|---|---|---|
| `plan` | Yes | The plan body has no room for per-target errors; `PrintErrors` is the only full-text surface. |
| `import` | Yes | The YAML body dumped to stdout says nothing about failed targets. |
| `import --into` | Yes | `printImportPlan` only covers matched, successful diffs. |
| `apply` | No | `printApplyResults` already renders `✗ <field>  <error>` lines per failed change. A leading `PrintErrors` block would duplicate every error. |

When adding a new command, the deciding question is **"does the body output already surface per-target failures in full?"**. If yes, skip `PrintErrors`. If no, call it. Leave a short comment at the `tracker.Wait()` site if you opt out — `internal/infra/apply.go` is the reference.

### "No changes" must distinguish clean no-op from skipped targets

When the body output ends up empty because errors skipped every target, don't pretend everything is up-to-date:

```go
if !result.HasChanges {
    if len(tracker.Errors()) > 0 { // or an equivalent skipped counter
        p.Message("\nNo changes computed. Some repositories were skipped due to errors above.")
    } else {
        p.Message("\nNo changes. Infrastructure is up-to-date.")
    }
    return result, nil
}
```

`import --into` follows the same pattern via `ImportDiff.Skipped`. If you add a new command that can return "no changes", wire an equivalent counter through its result type.

## Path display

Use `internal/infra.tildePath` for any user-facing display of manifest paths (phase messages, summaries, errors). Never expose the raw `/Users/<user>/…` form — it's noisy and leaks the developer's home directory.

## Printer output methods — quick reference

| Method | Destination | Used for |
|---|---|---|
| `p.Phase(msg)` | stderr | Phase announcements ("Reading desired state from ...", "Fetching ...") |
| `p.BlankLine()` | stderr | Spacing before the spinner |
| `p.Separator()` | stdout | Horizontal rule before body output |
| `p.Message(msg)` | stdout | Terminal no-op messages ("No changes. ...") |
| `p.Summary(msg)` | stderr | Final completion line ("Plan: ...", "Apply complete! ...") |
| `p.Warning(name, detail)` | stderr | Single-line yellow warning. Use only outside the spinner window. |
| `p.Error(name, detail)` | stdout | Single-line red error. Rarely used directly — prefer the tracker path. |
| `p.ErrorReport(name, detail)` | stderr | Multi-line per-target error block with soft word-wrap. Invoked by `tracker.PrintErrors`. |

## Checklist when touching these commands

- [ ] Phase messages match the skeleton and the order above.
- [ ] All path displays go through `tildePath`.
- [ ] No stderr/stdout writes happen between `ui.RunRefresh` and `tracker.Wait`.
- [ ] Per-target failures go through `tracker.Error` (or `tracker.Fail` with a documented justification).
- [ ] The `PrintErrors` decision (call it or not) matches the table above.
- [ ] The "no changes" terminal message distinguishes clean no-op from "skipped due to errors".
- [ ] The summary uses `p.Summary` and mirrors the existing commands' phrasing.
