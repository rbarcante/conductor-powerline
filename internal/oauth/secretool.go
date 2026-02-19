package oauth

import (
	"errors"
	"os/exec"
	"strings"
)

// secretoolServiceName is the GNOME Keyring service name for Claude credentials.
const secretoolServiceName = "claude.ai"

// secretoolTokenType is the GNOME Keyring attribute type for the OAuth token.
const secretoolTokenType = "oauth_token"

// secretoolCommandRunner executes shell commands for Linux secret-tool access.
var secretoolCommandRunner = runSecretoolCommand

// getSecretoolToken retrieves the Claude OAuth token from Linux GNOME Keyring
// using the secret-tool lookup command.
func getSecretoolToken() (string, error) {
	output, err := secretoolCommandRunner(
		"lookup",
		"service", secretoolServiceName,
		"type", secretoolTokenType,
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
