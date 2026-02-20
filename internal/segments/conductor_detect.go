package segments

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ConductorStatus represents the detection state of the conductor plugin.
type ConductorStatus int

const (
	// ConductorNone means neither the plugin nor the marketplace is present.
	ConductorNone ConductorStatus = iota
	// ConductorMarketplace means the marketplace dir exists but the plugin
	// is not registered in installed_plugins.json.
	ConductorMarketplace
	// ConductorInstalled means the plugin is in installed_plugins.json
	// but no conductor/ folder exists in the current project.
	ConductorInstalled
	// ConductorActive means the plugin is installed AND the current project
	// has a conductor/ directory (fully set up).
	ConductorActive
)

// installedPluginsFile is the registry file Claude Code maintains for installed plugins.
type installedPluginsFile struct {
	Plugins map[string]json.RawMessage `json:"plugins"`
}

// DetectConductorStatus checks the conductor plugin installation state.
// baseDir is the user's home directory (empty string uses os.UserHomeDir).
// projectDir is the current working directory to check for a conductor/ folder.
func DetectConductorStatus(baseDir string, projectDir string) ConductorStatus {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ConductorNone
		}
		baseDir = home
	}

	claudeDir := filepath.Join(baseDir, ".claude")

	inRegistry := detectViaRegistry(claudeDir)
	if inRegistry {
		if projectHasConductor(projectDir) {
			return ConductorActive
		}
		return ConductorInstalled
	}

	if marketplaceExists(claudeDir) {
		return ConductorMarketplace
	}

	return ConductorNone
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
		if strings.HasPrefix(key, "conductor@claude-conductor") {
			return true
		}
	}
	return false
}

func marketplaceExists(claudeDir string) bool {
	marketDir := filepath.Join(claudeDir, "plugins", "marketplaces", "claude-conductor")
	if _, err := os.Stat(marketDir); err == nil {
		return true
	}
	return false
}

func projectHasConductor(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	condDir := filepath.Join(projectDir, "conductor")
	info, err := os.Stat(condDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}
