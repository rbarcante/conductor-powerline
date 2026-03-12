package oauth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", auth)
		}
		// Verify required headers
		if beta := r.Header.Get("anthropic-beta"); beta != "oauth-2025-04-20" {
			t.Errorf("expected anthropic-beta header, got %q", beta)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"five_hour": {"resets_at": "2026-02-19T18:00:00Z", "utilization": 72.5},
			"seven_day": {"resets_at": "2026-02-23T00:00:00Z", "utilization": 45.0},
			"seven_day_opus": {"resets_at": "2026-02-23T00:00:00Z", "utilization": 30.0},
			"seven_day_sonnet": {"resets_at": "2026-02-23T00:00:00Z", "utilization": 15.0}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	data, err := client.FetchUsageData("test-token")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if data.BlockPercentage != 72.5 {
		t.Errorf("expected block percentage 72.5, got %f", data.BlockPercentage)
	}
	if data.WeeklyPercentage != 45.0 {
		t.Errorf("expected weekly percentage 45.0, got %f", data.WeeklyPercentage)
	}
	if data.OpusPercentage != 30.0 {
		t.Errorf("expected opus percentage 30.0, got %f", data.OpusPercentage)
	}
	if data.SonnetPercentage != 15.0 {
		t.Errorf("expected sonnet percentage 15.0, got %f", data.SonnetPercentage)
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 50*time.Millisecond)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClientHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("bad-token")
	if err == nil {
		t.Error("expected error on 401 response")
	}
}

func TestClientMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Error("expected error on malformed JSON")
	}
}

func TestClientNullBuckets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Real API returns null for opus when not used
		_, _ = w.Write([]byte(`{
			"five_hour": {"resets_at": "2026-02-19T18:00:00Z", "utilization": 18.0},
			"seven_day": {"resets_at": "2026-02-23T00:00:00Z", "utilization": 14.0},
			"seven_day_opus": null,
			"seven_day_sonnet": {"resets_at": "2026-02-23T00:00:00Z", "utilization": 10.0}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	data, err := client.FetchUsageData("test-token")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if data.BlockPercentage != 18.0 {
		t.Errorf("expected block 18.0, got %f", data.BlockPercentage)
	}
	if data.OpusPercentage != 0 {
		t.Errorf("expected opus 0 (null bucket), got %f", data.OpusPercentage)
	}
	if data.SonnetPercentage != 10.0 {
		t.Errorf("expected sonnet 10.0, got %f", data.SonnetPercentage)
	}
}

func TestClientPartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Only five_hour present
		_, _ = w.Write([]byte(`{"five_hour": {"resets_at": "2026-02-19T18:00:00Z", "utilization": 50.0}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	data, err := client.FetchUsageData("test-token")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if data.BlockPercentage != 50.0 {
		t.Errorf("expected block 50.0, got %f", data.BlockPercentage)
	}
	if data.WeeklyPercentage != 0 {
		t.Errorf("expected weekly 0 (missing), got %f", data.WeeklyPercentage)
	}
}

func TestClientMalformedResetsAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"five_hour": {"resets_at": "not-a-date", "utilization": 60.0},
			"seven_day": {"resets_at": "also-invalid", "utilization": 30.0}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	data, err := client.FetchUsageData("test-token")
	if err != nil {
		t.Fatalf("expected success despite malformed timestamps, got error: %v", err)
	}

	if data.BlockPercentage != 60.0 {
		t.Errorf("expected block percentage 60.0, got %f", data.BlockPercentage)
	}
	if data.WeeklyPercentage != 30.0 {
		t.Errorf("expected weekly percentage 30.0, got %f", data.WeeklyPercentage)
	}
	if !data.BlockResetTime.IsZero() {
		t.Errorf("expected zero BlockResetTime for malformed timestamp, got %v", data.BlockResetTime)
	}
	if !data.WeekResetTime.IsZero() {
		t.Errorf("expected zero WeekResetTime for malformed timestamp, got %v", data.WeekResetTime)
	}
}

func TestClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestClient429ReturnsRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Fatal("expected error on 429 response")
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rle.RetryAfter != 30*time.Second {
		t.Errorf("expected RetryAfter 30s, got %v", rle.RetryAfter)
	}
	if rle.Body != `{"error":"rate_limited"}` {
		t.Errorf("expected body captured, got %q", rle.Body)
	}
}

func TestClient429NoRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rle.RetryAfter != 0 {
		t.Errorf("expected RetryAfter 0, got %v", rle.RetryAfter)
	}
}

func TestClientNon200DrainsBody(t *testing.T) {
	bodySent := `{"error":"internal_error","message":"something went wrong"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(bodySent))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
	// The body is drained internally (for connection reuse) — we just verify no panic/hang.
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"numeric seconds", "60", 60 * time.Second},
		{"zero", "0", 0},
		{"empty", "", 0},
		{"garbage", "not-a-number", 0},
		{"negative", "-5", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// --- RefreshOAuthToken tests ---

func TestRefreshError_Error(t *testing.T) {
	re := &RefreshError{StatusCode: 400, Body: "invalid_grant"}
	msg := re.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRateLimitError_Error(t *testing.T) {
	rle := &RateLimitError{RetryAfter: 30 * time.Second, Body: "rate limited"}
	msg := rle.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRefreshOAuthToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if req["grant_type"] != "refresh_token" {
			t.Errorf("expected grant_type=refresh_token, got %q", req["grant_type"])
		}
		if req["refresh_token"] != "old-refresh-token" {
			t.Errorf("expected refresh_token=old-refresh-token, got %q", req["refresh_token"])
		}
		if req["client_id"] != oauthClientID {
			t.Errorf("expected client_id=%s, got %q", oauthClientID, req["client_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"new-refresh"}`))
	}))
	defer server.Close()

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = makeRefreshOAuthToken(server.URL)

	creds, err := tokenRefresher("old-refresh-token")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if creds.AccessToken != "new-access" {
		t.Errorf("expected new-access, got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "new-refresh" {
		t.Errorf("expected new-refresh, got %q", creds.RefreshToken)
	}
}

func TestRefreshOAuthToken_InvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = makeRefreshOAuthToken(server.URL)

	_, err := tokenRefresher("bad-refresh-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	var re *RefreshError
	if !errors.As(err, &re) {
		t.Fatalf("expected *RefreshError, got %T: %v", err, err)
	}
	if re.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", re.StatusCode)
	}
}

func TestRefreshOAuthToken_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = makeRefreshOAuthToken(server.URL)

	_, err := tokenRefresher("bad-refresh-token")
	var re *RefreshError
	if !errors.As(err, &re) {
		t.Fatalf("expected *RefreshError, got %T: %v", err, err)
	}
	if re.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", re.StatusCode)
	}
}

func TestRefreshOAuthToken_NetworkError(t *testing.T) {
	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = makeRefreshOAuthToken("http://localhost:1") // connection refused

	_, err := tokenRefresher("some-token")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
	var re *RefreshError
	if errors.As(err, &re) {
		t.Error("expected generic error, not RefreshError")
	}
}

func TestRefreshOAuthToken_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = makeRefreshOAuthToken(server.URL)

	_, err := tokenRefresher("some-token")
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	// Use a date far in the future so time.Until returns > 0.
	future := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC1123)
	got := parseRetryAfter(future)
	if got <= 0 {
		t.Errorf("expected positive duration for future HTTP-date, got %v", got)
	}
	if got > 3*time.Hour {
		t.Errorf("expected roughly 2h, got %v", got)
	}
}
