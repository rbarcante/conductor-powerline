package segments

import (
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestDirectoryFromWorkspace(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Directory("/Users/dev/my-project", theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Text != "my-project" {
		t.Errorf("expected text 'my-project', got %q", seg.Text)
	}
	if seg.Name != "directory" {
		t.Errorf("expected name 'directory', got %q", seg.Name)
	}
	if seg.FG == "" || seg.BG == "" {
		t.Error("expected non-empty colors from theme")
	}
}

func TestDirectoryNestedPath(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Directory("/home/user/projects/deep/nested/repo", theme)

	if seg.Text != "repo" {
		t.Errorf("expected text 'repo', got %q", seg.Text)
	}
}

func TestDirectoryRootPath(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Directory("/", theme)

	if seg.Text != "/" {
		t.Errorf("expected text '/', got %q", seg.Text)
	}
}

func TestDirectoryEmptyWorkspace(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Directory("", theme)

	// Should fall back to something (cwd base name)
	if seg.Text == "" {
		t.Error("expected non-empty text even with empty workspace")
	}
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
}

func TestDirectoryTrailingSlash(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Directory("/Users/dev/project/", theme)

	if seg.Text != "project" {
		t.Errorf("expected text 'project', got %q", seg.Text)
	}
}
