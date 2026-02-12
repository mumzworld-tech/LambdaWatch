package config

import (
	"os"
	"testing"
)

// Helper to set environment variables and clean up after test
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, exists := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if exists {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, exists := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if exists {
			os.Setenv(key, old)
		}
	})
}

// clearAllEnvVars clears all LambdaWatch-related env vars
func clearAllEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{
		"LOKI_URL", "LOKI_USERNAME", "LOKI_PASSWORD", "LOKI_API_KEY",
		"LOKI_TENANT_ID", "LOKI_BATCH_SIZE", "LOKI_MAX_BATCH_SIZE_BYTES",
		"LOKI_FLUSH_INTERVAL_MS", "LOKI_IDLE_FLUSH_MULTIPLIER", "LOKI_MAX_RETRIES",
		"LOKI_CRITICAL_FLUSH_RETRIES", "LOKI_ENABLE_GZIP", "LOKI_COMPRESSION_THRESHOLD",
		"LOKI_LABELS", "BUFFER_SIZE", "LOKI_MAX_LINE_SIZE", "LOKI_EXTRACT_REQUEST_ID",
		"SERVICE_NAME",
	}
	for _, v := range vars {
		unsetEnv(t, v)
	}
}

// TC-1.1.2: Valid LOKI_URL
func TestLoad_ValidLokiURL(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com/loki/api/v1/push")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LokiEndpoint != "https://loki.example.com/loki/api/v1/push" {
		t.Errorf("LokiEndpoint = %v, want https://loki.example.com/loki/api/v1/push", cfg.LokiEndpoint)
	}
}

// TC-1.2.1: Basic Auth
func TestLoad_BasicAuth(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_USERNAME", "user123")
	setEnv(t, "LOKI_PASSWORD", "pass456")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LokiUsername != "user123" {
		t.Errorf("LokiUsername = %v, want user123", cfg.LokiUsername)
	}
	if cfg.LokiPassword != "pass456" {
		t.Errorf("LokiPassword = %v, want pass456", cfg.LokiPassword)
	}
}

// TC-1.2.2: Bearer Token Auth
func TestLoad_BearerTokenAuth(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_API_KEY", "glc_xxx123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LokiAPIKey != "glc_xxx123" {
		t.Errorf("LokiAPIKey = %v, want glc_xxx123", cfg.LokiAPIKey)
	}
}

// TC-1.2.3: Tenant ID
func TestLoad_TenantID(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_TENANT_ID", "my-tenant")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LokiTenantID != "my-tenant" {
		t.Errorf("LokiTenantID = %v, want my-tenant", cfg.LokiTenantID)
	}
}

// TC-1.3.1: Default Batch Size
func TestLoad_DefaultBatchSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BatchSize != 100 {
		t.Errorf("BatchSize = %v, want 100", cfg.BatchSize)
	}
}

// TC-1.3.2: Custom Batch Size
func TestLoad_CustomBatchSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_BATCH_SIZE", "50")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BatchSize != 50 {
		t.Errorf("BatchSize = %v, want 50", cfg.BatchSize)
	}
}

// TC-1.3.3: Default Max Batch Bytes
func TestLoad_DefaultMaxBatchBytes(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expected := 5 * 1024 * 1024 // 5MB
	if cfg.MaxBatchSizeBytes != expected {
		t.Errorf("MaxBatchSizeBytes = %v, want %v", cfg.MaxBatchSizeBytes, expected)
	}
}

// TC-1.3.4: Custom Max Batch Bytes
func TestLoad_CustomMaxBatchBytes(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_MAX_BATCH_SIZE_BYTES", "1048576")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxBatchSizeBytes != 1048576 {
		t.Errorf("MaxBatchSizeBytes = %v, want 1048576", cfg.MaxBatchSizeBytes)
	}
}

// TC-1.4.1: Default Flush Interval
func TestLoad_DefaultFlushInterval(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.FlushIntervalMs != 1000 {
		t.Errorf("FlushIntervalMs = %v, want 1000", cfg.FlushIntervalMs)
	}
}

// TC-1.4.2: Custom Flush Interval
func TestLoad_CustomFlushInterval(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_FLUSH_INTERVAL_MS", "500")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.FlushIntervalMs != 500 {
		t.Errorf("FlushIntervalMs = %v, want 500", cfg.FlushIntervalMs)
	}
}

// TC-1.4.3: Default Idle Multiplier
func TestLoad_DefaultIdleMultiplier(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.IdleFlushMultiplier != 3 {
		t.Errorf("IdleFlushMultiplier = %v, want 3", cfg.IdleFlushMultiplier)
	}
}

// TC-1.4.4: Custom Idle Multiplier
func TestLoad_CustomIdleMultiplier(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_IDLE_FLUSH_MULTIPLIER", "5")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.IdleFlushMultiplier != 5 {
		t.Errorf("IdleFlushMultiplier = %v, want 5", cfg.IdleFlushMultiplier)
	}
}

