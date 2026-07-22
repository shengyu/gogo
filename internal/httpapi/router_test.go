package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGreeting(t *testing.T) {
	router := NewRouter(RouterOptions{
		Environment: "test",
		Version:     "1.0.0",
		Commit:      "abc123",
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/greetings?name=Steven", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["message"] != "Hello, Steven!" {
		t.Fatalf("unexpected message: %q", body["message"])
	}
	if body["environment"] != "test" {
		t.Fatalf("unexpected environment: %q", body["environment"])
	}
}

func TestHealth(t *testing.T) {
	router := NewRouter(RouterOptions{Environment: "test"})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
