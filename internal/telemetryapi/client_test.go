package telemetryapi

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
		if r.Header.Get(extensionIDHeader) != "ext-456" {
			t.Errorf("expected ext-456, got %s", r.Header.Get(extensionIDHeader))
		}
		var req SubscribeRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.SchemaVersion != "2022-07-01" {
			t.Errorf("expected schema 2022-07-01, got %s", req.SchemaVersion)
		}
		if len(req.Types) != 3 {
			t.Errorf("expected 3 types, got %d", len(req.Types))
		}
		if req.Buffering.MaxItems != 1000 {
			t.Errorf("expected MaxItems=1000, got %d", req.Buffering.MaxItems)
		}
		if req.Buffering.MaxBytes != 262144 {
			t.Errorf("expected MaxBytes=262144, got %d", req.Buffering.MaxBytes)
		}
		if req.Buffering.TimeoutMs != 100 {
			t.Errorf("expected TimeoutMs=100, got %d", req.Buffering.TimeoutMs)
		}
		if req.Destination.Protocol != "HTTP" {
			t.Errorf("expected HTTP protocol, got %s", req.Destination.Protocol)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL,
		httpClient:  &http.Client{},
		extensionID: "ext-456",
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
		extensionID: "ext-456",
	}
	err := c.Subscribe(context.Background(), "http://sandbox.localdomain:8080")
	if err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestClient_Subscribe_NetworkError(t *testing.T) {
	c := &Client{
		baseURL:     "http://localhost:1",
		httpClient:  &http.Client{},
		extensionID: "ext-456",
	}
	err := c.Subscribe(context.Background(), "http://sandbox.localdomain:8080")
	if err == nil {
		t.Error("expected error on network failure")
	}
}
