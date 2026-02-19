package oauth

import (
	"errors"
	"testing"
)

func TestWincredSuccess(t *testing.T) {
	origRunner := wincredCommandRunner
	defer func() { wincredCommandRunner = origRunner }()

	wincredCommandRunner = func(args ...string) (string, error) {
		return "sk-ant-oauth-windows-token", nil
	}

	token, err := getWincredToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token != "sk-ant-oauth-windows-token" {
		t.Errorf("expected token value, got %q", token)
	}
}

func TestWincredNotFound(t *testing.T) {
	origRunner := wincredCommandRunner
	defer func() { wincredCommandRunner = origRunner }()

	wincredCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("Element not found")
	}

	_, err := getWincredToken()
	if err == nil {
		t.Error("expected error when credential not found")
	}
}

func TestWincredCommandError(t *testing.T) {
	origRunner := wincredCommandRunner
	defer func() { wincredCommandRunner = origRunner }()

	wincredCommandRunner = func(args ...string) (string, error) {
		return "", errors.New("cmdkey: command not found")
	}

	_, err := getWincredToken()
	if err == nil {
		t.Error("expected error on command failure")
	}
}
