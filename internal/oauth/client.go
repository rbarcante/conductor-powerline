package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// NewClient creates a new API client with the given base URL and timeout.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
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
	defer resp.Body.Close()

	debug.Logf("api", "HTTP status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
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
