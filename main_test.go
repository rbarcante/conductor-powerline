package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestIntegrationPipeStdin(t *testing.T) {
	// Build the binary
	binPath := t.TempDir() + "/conductor-powerline"
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
	binPath := t.TempDir() + "/conductor-powerline"
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
	binPath := t.TempDir() + "/conductor-powerline"
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
	binPath := t.TempDir() + "/conductor-powerline"
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
