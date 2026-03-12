package oauth

import (
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
	body       string // unexported: contains full API response, not for external consumption
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

// NewClient creates a new API client with the given base URL and timeout.
// The transport forces HTTP/1.1 to avoid rate-limit issues observed with
// Go's default HTTP/2 in short-lived processes (new TLS handshake each run).
func NewClient(baseURL string, timeout time.Duration) *Client {
	transport := &http.Transport{
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
	req.Header.Set("Accept", "application/json")
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
				body:       string(errBody),
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

	return mapAPIResponse(&apiResp), nil
}

// mapAPIResponse converts an apiResponse into a UsageData struct.
func mapAPIResponse(apiResp *apiResponse) *UsageData {
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

	return data
}
