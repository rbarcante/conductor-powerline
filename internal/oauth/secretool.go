package oauth

import (
	"errors"
	"os/exec"
	"strings"
)

// secretoolCommandRunner executes shell commands for Linux secret-tool access.
var secretoolCommandRunner = runSecretoolCommand

// getSecretoolToken retrieves the Claude OAuth token from Linux GNOME Keyring
// using the secret-tool lookup command.
func getSecretoolToken() (string, error) {
	output, err := secretoolCommandRunner(
		"lookup",
		"service", "claude.ai",
		"type", "oauth_token",
	)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(output)
	if token == "" {
		return "", errors.New("oauth: empty token from secret-tool")
	}
	return token, nil
}

func runSecretoolCommand(args ...string) (string, error) {
	cmd := exec.Command("secret-tool", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
