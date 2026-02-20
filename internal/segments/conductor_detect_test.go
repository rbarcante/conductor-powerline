package segments

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// makeRegistry writes a minimal installed_plugins.json to base/.claude/plugins/
func makeRegistry(t *testing.T, base string, plugins map[string]any) {
	t.Helper()
	dir := filepath.Join(base, ".claude", "plugins")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	type registryFile struct {
		Version int            `json:"version"`
		Plugins map[string]any `json:"plugins"`
	}
	data, err := json.Marshal(registryFile{Version: 2, Plugins: plugins})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "installed_plugins.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

// --- Strategy 1: installed_plugins.json registry ---

func TestDetectConductorPlugin_RegistryFound(t *testing.T) {
	base := t.TempDir()
	makeRegistry(t, base, map[string]any{
		"claude-conductor@some-marketplace": []any{},
	})
	if !DetectConductorPlugin(base) {
		t.Error("expected true when installed_plugins.json contains claude-conductor key")
	}
}

func TestDetectConductorPlugin_RegistryFoundOtherPlugins(t *testing.T) {
	base := t.TempDir()
	// Registry has other plugins but NOT conductor
	makeRegistry(t, base, map[string]any{
		"context7@claude-plugins-official":   []any{},
		"frontend-design@claude-plugins-official": []any{},
	})
	if DetectConductorPlugin(base) {
		t.Error("expected false when registry contains no claude-conductor key")
	}
}

func TestDetectConductorPlugin_RegistryMalformed(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, ".claude", "plugins")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write invalid JSON — should fall through to next strategy
	if err := os.WriteFile(filepath.Join(dir, "installed_plugins.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	// No other locations exist — should return false without panicking
	if DetectConductorPlugin(base) {
		t.Error("expected false for malformed registry with no other detection paths")
	}
}

// --- Strategy 2: plugins/cache directory scan ---

func TestDetectConductorPlugin_CacheFound(t *testing.T) {
	base := t.TempDir()
	cacheDir := filepath.Join(base, ".claude", "plugins", "cache", "some-marketplace", "claude-conductor")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !DetectConductorPlugin(base) {
		t.Error("expected true when cache/marketplace/claude-conductor dir exists")
	}
}

func TestDetectConductorPlugin_MarketplacesDir(t *testing.T) {
	base := t.TempDir()
	// ~/.claude/plugins/marketplaces/claude-conductor/ — local marketplace layout
	marketDir := filepath.Join(base, ".claude", "plugins", "marketplaces", "claude-conductor")
	if err := os.MkdirAll(marketDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !DetectConductorPlugin(base) {
		t.Error("expected true when plugins/marketplaces/claude-conductor dir exists")
	}
}

func TestDetectConductorPlugin_CacheOtherPluginsOnly(t *testing.T) {
	base := t.TempDir()
	// Other plugins in cache but not conductor
	otherDir := filepath.Join(base, ".claude", "plugins", "cache", "some-marketplace", "other-plugin")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	if DetectConductorPlugin(base) {
		t.Error("expected false when cache has no claude-conductor dir")
	}
}

// --- Strategy 3: legacy flat directories ---

func TestDetectConductorPlugin_LegacyPluginsDir(t *testing.T) {
	base := t.TempDir()
	pluginDir := filepath.Join(base, ".claude", "plugins", "claude-conductor")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !DetectConductorPlugin(base) {
		t.Error("expected true for legacy plugins/claude-conductor dir")
	}
}

func TestDetectConductorPlugin_LegacyMarketplaceDir(t *testing.T) {
	base := t.TempDir()
	marketDir := filepath.Join(base, ".claude", "marketplace", "claude-conductor")
	if err := os.MkdirAll(marketDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !DetectConductorPlugin(base) {
		t.Error("expected true for legacy marketplace/claude-conductor dir")
	}
}

// --- Not found cases ---

func TestDetectConductorPlugin_NotFound(t *testing.T) {
	base := t.TempDir()
	// .claude exists but no conductor anywhere
	if err := os.MkdirAll(filepath.Join(base, ".claude", "plugins"), 0755); err != nil {
		t.Fatal(err)
	}
	if DetectConductorPlugin(base) {
		t.Error("expected false when no claude-conductor is found anywhere")
	}
}

func TestDetectConductorPlugin_NoDotClaude(t *testing.T) {
	base := t.TempDir()
	if DetectConductorPlugin(base) {
		t.Error("expected false when .claude dir doesn't exist")
	}
}

func TestDetectConductorPlugin_EmptyBaseUsesHomeDir(t *testing.T) {
	// Smoke test: passing "" should use os.UserHomeDir() without panicking
	result := DetectConductorPlugin("")
	_ = result
}
