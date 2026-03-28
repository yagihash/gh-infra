package ui

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// refreshModel message handling (unit tests for Update logic)
// ---------------------------------------------------------------------------

// helper: create a refreshModel with given task names and pending counts.
func makeModel(tasks ...struct {
	name    string
	pending int
}) refreshModel {
	var rt []RefreshTask
	for _, t := range tasks {
		rt = append(rt, RefreshTask{Name: t.name, Pending: t.pending})
	}
	return newRefreshModel(rt)
}

func task(name string, pending int) struct {
	name    string
	pending int
} {
	return struct {
		name    string
		pending int
	}{name, pending}
}

// update is a test helper that calls Update, asserts the result type, and returns it.
func update(t *testing.T, m refreshModel, msg any) refreshModel {
	t.Helper()
	result, cmd := m.Update(msg)
	_ = cmd // tea.Cmd is not exercised in unit tests
	rm, ok := result.(refreshModel)
	if !ok {
		t.Fatalf("Update returned %T, want refreshModel", result)
	}
	return rm
}

func TestRefreshModel_DoneSimple(t *testing.T) {
	m := makeModel(task("a", 1))
	m = update(t, m, taskDoneMsg{name: "a"})

	if m.items[0].status != taskDone {
		t.Errorf("status = %v, want taskDone", m.items[0].status)
	}
	if m.remaining != 0 {
		t.Errorf("remaining = %d, want 0", m.remaining)
	}
}

func TestRefreshModel_DoneRefCounted(t *testing.T) {
	m := makeModel(task("a", 2))

	// First Done: should still be running
	m = update(t, m, taskDoneMsg{name: "a"})
	if m.items[0].status != taskRunning {
		t.Errorf("after 1st Done: status = %v, want taskRunning", m.items[0].status)
	}
	if m.remaining != 1 {
		t.Errorf("after 1st Done: remaining = %d, want 1", m.remaining)
	}

	// Second Done: should complete
	m = update(t, m, taskDoneMsg{name: "a"})
	if m.items[0].status != taskDone {
		t.Errorf("after 2nd Done: status = %v, want taskDone", m.items[0].status)
	}
	if m.remaining != 0 {
		t.Errorf("after 2nd Done: remaining = %d, want 0", m.remaining)
	}
}

func TestRefreshModel_ErrorThenDone(t *testing.T) {
	// Regression test: Error on one source, then Done on the other.
	// The tracker must not hang.
	m := makeModel(task("a", 2))

	// Error from first source
	m = update(t, m, taskErrorMsg{name: "a", err: fmt.Errorf("fail")})
	if m.items[0].status != taskError {
		t.Errorf("after Error: status = %v, want taskError", m.items[0].status)
	}
	if m.remaining != 1 {
		t.Errorf("after Error: remaining = %d, want 1", m.remaining)
	}

	// Done from second source: should decrement remaining
	m = update(t, m, taskDoneMsg{name: "a"})
	if m.remaining != 0 {
		t.Errorf("after Done: remaining = %d, want 0", m.remaining)
	}
	// Status stays as taskError (first error wins)
	if m.items[0].status != taskError {
		t.Errorf("after Done: status = %v, want taskError", m.items[0].status)
	}
}

func TestRefreshModel_FailThenDone(t *testing.T) {
	m := makeModel(task("a", 2))

	m = update(t, m, taskFailMsg{name: "a"})
	if m.items[0].status != taskFailed {
		t.Errorf("after Fail: status = %v, want taskFailed", m.items[0].status)
	}
	if m.remaining != 1 {
		t.Errorf("after Fail: remaining = %d, want 1", m.remaining)
	}

	m = update(t, m, taskDoneMsg{name: "a"})
	if m.remaining != 0 {
		t.Errorf("after Done: remaining = %d, want 0", m.remaining)
	}
}

