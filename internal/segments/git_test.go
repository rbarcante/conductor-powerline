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
		switch args[0] {
		case "rev-parse":
			return "main", nil
		case "status":
			return "", nil
		}
		return "", nil
	}

	seg := Git(theme)

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
		switch args[0] {
		case "rev-parse":
			return "feature/my-branch", nil
		case "status":
			return " M file.go\n", nil
		}
		return "", nil
	}

	seg := Git(theme)

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

	seg := Git(theme)

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

	seg := Git(theme)

	if seg.Enabled {
		t.Error("expected segment disabled when not in a git repo")
	}
}

// testError implements the error interface for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
