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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"blockUsage": {"percentUsed": 72.5, "resetAt": "2026-02-19T18:00:00Z"},
			"weeklyUsage": {"percentUsed": 45.0, "opusPercent": 30.0, "sonnetPercent": 15.0, "resetAt": "2026-02-23T00:00:00Z"}
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
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.FetchUsageData("test-token")
	if err == nil {
		t.Error("expected error on malformed JSON")
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
