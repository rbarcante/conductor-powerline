package debug

import (
	"fmt"
	"io"
	"os"
)

var (
	enabled bool
	output  io.Writer
)

// Init reads the CONDUCTOR_DEBUG environment variable and sets up
// the debug logger. Must be called once at program startup.
func Init() {
	enabled = os.Getenv("CONDUCTOR_DEBUG") == "1"
	output = os.Stderr
}

// Enabled reports whether debug logging is active.
func Enabled() bool {
	return enabled
}

// Logf writes a formatted debug message to stderr when debug mode is enabled.
// The tag identifies the subsystem (e.g. "oauth", "token", "cache").
func Logf(tag, format string, args ...any) {
	if !enabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(output, "[conductor:%s] %s\n", tag, msg)
}
