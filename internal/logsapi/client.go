package logsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	extensionIDHeader = "Lambda-Extension-Identifier"
	logsAPIVersion    = "2020-08-15"
)

// Client is a Lambda Logs API client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	extensionID string
}

// NewClient creates a new Logs API client
func NewClient(extensionID string) *Client {
	runtimeAPI := os.Getenv("AWS_LAMBDA_RUNTIME_API")

	return &Client{
		baseURL:     fmt.Sprintf("http://%s/%s/logs", runtimeAPI, logsAPIVersion),
		httpClient:  &http.Client{},
		extensionID: extensionID,
	}
}

// Subscribe subscribes to the Lambda Logs API
func (c *Client) Subscribe(ctx context.Context, listenerURI string) error {
	req := SubscribeRequest{
		SchemaVersion: "2021-03-18",
		Types:         []string{LogTypePlatform, LogTypeFunction, LogTypeExtension},
		Buffering: BufferConfig{
			MaxItems:  1000,
			MaxBytes:  262144,
			TimeoutMs: 100,
		},
		Destination: Destination{
			Protocol: "HTTP",
			URI:      listenerURI,
		},
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create subscribe request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(extensionIDHeader, c.extensionID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscribe failed with status: %d", resp.StatusCode)
	}

	return nil
}