func TestRefreshModel_StatusUpdate(t *testing.T) {
	m := makeModel(task("a", 1))

	m = update(t, m, taskStatusMsg{name: "a", status: "fetching secrets..."})
	if m.items[0].statusText != "fetching secrets..." {
		t.Errorf("statusText = %q, want %q", m.items[0].statusText, "fetching secrets...")
	}

	// Status cleared on Done
	m = update(t, m, taskDoneMsg{name: "a"})
	if m.items[0].statusText != "" {
		t.Errorf("statusText after Done = %q, want empty", m.items[0].statusText)
	}
}

func TestRefreshModel_StatusIgnoredAfterError(t *testing.T) {
	m := makeModel(task("a", 1))

	m = update(t, m, taskErrorMsg{name: "a", err: fmt.Errorf("boom")})
	m = update(t, m, taskStatusMsg{name: "a", status: "should be ignored"})

	// statusText should not be set because the task is no longer running
	if m.items[0].statusText != "" {
		t.Errorf("statusText = %q, want empty (task already errored)", m.items[0].statusText)
	}
}

func TestRefreshModel_MultipleTasks(t *testing.T) {
	m := makeModel(task("a", 1), task("b", 2))

	m = update(t, m, taskDoneMsg{name: "a"})
	if m.remaining != 1 {
		t.Errorf("remaining = %d, want 1", m.remaining)
	}

	m = update(t, m, taskDoneMsg{name: "b"})
	if m.remaining != 1 {
		t.Errorf("remaining = %d, want 1 (b has pending=2)", m.remaining)
	}

	m = update(t, m, taskDoneMsg{name: "b"})
	if m.remaining != 0 {
		t.Errorf("remaining = %d, want 0", m.remaining)
	}
}

func TestRefreshModel_ErrorClearsStatusText(t *testing.T) {
	m := makeModel(task("a", 2))

	m = update(t, m, taskStatusMsg{name: "a", status: "fetching..."})
	if m.items[0].statusText != "fetching..." {
		t.Fatalf("statusText not set")
	}

	m = update(t, m, taskErrorMsg{name: "a", err: fmt.Errorf("403")})
	if m.items[0].statusText != "" {
		t.Errorf("statusText after Error = %q, want empty", m.items[0].statusText)
	}
}

// ---------------------------------------------------------------------------
// NewSpinnerReporterWith
// ---------------------------------------------------------------------------

func TestNewSpinnerReporterWith_TaskMapping(t *testing.T) {
	// Use a nil-program tracker (fallback mode) to avoid starting bubbletea.
	tracker := &RefreshTracker{fallback: true, done: closedChan()}
	reporter := NewSpinnerReporterWith(tracker, []string{"org/repo1", "org/repo2"})

	if !reporter.shared {
		t.Error("expected shared=true")
	}
	if _, ok := reporter.tasks["org/repo1"]; !ok {
		t.Error("missing task for org/repo1")
	}
	if _, ok := reporter.tasks["org/repo2"]; !ok {
		t.Error("missing task for org/repo2")
	}
	if reporter.tasks["org/repo1"].Name != "org/repo1" {
		t.Errorf("task name = %q, want org/repo1", reporter.tasks["org/repo1"].Name)
	}
}

func TestSpinnerReporterWith_WaitIsNoop(t *testing.T) {
	tracker := &RefreshTracker{fallback: true, done: closedChan()}
	reporter := NewSpinnerReporterWith(tracker, []string{"a"})

	// Should not block (shared=true means Wait is a no-op)
	reporter.Wait()
}

// ---------------------------------------------------------------------------
// PendingDefault
// ---------------------------------------------------------------------------

func TestNewRefreshModel_DefaultPending(t *testing.T) {
	m := newRefreshModel([]RefreshTask{{Name: "a"}})
	if m.items[0].pending != 1 {
		t.Errorf("default pending = %d, want 1", m.items[0].pending)
	}
}

func TestNewRefreshModel_ExplicitPending(t *testing.T) {
	m := newRefreshModel([]RefreshTask{{Name: "a", Pending: 3}})
	if m.items[0].pending != 3 {
		t.Errorf("pending = %d, want 3", m.items[0].pending)
	}
}
