package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// usageFetcher defines the interface for fetching usage data.
type usageFetcher interface {
	FetchUsageData(token string) (*UsageData, error)
}

// Client is an HTTP client for the Anthropic usage API endpoint.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client with the given base URL and timeout.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// apiResponse mirrors the JSON structure from the Anthropic usage endpoint.
type apiResponse struct {
	BlockUsage struct {
		PercentUsed float64 `json:"percentUsed"`
		ResetAt     string  `json:"resetAt"`
	} `json:"blockUsage"`
	WeeklyUsage struct {
		PercentUsed   float64 `json:"percentUsed"`
		OpusPercent   float64 `json:"opusPercent"`
		SonnetPercent float64 `json:"sonnetPercent"`
		ResetAt       string  `json:"resetAt"`
	} `json:"weeklyUsage"`
}

// FetchUsageData calls the Anthropic usage endpoint and returns structured usage data.
func (c *Client) FetchUsageData(token string) (*UsageData, error) {
	debug.Logf("api", "GET %s", c.baseURL)
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		debug.Logf("api", "HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	debug.Logf("api", "HTTP status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth: API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	debug.Logf("api", "response body length: %d bytes", len(body))

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		debug.Logf("api", "JSON parse error: %v", err)
		return nil, err
	}

	blockReset, _ := time.Parse(time.RFC3339, apiResp.BlockUsage.ResetAt)
	weekReset, _ := time.Parse(time.RFC3339, apiResp.WeeklyUsage.ResetAt)

	return &UsageData{
		BlockPercentage:  apiResp.BlockUsage.PercentUsed,
		BlockResetTime:   blockReset,
		WeeklyPercentage: apiResp.WeeklyUsage.PercentUsed,
		OpusPercentage:   apiResp.WeeklyUsage.OpusPercent,
		SonnetPercentage: apiResp.WeeklyUsage.SonnetPercent,
		WeekResetTime:    weekReset,
		FetchedAt:        time.Now(),
	}, nil
}
