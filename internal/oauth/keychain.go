package oauth

import (
	"os/exec"
)

// keychainServiceName is the macOS Keychain service name for Claude Code credentials.
const keychainServiceName = "Claude Code-credentials"

// keychainCommandRunner executes shell commands for macOS Keychain access.
var keychainCommandRunner = runKeychainCommand

// getKeychainToken retrieves the Claude OAuth token from macOS Keychain.
func getKeychainToken() (string, error) {
	output, err := keychainCommandRunner(
		"find-generic-password",
		"-s", keychainServiceName,
		"-w",
	)
	if err != nil {
		return "", err
	}

	return extractTokenFromCredentialJSON([]byte(output))
}

func runKeychainCommand(args ...string) (string, error) {
	cmd := exec.Command("security", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
