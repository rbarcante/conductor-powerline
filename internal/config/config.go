package config

import (
	"encoding/json"
	"os"
	"time"
)

// DefaultConfig returns the default configuration with all segments enabled.
func DefaultConfig() Config {
	return Config{
		Theme: "dark",
		Display: DisplayConfig{
			NerdFonts:    boolPtr(true),
			CompactWidth: 100,
		},
		Segments: map[string]SegmentConfig{
			"directory": {Enabled: true},
			"git":       {Enabled: true},
			"model":     {Enabled: true},
			"conductor": {Enabled: true},
			"block":     {Enabled: true},
			"weekly":    {Enabled: true},
			"context":   {Enabled: true},
		},
		SegmentOrder:   []string{"directory", "git", "model", "block", "weekly", "context", "conductor"},
		APITimeout:     Duration{5 * time.Second},
		CacheTTL:       Duration{30 * time.Second},
		TrendThreshold: 2.0,
	}
}

// LoadFromFile reads and parses a JSON config file. Returns a zero-value Config
// if the file does not exist. Returns an error for malformed JSON.
func LoadFromFile(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// MergeConfig merges an override config on top of a base config.
// Non-zero override values replace base values. Segment maps are merged
// key-by-key so that base segments not mentioned in override are preserved.
func MergeConfig(base, override Config) Config {
	merged := base

	if override.Theme != "" {
		merged.Theme = override.Theme
	}

	if override.Display.CompactWidth != 0 {
		merged.Display.CompactWidth = override.Display.CompactWidth
	}
	if override.Display.NerdFonts != nil {
		merged.Display.NerdFonts = override.Display.NerdFonts
	}

	if override.Segments != nil {
		if merged.Segments == nil {
			merged.Segments = make(map[string]SegmentConfig)
		}
		for k, v := range override.Segments {
			merged.Segments[k] = v
		}
	}

	if len(override.SegmentOrder) > 0 {
		merged.SegmentOrder = override.SegmentOrder
	}

	if override.APITimeout.Duration != 0 {
		merged.APITimeout = override.APITimeout
	}

	if override.CacheTTL.Duration != 0 {
		merged.CacheTTL = override.CacheTTL
	}

	if override.TrendThreshold != 0 {
		merged.TrendThreshold = override.TrendThreshold
	}

	return merged
}

// Load resolves configuration by loading project-level config, then user-level
// config, and merging both on top of defaults. Pass empty strings to skip a level.
func Load(projectPath, userPath string) Config {
	cfg := DefaultConfig()

	if userPath != "" {
		userCfg, err := LoadFromFile(userPath)
		if err == nil {
			cfg = MergeConfig(cfg, userCfg)
		}
	}

	if projectPath != "" {
		projectCfg, err := LoadFromFile(projectPath)
		if err == nil {
			cfg = MergeConfig(cfg, projectCfg)
		}
	}

	return cfg
}
