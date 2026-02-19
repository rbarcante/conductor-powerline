// Package hook parses Claude Code hook JSON data from stdin.
package hook

import (
	"encoding/json"
	"io"
	"math"
)

// ContextWindowUsage holds token usage counts from the context window.
type ContextWindowUsage struct {
	InputTokens                int `json:"input_tokens"`
	CacheCreationInputTokens   int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens       int `json:"cache_read_input_tokens"`
}

// ContextWindow holds context window data from Claude Code's hook JSON.
type ContextWindow struct {
	CurrentUsage      ContextWindowUsage `json:"current_usage"`
	ContextWindowSize int                `json:"context_window_size"`
}

// Data holds the parsed hook input from Claude Code.
// Supports both legacy string format and Claude Code's object format
// for model and workspace fields. Fields are resolved once during unmarshal.
type Data struct {
	modelID          string
	modelDisplayName string
	workspacePath    string
	contextWindow    *ContextWindow

	Context json.RawMessage `json:"context"`
}

// modelObject represents the object form of the model field.
type modelObject struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// workspaceObject represents the object form of the workspace field.
type workspaceObject struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// UnmarshalJSON implements custom unmarshaling to handle both string and object
// forms for model and workspace fields. Values are resolved eagerly.
func (d *Data) UnmarshalJSON(data []byte) error {
	type alias struct {
		Model         json.RawMessage `json:"model"`
		Workspace     json.RawMessage `json:"workspace"`
		Context       json.RawMessage `json:"context"`
		ContextWindow json.RawMessage `json:"context_window"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	d.Context = a.Context
	d.modelID, d.modelDisplayName = resolveModel(a.Model)
	d.workspacePath = resolveWorkspace(a.Workspace)
	d.contextWindow = resolveContextWindow(a.ContextWindow)
	return nil
}

func resolveModel(raw json.RawMessage) (id, displayName string) {
	if len(raw) == 0 {
		return "", ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, ""
	}
	var obj modelObject
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj.ID, obj.DisplayName
	}
	return "", ""
}

func resolveWorkspace(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj workspaceObject
	if err := json.Unmarshal(raw, &obj); err == nil {
		if obj.ProjectDir != "" {
			return obj.ProjectDir
		}
		return obj.CurrentDir
	}
	return ""
}

func resolveContextWindow(raw json.RawMessage) *ContextWindow {
	if len(raw) == 0 {
		return nil
	}
	var cw ContextWindow
	if err := json.Unmarshal(raw, &cw); err != nil {
		return nil
	}
	return &cw
}

// ContextWindow returns the parsed context window data, or nil if absent.
func (d Data) ContextWindow() *ContextWindow {
	return d.contextWindow
}

// ContextPercent returns the context window usage as a rounded percentage (0-100).
// Returns -1 if context window data is missing or window size is zero.
func (d Data) ContextPercent() int {
	if d.contextWindow == nil || d.contextWindow.ContextWindowSize == 0 {
		return -1
	}
	u := d.contextWindow.CurrentUsage
	total := float64(u.InputTokens + u.CacheCreationInputTokens + u.CacheReadInputTokens)
	pct := total / float64(d.contextWindow.ContextWindowSize) * 100
	return int(math.Round(pct))
}

// ModelID returns the model identifier.
func (d Data) ModelID() string {
	return d.modelID
}

// ModelDisplayName returns the friendly display name from the model object.
// Returns empty string if model was a plain string or if display_name is not set.
func (d Data) ModelDisplayName() string {
	return d.modelDisplayName
}

// WorkspacePath returns the project directory path.
func (d Data) WorkspacePath() string {
	return d.workspacePath
}

// Parse reads JSON from r and returns the parsed hook data.
// Returns a zero-value Data for empty input. Returns an error for malformed JSON.
func Parse(r io.Reader) (Data, error) {
	var data Data

	raw, err := io.ReadAll(r)
	if err != nil {
		return data, err
	}

	if len(raw) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return data, err
	}
	return data, nil
}
