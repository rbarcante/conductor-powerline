package oauth

import (
	"errors"
	"os/exec"
	"strings"
)

// keychainCommandRunner executes shell commands for macOS Keychain access.
var keychainCommandRunner = runKeychainCommand

// getKeychainToken retrieves the Claude OAuth token from macOS Keychain
// using the security find-generic-password command.
func getKeychainToken() (string, error) {
	output, err := keychainCommandRunner(
		"find-generic-password",
		"-s", "claude.ai",
		"-a", "oauth_token",
		"-w",
	)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(output)
	if token == "" {
		return "", errors.New("oauth: empty token from keychain")
	}
	return token, nil
}

func runKeychainCommand(args ...string) (string, error) {
	cmd := exec.Command("security", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
