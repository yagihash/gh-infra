package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/term"
)

type taskStatus int

const (
	taskRunning taskStatus = iota
	taskDone
	taskError
)

// Braille spinner frames
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type refreshItem struct {
	name   string
	status taskStatus
	errMsg string
}

// RefreshTracker tracks parallel task progress with spinners or plain text.
type RefreshTracker struct {
	mu       sync.Mutex
	items    []refreshItem
	out      *os.File
	done     chan struct{}
	fallback bool
	frame    int
	stopTick chan struct{}
	closeOnce sync.Once
}

// RunRefresh starts a spinner display for the given task names.
// Returns a non-nil tracker in all cases.
func RunRefresh(names []string) *RefreshTracker {
	if len(names) == 0 {
		return &RefreshTracker{fallback: true, done: closedChan()}
	}

	f, ok := DefaultPrinter.err.(*os.File)
	if !ok || !term.IsTerminal(f.Fd()) {
		for _, name := range names {
			Refreshing(name)
		}
		return &RefreshTracker{fallback: true, done: closedChan()}
	}

	items := make([]refreshItem, len(names))
	for i, name := range names {
		items[i] = refreshItem{name: name, status: taskRunning}
	}

	t := &RefreshTracker{
		items:    items,
		out:      f,
		done:     make(chan struct{}),
		stopTick: make(chan struct{}),
	}

	// Print initial lines
	t.render()

	// Start ticker for spinner animation
	go t.tick()

	return t
}

func (t *RefreshTracker) tick() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			t.frame++
			t.mu.Unlock()
			t.redraw()
		case <-t.stopTick:
			return
		}
	}
}

func (t *RefreshTracker) render() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(t.out, t.view())
}

func (t *RefreshTracker) redraw() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if any are still running
	anyRunning := false
	for _, item := range t.items {
		if item.status == taskRunning {
			anyRunning = true
			break
		}
	}

	// Move cursor up N lines and clear each line
	n := len(t.items)
	fmt.Fprintf(t.out, "\x1b[%dA", n) // move up
	fmt.Fprint(t.out, t.view())

	if !anyRunning {
		t.closeOnce.Do(func() {
			close(t.stopTick)
			close(t.done)
		})
	}
}

func (t *RefreshTracker) view() string {
	var b strings.Builder
	frame := spinnerFrames[t.frame%len(spinnerFrames)]
	for _, item := range t.items {
		// Clear line
		b.WriteString("\x1b[2K")
		switch item.status {
		case taskDone:
			fmt.Fprintf(&b, "  %s %s\n", Green.Render("✓"), item.name)
		case taskError:
			fmt.Fprintf(&b, "  %s %s: %s\n", Red.Render("✗"), Bold.Render(item.name), item.errMsg)
		case taskRunning:
			fmt.Fprintf(&b, "  %s %s...\n", Cyan.Render(frame), item.name)
		}
	}
	return b.String()
}

// Done marks a task as successfully completed.
func (t *RefreshTracker) Done(name string) {
	if t.fallback {
		return
	}
	t.mu.Lock()
	for i := range t.items {
		if t.items[i].name == name && t.items[i].status == taskRunning {
			t.items[i].status = taskDone
			break
		}
	}
	t.mu.Unlock()
	t.redraw()
}

// Error marks a task as failed.
func (t *RefreshTracker) Error(name string, err error) {
	if t.fallback {
		RefreshError(name, err)
		return
	}
	t.mu.Lock()
	for i := range t.items {
		if t.items[i].name == name && t.items[i].status == taskRunning {
			t.items[i].status = taskError
			t.items[i].errMsg = err.Error()
			break
		}
	}
	t.mu.Unlock()
	t.redraw()
}

// Wait blocks until all tasks are reported and the display finishes.
func (t *RefreshTracker) Wait() {
	<-t.done
}

func closedChan() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
