// Package hook parses Claude Code hook JSON data from stdin.
package hook

import (
	"encoding/json"
	"io"
)

// Data holds the parsed hook input from Claude Code.
// Supports both legacy string format and Claude Code's object format
// for model and workspace fields.
type Data struct {
	// model can be a string ("claude-opus-4-6") or an object ({"id": "...", "display_name": "..."})
	model rawField
	// workspace can be a string ("/path") or an object ({"current_dir": "...", "project_dir": "..."})
	workspace rawField

	Context json.RawMessage `json:"context"`
}

// rawField holds a JSON value that can be either a string or an object.
type rawField struct {
	raw json.RawMessage
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
// forms for model and workspace fields.
func (d *Data) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type alias struct {
		Model     json.RawMessage `json:"model"`
		Workspace json.RawMessage `json:"workspace"`
		Context   json.RawMessage `json:"context"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	d.model.raw = a.Model
	d.workspace.raw = a.Workspace
	d.Context = a.Context
	return nil
}

// Model returns the model string for backward compatibility.
// If model was an object, returns the id field. If a string, returns as-is.
// Deprecated: use ModelID() instead.
func (d Data) Model() string {
	return d.ModelID()
}

// ModelID returns the model identifier.
// For object form: returns the "id" field.
// For string form: returns the string value.
func (d Data) ModelID() string {
	if len(d.model.raw) == 0 {
		return ""
	}

	// Try string first
	var s string
	if err := json.Unmarshal(d.model.raw, &s); err == nil {
		return s
	}

	// Try object
	var obj modelObject
	if err := json.Unmarshal(d.model.raw, &obj); err == nil {
		return obj.ID
	}

	return ""
}

// ModelDisplayName returns the friendly display name from the model object.
// Returns empty string if model was a plain string or if display_name is not set.
func (d Data) ModelDisplayName() string {
	if len(d.model.raw) == 0 {
		return ""
	}

	var obj modelObject
	if err := json.Unmarshal(d.model.raw, &obj); err == nil {
		return obj.DisplayName
	}

	return ""
}

// Workspace returns the workspace path for backward compatibility.
// Deprecated: use WorkspacePath() instead.
func (d Data) Workspace() string {
	return d.WorkspacePath()
}

// WorkspacePath returns the project directory path.
// For object form: returns "project_dir", falling back to "current_dir".
// For string form: returns the string value.
func (d Data) WorkspacePath() string {
	if len(d.workspace.raw) == 0 {
		return ""
	}

	// Try string first
	var s string
	if err := json.Unmarshal(d.workspace.raw, &s); err == nil {
		return s
	}

	// Try object
	var obj workspaceObject
	if err := json.Unmarshal(d.workspace.raw, &obj); err == nil {
		if obj.ProjectDir != "" {
			return obj.ProjectDir
		}
		return obj.CurrentDir
	}

	return ""
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
