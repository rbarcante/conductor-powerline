package hook

import (
	"strings"
	"testing"
)

func TestParseValid(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"workspace": "/Users/dev/my-project",
		"context": {"session_id": "abc123"}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Model != "claude-opus-4-6" {
		t.Errorf("expected model 'claude-opus-4-6', got %q", data.Model)
	}
	if data.Workspace != "/Users/dev/my-project" {
		t.Errorf("expected workspace '/Users/dev/my-project', got %q", data.Workspace)
	}
}

func TestParseEmpty(t *testing.T) {
	data, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("empty stdin should not return error, got: %v", err)
	}
	if data.Model != "" {
		t.Errorf("expected empty model, got %q", data.Model)
	}
	if data.Workspace != "" {
		t.Errorf("expected empty workspace, got %q", data.Workspace)
	}
}

func TestParseMalformed(t *testing.T) {
	_, err := Parse(strings.NewReader("{not valid json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseMissingFields(t *testing.T) {
	input := `{"model": "claude-sonnet-4-6"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Model != "claude-sonnet-4-6" {
		t.Errorf("expected model 'claude-sonnet-4-6', got %q", data.Model)
	}
	if data.Workspace != "" {
		t.Errorf("expected empty workspace, got %q", data.Workspace)
	}
}

func TestParseExtraFields(t *testing.T) {
	input := `{"model": "claude-haiku-4-5", "unknown_field": "value", "workspace": "/tmp"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Model != "claude-haiku-4-5" {
		t.Errorf("expected model 'claude-haiku-4-5', got %q", data.Model)
	}
	if data.Workspace != "/tmp" {
		t.Errorf("expected workspace '/tmp', got %q", data.Workspace)
	}
}
