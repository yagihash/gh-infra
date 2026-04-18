package manifest

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestPullRequests_UnmarshalYAML_Bool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"pull_requests: true", true},
		{"pull_requests: false", false},
	}
	for _, tt := range tests {
		var out struct {
			PR *PullRequests `yaml:"pull_requests"`
		}
		if err := yaml.Unmarshal([]byte(tt.input), &out); err != nil {
			t.Fatalf("unmarshal %q: %v", tt.input, err)
		}
		if out.PR == nil {
			t.Fatalf("expected PullRequests, got nil for %q", tt.input)
		}
		if out.PR.Enabled == nil || *out.PR.Enabled != tt.want {
			t.Errorf("Enabled = %v, want %v", out.PR.Enabled, tt.want)
		}
		if out.PR.Creation != nil {
			t.Errorf("Creation should be nil for bool form, got %v", *out.PR.Creation)
		}
	}
}

func TestPullRequests_UnmarshalYAML_Object(t *testing.T) {
	input := `
pull_requests:
  creation: collaborators_only
`
	var out struct {
		PR *PullRequests `yaml:"pull_requests"`
	}
	if err := yaml.Unmarshal([]byte(input), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.PR == nil {
		t.Fatal("expected PullRequests, got nil")
	}
	if out.PR.Enabled == nil || !*out.PR.Enabled {
		t.Error("Enabled should be implicitly true for object form")
	}
	if out.PR.Creation == nil || *out.PR.Creation != PullRequestCreationCollaboratorsOnly {
		t.Errorf("Creation = %v, want collaborators_only", out.PR.Creation)
	}
}

func TestPullRequests_MarshalYAML_BoolTrue(t *testing.T) {
	pr := PullRequests{Enabled: Ptr(true)}
	data, err := yaml.Marshal(&pr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := strings.TrimSpace(string(data))
	if s != "true" {
		t.Errorf("expected bare bool 'true', got %q", s)
	}
}

func TestPullRequests_MarshalYAML_BoolFalse(t *testing.T) {
	pr := PullRequests{Enabled: Ptr(false)}
	data, err := yaml.Marshal(&pr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := strings.TrimSpace(string(data))
	if s != "false" {
		t.Errorf("expected bare bool 'false', got %q", s)
	}
}

func TestPullRequests_MarshalYAML_ObjectForm(t *testing.T) {
	pr := PullRequests{
		Enabled:  Ptr(true),
		Creation: Ptr(PullRequestCreationCollaboratorsOnly),
	}
	data, err := yaml.Marshal(&pr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "enabled") {
		t.Errorf("should not contain 'enabled' in output: %q", s)
	}
	if !strings.Contains(s, "creation: collaborators_only") {
		t.Errorf("expected 'creation: collaborators_only' in %q", s)
	}
}

func TestPullRequests_MarshalRoundTrip(t *testing.T) {
	original := Features{
		Issues: Ptr(true),
		PullRequests: &PullRequests{
			Enabled:  Ptr(true),
			Creation: Ptr(PullRequestCreationCollaboratorsOnly),
		},
	}
	data, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Features
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PullRequests == nil {
		t.Fatal("PullRequests nil after round-trip")
	}
	if got.PullRequests.Enabled == nil || !*got.PullRequests.Enabled {
		t.Error("Enabled should be true after round-trip")
	}
	if got.PullRequests.Creation == nil || *got.PullRequests.Creation != PullRequestCreationCollaboratorsOnly {
		t.Errorf("Creation = %v, want collaborators_only", got.PullRequests.Creation)
	}
}