// TC-1.5.1: Default Max Retries
func TestLoad_DefaultMaxRetries(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", cfg.MaxRetries)
	}
}

// TC-1.5.2: Custom Max Retries
func TestLoad_CustomMaxRetries(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_MAX_RETRIES", "5")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %v, want 5", cfg.MaxRetries)
	}
}

// TC-1.5.3: Default Critical Retries
func TestLoad_DefaultCriticalRetries(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CriticalFlushRetries != 5 {
		t.Errorf("CriticalFlushRetries = %v, want 5", cfg.CriticalFlushRetries)
	}
}

// TC-1.5.4: Custom Critical Retries
func TestLoad_CustomCriticalRetries(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_CRITICAL_FLUSH_RETRIES", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CriticalFlushRetries != 10 {
		t.Errorf("CriticalFlushRetries = %v, want 10", cfg.CriticalFlushRetries)
	}
}

// TC-1.6.1: Gzip Enabled (Default)
func TestLoad_GzipEnabledDefault(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.EnableGzip {
		t.Errorf("EnableGzip = %v, want true", cfg.EnableGzip)
	}
}

// TC-1.6.2: Gzip Disabled
func TestLoad_GzipDisabled(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_ENABLE_GZIP", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.EnableGzip {
		t.Errorf("EnableGzip = %v, want false", cfg.EnableGzip)
	}
}

// TC-1.6.3: Compression Threshold Default
func TestLoad_CompressionThresholdDefault(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CompressionThreshold != 1024 {
		t.Errorf("CompressionThreshold = %v, want 1024", cfg.CompressionThreshold)
	}
}

// TC-1.6.3: Compression Threshold Custom
func TestLoad_CompressionThresholdCustom(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_COMPRESSION_THRESHOLD", "2048")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CompressionThreshold != 2048 {
		t.Errorf("CompressionThreshold = %v, want 2048", cfg.CompressionThreshold)
	}
}

// TC-1.7.1: Custom Labels JSON
func TestLoad_CustomLabelsJSON(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_LABELS", `{"env":"prod","team":"platform"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Labels["env"] != "prod" {
		t.Errorf("Labels[env] = %v, want prod", cfg.Labels["env"])
	}
	if cfg.Labels["team"] != "platform" {
		t.Errorf("Labels[team] = %v, want platform", cfg.Labels["team"])
	}
}

// TC-1.7.2: Invalid Labels JSON
func TestLoad_InvalidLabelsJSON(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_LABELS", "invalid-json")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

// TC-1.7.3: SERVICE_NAME Label
func TestLoad_ServiceNameLabel(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "SERVICE_NAME", "my-service")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Labels["service_name"] != "my-service" {
		t.Errorf("Labels[service_name] = %v, want my-service", cfg.Labels["service_name"])
	}
}

// TC-1.8.1: Default Buffer Size
func TestLoad_DefaultBufferSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BufferSize != 10000 {
		t.Errorf("BufferSize = %v, want 10000", cfg.BufferSize)
	}
}

// TC-1.8.2: Custom Buffer Size
func TestLoad_CustomBufferSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "BUFFER_SIZE", "5000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BufferSize != 5000 {
		t.Errorf("BufferSize = %v, want 5000", cfg.BufferSize)
	}
}

// TC-1.9.1: Default Max Line Size
func TestLoad_DefaultMaxLineSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxLineSize != 204800 {
		t.Errorf("MaxLineSize = %v, want 204800", cfg.MaxLineSize)
	}
}

// TC-1.9.2: Custom Max Line Size
func TestLoad_CustomMaxLineSize(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_MAX_LINE_SIZE", "1024")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxLineSize != 1024 {
		t.Errorf("MaxLineSize = %v, want 1024", cfg.MaxLineSize)
	}
}

// TC-1.9.3: Request ID Extraction Enabled (Default)
func TestLoad_ExtractRequestIDDefault(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.ExtractRequestID {
		t.Errorf("ExtractRequestID = %v, want true", cfg.ExtractRequestID)
	}
}

// TC-1.9.4: Request ID Extraction Disabled
func TestLoad_ExtractRequestIDDisabled(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_EXTRACT_REQUEST_ID", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ExtractRequestID {
		t.Errorf("ExtractRequestID = %v, want false", cfg.ExtractRequestID)
	}
}

// Test invalid integer value falls back to default
func TestLoad_InvalidIntegerFallsBackToDefault(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_BATCH_SIZE", "not-a-number")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BatchSize != 100 {
		t.Errorf("BatchSize = %v, want 100 (default)", cfg.BatchSize)
	}
}

// Test invalid boolean value falls back to default
func TestLoad_InvalidBooleanFallsBackToDefault(t *testing.T) {
	clearAllEnvVars(t)
	setEnv(t, "LOKI_URL", "https://loki.example.com")
	setEnv(t, "LOKI_ENABLE_GZIP", "not-a-bool")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.EnableGzip {
		t.Errorf("EnableGzip = %v, want true (default)", cfg.EnableGzip)
	}
}
