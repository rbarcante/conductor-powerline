package segments

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// installedPluginsFile is the registry file Claude Code maintains for installed plugins.
type installedPluginsFile struct {
	Plugins map[string]json.RawMessage `json:"plugins"`
}

// DetectConductorPlugin checks whether the claude-conductor plugin is installed
// in the user's Claude Code plugin directories. It accepts a base directory for
// testability; pass an empty string to use the real home directory.
//
// Detection strategy (in order):
//  1. Parse <base>/.claude/plugins/installed_plugins.json — look for any key
//     containing "claude-conductor" (e.g. "claude-conductor@some-marketplace")
//  2. Scan <base>/.claude/plugins/cache/**/claude-conductor/ directories
//  3. Legacy: check <base>/.claude/plugins/claude-conductor/ or
//     <base>/.claude/marketplace/claude-conductor/
func DetectConductorPlugin(baseDir string) bool {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		baseDir = home
	}

	claudeDir := filepath.Join(baseDir, ".claude")

	// Strategy 1: parse installed_plugins.json
	if detectViaRegistry(claudeDir) {
		return true
	}

	// Strategy 2: scan plugins/cache for claude-conductor directories
	if detectViaCache(claudeDir) {
		return true
	}

	// Strategy 3: legacy flat-directory layout
	legacy := []string{
		filepath.Join(claudeDir, "plugins", "claude-conductor"),
		filepath.Join(claudeDir, "marketplace", "claude-conductor"),
	}
	for _, loc := range legacy {
		if _, err := os.Stat(loc); err == nil {
			return true
		}
	}

	return false
}

func detectViaRegistry(claudeDir string) bool {
	registryPath := filepath.Join(claudeDir, "plugins", "installed_plugins.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return false
	}

	var registry installedPluginsFile
	if err := json.Unmarshal(data, &registry); err != nil {
		return false
	}

	for key := range registry.Plugins {
		// Keys look like "claude-conductor@marketplace-name"
		if strings.HasPrefix(key, "claude-conductor") {
			return true
		}
	}
	return false
}

func detectViaCache(claudeDir string) bool {
	// Check both cache/<marketplace>/claude-conductor/ and
	// marketplaces/claude-conductor/ (the layout used by local marketplace installs).
	scanDirs := []string{
		filepath.Join(claudeDir, "plugins", "cache"),
		filepath.Join(claudeDir, "plugins", "marketplaces"),
	}

	for _, scanDir := range scanDirs {
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			// marketplaces/claude-conductor/ — direct match
			if entry.Name() == "claude-conductor" {
				return true
			}
			// cache/<marketplace>/claude-conductor/ — one level deeper
			pluginDir := filepath.Join(scanDir, entry.Name(), "claude-conductor")
			if _, err := os.Stat(pluginDir); err == nil {
				return true
			}
		}
	}
	return false
}
