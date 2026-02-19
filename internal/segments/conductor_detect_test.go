package segments

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConductorPluginFound_Plugins(t *testing.T) {
	base := t.TempDir()
	pluginDir := filepath.Join(base, ".claude", "plugins", "claude-conductor")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !DetectConductorPlugin(base) {
		t.Error("expected DetectConductorPlugin to return true when plugin dir exists in plugins/")
	}
}

func TestDetectConductorPluginFound_Marketplace(t *testing.T) {
	base := t.TempDir()
	marketDir := filepath.Join(base, ".claude", "marketplace", "claude-conductor")
	if err := os.MkdirAll(marketDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !DetectConductorPlugin(base) {
		t.Error("expected DetectConductorPlugin to return true when plugin dir exists in marketplace/")
	}
}

func TestDetectConductorPluginNotFound(t *testing.T) {
	base := t.TempDir()
	// Create .claude dir but no claude-conductor subdirectory
	if err := os.MkdirAll(filepath.Join(base, ".claude", "plugins"), 0755); err != nil {
		t.Fatal(err)
	}

	if DetectConductorPlugin(base) {
		t.Error("expected DetectConductorPlugin to return false when no claude-conductor dir exists")
	}
}

func TestDetectConductorPluginNoDotClaude(t *testing.T) {
	base := t.TempDir()
	// No .claude directory at all

	if DetectConductorPlugin(base) {
		t.Error("expected DetectConductorPlugin to return false when .claude dir doesn't exist")
	}
}

func TestDetectConductorPluginBothLocations(t *testing.T) {
	base := t.TempDir()
	// Both locations exist — should return true
	if err := os.MkdirAll(filepath.Join(base, ".claude", "plugins", "claude-conductor"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, ".claude", "marketplace", "claude-conductor"), 0755); err != nil {
		t.Fatal(err)
	}

	if !DetectConductorPlugin(base) {
		t.Error("expected DetectConductorPlugin to return true when both locations exist")
	}
}
