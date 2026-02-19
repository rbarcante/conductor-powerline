package oauth

import (
	"errors"
	"testing"
)

func TestKeychainSuccess(t *testing.T) {
	origRunner := keychainCommandRunner
	defer func() { keychainCommandRunner = origRunner }()

	// Keychain returns JSON with the Claude Code credential format
	keychainCommandRunner = func(args ...string) (string, error) {
		return `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-test-token"}}`, nil
	}

	token, err := getKeychainToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token != "sk-ant-oat01-test-token" {
		t.Errorf("expected access token, got %q", token)
	}
}

func TestKeychainRawToken(t *testing.T) {
	origRunner := keychainCommandRunner
	defer func() { keychainCommandRunner = origRunner }()

	// Some setups may store the raw token directly
	keychainCommandRunner = func(args ...string) (string, error) {
		return "sk-ant-oat01-raw-token-12345", nil
	}

	token, err := getKeychainToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token != "sk-ant-oat01-raw-token-12345" {
		t.Errorf("expected raw token, got %q", token)
	}
}

func TestKeychainNotFound(t *testing.T) {
	origRunner := keychainCommandRunner
	defer func() { keychainCommandRunner = origRunner }()

	keychainCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.")
	}

	_, err := getKeychainToken()
	if err == nil {
		t.Error("expected error when keychain item not found")
	}
}

func TestKeychainCommandError(t *testing.T) {
	origRunner := keychainCommandRunner
	defer func() { keychainCommandRunner = origRunner }()

	keychainCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("security: command not found")
	}

	_, err := getKeychainToken()
	if err == nil {
		t.Error("expected error on command failure")
	}
}

func TestKeychainEmptyOutput(t *testing.T) {
	origRunner := keychainCommandRunner
	defer func() { keychainCommandRunner = origRunner }()

	keychainCommandRunner = func(args ...string) (string, error) {
		return "  \n  ", nil
	}

	_, err := getKeychainToken()
	if err == nil {
		t.Error("expected error for empty token output")
	}
}
