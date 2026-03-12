package oauth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTokenFromCredentialJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name: "claude code nested format",
			data: `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-test","refreshToken":"rt"}}`,
			want: "sk-ant-oat01-test",
		},
		{
			name: "legacy flat format",
			data: `{"oauthToken":"sk-ant-oauth-legacy"}`,
			want: "sk-ant-oauth-legacy",
		},
		{
			name: "raw token prefix",
			data: "sk-ant-oat01-raw-token-12345",
			want: "sk-ant-oat01-raw-token-12345",
		},
		{
			name:    "malformed JSON",
			data:    "{not valid json}",
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			data:    "  \n  ",
			wantErr: true,
		},
		{
			name:    "empty access token",
			data:    `{"claudeAiOauth":{"accessToken":"","refreshToken":"rt"}}`,
			wantErr: true,
		},
		{
			name:    "no token fields",
			data:    `{"someOtherField":"value"}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTokenFromCredentialJSON([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got token %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

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

func TestCredfileClaudeCodeEmptyAccessToken(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".credentials.json")

	content := `{"claudeAiOauth":{"accessToken":"","refreshToken":"some-refresh"}}`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	origResolver := credfilePathResolver
	defer func() { credfilePathResolver = origResolver }()
	credfilePathResolver = func() string { return credPath }

	_, err := getCredfileToken()
	if err == nil {
		t.Error("expected error for empty accessToken in Claude Code format")
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
