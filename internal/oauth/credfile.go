package oauth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// credfilePathResolver returns the path to the credentials file.
var credfilePathResolver = defaultCredfilePath

// claudeAiOAuthEntry represents the nested OAuth object in Claude Code's credentials file.
type claudeAiOAuthEntry struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// credentialFile represents the JSON structure of ~/.claude/.credentials.json.
// Supports both Claude Code's format {"claudeAiOauth":{"accessToken":"..."}}
// and the legacy flat format {"oauthToken":"..."}.
type credentialFile struct {
	ClaudeAiOAuth *claudeAiOAuthEntry `json:"claudeAiOauth"`
	OAuthToken    string              `json:"oauthToken"`
}

// extractTokenFromCredentialJSON parses credential JSON data and extracts the
// OAuth access token. Supports Claude Code's nested format, legacy flat format,
// and raw token strings prefixed with "sk-ant-oat".
func extractTokenFromCredentialJSON(data []byte) (string, error) {
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return "", errors.New("oauth: empty credential data")
	}

	var cred credentialFile
	if err := json.Unmarshal(data, &cred); err != nil {
		// If it's not JSON, check if it's a raw token
		if strings.HasPrefix(raw, "sk-ant-oat") {
			return raw, nil
		}
		return "", errors.New("oauth: could not parse credential data")
	}

	// Try Claude Code's nested format first
	if cred.ClaudeAiOAuth != nil && cred.ClaudeAiOAuth.AccessToken != "" {
		return cred.ClaudeAiOAuth.AccessToken, nil
	}

	// Fall back to legacy flat format
	if cred.OAuthToken != "" {
		return cred.OAuthToken, nil
	}

	return "", errors.New("oauth: no access token in credential data")
}

// getCredfileToken reads the Claude OAuth token from ~/.claude/.credentials.json.
func getCredfileToken() (string, error) {
	path := credfilePathResolver()

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return extractTokenFromCredentialJSON(data)
}

func defaultCredfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}
