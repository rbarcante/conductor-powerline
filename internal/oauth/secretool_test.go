package oauth

import (
	"errors"
	"testing"
)

func TestSecretoolSuccess(t *testing.T) {
	origRunner := secretoolCommandRunner
	defer func() { secretoolCommandRunner = origRunner }()

	secretoolCommandRunner = func(args ...string) (string, error) {
		return "sk-ant-oauth-linux-token", nil
	}

	token, err := getSecretoolToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token != "sk-ant-oauth-linux-token" {
		t.Errorf("expected token value, got %q", token)
	}
}

func TestSecretoolNotFound(t *testing.T) {
	origRunner := secretoolCommandRunner
	defer func() { secretoolCommandRunner = origRunner }()

	secretoolCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("secret-tool: No matching items")
	}

	_, err := getSecretoolToken()
	if err == nil {
		t.Error("expected error when secret not found")
	}
}

func TestSecretoolCommandError(t *testing.T) {
	origRunner := secretoolCommandRunner
	defer func() { secretoolCommandRunner = origRunner }()

	secretoolCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("secret-tool: command not found")
	}

	_, err := getSecretoolToken()
	if err == nil {
		t.Error("expected error on command failure")
	}
}
