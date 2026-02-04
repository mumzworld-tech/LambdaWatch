package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const (
	extensionNameHeader = "Lambda-Extension-Name"
	extensionIDHeader   = "Lambda-Extension-Identifier"
)

// Client is a Lambda Extensions API client
type Client struct {
	baseURL      string
	httpClient   *http.Client
	extensionID  string
	extensionName string
}

// NewClient creates a new Extensions API client
func NewClient() *Client {
	runtimeAPI := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	extensionName := filepath.Base(os.Args[0])

	return &Client{
		baseURL:      fmt.Sprintf("http://%s/2020-01-01/extension", runtimeAPI),
		httpClient:   &http.Client{},
		extensionName: extensionName,
	}
}

// Register registers the extension with Lambda
func (c *Client) Register(ctx context.Context) (*RegisterResponse, error) {
	body := map[string][]string{
		"events": {string(Invoke), string(Shutdown)},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/register", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create register request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(extensionNameHeader, c.extensionName)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register extension: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("register failed with status: %d", resp.StatusCode)
	}

	c.extensionID = resp.Header.Get(extensionIDHeader)
	if c.extensionID == "" {
		return nil, fmt.Errorf("no extension ID in response")
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode register response: %w", err)
	}

	return &result, nil
}

// NextEvent blocks waiting for the next Lambda event
func (c *Client) NextEvent(ctx context.Context) (*NextEventResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/event/next", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create next event request: %w", err)
	}

	req.Header.Set(extensionIDHeader, c.extensionID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get next event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("next event failed with status: %d", resp.StatusCode)
	}

	var event NextEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode next event: %w", err)
	}

	return &event, nil
}

// GetExtensionID returns the extension identifier
func (c *Client) GetExtensionID() string {
	return c.extensionID
}
