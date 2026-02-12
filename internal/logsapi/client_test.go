package logsapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Subscribe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get(extensionIDHeader) != "ext-123" {
			t.Errorf("expected extension ID header ext-123, got %s", r.Header.Get(extensionIDHeader))
		}
		var req SubscribeRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Types) != 3 {
			t.Errorf("expected 3 types, got %d", len(req.Types))
		}
		if req.Buffering.MaxItems != 1000 {
			t.Errorf("expected MaxItems=1000, got %d", req.Buffering.MaxItems)
		}
		if req.Destination.URI != "http://sandbox.localdomain:8080" {
			t.Errorf("unexpected URI: %s", req.Destination.URI)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		extensionID: "ext-123",
	}
	err := c.Subscribe(context.Background(), "http://sandbox.localdomain:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Subscribe_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		extensionID: "ext-123",
	}
	err := c.Subscribe(context.Background(), "http://sandbox.localdomain:8080")
	if err == nil {
		t.Error("expected error on 500 response")
	}
}
