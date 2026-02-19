package segments

import (
	"os/exec"
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// BranchIcon is the Nerd Font glyph for git branch.
const BranchIcon = "\ue0a0"

// gitCommandRunner is the function used to execute git commands.
// It is a package-level variable to allow testing with mocks.
var gitCommandRunner = runGitCommand

// Git returns a segment displaying the current git branch and dirty state.
// When workspace is non-empty, git commands target that directory via -C.
// Returns a disabled segment if git is unavailable or not in a repo.
func Git(workspace string, theme themes.Theme) Segment {
	colors := theme.Segments["git"]

	branch, err := gitCommandRunner(gitArgs(workspace, "rev-parse", "--abbrev-ref", "HEAD")...)
	if err != nil {
		return Segment{Name: "git", Enabled: false}
	}

	branch = strings.TrimSpace(branch)
	text := BranchIcon + " " + branch

	dirty, err := gitCommandRunner(gitArgs(workspace, "status", "--porcelain")...)
	if err == nil && strings.TrimSpace(dirty) != "" {
		text += " *"
	}

	return Segment{
		Name:    "git",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// gitArgs prepends -C <workspace> to git arguments when workspace is non-empty.
func gitArgs(workspace string, args ...string) []string {
	if workspace == "" {
		return args
	}
	return append([]string{"-C", workspace}, args...)
}

func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
