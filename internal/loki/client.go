package loki

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/config"
)

// Client is a Loki HTTP client
type Client struct {
	endpoint             string
	httpClient           *http.Client
	username             string
	password             string
	apiKey               string
	tenantID             string
	enableGzip           bool
	compressionThreshold int
	maxRetries           int
	criticalRetries      int
}

// NewClient creates a new Loki client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		endpoint:             cfg.LokiEndpoint,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
		username:             cfg.LokiUsername,
		password:             cfg.LokiPassword,
		apiKey:               cfg.LokiAPIKey,
		tenantID:             cfg.LokiTenantID,
		enableGzip:           cfg.EnableGzip,
		compressionThreshold: cfg.CompressionThreshold,
		maxRetries:           cfg.MaxRetries,
		criticalRetries:      cfg.CriticalFlushRetries,
	}
}

// Push sends a push request to Loki with retries (regular flush)
func (c *Client) Push(ctx context.Context, req *PushRequest) error {
	return c.push(ctx, req, false)
}

// PushCritical sends a push request with higher retry count (shutdown/runtimeDone)
func (c *Client) PushCritical(ctx context.Context, req *PushRequest) error {
	return c.push(ctx, req, true)
}

func (c *Client) push(ctx context.Context, req *PushRequest, isCritical bool) error {
	if req == nil || len(req.Streams) == 0 {
		return nil
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal push request: %w", err)
	}

	var body io.Reader
	var contentEncoding string

	// Only compress if enabled AND payload exceeds threshold
	if c.enableGzip && len(jsonBody) > c.compressionThreshold {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(jsonBody); err != nil {
			return fmt.Errorf("failed to gzip body: %w", err)
		}
		if err := gw.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
		body = &buf
		contentEncoding = "gzip"
	} else {
		body = bytes.NewReader(jsonBody)
	}

	return c.pushWithRetry(ctx, body, contentEncoding, isCritical)
}

func (c *Client) pushWithRetry(ctx context.Context, body io.Reader, contentEncoding string, isCritical bool) error {
	var lastErr error

	// Use higher retry count for critical flushes
	retries := c.maxRetries
	if isCritical {
		retries = c.criticalRetries
	}

	// Read body into buffer for retries
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms, ...
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.doPush(ctx, bytes.NewReader(bodyBytes), contentEncoding)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on non-retryable errors
		if !isRetryable(err) {
			return err
		}
	}

	return fmt.Errorf("push failed after %d retries: %w", retries, lastErr)
}

func (c *Client) doPush(ctx context.Context, body io.Reader, contentEncoding string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if contentEncoding != "" {
		req.Header.Set("Content-Encoding", contentEncoding)
	}

	// Set authentication
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Set tenant ID for multi-tenant Loki
	if c.tenantID != "" {
		req.Header.Set("X-Scope-OrgID", c.tenantID)
	}

	log.Printf("Sending HTTP request to Loki...")
	resp, err := c.httpClient.Do(req)
	log.Printf("HTTP request completed, err: %v", err)
	if err != nil {
		return &retryableError{err: fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Loki response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Loki push successful: %d", resp.StatusCode)
		return nil
	}

	err = fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(respBody))
	log.Printf("Loki push failed: %v", err)

	// Retry on 429 (rate limited) or 5xx (server errors)
	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		return &retryableError{err: err}
	}

	return err
}

type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func isRetryable(err error) bool {
	_, ok := err.(*retryableError)
	return ok
}
