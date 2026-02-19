package oauth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// credfilePathResolver returns the path to the credentials file.
var credfilePathResolver = defaultCredfilePath

// credentialFile represents the JSON structure of ~/.claude/.credentials.json.
type credentialFile struct {
	OAuthToken string `json:"oauthToken"`
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

	if cred.OAuthToken == "" {
		return "", errors.New("oauth: empty token in credentials file")
	}
	return cred.OAuthToken, nil
}

func defaultCredfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}
