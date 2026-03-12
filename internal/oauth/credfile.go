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

// getCredfileCredentials reads credentials from ~/.claude/.credentials.json,
// returning both access token and refresh token (if present).
func getCredfileCredentials() (*TokenCredentials, error) {
	path := credfilePathResolver()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cred credentialFile
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, err
	}

	// Try Claude Code's nested format first
	if cred.ClaudeAiOAuth != nil && cred.ClaudeAiOAuth.AccessToken != "" {
		return &TokenCredentials{
			AccessToken:  cred.ClaudeAiOAuth.AccessToken,
			RefreshToken: cred.ClaudeAiOAuth.RefreshToken,
		}, nil
	}

	// Fall back to legacy flat format (no refresh token available)
	if cred.OAuthToken != "" {
		return &TokenCredentials{AccessToken: cred.OAuthToken}, nil
	}

	return nil, errors.New("oauth: empty token in credentials file")
}

// updateCredfileTokens reads the existing credfile, updates the access and
// refresh tokens, and writes it back. This preserves other fields like
// scopes, subscriptionType, etc. that Claude Code manages.
func updateCredfileTokens(creds *TokenCredentials) error {
	path := credfilePathResolver()
	if path == "" {
		return errors.New("oauth: no credfile path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse into a generic map to preserve all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	oauthRaw, ok := raw["claudeAiOauth"]
	if !ok {
		return errors.New("oauth: no claudeAiOauth in credfile")
	}

	var oauthMap map[string]json.RawMessage
	if err := json.Unmarshal(oauthRaw, &oauthMap); err != nil {
		return err
	}

	// Update tokens
	accessJSON, _ := json.Marshal(creds.AccessToken)
	refreshJSON, _ := json.Marshal(creds.RefreshToken)
	oauthMap["accessToken"] = accessJSON
	oauthMap["refreshToken"] = refreshJSON

	updatedOAuth, err := json.Marshal(oauthMap)
	if err != nil {
		return err
	}
	raw["claudeAiOauth"] = updatedOAuth

	updatedData, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return os.WriteFile(path, updatedData, 0o600)
}

func defaultCredfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}
