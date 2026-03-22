package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Printer handles all user-facing output. All output goes through this
// to ensure consistent styling and stderr/stdout separation.
type Printer struct {
	out io.Writer // stdout (plan results, apply results)
	err io.Writer // stderr (progress, status messages)
}

// DefaultPrinter is the package-level printer.
var DefaultPrinter = &Printer{out: os.Stdout, err: os.Stderr}

// OutWriter returns the stdout writer.
func (p *Printer) OutWriter() io.Writer { return p.out }

// ErrWriter returns the stderr writer.
func (p *Printer) ErrWriter() io.Writer { return p.err }

// SetWriters overrides output destinations (for testing).
func SetWriters(out, err io.Writer) {
	DefaultPrinter.out = out
	DefaultPrinter.err = err
}

// ResetWriters restores default stdout/stderr.
func ResetWriters() {
	DefaultPrinter.out = os.Stdout
	DefaultPrinter.err = os.Stderr
}

// --- Progress messages (stderr) ---

func StartPhase(path string) {
	p := DefaultPrinter
	fmt.Fprintf(p.err, "Reading desired state from %s ...\n", path)
	fmt.Fprintf(p.err, "Fetching current state from GitHub API ...\n\n")
}

func Refreshing(name string) {
	fmt.Fprintf(DefaultPrinter.err, "  Refreshing %s...\n", name)
}

func RefreshingFileSet(fileSet, repo string) {
	fmt.Fprintf(DefaultPrinter.err, "  Refreshing %s → %s...\n", fileSet, repo)
}

func Creating(name, field string) {
	fmt.Fprintf(DefaultPrinter.err, "  Creating %s %s...\n", name, field)
}

func Updating(name, field string) {
	fmt.Fprintf(DefaultPrinter.err, "  Updating %s %s...\n", name, field)
}

func Destroying(name, field string) {
	fmt.Fprintf(DefaultPrinter.err, "  Destroying %s %s...\n", name, field)
}

func Committing(repo string, fileCount int) {
	fmt.Fprintf(DefaultPrinter.err, "  Committing %s (%d files)...\n", repo, fileCount)
}

func Importing(name string) {
	fmt.Fprintf(DefaultPrinter.err, "Importing %s ...\n", name)
}

func SkipManagedBySelf(name string) {
	fmt.Fprintf(DefaultPrinter.err, "  ⚠ %s: managed_by=self, skipping\n", name)
}

func SkipImportError(name string, err error) {
	fmt.Fprintf(DefaultPrinter.err, "  ⚠ skipping %s: %v\n", name, err)
}

// --- Plan output (stdout) ---

func PlanHeader(creates, updates, deletes int) {
	fmt.Fprintf(DefaultPrinter.out, "\nPlan: %d to create, %d to update, %d to destroy\n\n", creates, updates, deletes)
}

func PlanRepoGroup(name string) {
	fmt.Fprintf(DefaultPrinter.out, "  %s %s\n", Yellow.Render("~"), Bold.Render(name))
}

func PlanRepoGroupNew(name string) {
	fmt.Fprintf(DefaultPrinter.out, "  %s %s %s\n",
		Green.Render("+"), Bold.Render(name), Green.Render("(new)"))
}

func PlanFileSetGroup(fileSet, repo string) {
	fmt.Fprintf(DefaultPrinter.out, "  %s FileSet: %s → %s\n",
		Yellow.Render("~"), Bold.Render(fileSet), Bold.Render(repo))
}

func PlanCreate(field string, value any) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s: %s\n",
		Green.Render("+"), field, Green.Render(fmt.Sprintf("%v", value)))
}

func PlanUpdate(field string, oldVal, newVal string) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s: %s → %s\n",
		Yellow.Render("~"), field, Dim.Render(oldVal), Bold.Render(newVal))
}

func PlanDelete(field string, value any) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s: %s\n",
		Red.Render("-"), field, Red.Render(fmt.Sprintf("%v", value)))
}

func PlanFileCreate(path string) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s  %s\n",
		Green.Render("+"), path, Green.Render("(new file)"))
}

func PlanFileUpdate(path string) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s  %s\n",
		Yellow.Render("~"), path, Yellow.Render("(content changed)"))
}

func PlanFileDrift(path, onDrift string) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s  %s on_drift: %s → skipping apply\n",
		Yellow.Render("⚠"), path, Yellow.Render("[drift detected]"), onDrift)
}

func PlanFileSkip(path string) {
	fmt.Fprintf(DefaultPrinter.out, "      %s %s  %s on_drift: skip → ignored\n",
		Dim.Render("-"), path, Dim.Render("[drift detected]"))
}

func PlanGroupEnd() {
	fmt.Fprintln(DefaultPrinter.out)
}

func PlanSeparator() {
	fmt.Fprintln(DefaultPrinter.out, Dim.Render(strings.Repeat("─", 50)))
}

func PlanFooter() {
	fmt.Fprintf(DefaultPrinter.out, "To apply these changes, run: %s\n", Bold.Render("gh infra apply"))
}

// --- Apply results (stdout) ---

func ResultSuccess(name, field string, changeType any) {
	fmt.Fprintf(DefaultPrinter.out, "  %s %s  %s %sd\n",
		Green.Render("✓"), Bold.Render(name), field, changeType)
}

func ResultError(name, field string, err error) {
	msg := strings.ReplaceAll(err.Error(), "\n", "\n    ")
	fmt.Fprintf(DefaultPrinter.out, "  %s %s  %s: %s\n",
		Red.Render("✗"), Bold.Render(name), field, msg)
}

func ResultSkipped(name, path, onDrift string) {
	fmt.Fprintf(DefaultPrinter.out, "  %s %s %s  drift detected, skipped (on_drift: %s)\n",
		Yellow.Render("⚠"), Bold.Render(name), path, onDrift)
}

func ApplySummary(succeeded, failed int) {
	fmt.Fprintf(DefaultPrinter.out, "\nApply complete! %d changes applied", succeeded)
	if failed > 0 {
		fmt.Fprintf(DefaultPrinter.out, ", %d failed", failed)
	}
	fmt.Fprintln(DefaultPrinter.out, ".")
}

// --- Status messages (stdout) ---

func NoChanges() {
	fmt.Fprintln(DefaultPrinter.out, "\nNo changes. Infrastructure is up-to-date.")
}

func NoResources(path string) {
	fmt.Fprintln(DefaultPrinter.out, "No resources found in", path)
}

func ApplyCancelled() {
	fmt.Fprintln(DefaultPrinter.out, "Apply cancelled.")
}

func ConfirmPrompt() string {
	fmt.Fprint(DefaultPrinter.out, "\nDo you want to apply these changes? (yes/no): ")
	return "" // caller reads stdin
}

// --- Validate output (stdout) ---

func ValidateSummary(repos, filesets int) {
	fmt.Fprintf(DefaultPrinter.out, "✓ Valid: %d repositories, %d filesets defined\n", repos, filesets)
}

func ValidateRepo(name string) {
	fmt.Fprintf(DefaultPrinter.out, "  - Repository: %s\n", name)
}

func ValidateFileSet(name string, files, repos int) {
	fmt.Fprintf(DefaultPrinter.out, "  - FileSet: %s (%d files → %d repositories)\n", name, files, repos)
}

// --- Format helpers ---

func FormatValue(v any) string {
	switch val := v.(type) {
	case []string:
		return "[" + strings.Join(val, ", ") + "]"
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
