package radiobrowser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"
)

// apiJSON returns a minimal JSON array that matches the radio-browser wire format.
const singleStationJSON = `[{
	"stationuuid": "uuid-1",
	"name":        "Test FM",
	"url":         "http://stream.test/1",
	"url_resolved":"http://stream.test/resolved",
	"countrycode": "US",
	"codec":       "MP3",
	"bitrate":     128,
	"votes":       42,
	"lastcheckok": 1
}]`

func TestClient_FetchStations_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/stations/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleStationJSON))
	}))
	defer srv.Close()

	client := radiobrowser.NewClient(srv.URL)
	got, err := client.FetchStations(context.Background(), radiobrowser.FetchParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 station, got %d", len(got))
	}
	if got[0].UUID != "uuid-1" {
		t.Errorf("UUID: got %q, want %q", got[0].UUID, "uuid-1")
	}
	if got[0].Name != "Test FM" {
		t.Errorf("Name: got %q, want %q", got[0].Name, "Test FM")
	}
	if !got[0].LastCheckOK {
		t.Errorf("LastCheckOK: expected true (converted from lastcheckok=1)")
	}
}

func TestClient_FetchStations_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	client := radiobrowser.NewClient(srv.URL)
	got, err := client.FetchStations(context.Background(), radiobrowser.FetchParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d stations", len(got))
	}
}

func TestClient_FetchStations_NonTransientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// 400 is not a transient error — should fail immediately, no retry.
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	// WithMaxRetries(0) just to be explicit; 400 would not be retried anyway.
	client := radiobrowser.NewClient(srv.URL, radiobrowser.WithMaxRetries(0))
	_, err := client.FetchStations(context.Background(), radiobrowser.FetchParams{Limit: 10})
	if err == nil {
		t.Fatal("expected error for 400 response, got nil")
	}
}

func TestClient_FetchStations_RetriesOnTransient(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			// First two calls fail with a transient 503.
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Third call succeeds.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleStationJSON))
	}))
	defer srv.Close()

	// Allow up to 3 retries (4 attempts total); use a zero-delay HTTP client
	// by overriding with a fast client — backoff is the only delay.
	client := radiobrowser.NewClient(srv.URL,
		radiobrowser.WithMaxRetries(3),
		radiobrowser.WithHTTPClient(&http.Client{}),
	)

	got, err := client.FetchStations(context.Background(), radiobrowser.FetchParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 station after retry, got %d", len(got))
	}
	if calls.Load() != 3 {
		t.Errorf("expected 3 total attempts, got %d", calls.Load())
	}
}

func TestClient_FetchStations_ExhaustsRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := radiobrowser.NewClient(srv.URL, radiobrowser.WithMaxRetries(2))
	_, err := client.FetchStations(context.Background(), radiobrowser.FetchParams{Limit: 10})
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
}
