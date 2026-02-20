package oauth

import (
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
