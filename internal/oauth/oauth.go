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
			debug.Logf("token", "%s retriever succeeded (token length=%d)", platformName, len(token))
			return token, nil
		}
		debug.Logf("token", "%s retriever failed: %v", platformName, err)
	}

	// Fallback to credential file
	debug.Logf("token", "trying credfile fallback")
	token, err := credfileRetriever()
	if err == nil {
		debug.Logf("token", "credfile retriever succeeded (token length=%d)", len(token))
		return token, nil
	}
	debug.Logf("token", "credfile retriever failed: %v", err)

	return "", errors.New("oauth: no token found in any credential source")
}
