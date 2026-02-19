package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got %q", cfg.Theme)
	}
	if !cfg.Display.NerdFonts {
		t.Error("expected NerdFonts enabled by default")
	}
	if cfg.Display.CompactWidth != 80 {
		t.Errorf("expected default CompactWidth 80, got %d", cfg.Display.CompactWidth)
	}
	expectedOrder := []string{"directory", "git", "model", "block", "weekly"}
	if len(cfg.SegmentOrder) != len(expectedOrder) {
		t.Fatalf("expected %d segment order items, got %d", len(expectedOrder), len(cfg.SegmentOrder))
	}
	for i, name := range expectedOrder {
		if cfg.SegmentOrder[i] != name {
			t.Errorf("segment order[%d]: expected %q, got %q", i, name, cfg.SegmentOrder[i])
		}
	}
	for _, name := range expectedOrder {
		seg, ok := cfg.Segments[name]
		if !ok {
			t.Errorf("expected segment %q in defaults", name)
			continue
		}
		if !seg.Enabled {
			t.Errorf("expected segment %q enabled by default", name)
		}
	}

	if cfg.APITimeout.Duration != 5*time.Second {
		t.Errorf("expected default APITimeout 5s, got %v", cfg.APITimeout.Duration)
	}
	if cfg.CacheTTL.Duration != 30*time.Second {
		t.Errorf("expected default CacheTTL 30s, got %v", cfg.CacheTTL.Duration)
	}
	if cfg.TrendThreshold != 2.0 {
		t.Errorf("expected default TrendThreshold 2.0, got %f", cfg.TrendThreshold)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".conductor-powerline.json")

	content := `{
		"theme": "nord",
		"display": { "nerdFonts": false },
		"segments": { "git": { "enabled": false } }
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "nord" {
		t.Errorf("expected theme 'nord', got %q", cfg.Theme)
	}
	if cfg.Display.NerdFonts {
		t.Error("expected NerdFonts disabled")
	}
	gitSeg, ok := cfg.Segments["git"]
	if !ok {
		t.Fatal("expected git segment in config")
	}
	if gitSeg.Enabled {
		t.Error("expected git segment disabled")
	}
}

func TestLoadFromFileMissing(t *testing.T) {
	cfg, err := LoadFromFile("/nonexistent/path/.conductor-powerline.json")
	if err != nil {
		t.Fatalf("missing file should not return error, got: %v", err)
	}
	// Should return zero-value config
	if cfg.Theme != "" {
		t.Errorf("expected empty theme for missing file, got %q", cfg.Theme)
	}
}

func TestLoadFromFileMalformed(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".conductor-powerline.json")
	if err := os.WriteFile(cfgPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFromFile(cfgPath)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestMergeConfig(t *testing.T) {
	base := DefaultConfig()
	override := Config{
		Theme: "gruvbox",
		Display: DisplayConfig{
			CompactWidth: 60,
		},
		Segments: map[string]SegmentConfig{
			"git": {Enabled: false},
		},
		SegmentOrder: []string{"model", "directory"},
	}

	merged := MergeConfig(base, override)

	if merged.Theme != "gruvbox" {
		t.Errorf("expected theme 'gruvbox', got %q", merged.Theme)
	}
	if merged.Display.CompactWidth != 60 {
		t.Errorf("expected CompactWidth 60, got %d", merged.Display.CompactWidth)
	}
	// NerdFonts should be overridden to false (zero value from override)
	if merged.Display.NerdFonts {
		t.Error("expected NerdFonts false after merge with override")
	}
	gitSeg, ok := merged.Segments["git"]
	if !ok {
		t.Fatal("expected git segment in merged config")
	}
	if gitSeg.Enabled {
		t.Error("expected git segment disabled after merge")
	}
	// directory segment should still exist from base
	dirSeg, ok := merged.Segments["directory"]
	if !ok {
		t.Fatal("expected directory segment preserved from base")
	}
	if !dirSeg.Enabled {
		t.Error("expected directory segment still enabled")
	}
	// Segment order should be overridden
	if len(merged.SegmentOrder) != 2 || merged.SegmentOrder[0] != "model" {
		t.Errorf("expected overridden segment order, got %v", merged.SegmentOrder)
	}
}

func TestMergeConfigPartialOverride(t *testing.T) {
	base := DefaultConfig()
	// Only override theme — everything else should stay default
	override := Config{
		Theme: "light",
	}

	merged := MergeConfig(base, override)

	if merged.Theme != "light" {
		t.Errorf("expected theme 'light', got %q", merged.Theme)
	}
	// Display should keep base defaults
	if !merged.Display.NerdFonts {
		t.Error("expected NerdFonts to remain true from base")
	}
	if merged.Display.CompactWidth != 80 {
		t.Errorf("expected CompactWidth 80 from base, got %d", merged.Display.CompactWidth)
	}
	// Segment order should stay default
	if len(merged.SegmentOrder) != 5 {
		t.Errorf("expected 5 segment order items from base, got %d", len(merged.SegmentOrder))
	}
}

func TestMergeConfigNewFields(t *testing.T) {
	base := DefaultConfig()
	override := Config{
		APITimeout:     Duration{10 * time.Second},
		CacheTTL:       Duration{60 * time.Second},
		TrendThreshold: 5.0,
	}

	merged := MergeConfig(base, override)

	if merged.APITimeout.Duration != 10*time.Second {
		t.Errorf("expected APITimeout 10s, got %v", merged.APITimeout.Duration)
	}
	if merged.CacheTTL.Duration != 60*time.Second {
		t.Errorf("expected CacheTTL 60s, got %v", merged.CacheTTL.Duration)
	}
	if merged.TrendThreshold != 5.0 {
		t.Errorf("expected TrendThreshold 5.0, got %f", merged.TrendThreshold)
	}
}

func TestLoadFromFileWithNewFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".conductor-powerline.json")

	content := `{
		"apiTimeout": "10s",
		"cacheTTL": "1m",
		"trendThreshold": 3.5
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APITimeout.Duration != 10*time.Second {
		t.Errorf("expected APITimeout 10s, got %v", cfg.APITimeout.Duration)
	}
	if cfg.CacheTTL.Duration != 1*time.Minute {
		t.Errorf("expected CacheTTL 1m, got %v", cfg.CacheTTL.Duration)
	}
	if cfg.TrendThreshold != 3.5 {
		t.Errorf("expected TrendThreshold 3.5, got %f", cfg.TrendThreshold)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".conductor-powerline.json")

	content := `{ "theme": "tokyo-night", "segments": { "model": { "enabled": false } } }`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(cfgPath, "")

	if cfg.Theme != "tokyo-night" {
		t.Errorf("expected theme 'tokyo-night', got %q", cfg.Theme)
	}
	modelSeg := cfg.Segments["model"]
	if modelSeg.Enabled {
		t.Error("expected model segment disabled")
	}
	// directory should still be enabled from defaults
	dirSeg := cfg.Segments["directory"]
	if !dirSeg.Enabled {
		t.Error("expected directory segment enabled from defaults")
	}
}

func TestLoadNoFiles(t *testing.T) {
	cfg := Load("/nonexistent/project", "/nonexistent/user")

	// Should return defaults
	if cfg.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got %q", cfg.Theme)
	}
	if len(cfg.SegmentOrder) != 5 {
		t.Errorf("expected 5 default segments, got %d", len(cfg.SegmentOrder))
	}
}
