package segments

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// Directory returns a segment displaying the project/directory name.
// It extracts the base name from the workspace path, falling back to cwd.
func Directory(workspace string, theme themes.Theme) Segment {
	colors := theme.Segments["directory"]

	name := extractDirName(workspace)

	return Segment{
		Name:    "directory",
		Text:    name,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

func extractDirName(workspace string) string {
	if workspace == "" {
		dir, err := os.Getwd()
		if err != nil {
			return "?"
		}
		return filepath.Base(dir)
	}

	if workspace == "/" {
		return "/"
	}

	workspace = strings.TrimRight(workspace, "/")
	return filepath.Base(workspace)
}
