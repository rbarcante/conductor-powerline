package segments

import (
	"os"
	"path/filepath"
)

// DetectConductorPlugin checks whether the claude-conductor plugin is installed
// in the user's Claude Code plugin directories. It accepts a base directory for
// testability; pass an empty string to use the real home directory.
//
// Detection locations checked (filesystem stat only, no network calls):
//   - <base>/.claude/plugins/claude-conductor/
//   - <base>/.claude/marketplace/claude-conductor/
func DetectConductorPlugin(baseDir string) bool {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		baseDir = home
	}

	locations := []string{
		filepath.Join(baseDir, ".claude", "plugins", "claude-conductor"),
		filepath.Join(baseDir, ".claude", "marketplace", "claude-conductor"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return true
		}
	}
	return false
}
