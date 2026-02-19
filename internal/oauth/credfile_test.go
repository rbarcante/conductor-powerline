package oauth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredfileValidJSON(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".credentials.json")

	content := `{"oauthToken":"sk-ant-oauth-credfile-token"}`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return credPath }

	token, err := getCredfileToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token != "sk-ant-oauth-credfile-token" {
		t.Errorf("expected credfile token, got %q", token)
	}
}

func TestCredfileMissingFile(t *testing.T) {
	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return "/nonexistent/path/.credentials.json" }

	_, err := getCredfileToken()
	if err == nil {
		t.Error("expected error for missing credentials file")
	}
}

func TestCredfileMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".credentials.json")

	if err := os.WriteFile(credPath, []byte("{not valid json}"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return credPath }

	_, err := getCredfileToken()
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestCredfileClaudeCodeFormat(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".credentials.json")

	// This is the actual format Claude Code writes to ~/.claude/.credentials.json
	content := `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-test-token","refreshToken":"sk-ant-ort01-refresh","expiresAt":1771535255460,"scopes":["user:inference"],"subscriptionType":"max"}}`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return credPath }

	token, err := getCredfileToken()
	if err != nil {
		t.Fatalf("expected token from Claude Code format, got error: %v", err)
	}
	if token != "sk-ant-oat01-test-token" {
		t.Errorf("expected access token, got %q", token)
	}
}

func TestCredfileEmptyToken(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".credentials.json")

	content := `{"oauthToken":""}`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return credPath }

	_, err := getCredfileToken()
	if err == nil {
		t.Error("expected error for empty token in credentials file")
	}
}
