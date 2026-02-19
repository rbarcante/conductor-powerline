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
	if data.ModelID() != "claude-opus-4-6" {
		t.Errorf("expected model 'claude-opus-4-6', got %q", data.ModelID())
	}
	if data.WorkspacePath() != "/Users/dev/my-project" {
		t.Errorf("expected workspace '/Users/dev/my-project', got %q", data.WorkspacePath())
	}
}

func TestParseEmpty(t *testing.T) {
	data, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("empty stdin should not return error, got: %v", err)
	}
	if data.ModelID() != "" {
		t.Errorf("expected empty model, got %q", data.ModelID())
	}
	if data.WorkspacePath() != "" {
		t.Errorf("expected empty workspace, got %q", data.WorkspacePath())
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
	if data.ModelID() != "claude-sonnet-4-6" {
		t.Errorf("expected model 'claude-sonnet-4-6', got %q", data.ModelID())
	}
	if data.WorkspacePath() != "" {
		t.Errorf("expected empty workspace, got %q", data.WorkspacePath())
	}
}

func TestParseExtraFields(t *testing.T) {
	input := `{"model": "claude-haiku-4-5", "unknown_field": "value", "workspace": "/tmp"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.ModelID() != "claude-haiku-4-5" {
		t.Errorf("expected model 'claude-haiku-4-5', got %q", data.ModelID())
	}
	if data.WorkspacePath() != "/tmp" {
		t.Errorf("expected workspace '/tmp', got %q", data.WorkspacePath())
	}
}

// --- Tests for Claude Code's actual stdin schema (model/workspace as objects) ---

func TestParseModelAsObject(t *testing.T) {
	input := `{
		"model": {"id": "claude-opus-4-6", "display_name": "Claude Opus 4.6 (Thinking)"},
		"workspace": "/Users/dev/my-project"
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.ModelID() != "claude-opus-4-6" {
		t.Errorf("ModelID() = %q, want %q", data.ModelID(), "claude-opus-4-6")
	}
	if data.ModelDisplayName() != "Claude Opus 4.6 (Thinking)" {
		t.Errorf("ModelDisplayName() = %q, want %q", data.ModelDisplayName(), "Claude Opus 4.6 (Thinking)")
	}
}

func TestParseWorkspaceAsObject(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"workspace": {"current_dir": "/Users/dev/my-project/src", "project_dir": "/Users/dev/my-project"}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.WorkspacePath() != "/Users/dev/my-project" {
		t.Errorf("WorkspacePath() = %q, want %q", data.WorkspacePath(), "/Users/dev/my-project")
	}
}

func TestParseClaudeCodeFullSchema(t *testing.T) {
	input := `{
		"model": {"id": "claude-sonnet-4-6", "display_name": "Claude Sonnet 4.6"},
		"workspace": {"current_dir": "/Users/dev/project/src", "project_dir": "/Users/dev/project"},
		"cwd": "/Users/dev/project",
		"session_id": "abc-123",
		"version": "1.0.30"
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.ModelID() != "claude-sonnet-4-6" {
		t.Errorf("ModelID() = %q, want %q", data.ModelID(), "claude-sonnet-4-6")
	}
	if data.ModelDisplayName() != "Claude Sonnet 4.6" {
		t.Errorf("ModelDisplayName() = %q, want %q", data.ModelDisplayName(), "Claude Sonnet 4.6")
	}
	if data.WorkspacePath() != "/Users/dev/project" {
		t.Errorf("WorkspacePath() = %q, want %q", data.WorkspacePath(), "/Users/dev/project")
	}
}

func TestParseModelAsStringAccessors(t *testing.T) {
	input := `{"model": "claude-opus-4-6", "workspace": "/Users/dev/project"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When model is a string, ModelID() should return it, ModelDisplayName() empty
	if data.ModelID() != "claude-opus-4-6" {
		t.Errorf("ModelID() = %q, want %q", data.ModelID(), "claude-opus-4-6")
	}
	if data.ModelDisplayName() != "" {
		t.Errorf("ModelDisplayName() = %q, want empty", data.ModelDisplayName())
	}
	// When workspace is a string, WorkspacePath() should return it
	if data.WorkspacePath() != "/Users/dev/project" {
		t.Errorf("WorkspacePath() = %q, want %q", data.WorkspacePath(), "/Users/dev/project")
	}
}

// --- Tests for context_window parsing ---

func TestContextWindowParsing(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {
				"input_tokens": 5000,
				"cache_creation_input_tokens": 2000,
				"cache_read_input_tokens": 3000
			},
			"context_window_size": 200000
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cw := data.ContextWindow()
	if cw == nil {
		t.Fatal("expected non-nil ContextWindow")
	}
	if cw.CurrentUsage.InputTokens != 5000 {
		t.Errorf("InputTokens = %d, want 5000", cw.CurrentUsage.InputTokens)
	}
	if cw.CurrentUsage.CacheCreationInputTokens != 2000 {
		t.Errorf("CacheCreationInputTokens = %d, want 2000", cw.CurrentUsage.CacheCreationInputTokens)
	}
	if cw.CurrentUsage.CacheReadInputTokens != 3000 {
		t.Errorf("CacheReadInputTokens = %d, want 3000", cw.CurrentUsage.CacheReadInputTokens)
	}
	if cw.ContextWindowSize != 200000 {
		t.Errorf("ContextWindowSize = %d, want 200000", cw.ContextWindowSize)
	}
}

func TestContextWindowMissing(t *testing.T) {
	input := `{"model": "claude-opus-4-6"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.ContextWindow() != nil {
		t.Error("expected nil ContextWindow when field is absent")
	}
}

func TestContextWindowZeroSize(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {"input_tokens": 100},
			"context_window_size": 0
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cw := data.ContextWindow()
	if cw == nil {
		t.Fatal("expected non-nil ContextWindow even with zero size")
	}
	if cw.ContextWindowSize != 0 {
		t.Errorf("ContextWindowSize = %d, want 0", cw.ContextWindowSize)
	}
}

func TestContextWindowPartialUsage(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {"input_tokens": 5000},
			"context_window_size": 200000
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cw := data.ContextWindow()
	if cw == nil {
		t.Fatal("expected non-nil ContextWindow")
	}
	if cw.CurrentUsage.InputTokens != 5000 {
		t.Errorf("InputTokens = %d, want 5000", cw.CurrentUsage.InputTokens)
	}
	if cw.CurrentUsage.CacheCreationInputTokens != 0 {
		t.Errorf("CacheCreationInputTokens = %d, want 0", cw.CurrentUsage.CacheCreationInputTokens)
	}
	if cw.CurrentUsage.CacheReadInputTokens != 0 {
		t.Errorf("CacheReadInputTokens = %d, want 0", cw.CurrentUsage.CacheReadInputTokens)
	}
}

// --- Tests for ContextPercent() ---

func TestContextPercent(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {
				"input_tokens": 5000,
				"cache_creation_input_tokens": 2000,
				"cache_read_input_tokens": 3000
			},
			"context_window_size": 200000
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (5000 + 2000 + 3000) / 200000 * 100 = 5%
	got := data.ContextPercent()
	if got != 5 {
		t.Errorf("ContextPercent() = %d, want 5", got)
	}
}

func TestContextPercentMissingData(t *testing.T) {
	input := `{"model": "claude-opus-4-6"}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := data.ContextPercent()
	if got != -1 {
		t.Errorf("ContextPercent() = %d, want -1 for missing data", got)
	}
}

func TestContextPercentZeroWindowSize(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {"input_tokens": 100},
			"context_window_size": 0
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := data.ContextPercent()
	if got != -1 {
		t.Errorf("ContextPercent() = %d, want -1 for zero window size", got)
	}
}

func TestContextPercentRounding(t *testing.T) {
	// 1 / 3 * 100 = 33.333... → 33
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {"input_tokens": 1},
			"context_window_size": 3
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := data.ContextPercent()
	if got != 33 {
		t.Errorf("ContextPercent() = %d, want 33", got)
	}
}

func TestContextPercentFull(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"context_window": {
			"current_usage": {"input_tokens": 200000},
			"context_window_size": 200000
		}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := data.ContextPercent()
	if got != 100 {
		t.Errorf("ContextPercent() = %d, want 100", got)
	}
}

func TestParseWorkspaceFallbackToCurrentDir(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6",
		"workspace": {"current_dir": "/Users/dev/project/sub"}
	}`

	data, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When project_dir is empty, should fall back to current_dir
	if data.WorkspacePath() != "/Users/dev/project/sub" {
		t.Errorf("WorkspacePath() = %q, want %q", data.WorkspacePath(), "/Users/dev/project/sub")
	}
}
