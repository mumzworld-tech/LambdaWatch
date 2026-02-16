package loki

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mumzworld-tech/lambdawatch/internal/config"
)

func newTestConfig(endpoint string) *config.Config {
	return &config.Config{
		LokiEndpoint:         endpoint,
		EnableGzip:           true,
		CompressionThreshold: 1024,
		MaxRetries:           3,
		CriticalFlushRetries: 5,
	}
}

func newTestRequest() *PushRequest {
	return &PushRequest{
		Streams: []Stream{
			{
				Stream: map[string]string{"test": "label"},
				Values: [][]string{{"1234567890", "test message"}},
			},
		},
	}
}

// TC-5.1.1: Successful Push
func TestClient_Push_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
}

// TC-5.1.2: Push Empty Request
func TestClient_Push_EmptyRequest(t *testing.T) {
	client := NewClient(newTestConfig("http://unused"))
	err := client.Push(context.Background(), nil)

	if err != nil {
		t.Errorf("Push(nil) error = %v, want nil", err)
	}
}

// TC-5.1.3: Push Empty Streams
func TestClient_Push_EmptyStreams(t *testing.T) {
	client := NewClient(newTestConfig("http://unused"))
	err := client.Push(context.Background(), &PushRequest{Streams: []Stream{}})

	if err != nil {
		t.Errorf("Push(empty streams) error = %v, want nil", err)
	}
}

// TC-5.2.1: Retry on 500 Error
func TestClient_Push_RetryOn500(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

// TC-5.2.2: Retry on 429 (Rate Limited)
func TestClient_Push_RetryOn429(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
	if atomic.LoadInt32(&attempts) < 2 {
		t.Errorf("attempts = %d, want >= 2", attempts)
	}
}

// TC-5.2.3: No Retry on 400 (Bad Request)
func TestClient_Push_NoRetryOn400(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err == nil {
		t.Error("Push() error = nil, want error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry)", attempts)
	}
}

// TC-5.2.4: No Retry on 401 (Unauthorized)
func TestClient_Push_NoRetryOn401(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err == nil {
		t.Error("Push() error = nil, want error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry)", attempts)
	}
}

// TC-5.2.5: Max Retries Exhausted
func TestClient_Push_MaxRetriesExhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.MaxRetries = 3
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err == nil {
		t.Error("Push() error = nil, want error")
	}
	// 1 initial + 3 retries = 4 attempts
	if atomic.LoadInt32(&attempts) != 4 {
		t.Errorf("attempts = %d, want 4", attempts)
	}
	if !strings.Contains(err.Error(), "after 3 retries") {
		t.Errorf("error = %v, want 'after 3 retries'", err)
	}
}

// TC-5.2.6: Critical Push Higher Retries
func TestClient_PushCritical_HigherRetries(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.CriticalFlushRetries = 5
	client := NewClient(cfg)

	err := client.PushCritical(context.Background(), newTestRequest())

	if err == nil {
		t.Error("PushCritical() error = nil, want error")
	}
	// 1 initial + 5 retries = 6 attempts
	if atomic.LoadInt32(&attempts) != 6 {
		t.Errorf("attempts = %d, want 6", attempts)
	}
}

// TC-5.3.2: Context Cancellation During Backoff
func TestClient_Push_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := client.Push(ctx, newTestRequest())

	if err == nil {
		t.Error("Push() error = nil, want context error")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("error = %v, want context-related error", err)
	}
}

// TC-5.4.1: Gzip Enabled Above Threshold
func TestClient_Push_GzipAboveThreshold(t *testing.T) {
	var receivedContentEncoding string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentEncoding = r.Header.Get("Content-Encoding")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = true
	cfg.CompressionThreshold = 100
	client := NewClient(cfg)

	// Create request larger than threshold
	req := &PushRequest{
		Streams: []Stream{
			{
				Stream: map[string]string{"test": "label"},
				Values: [][]string{
					{"1234567890", strings.Repeat("a", 200)},
				},
			},
		},
	}

	err := client.Push(context.Background(), req)

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedContentEncoding != "gzip" {
		t.Errorf("Content-Encoding = %s, want gzip", receivedContentEncoding)
	}
}

// TC-5.4.2: No Compression Below Threshold
func TestClient_Push_NoGzipBelowThreshold(t *testing.T) {
	var receivedContentEncoding string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentEncoding = r.Header.Get("Content-Encoding")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = true
	cfg.CompressionThreshold = 10000 // High threshold
	client := NewClient(cfg)

	// Small request below threshold
	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedContentEncoding != "" {
		t.Errorf("Content-Encoding = %s, want empty", receivedContentEncoding)
	}
}

// TC-5.4.3: Gzip Disabled
func TestClient_Push_GzipDisabled(t *testing.T) {
	var receivedContentEncoding string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentEncoding = r.Header.Get("Content-Encoding")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = false
	client := NewClient(cfg)

	// Large request
	req := &PushRequest{
		Streams: []Stream{
			{
				Stream: map[string]string{"test": "label"},
				Values: [][]string{{"1234567890", strings.Repeat("a", 5000)}},
			},
		},
	}

	err := client.Push(context.Background(), req)

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedContentEncoding != "" {
		t.Errorf("Content-Encoding = %s, want empty (gzip disabled)", receivedContentEncoding)
	}
}

