package debug

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogf_Enabled(t *testing.T) {
	t.Setenv("CONDUCTOR_DEBUG", "1")
	Init()

	var buf bytes.Buffer
	output = &buf

	Logf("token", "got token: %s", "abc123")

	got := buf.String()
	if !strings.Contains(got, "[conductor:token]") {
		t.Errorf("expected tag [conductor:token], got %q", got)
	}
	if !strings.Contains(got, "got token: abc123") {
		t.Errorf("expected message 'got token: abc123', got %q", got)
	}
}

func TestLogf_Disabled(t *testing.T) {
	t.Setenv("CONDUCTOR_DEBUG", "")
	Init()

	var buf bytes.Buffer
	output = &buf

	Logf("oauth", "this should not appear")

	if buf.Len() != 0 {
		t.Errorf("expected no output when disabled, got %q", buf.String())
	}
}

func TestLogf_DisabledByDefault(t *testing.T) {
	os.Unsetenv("CONDUCTOR_DEBUG")
	Init()

	var buf bytes.Buffer
	output = &buf

	Logf("test", "should not appear")

	if buf.Len() != 0 {
		t.Errorf("expected no output by default, got %q", buf.String())
	}
}

func TestEnabled(t *testing.T) {
	t.Setenv("CONDUCTOR_DEBUG", "1")
	Init()
	if !Enabled() {
		t.Error("expected Enabled() == true when CONDUCTOR_DEBUG=1")
	}

	t.Setenv("CONDUCTOR_DEBUG", "")
	Init()
	if Enabled() {
		t.Error("expected Enabled() == false when CONDUCTOR_DEBUG is empty")
	}
}

func TestInit_OutputDefaultsToStderr(t *testing.T) {
	t.Setenv("CONDUCTOR_DEBUG", "1")
	output = nil
	Init()

	if output != os.Stderr {
		t.Error("expected output to default to os.Stderr after Init()")
	}
}
