package oauth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// credfilePathResolver returns the path to the credentials file.
var credfilePathResolver = defaultCredfilePath

// claudeAiOAuthEntry represents the nested OAuth object in Claude Code's credentials file.
type claudeAiOAuthEntry struct {
	AccessToken string `json:"accessToken"`
}

// credentialFile represents the JSON structure of ~/.claude/.credentials.json.
// Supports both Claude Code's format {"claudeAiOauth":{"accessToken":"..."}}
// and the legacy flat format {"oauthToken":"..."}.
type credentialFile struct {
	ClaudeAiOAuth *claudeAiOAuthEntry `json:"claudeAiOauth"`
	OAuthToken    string              `json:"oauthToken"`
}

// getCredfileToken reads the Claude OAuth token from ~/.claude/.credentials.json.
func getCredfileToken() (string, error) {
	path := credfilePathResolver()

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var cred credentialFile
	if err := json.Unmarshal(data, &cred); err != nil {
		return "", err
	}

	// Try Claude Code's nested format first
	if cred.ClaudeAiOAuth != nil && cred.ClaudeAiOAuth.AccessToken != "" {
		return cred.ClaudeAiOAuth.AccessToken, nil
	}

	// Fall back to legacy flat format
	if cred.OAuthToken != "" {
		return cred.OAuthToken, nil
	}

	return "", errors.New("oauth: empty token in credentials file")
}

func defaultCredfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}
