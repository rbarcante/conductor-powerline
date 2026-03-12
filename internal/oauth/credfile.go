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
	creds, err := getCredfileCredentials()
	if err != nil {
		return "", err
	}
	return creds.AccessToken, nil
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
// refresh tokens, and writes it back atomically. This preserves other fields
// like scopes, subscriptionType, etc. that Claude Code manages.
func updateCredfileTokens(creds *TokenCredentials) error {
	path := credfilePathResolver()
	if path == "" {
		return errors.New("oauth: no credfile path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updatedData, err := mergeTokensIntoJSON(data, creds)
	if err != nil {
		return err
	}

	// Atomic write: temp file + rename to avoid corrupting the credfile on crash
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-cred-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(updatedData); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// mergeTokensIntoJSON updates the access and refresh tokens in a Claude Code
// credential JSON blob, preserving all other fields (scopes, subscriptionType, etc.).
func mergeTokensIntoJSON(data []byte, creds *TokenCredentials) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	oauthRaw, ok := raw["claudeAiOauth"]
	if !ok {
		return nil, errors.New("oauth: no claudeAiOauth in credential data")
	}

	var oauthMap map[string]json.RawMessage
	if err := json.Unmarshal(oauthRaw, &oauthMap); err != nil {
		return nil, err
	}

	// json.Marshal on a plain string cannot fail
	accessJSON, _ := json.Marshal(creds.AccessToken)
	refreshJSON, _ := json.Marshal(creds.RefreshToken)
	oauthMap["accessToken"] = accessJSON
	oauthMap["refreshToken"] = refreshJSON

	updatedOAuth, err := json.Marshal(oauthMap)
	if err != nil {
		return nil, err
	}
	raw["claudeAiOauth"] = updatedOAuth

	return json.Marshal(raw)
}

func defaultCredfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}
