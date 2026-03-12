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

// --- GetCredentials tests ---

func TestGetCredentials_PlatformRetriever(t *testing.T) {
	origGetter := credentialsGetter
	defer func() { credentialsGetter = origGetter }()

	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "platform-token", RefreshToken: "platform-refresh"}, nil
	}

	creds, err := GetCredentials()
	if err != nil {
		t.Fatalf("expected credentials, got error: %v", err)
	}
	if creds.AccessToken != "platform-token" {
		t.Errorf("expected platform-token, got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "platform-refresh" {
		t.Errorf("expected platform-refresh, got %q", creds.RefreshToken)
	}
}

func TestGetCredentials_FallbackToCredfile(t *testing.T) {
	origKeychain := keychainCredentialsRetriever
	origCredfile := credfileCredentialsRetriever
	defer func() {
		keychainCredentialsRetriever = origKeychain
		credfileCredentialsRetriever = origCredfile
	}()

	keychainCredentialsRetriever = func() (*TokenCredentials, error) {
		return nil, errors.New("no keychain")
	}
	credfileCredentialsRetriever = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "credfile-token", RefreshToken: "credfile-refresh"}, nil
	}

	// Reset credentialsGetter to use the real dispatch logic
	origGetter := credentialsGetter
	defer func() { credentialsGetter = origGetter }()
	credentialsGetter = getCredentialsDefault

	creds, err := GetCredentials()
	if err != nil {
		t.Fatalf("expected credentials, got error: %v", err)
	}
	if creds.AccessToken != "credfile-token" {
		t.Errorf("expected credfile-token, got %q", creds.AccessToken)
	}
}

func TestGetCredentials_RotatedTokenFirst(t *testing.T) {
	dir := t.TempDir()

	// Store a rotated token
	if err := StoreRotatedToken(dir, &TokenCredentials{
		AccessToken: "rotated-access", RefreshToken: "rotated-refresh",
	}); err != nil {
		t.Fatal(err)
	}

	origDir := rotatedTokenDir
	defer func() { rotatedTokenDir = origDir }()
	rotatedTokenDir = dir

	origGetter := credentialsGetter
	defer func() { credentialsGetter = origGetter }()
	credentialsGetter = getCredentialsDefault

	// Mock all other sources to fail
	origKeychain := keychainCredentialsRetriever
	origCredfile := credfileCredentialsRetriever
	defer func() {
		keychainCredentialsRetriever = origKeychain
		credfileCredentialsRetriever = origCredfile
	}()
	keychainCredentialsRetriever = func() (*TokenCredentials, error) {
		return nil, errors.New("no keychain")
	}
	credfileCredentialsRetriever = func() (*TokenCredentials, error) {
		return nil, errors.New("no credfile")
	}

	creds, err := GetCredentials()
	if err != nil {
		t.Fatalf("expected rotated credentials, got error: %v", err)
	}
	if creds.AccessToken != "rotated-access" {
		t.Errorf("expected rotated-access, got %q", creds.AccessToken)
	}
}

func TestGetCredentials_AllFail(t *testing.T) {
	origKeychain := keychainCredentialsRetriever
	origCredfile := credfileCredentialsRetriever
	origWincred := wincredRetriever
	origSecretool := secretoolRetriever
	defer func() {
		keychainCredentialsRetriever = origKeychain
		credfileCredentialsRetriever = origCredfile
		wincredRetriever = origWincred
		secretoolRetriever = origSecretool
	}()

	keychainCredentialsRetriever = func() (*TokenCredentials, error) {
		return nil, errors.New("no keychain")
	}
	wincredRetriever = func() (string, error) { return "", errors.New("no wincred") }
	secretoolRetriever = func() (string, error) { return "", errors.New("no secretool") }
	credfileCredentialsRetriever = func() (*TokenCredentials, error) {
		return nil, errors.New("no credfile")
	}

	origGetter := credentialsGetter
	defer func() { credentialsGetter = origGetter }()
	credentialsGetter = getCredentialsDefault

	_, err := GetCredentials()
	if err == nil {
		t.Error("expected error when all credential sources fail")
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
