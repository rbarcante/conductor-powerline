package oauth

import (
	"errors"
	"os/exec"
	"strings"
)

// wincredCommandRunner executes shell commands for Windows Credential Manager access.
var wincredCommandRunner = runWincredCommand

// getWincredToken retrieves the Claude OAuth token from Windows Credential Manager
// using cmdkey and PowerShell to extract the credential value.
func getWincredToken() (string, error) {
	output, err := wincredCommandRunner(
		"-Command",
		"(Get-StoredCredential -Target 'claude.ai').Password | ConvertFrom-SecureString -AsPlainText",
	)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(output)
	if token == "" {
		return "", errors.New("oauth: empty token from wincred")
	}
	return token, nil
}

func runWincredCommand(args ...string) (string, error) {
	cmd := exec.Command("powershell", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
