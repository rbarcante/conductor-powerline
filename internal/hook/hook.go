// Package hook parses Claude Code hook JSON data from stdin.
package hook

import (
	"encoding/json"
	"io"
)

// Data holds the parsed hook input from Claude Code.
type Data struct {
	Model     string          `json:"model"`
	Workspace string          `json:"workspace"`
	Context   json.RawMessage `json:"context"`
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
