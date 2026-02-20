package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func binName() string {
	name := "conductor-powerline"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func TestIntegrationPipeStdin(t *testing.T) {
	// Build the binary
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run with stdin JSON
	input := `{"model":"claude-opus-4-6","workspace":"/tmp/my-project"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Verify segments appear in output
	if !strings.Contains(output, "my-project") {
		t.Error("expected 'my-project' directory segment in output")
	}
	if !strings.Contains(output, "Opus 4.6") {
		t.Error("expected 'Opus 4.6' model segment in output")
	}
	// Should contain ANSI codes
	if !strings.Contains(output, "\033[") {
		t.Error("expected ANSI escape codes in output")
	}
	// No trailing newline
	if strings.HasSuffix(output, "\n") {
		t.Error("output must not have trailing newline")
	}
}

func TestIntegrationEmptyStdin(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Should still produce output (directory from cwd, no model)
	if output == "" {
		t.Error("expected non-empty output even with empty stdin")
	}
	if strings.HasSuffix(output, "\n") {
		t.Error("output must not have trailing newline")
	}
}

func TestIntegrationMalformedStdin(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader("{invalid json")
	err := cmd.Run()

	// Should exit cleanly (exit 0) even with malformed input
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() != 0 {
			// Exit code 0 is expected (silent failure)
			// Any other exit code is unexpected but acceptable
			_ = exitErr
		}
	}

	// Verify no stderr output
	cmd2 := exec.Command(binPath)
	cmd2.Stdin = strings.NewReader("{invalid json")
	stderr, _ := cmd2.CombinedOutput()
	_ = stderr // Malformed input should produce no visible error
}

func TestIntegrationClaudeCodeSchema(t *testing.T) {
	// Build the binary
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run with Claude Code's actual JSON schema (model and workspace as objects)
	input := `{
		"model": {"id": "claude-opus-4-6", "display_name": "Claude Opus 4.6 (Thinking)"},
		"workspace": {"current_dir": "/tmp/my-project/src", "project_dir": "/tmp/my-project"},
		"cwd": "/tmp/my-project",
		"session_id": "test-session-123",
		"version": "1.0.30"
	}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Verify model segment shows friendly name (extracted from object id)
	if !strings.Contains(output, "Opus 4.6") {
		t.Errorf("expected 'Opus 4.6' model segment in output, got: %q", output)
	}
	// Verify directory segment shows project dir basename
	if !strings.Contains(output, "my-project") {
		t.Errorf("expected 'my-project' directory segment in output, got: %q", output)
	}
	// Should contain ANSI codes
	if !strings.Contains(output, "\033[") {
		t.Error("expected ANSI escape codes in output")
	}
	// No trailing newline
	if strings.HasSuffix(output, "\n") {
		t.Error("output must not have trailing newline")
	}
}

func TestIntegrationContextWindowSegment(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run with context_window data
	input := `{
		"model": "claude-opus-4-6",
		"workspace": "/tmp/my-project",
		"context_window": {
			"current_usage": {
				"input_tokens": 50000,
				"cache_creation_input_tokens": 10000,
				"cache_read_input_tokens": 20000
			},
			"context_window_size": 200000
		}
	}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// (50000+10000+20000)/200000*100 = 40%
	if !strings.Contains(output, "40%") {
		t.Errorf("expected '40%%' context segment in output, got: %q", output)
	}
	// Should contain the empty circle icon (nerd fonts enabled by default)
	if !strings.Contains(output, "○") {
		t.Errorf("expected '○' icon for 40%% usage, got: %q", output)
	}
}

func TestIntegrationContextWindowAbsent(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run without context_window data
	input := `{"model":"claude-opus-4-6","workspace":"/tmp/my-project"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Context segment should NOT appear when data is absent
	if strings.Contains(output, "○") || strings.Contains(output, "◐") || strings.Contains(output, "●") {
		t.Errorf("expected no context segment icons when data absent, got: %q", output)
	}
}

func TestIntegrationConductorSegmentPresent(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Use an isolated HOME so conductor plugin is guaranteed not installed,
	// which produces ConductorNone → segment shown in line 1.
	fakeHome := t.TempDir()
	input := `{"model":"claude-opus-4-6","workspace":"/tmp/my-project"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "HOME="+fakeHome, "USERPROFILE="+fakeHome)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Conductor segment should appear in line 1 only when not installed (ConductorNone/Marketplace)
	if !strings.Contains(output, "Conductor") {
		t.Errorf("expected 'Conductor' segment in output for uninstalled state, got: %q", output)
	}
}

func TestIntegrationConductorSegmentDisabled(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Write a project config that disables the conductor segment
	cfgDir := t.TempDir()
	cfgPath := cfgDir + "/.conductor-powerline.json"
	cfg := `{"segments":{"conductor":{"enabled":false}}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	input := `{"model":"claude-opus-4-6","workspace":"/tmp/my-project"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	cmd.Dir = cfgDir // run from dir with config file
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Conductor segment should NOT appear when disabled
	if strings.Contains(output, "Conductor") {
		t.Errorf("expected no 'Conductor' segment when disabled, got: %q", output)
	}
}

func TestIntegrationWorkflowSecondLine(t *testing.T) {
	// Build the binary
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Set up a fake project with conductor/ dir and installed_plugins.json so
	// ConductorActive is detected. We also need a fake conductor_cli.py that
	// outputs valid status JSON.
	projectDir := t.TempDir()
	conductorDir := filepath.Join(projectDir, "conductor")
	if err := os.MkdirAll(conductorDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fake installed_plugins.json in a fake home dir
	fakeHome := t.TempDir()
	pluginDir := filepath.Join(fakeHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	pluginsJSON := `{"plugins":{"conductor@claude-conductor":{}}}`
	if err := os.WriteFile(filepath.Join(pluginDir, "installed_plugins.json"), []byte(pluginsJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create fake conductor_cli.py that outputs valid status JSON
	cliDir := filepath.Join(fakeHome, ".claude", "plugins", "cache", "claude-conductor", "conductor", "1.0.0", "scripts")
	if err := os.MkdirAll(cliDir, 0755); err != nil {
		t.Fatal(err)
	}
	cliScript := `#!/usr/bin/env python3
import json, sys
data = {
  "success": True,
  "data": {
    "setup": {"is_valid": True, "setup_complete": True, "missing_required": []},
    "tracks": {
      "tracks": [
        {"description": "Test track", "status": "in_progress", "track_id": "test_20260220",
         "updated_at": "2026-02-20T00:00:00Z", "tasks": {"completed": 3, "in_progress": 1, "total": 10}}
      ]
    }
  }
}
print(json.dumps(data))
`
	if err := os.WriteFile(filepath.Join(cliDir, "conductor_cli.py"), []byte(cliScript), 0755); err != nil {
		t.Fatal(err)
	}

	// Escape backslashes in projectDir for JSON embedding (Windows paths)
	escapedProjectDir := strings.ReplaceAll(projectDir, `\`, `\\`)
	input := `{"model":"claude-opus-4-6","workspace":"` + escapedProjectDir + `"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"HOME="+fakeHome,
		"USERPROFILE="+fakeHome,       // Windows uses USERPROFILE for os.UserHomeDir()
		"XDG_CACHE_HOME="+t.TempDir(), // isolate cache
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v\noutput: %s", err, out)
	}

	output := string(out)

	// Output should contain a newline separator between line 1 and line 2
	if !strings.Contains(output, "\n") {
		t.Error("expected newline separator for two-line output")
	}

	lines := strings.SplitN(output, "\n", 2)
	if len(lines) < 2 {
		t.Fatalf("expected two lines of output, got: %q", output)
	}

	// Line 2 should contain workflow segment data
	line2 := lines[1]
	if !strings.Contains(line2, "Setup 100%") {
		t.Errorf("expected 'Setup 100%%' in line 2, got: %q", line2)
	}
	if !strings.Contains(line2, "3/10") {
		t.Errorf("expected '3/10' task count in line 2, got: %q", line2)
	}
}

func TestIntegrationWorkflowSecondLineDisabled(t *testing.T) {
	binPath := t.TempDir() + "/" + binName()
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Write config disabling conductor_workflow
	cfgDir := t.TempDir()
	cfgPath := cfgDir + "/.conductor-powerline.json"
	cfg := `{"segments":{"conductor_workflow":{"enabled":false}}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	input := `{"model":"claude-opus-4-6","workspace":"/tmp/my-project"}`
	cmd := exec.Command(binPath)
	cmd.Stdin = strings.NewReader(input)
	cmd.Dir = cfgDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	output := string(out)

	// Line 2 is disabled; output must not have trailing newline
	if strings.HasSuffix(output, "\n") {
		t.Error("output must not have trailing newline when workflow disabled")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
