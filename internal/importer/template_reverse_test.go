package importer

import "testing"

func TestReverseTemplateContent_SimplePlaceholder(t *testing.T) {
	template := "module github.com/<% .Repo.FullName %>\n\ngo 1.26.0\n"
	remote := "module github.com/hoge/fuga\n\ngo 1.27.0\n"

	got, ok := reverseTemplateContent(template, "hoge/fuga", nil, remote)
	if !ok {
		t.Fatal("expected reverseTemplateContent to succeed")
	}

	want := "module github.com/<% .Repo.FullName %>\n\ngo 1.27.0\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReverseTemplateContent_UnsupportedControlSyntax(t *testing.T) {
	template := "<% if .Repo.Name %>enabled<% end %>\n"
	remote := "enabled\n"

	if _, ok := reverseTemplateContent(template, "org/repo", nil, remote); ok {
		t.Fatal("expected reverseTemplateContent to reject control syntax")
	}
}

func TestReverseTemplateContent_ConsecutivePlaceholders(t *testing.T) {
	template := "<% .Repo.Owner %><% .Repo.Name %>\n"
	remote := "babarotgh-infra\n"

	got, ok := reverseTemplateContent(template, "babarot/gh-infra", nil, remote)
	if !ok {
		t.Fatal("expected reverseTemplateContent to support consecutive placeholders")
	}
	if got != template {
		t.Fatalf("got %q, want %q", got, template)
	}
}

func TestReverseTemplateContent_ChangedVarsPlaceholderRejected(t *testing.T) {
	template := "GO_VERSION=<% .Vars.go_version %>\n"
	remote := "GO_VERSION=1.27.3\n"

	if _, ok := reverseTemplateContent(template, "org/repo", map[string]string{"go_version": "1.26.1"}, remote); ok {
		t.Fatal("expected reverseTemplateContent to reject changed .Vars placeholders")
	}
}
