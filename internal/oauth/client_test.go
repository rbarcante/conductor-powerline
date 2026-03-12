package oauth

import (
	"errors"
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
	// Body is unexported — verify via Error() that the error is well-formed.
	if rle.Error() == "" {
		t.Error("expected non-empty error message from RateLimitError")
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal_error","message":"something went wrong"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
	// The body is drained internally (for connection reuse) — we just verify no panic/hang.
}

func TestMapAPIResponse(t *testing.T) {
	t.Run("full response", func(t *testing.T) {
		resp := &apiResponse{
			FiveHour:       &usageBucket{ResetsAt: "2026-02-19T18:00:00Z", Utilization: 72.5},
			SevenDay:       &usageBucket{ResetsAt: "2026-02-23T00:00:00Z", Utilization: 45.0},
			SevenDayOpus:   &usageBucket{ResetsAt: "2026-02-23T00:00:00Z", Utilization: 30.0},
			SevenDaySonnet: &usageBucket{ResetsAt: "2026-02-23T00:00:00Z", Utilization: 15.0},
		}
		data := mapAPIResponse(resp)
		if data.BlockPercentage != 72.5 {
			t.Errorf("expected block 72.5, got %f", data.BlockPercentage)
		}
		if data.WeeklyPercentage != 45.0 {
			t.Errorf("expected weekly 45.0, got %f", data.WeeklyPercentage)
		}
		if data.OpusPercentage != 30.0 {
			t.Errorf("expected opus 30.0, got %f", data.OpusPercentage)
		}
		if data.SonnetPercentage != 15.0 {
			t.Errorf("expected sonnet 15.0, got %f", data.SonnetPercentage)
		}
		if data.BlockResetTime.IsZero() {
			t.Error("expected non-zero BlockResetTime")
		}
		if data.FetchedAt.IsZero() {
			t.Error("expected non-zero FetchedAt")
		}
	})

	t.Run("nil buckets", func(t *testing.T) {
		resp := &apiResponse{}
		data := mapAPIResponse(resp)
		if data.BlockPercentage != 0 {
			t.Errorf("expected block 0, got %f", data.BlockPercentage)
		}
		if data.WeeklyPercentage != 0 {
			t.Errorf("expected weekly 0, got %f", data.WeeklyPercentage)
		}
	})

	t.Run("malformed resets_at", func(t *testing.T) {
		resp := &apiResponse{
			FiveHour: &usageBucket{ResetsAt: "not-a-date", Utilization: 60.0},
		}
		data := mapAPIResponse(resp)
		if data.BlockPercentage != 60.0 {
			t.Errorf("expected block 60.0, got %f", data.BlockPercentage)
		}
		if !data.BlockResetTime.IsZero() {
			t.Errorf("expected zero BlockResetTime for malformed date, got %v", data.BlockResetTime)
		}
	})
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

func TestRateLimitError_Error(t *testing.T) {
	rle := &RateLimitError{RetryAfter: 30 * time.Second, body: "rate limited"}
	msg := rle.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
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
