package repository

import (
	"github.com/babarot/gh-infra/internal/ui"
)

// HasRealChanges returns true if there are any non-noop changes.
func HasRealChanges(changes []Change) bool {
	for _, c := range changes {
		if c.Type != ChangeNoOp {
			return true
		}
	}
	return false
}

// PrintPlanChanges prints the repository change details (without header/footer).
func PrintPlanChanges(changes []Change) {
	grouped := groupByName(changes)
	for _, group := range grouped {
		if len(group.changes) == 0 {
			continue
		}
		if isNewRepo(group.changes) {
			ui.PlanRepoGroupNew(group.name)
		} else {
			ui.PlanRepoGroup(group.name)
		}
		for _, c := range group.changes {
			switch c.Type {
			case ChangeCreate:
				ui.PlanCreate(c.Field, c.NewValue)
			case ChangeUpdate:
				ui.PlanUpdate(c.Field, ui.FormatValue(c.OldValue), ui.FormatValue(c.NewValue))
			case ChangeDelete:
				ui.PlanDelete(c.Field, c.OldValue)
			}
		}
		ui.PlanGroupEnd()
	}
}

// CountChanges returns the number of creates, updates, and deletes.
func CountChanges(changes []Change) (creates, updates, deletes int) {
	return countChanges(changes)
}

// PrintApplyResults prints individual apply result lines (no summary).
func PrintApplyResults(results []ApplyResult) {
	for _, r := range results {
		if r.Err != nil {
			ui.ResultError(r.Change.Name, r.Change.Field, r.Err)
		} else {
			ui.ResultSuccess(r.Change.Name, r.Change.Field, r.Change.Type)
		}
	}
}

// CountApplyResults returns succeeded and failed counts.
func CountApplyResults(results []ApplyResult) (succeeded, failed int) {
	for _, r := range results {
		if r.Err != nil {
			failed++
		} else {
			succeeded++
		}
	}
	return
}

type changeGroup struct {
	name    string
	changes []Change
}

func groupByName(changes []Change) []changeGroup {
	seen := make(map[string]int)
	var groups []changeGroup

	for _, c := range changes {
		if c.Type == ChangeNoOp {
			continue
		}
		idx, ok := seen[c.Name]
		if !ok {
			idx = len(groups)
			seen[c.Name] = idx
			groups = append(groups, changeGroup{name: c.Name})
		}
		groups[idx].changes = append(groups[idx].changes, c)
	}
	return groups
}

func isNewRepo(changes []Change) bool {
	for _, c := range changes {
		if c.Type == ChangeCreate && c.Field == "repository" {
			return true
		}
	}
	return false
}

func countChanges(changes []Change) (creates, updates, deletes int) {
	for _, c := range changes {
		switch c.Type {
		case ChangeCreate:
			creates++
		case ChangeUpdate:
			updates++
		case ChangeDelete:
			deletes++
		}
	}
	return
}
