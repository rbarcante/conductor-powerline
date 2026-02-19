package segments

import (
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestGitCleanBranch(t *testing.T) {
	theme, _ := themes.Get("dark")

	// Mock git commands
	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	gitCommandRunner = func(args ...string) (string, error) {
		switch args[len(args)-1] {
		case "HEAD":
			return "main", nil
		case "--porcelain":
			return "", nil
		}
		return "", nil
	}

	seg := Git("", theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Text != "\ue0a0 main" {
		t.Errorf("expected text with branch icon, got %q", seg.Text)
	}
	if seg.Name != "git" {
		t.Errorf("expected name 'git', got %q", seg.Name)
	}
}

func TestGitDirtyBranch(t *testing.T) {
	theme, _ := themes.Get("dark")

	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	gitCommandRunner = func(args ...string) (string, error) {
		switch args[len(args)-1] {
		case "HEAD":
			return "feature/my-branch", nil
		case "--porcelain":
			return " M file.go\n", nil
		}
		return "", nil
	}

	seg := Git("", theme)

	if seg.Text != "\ue0a0 feature/my-branch *" {
		t.Errorf("expected dirty indicator, got %q", seg.Text)
	}
}

func TestGitUnavailable(t *testing.T) {
	theme, _ := themes.Get("dark")

	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	gitCommandRunner = func(args ...string) (string, error) {
		return "", &testError{msg: "git not found"}
	}

	seg := Git("", theme)

	if seg.Enabled {
		t.Error("expected segment disabled when git unavailable")
	}
}

func TestGitNotARepo(t *testing.T) {
	theme, _ := themes.Get("dark")

	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	gitCommandRunner = func(args ...string) (string, error) {
		return "", &testError{msg: "not a git repository"}
	}

	seg := Git("", theme)

	if seg.Enabled {
		t.Error("expected segment disabled when not in a git repo")
	}
}

func TestGitWithWorkspacePath(t *testing.T) {
	theme, _ := themes.Get("dark")

	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	var capturedArgs []string
	gitCommandRunner = func(args ...string) (string, error) {
		capturedArgs = append(capturedArgs, args...)
		switch args[len(args)-1] {
		case "HEAD":
			return "develop", nil
		case "--porcelain":
			return "", nil
		}
		return "", nil
	}

	seg := Git("/home/user/my-project", theme)

	if !seg.Enabled {
		t.Error("expected segment enabled with workspace path")
	}
	if seg.Text != "\ue0a0 develop" {
		t.Errorf("expected branch 'develop', got %q", seg.Text)
	}

	// Verify -C flag was passed for both commands
	foundC := false
	for i, arg := range capturedArgs {
		if arg == "-C" && i+1 < len(capturedArgs) && capturedArgs[i+1] == "/home/user/my-project" {
			foundC = true
			break
		}
	}
	if !foundC {
		t.Errorf("expected -C /home/user/my-project in git args, got %v", capturedArgs)
	}
}

func TestGitWorkspacePathIgnoredWhenEmpty(t *testing.T) {
	theme, _ := themes.Get("dark")

	origRunner := gitCommandRunner
	defer func() { gitCommandRunner = origRunner }()

	var capturedArgs []string
	gitCommandRunner = func(args ...string) (string, error) {
		capturedArgs = append(capturedArgs, args...)
		switch args[len(args)-1] {
		case "HEAD":
			return "main", nil
		case "--porcelain":
			return "", nil
		}
		return "", nil
	}

	seg := Git("", theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}

	// Verify -C flag was NOT passed
	for _, arg := range capturedArgs {
		if arg == "-C" {
			t.Error("expected no -C flag when workspace is empty")
			break
		}
	}
}

// testError implements the error interface for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
