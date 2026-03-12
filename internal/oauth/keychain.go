package oauth

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// keychainServiceName is the macOS Keychain service name for Claude Code credentials.
const keychainServiceName = "Claude Code-credentials"

// keychainCommandRunner executes shell commands for macOS Keychain access.
var keychainCommandRunner = runKeychainCommand

// getKeychainToken retrieves the Claude OAuth token from macOS Keychain
// using the security find-generic-password command.
func getKeychainToken() (string, error) {
	output, err := keychainCommandRunner(
		"find-generic-password",
		"-s", keychainServiceName,
		"-w",
	)
	if err != nil {
		return "", err
	}

	raw := strings.TrimSpace(output)
	if raw == "" {
		return "", errors.New("oauth: empty output from keychain")
	}

	// The keychain stores the full credentials JSON object.
	// Parse it to extract the access token.
	var cred credentialFile
	if err := json.Unmarshal([]byte(raw), &cred); err != nil {
		// If it's not JSON, check if it's a raw token
		if strings.HasPrefix(raw, "sk-ant-oat") {
			return raw, nil
		}
		return "", errors.New("oauth: could not parse keychain data")
	}

	if cred.ClaudeAiOAuth != nil && cred.ClaudeAiOAuth.AccessToken != "" {
		return cred.ClaudeAiOAuth.AccessToken, nil
	}

	return "", errors.New("oauth: no access token in keychain data")
}

// getKeychainCredentials retrieves credentials from macOS Keychain,
// including the refresh token if present in the JSON blob.
func getKeychainCredentials() (*TokenCredentials, error) {
	output, err := keychainCommandRunner(
		"find-generic-password",
		"-s", keychainServiceName,
		"-w",
	)
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(output)
	if raw == "" {
		return nil, errors.New("oauth: empty output from keychain")
	}

	// The keychain stores the full credentials JSON object.
	var cred credentialFile
	if err := json.Unmarshal([]byte(raw), &cred); err != nil {
		// If it's not JSON, check if it's a raw token
		if strings.HasPrefix(raw, "sk-ant-oat") {
			return &TokenCredentials{AccessToken: raw}, nil
		}
		return nil, errors.New("oauth: could not parse keychain data")
	}

	if cred.ClaudeAiOAuth != nil && cred.ClaudeAiOAuth.AccessToken != "" {
		return &TokenCredentials{
			AccessToken:  cred.ClaudeAiOAuth.AccessToken,
			RefreshToken: cred.ClaudeAiOAuth.RefreshToken,
		}, nil
	}

	return nil, errors.New("oauth: no access token in keychain data")
}

// updateKeychainTokens reads the existing Keychain JSON, updates the tokens,
// and writes it back using security add-generic-password -U.
func updateKeychainTokens(creds *TokenCredentials) error {
	// Read existing data
	output, err := keychainCommandRunner(
		"find-generic-password",
		"-s", keychainServiceName,
		"-w",
	)
	if err != nil {
		return err
	}

	raw := strings.TrimSpace(output)
	if raw == "" {
		return errors.New("oauth: empty keychain data")
	}

	// Parse into generic map to preserve all fields
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &jsonMap); err != nil {
		return err
	}

	oauthRaw, ok := jsonMap["claudeAiOauth"]
	if !ok {
		return errors.New("oauth: no claudeAiOauth in keychain data")
	}

	var oauthMap map[string]json.RawMessage
	if err := json.Unmarshal(oauthRaw, &oauthMap); err != nil {
		return err
	}

	accessJSON, _ := json.Marshal(creds.AccessToken)
	refreshJSON, _ := json.Marshal(creds.RefreshToken)
	oauthMap["accessToken"] = accessJSON
	oauthMap["refreshToken"] = refreshJSON

	updatedOAuth, err := json.Marshal(oauthMap)
	if err != nil {
		return err
	}
	jsonMap["claudeAiOauth"] = updatedOAuth

	updatedData, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}

	// Write back to Keychain using -U (update) flag
	_, err = keychainCommandRunner(
		"add-generic-password",
		"-s", keychainServiceName,
		"-a", keychainServiceName,
		"-w", string(updatedData),
		"-U",
	)
	return err
}

func runKeychainCommand(args ...string) (string, error) {
	cmd := exec.Command("security", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