// TC-5.4.4: Compression Reduces Size
func TestClient_Push_CompressionReducesSize(t *testing.T) {
	var receivedSize int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedSize = len(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = true
	cfg.CompressionThreshold = 100
	client := NewClient(cfg)

	// Highly compressible content
	req := &PushRequest{
		Streams: []Stream{
			{
				Stream: map[string]string{"test": "label"},
				Values: [][]string{{"1234567890", strings.Repeat("a", 1000)}},
			},
		},
	}

	originalJSON, _ := json.Marshal(req)
	originalSize := len(originalJSON)

	err := client.Push(context.Background(), req)

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedSize >= originalSize {
		t.Errorf("Compressed size %d >= original %d", receivedSize, originalSize)
	}
}

// TC-5.5.1: Basic Auth Header
func TestClient_Push_BasicAuth(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.LokiUsername = "user"
	cfg.LokiPassword = "pass"
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if !strings.HasPrefix(receivedAuth, "Basic ") {
		t.Errorf("Authorization = %s, want Basic auth", receivedAuth)
	}
}

// TC-5.5.2: Bearer Token Header
func TestClient_Push_BearerToken(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.LokiAPIKey = "my-token"
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedAuth != "Bearer my-token" {
		t.Errorf("Authorization = %s, want 'Bearer my-token'", receivedAuth)
	}
}

// TC-5.5.3: Tenant ID Header
func TestClient_Push_TenantID(t *testing.T) {
	var receivedTenantID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTenantID = r.Header.Get("X-Scope-OrgID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.LokiTenantID = "tenant-123"
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedTenantID != "tenant-123" {
		t.Errorf("X-Scope-OrgID = %s, want 'tenant-123'", receivedTenantID)
	}
}

// TC-5.5.4: All Auth Combined
func TestClient_Push_AllAuthCombined(t *testing.T) {
	var receivedAuth, receivedTenantID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedTenantID = r.Header.Get("X-Scope-OrgID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.LokiAPIKey = "my-token"
	cfg.LokiTenantID = "tenant-123"
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedAuth != "Bearer my-token" {
		t.Errorf("Authorization = %s, want 'Bearer my-token'", receivedAuth)
	}
	if receivedTenantID != "tenant-123" {
		t.Errorf("X-Scope-OrgID = %s, want 'tenant-123'", receivedTenantID)
	}
}

// TC-5.5.5: No Auth
func TestClient_Push_NoAuth(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	// No auth configured
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedAuth != "" {
		t.Errorf("Authorization = %s, want empty", receivedAuth)
	}
}

// TC-5.6.2: Content-Type Header
func TestClient_Push_ContentType(t *testing.T) {
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(newTestConfig(server.URL))
	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type = %s, want 'application/json'", receivedContentType)
	}
}

// TC-5.6.3: Network Error
func TestClient_Push_NetworkError(t *testing.T) {
	cfg := newTestConfig("http://localhost:99999") // Invalid port
	cfg.MaxRetries = 0                             // No retries for faster test
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err == nil {
		t.Error("Push() error = nil, want network error")
	}
}

// TC-5.7.1: Valid JSON Body
func TestClient_Push_ValidJSONBody(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = false // Disable compression for this test
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	var received PushRequest
	if err := json.Unmarshal(receivedBody, &received); err != nil {
		t.Errorf("Invalid JSON body: %v", err)
	}
}

// TC-5.7.2: Body Preserved Across Retries
func TestClient_Push_BodyPreservedAcrossRetries(t *testing.T) {
	var bodies []string
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = false
	client := NewClient(cfg)

	err := client.Push(context.Background(), newTestRequest())

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(bodies))
	}
	if bodies[0] != bodies[1] {
		t.Errorf("Body changed between retries:\nFirst:  %s\nSecond: %s", bodies[0], bodies[1])
	}
}

// Test gzip body can be decompressed
func TestClient_Push_GzipBodyDecompresses(t *testing.T) {
	var receivedBody []byte
	var contentEncoding string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding = r.Header.Get("Content-Encoding")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := newTestConfig(server.URL)
	cfg.EnableGzip = true
	cfg.CompressionThreshold = 10 // Low threshold
	client := NewClient(cfg)

	req := &PushRequest{
		Streams: []Stream{
			{
				Stream: map[string]string{"test": "label"},
				Values: [][]string{{"1234567890", strings.Repeat("test message ", 100)}},
			},
		},
	}

	err := client.Push(context.Background(), req)

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if contentEncoding != "gzip" {
		t.Skip("Body not gzipped, skipping decompression test")
	}

	// Decompress and verify
	reader, err := gzip.NewReader(strings.NewReader(string(receivedBody)))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	var received PushRequest
	if err := json.Unmarshal(decompressed, &received); err != nil {
		t.Errorf("Decompressed body is not valid JSON: %v", err)
	}
}

// Test isRetryable function
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "retryable error",
			err:      &retryableError{err: io.EOF},
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      io.EOF,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.expected {
				t.Errorf("isRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test retryableError.Error()
func TestRetryableError_Error(t *testing.T) {
	err := &retryableError{err: io.EOF}
	if err.Error() != io.EOF.Error() {
		t.Errorf("Error() = %s, want %s", err.Error(), io.EOF.Error())
	}
}

// Test retryableError.Unwrap()
func TestRetryableError_Unwrap(t *testing.T) {
	err := &retryableError{err: io.EOF}
	if err.Unwrap() != io.EOF {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), io.EOF)
	}
}
