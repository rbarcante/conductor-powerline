// Package oauth handles OAuth token retrieval from platform credential stores
// and API usage data fetching for the Anthropic usage endpoint.
package oauth

import (
	"errors"
	"runtime"
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

	switch runtime.GOOS {
	case "darwin":
		platformRetriever = keychainRetriever
	case "windows":
		platformRetriever = wincredRetriever
	case "linux":
		platformRetriever = secretoolRetriever
	}

	// Try platform-specific retriever first
	if platformRetriever != nil {
		token, err := platformRetriever()
		if err == nil {
			return token, nil
		}
	}

	// Fallback to credential file
	token, err := credfileRetriever()
	if err == nil {
		return token, nil
	}

	return "", errors.New("oauth: no token found in any credential source")
}
