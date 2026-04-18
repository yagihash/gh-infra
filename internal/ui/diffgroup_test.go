package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderDiffGroups_BareItems(t *testing.T) {
	var buf bytes.Buffer
	p := NewStandardPrinterWith(&buf, &buf)

	groups := []DiffGroup{
		{Items: []DiffItem{{Icon: IconChange, Field: "description", Old: `""`, New: "My repo"}}},
		{Items: []DiffItem{{Icon: IconChange, Field: "visibility", Old: "private", New: "public"}}},
	}
	p.SetColumnWidth(DiffGroupFieldWidth(groups))
	RenderDiffGroups(p, groups)

	out := buf.String()
	if !strings.Contains(out, "description") || !strings.Contains(out, "visibility") {
		t.Errorf("expected bare items, got:\n%s", out)
	}
}

func TestRenderDiffGroups_GroupWithHeader(t *testing.T) {
	var buf bytes.Buffer
	p := NewStandardPrinterWith(&buf, &buf)

	groups := []DiffGroup{
		{
			Header: "features",
			Icon:   IconChange,
			Items: []DiffItem{
				{Icon: IconChange, Field: "issues", Old: "false", New: "true"},
				{Icon: IconChange, Field: "wiki", Old: "true", New: "false"},
			},
		},
	}
	p.SetColumnWidth(DiffGroupFieldWidth(groups))
	RenderDiffGroups(p, groups)

	out := buf.String()
	if !strings.Contains(out, "features") {
		t.Errorf("expected 'features' header, got:\n%s", out)
	}
	if !strings.Contains(out, "issues") || !strings.Contains(out, "wiki") {
		t.Errorf("expected child items, got:\n%s", out)
	}
}

func TestRenderDiffGroups_Mixed(t *testing.T) {
	var buf bytes.Buffer
	p := NewStandardPrinterWith(&buf, &buf)

	groups := []DiffGroup{
		{Items: []DiffItem{{Icon: IconChange, Field: "description", Old: `""`, New: "test"}}},
		{
			Header: "labels",
			Icon:   IconAdd,
			Items: []DiffItem{
				{Icon: IconAdd, Field: "kind/bug", Value: "#d73a4a"},
			},
		},
	}
	p.SetColumnWidth(DiffGroupFieldWidth(groups))
	RenderDiffGroups(p, groups)

	out := buf.String()
	if !strings.Contains(out, "description") {
		t.Errorf("expected bare item")
	}
	if !strings.Contains(out, "labels") {
		t.Errorf("expected labels header")
	}
	if !strings.Contains(out, "kind/bug") {
		t.Errorf("expected label item")
	}
}

func TestRenderDiffGroups_Empty(t *testing.T) {
	var buf bytes.Buffer
	p := NewStandardPrinterWith(&buf, &buf)
	RenderDiffGroups(p, nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty groups")
	}
}

func TestDiffGroupFieldWidth(t *testing.T) {
	groups := []DiffGroup{
		{Items: []DiffItem{{Field: "ab"}, {Field: "abcdef"}}},
		{Header: "x", Items: []DiffItem{{Field: "abc"}}},
	}
	if w := DiffGroupFieldWidth(groups); w != 6 {
		t.Errorf("expected width 6, got %d", w)
	}
}

func TestDiffGroupFieldWidth_Empty(t *testing.T) {
	if w := DiffGroupFieldWidth(nil); w != 0 {
		t.Errorf("expected 0, got %d", w)
	}
}
