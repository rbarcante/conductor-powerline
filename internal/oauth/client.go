package oauth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

const version = "conductor-powerline/1.0.0"

// oauthClientID is the public client_id for Claude Code's OAuth application.
const oauthClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"

// oauthTokenURL is the token endpoint for refreshing OAuth tokens.
const oauthTokenURL = "https://console.anthropic.com/v1/oauth/token"

// refreshTimeout is the HTTP timeout for token refresh requests.
const refreshTimeout = 5 * time.Second

// tokenRefresher is the function used to refresh an OAuth token.
// Package-level variable for testability.
var tokenRefresher = defaultRefreshOAuthToken

// maxResponseBody is the maximum size of an API response body (64KB).
const maxResponseBody = 64 * 1024

// UsageFetcher defines the interface for fetching usage data.
type UsageFetcher interface {
	FetchUsageData(token string) (*UsageData, error)
}

// Client is an HTTP client for the Anthropic usage API endpoint.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// RateLimitError is returned when the API responds with HTTP 429.
type RateLimitError struct {
	RetryAfter time.Duration
	Body       string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("oauth: rate limited (429), retry after %v", e.RetryAfter)
}

// parseRetryAfter interprets a Retry-After header value as either a number of
// seconds or an HTTP-date (RFC1123). Returns 0 if the value is empty or unparseable.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(value); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := time.Parse(time.RFC1123, value); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

// RefreshError is returned when the token refresh endpoint responds with
// HTTP 400 or 401, indicating the refresh token is invalid or expired.
type RefreshError struct {
	StatusCode int
	Body       string
}

func (e *RefreshError) Error() string {
	return fmt.Sprintf("oauth: token refresh failed (HTTP %d): %s", e.StatusCode, e.Body)
}

// refreshRequest is the JSON body sent to the token refresh endpoint.
type refreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
}

// refreshResponse is the JSON response from the token refresh endpoint.
type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// defaultRefreshOAuthToken calls the production token endpoint.
func defaultRefreshOAuthToken(refreshToken string) (*TokenCredentials, error) {
	return makeRefreshOAuthToken(oauthTokenURL)(refreshToken)
}

// makeRefreshOAuthToken returns a refresh function targeting the given URL.
// This factory enables tests to use httptest servers.
func makeRefreshOAuthToken(baseURL string) func(string) (*TokenCredentials, error) {
	transport := &http.Transport{
		ForceAttemptHTTP2: false,
		TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		MaxIdleConns:      1,
	}
	client := &http.Client{
		Timeout:   refreshTimeout,
		Transport: transport,
	}

	return func(refreshToken string) (*TokenCredentials, error) {
		reqBody := refreshRequest{
			GrantType:    "refresh_token",
			RefreshToken: refreshToken,
			ClientID:     oauthClientID,
		}
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", baseURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", version)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))

		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
			return nil, &RefreshError{
				StatusCode: resp.StatusCode,
				Body:       string(respBody),
			}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("oauth: token refresh returned status %d", resp.StatusCode)
		}

		var tokenResp refreshResponse
		if err := json.Unmarshal(respBody, &tokenResp); err != nil {
			return nil, fmt.Errorf("oauth: failed to parse refresh response: %w", err)
		}

		return &TokenCredentials{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
		}, nil
	}
}

// NewClient creates a new API client with the given base URL and timeout.
// The transport forces HTTP/1.1 to avoid rate-limit issues observed with
// Go's default HTTP/2 in short-lived processes (new TLS handshake each run).
func NewClient(baseURL string, timeout time.Duration) *Client {
	transport := &http.Transport{
		ForceAttemptHTTP2: false,
		// Empty non-nil map prevents h2 ALPN negotiation — definitive HTTP/2 disable in Go.
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		MaxIdleConns: 1,
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// usageBucket represents a single usage bucket from the API response.
type usageBucket struct {
	ResetsAt    string  `json:"resets_at"`
	Utilization float64 `json:"utilization"`
}

// apiResponse mirrors the JSON structure from the Anthropic OAuth usage endpoint.
type apiResponse struct {
	FiveHour       *usageBucket `json:"five_hour"`
	SevenDay       *usageBucket `json:"seven_day"`
	SevenDayOpus   *usageBucket `json:"seven_day_opus"`
	SevenDaySonnet *usageBucket `json:"seven_day_sonnet"`
}

// FetchUsageData calls the Anthropic usage endpoint and returns structured usage data.
func (c *Client) FetchUsageData(token string) (*UsageData, error) {
	debug.Logf("api", "GET %s", c.baseURL)
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", version)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		debug.Logf("api", "HTTP error: %v", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	debug.Logf("api", "HTTP status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Always drain the body so the connection can be reused and we capture debug info.
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		debug.Logf("api", "error response body: %s", string(errBody))

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			debug.Logf("api", "rate limited (429), Retry-After: %v", retryAfter)
			return nil, &RateLimitError{
				RetryAfter: retryAfter,
				Body:       string(errBody),
			}
		}
		return nil, fmt.Errorf("oauth: API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, err
	}
	debug.Logf("api", "response body length: %d bytes", len(body))

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		debug.Logf("api", "JSON parse error: %v", err)
		return nil, err
	}

	data := &UsageData{
		FetchedAt: time.Now(),
	}

	if apiResp.FiveHour != nil {
		data.BlockPercentage = apiResp.FiveHour.Utilization
		if t, err := time.Parse(time.RFC3339, apiResp.FiveHour.ResetsAt); err == nil {
			data.BlockResetTime = t
		} else {
			debug.Logf("api", "failed to parse block reset time: %v", err)
		}
	}

	if apiResp.SevenDay != nil {
		data.WeeklyPercentage = apiResp.SevenDay.Utilization
		if t, err := time.Parse(time.RFC3339, apiResp.SevenDay.ResetsAt); err == nil {
			data.WeekResetTime = t
		} else {
			debug.Logf("api", "failed to parse weekly reset time: %v", err)
		}
	}

	if apiResp.SevenDayOpus != nil {
		data.OpusPercentage = apiResp.SevenDayOpus.Utilization
	}
	if apiResp.SevenDaySonnet != nil {
		data.SonnetPercentage = apiResp.SevenDaySonnet.Utilization
	}

	return data, nil
}
