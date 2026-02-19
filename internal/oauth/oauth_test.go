package oauth

import (
	"errors"
	"runtime"
	"testing"
)

func TestGetTokenPlatformDispatch(t *testing.T) {
	// Save and restore originals
	origKeychain := keychainRetriever
	origWincred := wincredRetriever
	origSecretool := secretoolRetriever
	origCredfile := credfileRetriever
	defer func() {
		keychainRetriever = origKeychain
		wincredRetriever = origWincred
		secretoolRetriever = origSecretool
		credfileRetriever = origCredfile
	}()

	// Mock all retrievers to fail
	keychainRetriever = func() (string, error) { return "", errors.New("no keychain") }
	wincredRetriever = func() (string, error) { return "", errors.New("no wincred") }
	secretoolRetriever = func() (string, error) { return "", errors.New("no secretool") }
	credfileRetriever = func() (string, error) { return "", errors.New("no credfile") }

	// Set the platform retriever to return a token
	switch runtime.GOOS {
	case "darwin":
		keychainRetriever = func() (string, error) { return "macos-token", nil }
	case "windows":
		wincredRetriever = func() (string, error) { return "windows-token", nil }
	case "linux":
		secretoolRetriever = func() (string, error) { return "linux-token", nil }
	}

	token, err := GetToken()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGetTokenFallbackToCredfile(t *testing.T) {
	origKeychain := keychainRetriever
	origWincred := wincredRetriever
	origSecretool := secretoolRetriever
	origCredfile := credfileRetriever
	defer func() {
		keychainRetriever = origKeychain
		wincredRetriever = origWincred
		secretoolRetriever = origSecretool
		credfileRetriever = origCredfile
	}()

	// All platform retrievers fail
	keychainRetriever = func() (string, error) { return "", errors.New("no keychain") }
	wincredRetriever = func() (string, error) { return "", errors.New("no wincred") }
	secretoolRetriever = func() (string, error) { return "", errors.New("no secretool") }

	// Credfile succeeds
	credfileRetriever = func() (string, error) { return "fallback-token", nil }

	token, err := GetToken()
	if err != nil {
		t.Fatalf("expected fallback token, got error: %v", err)
	}
	if token != "fallback-token" {
		t.Errorf("expected 'fallback-token', got %q", token)
	}
}

func TestGetTokenAllFail(t *testing.T) {
	origKeychain := keychainRetriever
	origWincred := wincredRetriever
	origSecretool := secretoolRetriever
	origCredfile := credfileRetriever
	defer func() {
		keychainRetriever = origKeychain
		wincredRetriever = origWincred
		secretoolRetriever = origSecretool
		credfileRetriever = origCredfile
	}()

	keychainRetriever = func() (string, error) { return "", errors.New("no keychain") }
	wincredRetriever = func() (string, error) { return "", errors.New("no wincred") }
	secretoolRetriever = func() (string, error) { return "", errors.New("no secretool") }
	credfileRetriever = func() (string, error) { return "", errors.New("no credfile") }

	_, err := GetToken()
	if err == nil {
		t.Error("expected error when all retrievers fail")
	}
}
