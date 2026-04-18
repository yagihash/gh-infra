package ui

// DiffGroup represents a block of related changes for rendering.
// When Header is empty, items are rendered at IndentItem level (bare fields).
// When Header is set, a SubGroupHeader is printed and items render at IndentSub.
type DiffGroup struct {
	Header string     // e.g. "features", "branch_protection[main]", "labels", or ""
	Icon   string     // aggregate icon for the header (IconAdd, IconChange, IconRemove)
	Items  []DiffItem // individual field changes within this group
}

// DiffItem represents a single field-level change within a DiffGroup.
type DiffItem struct {
	Icon  string // IconAdd, IconChange, IconRemove
	Field string // leaf field name (not full dotted path)
	Value any    // for create/delete
	Old   string // for update (pre-formatted)
	New   string // for update (pre-formatted)
}

// RenderDiffGroups renders a sequence of DiffGroups through the printer.
// Groups with a Header get a SubGroupHeader; bare groups render items directly.
func RenderDiffGroups(p Printer, groups []DiffGroup) {
	for _, g := range groups {
		if g.Header != "" {
			p.SubGroupHeader(g.Icon, g.Header)
			for _, item := range g.Items {
				p.PrintChange(itemToChangeItem(item, IndentSub))
			}
		} else {
			for _, item := range g.Items {
				p.PrintChange(itemToChangeItem(item, IndentItem))
			}
		}
	}
}

// DiffGroupFieldWidth returns the maximum field name width across all items.
func DiffGroupFieldWidth(groups []DiffGroup) int {
	w := 0
	for _, g := range groups {
		for _, item := range g.Items {
			if len(item.Field) > w {
				w = len(item.Field)
			}
		}
	}
	return w
}

func itemToChangeItem(item DiffItem, level IndentLevel) ChangeItem {
	return ChangeItem{
		Icon:  item.Icon,
		Field: item.Field,
		Value: item.Value,
		Old:   item.Old,
		New:   item.New,
		Level: level,
	}
}
