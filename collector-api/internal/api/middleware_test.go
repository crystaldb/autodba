package api_test

import (
	"collector-api/internal/api"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_DebugEnabled(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the LoggingMiddleware with debug enabled
	debug := true
	testServer := httptest.NewServer(api.LoggingMiddleware(handler, debug))
	defer testServer.Close()

	// Make a test request
	resp, err := http.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to make test request: %v", err)
	}

	// Ensure the response is OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestLoggingMiddleware_DebugDisabled(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the LoggingMiddleware with debug disabled
	debug := false
	testServer := httptest.NewServer(api.LoggingMiddleware(handler, debug))
	defer testServer.Close()

	// Make a test request
	resp, err := http.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to make test request: %v", err)
	}

	// Ensure the response is OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
