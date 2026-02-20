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

func makeMarketplace(t *testing.T, base string) {
	t.Helper()
	dir := filepath.Join(base, ".claude", "plugins", "marketplaces", "claude-conductor")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func makeProjectConductor(t *testing.T, projectDir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(projectDir, "conductor"), 0755); err != nil {
		t.Fatal(err)
	}
}

// --- ConductorActive: in registry + conductor/ dir in project ---

func TestDetect_Active(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	makeRegistry(t, base, map[string]any{
		"claude-conductor@some-marketplace": []any{},
	})
	makeProjectConductor(t, project)

	status := DetectConductorStatus(base, project)
	if status != ConductorActive {
		t.Errorf("expected ConductorActive, got %d", status)
	}
}

// --- ConductorInstalled: in registry, no conductor/ in project ---

func TestDetect_Installed(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir() // no conductor/ dir
	makeRegistry(t, base, map[string]any{
		"claude-conductor@some-marketplace": []any{},
	})

	status := DetectConductorStatus(base, project)
	if status != ConductorInstalled {
		t.Errorf("expected ConductorInstalled, got %d", status)
	}
}

func TestDetect_InstalledEmptyProject(t *testing.T) {
	base := t.TempDir()
	makeRegistry(t, base, map[string]any{
		"claude-conductor@some-marketplace": []any{},
	})

	// Empty project dir string
	status := DetectConductorStatus(base, "")
	if status != ConductorInstalled {
		t.Errorf("expected ConductorInstalled, got %d", status)
	}
}

// --- ConductorMarketplace: not in registry, marketplace dir exists ---

func TestDetect_Marketplace(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	makeMarketplace(t, base)

	status := DetectConductorStatus(base, project)
	if status != ConductorMarketplace {
		t.Errorf("expected ConductorMarketplace, got %d", status)
	}
}

func TestDetect_MarketplaceNoRegistry(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	// Registry has other plugins but not conductor
	makeRegistry(t, base, map[string]any{
		"context7@claude-plugins-official": []any{},
	})
	makeMarketplace(t, base)

	status := DetectConductorStatus(base, project)
	if status != ConductorMarketplace {
		t.Errorf("expected ConductorMarketplace, got %d", status)
	}
}

// --- ConductorNone: nothing present ---

func TestDetect_None(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	// Create .claude/plugins but no registry, no marketplace
	if err := os.MkdirAll(filepath.Join(base, ".claude", "plugins"), 0755); err != nil {
		t.Fatal(err)
	}

	status := DetectConductorStatus(base, project)
	if status != ConductorNone {
		t.Errorf("expected ConductorNone, got %d", status)
	}
}

func TestDetect_NoDotClaude(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()

	status := DetectConductorStatus(base, project)
	if status != ConductorNone {
		t.Errorf("expected ConductorNone, got %d", status)
	}
}

func TestDetect_MalformedRegistry(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	dir := filepath.Join(base, ".claude", "plugins")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "installed_plugins.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	status := DetectConductorStatus(base, project)
	if status != ConductorNone {
		t.Errorf("expected ConductorNone for malformed registry, got %d", status)
	}
}

func TestDetect_EmptyBaseUsesHomeDir(t *testing.T) {
	project := t.TempDir()
	// Smoke test: passing "" for baseDir uses os.UserHomeDir() without panicking
	status := DetectConductorStatus("", project)
	_ = status
}

func TestDetect_ProjectConductorIsFile(t *testing.T) {
	base := t.TempDir()
	project := t.TempDir()
	makeRegistry(t, base, map[string]any{
		"claude-conductor@some-marketplace": []any{},
	})
	// conductor exists but is a file, not a directory
	if err := os.WriteFile(filepath.Join(project, "conductor"), []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}

	status := DetectConductorStatus(base, project)
	if status != ConductorInstalled {
		t.Errorf("expected ConductorInstalled when conductor is a file not dir, got %d", status)
	}
}
