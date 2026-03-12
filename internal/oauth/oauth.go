// Package oauth handles OAuth token retrieval from platform credential stores
// and API usage data fetching for the Anthropic usage endpoint.
package oauth

import (
	"errors"
	"runtime"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// Retriever function variables for platform-specific token retrieval.
// These are package-level variables to allow testing with mocks.
var (
	keychainRetriever  = getKeychainToken
	wincredRetriever   = getWincredToken
	secretoolRetriever = getSecretoolToken
	credfileRetriever  = getCredfileToken
)

// Credentials retriever function variables — return TokenCredentials with
// both access token and refresh token. Used by the token rotation flow.
var (
	keychainCredentialsRetriever = getKeychainCredentials
	credfileCredentialsRetriever = getCredfileCredentials
	credentialsGetter            = getCredentialsDefault
)

// GetToken retrieves the Claude OAuth token by trying the platform credential
// store first (based on runtime.GOOS), then falling back to the credential file.
// Returns the token string or an error if all sources fail.
func GetToken() (string, error) {
	var platformRetriever func() (string, error)
	var platformName string

	switch runtime.GOOS {
	case "darwin":
		platformRetriever = keychainRetriever
		platformName = "keychain"
	case "windows":
		platformRetriever = wincredRetriever
		platformName = "wincred"
	case "linux":
		platformRetriever = secretoolRetriever
		platformName = "secret-tool"
	}

	debug.Logf("token", "platform=%s, trying %s retriever", runtime.GOOS, platformName)

	// Try platform-specific retriever first
	if platformRetriever != nil {
		token, err := platformRetriever()
		if err == nil {
			debug.Logf("token", "%s retriever succeeded", platformName)
			return token, nil
		}
		debug.Logf("token", "%s retriever failed: %v", platformName, err)
	}

	// Fallback to credential file
	debug.Logf("token", "trying credfile fallback")
	token, err := credfileRetriever()
	if err == nil {
		debug.Logf("token", "credfile retriever succeeded")
		return token, nil
	}
	debug.Logf("token", "credfile retriever failed: %v", err)

	return "", errors.New("oauth: no token found in any credential source")
}

// GetCredentials retrieves OAuth credentials including the refresh token.
// It delegates to the package-level credentialsGetter, which can be replaced
// for testing.
func GetCredentials() (*TokenCredentials, error) {
	return credentialsGetter()
}

// getCredentialsDefault is the real implementation of credential retrieval.
// Checks rotated token file first, then platform-specific stores, then credfile.
func getCredentialsDefault() (*TokenCredentials, error) {
	// Check for a previously rotated token first
	if rotatedTokenDir != "" {
		creds, err := LoadRotatedToken(rotatedTokenDir)
		if err == nil && creds != nil {
			debug.Logf("creds", "using rotated token from disk (hasRefresh=%v)", creds.RefreshToken != "")
			return creds, nil
		}
	}

	var platformName string

	switch runtime.GOOS {
	case "darwin":
		platformName = "keychain"
		debug.Logf("creds", "trying keychain credentials retriever")
		creds, err := keychainCredentialsRetriever()
		if err == nil {
			debug.Logf("creds", "keychain credentials retrieved (hasRefresh=%v)", creds.RefreshToken != "")
			return creds, nil
		}
		debug.Logf("creds", "keychain credentials failed: %v", err)
	case "windows":
		platformName = "wincred"
		debug.Logf("creds", "trying wincred (plain token only)")
		token, err := wincredRetriever()
		if err == nil {
			debug.Logf("creds", "wincred token retrieved")
			return &TokenCredentials{AccessToken: token}, nil
		}
		debug.Logf("creds", "wincred failed: %v", err)
	case "linux":
		platformName = "secret-tool"
		debug.Logf("creds", "trying secret-tool (plain token only)")
		token, err := secretoolRetriever()
		if err == nil {
			debug.Logf("creds", "secret-tool token retrieved")
			return &TokenCredentials{AccessToken: token}, nil
		}
		debug.Logf("creds", "secret-tool failed: %v", err)
	default:
		platformName = "unknown"
	}

	_ = platformName

	// Fallback to credential file
	debug.Logf("creds", "trying credfile fallback")
	creds, err := credfileCredentialsRetriever()
	if err == nil {
		debug.Logf("creds", "credfile credentials retrieved (hasRefresh=%v)", creds.RefreshToken != "")
		return creds, nil
	}
	debug.Logf("creds", "credfile failed: %v", err)

	return nil, errors.New("oauth: no credentials found in any source")
}
