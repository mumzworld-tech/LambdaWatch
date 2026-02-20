package config

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	// Loki endpoint (required)
	LokiEndpoint string

	// Authentication
	LokiUsername string
	LokiPassword string
	LokiAPIKey   string
	LokiTenantID string

	// Batching
	BatchSize           int
	MaxBatchSizeBytes   int // Max batch size in bytes (0 = no limit)
	FlushIntervalMs     int
	IdleFlushMultiplier int // Multiplier for flush interval when idle (default 3x)

	// Reliability
	MaxRetries           int
	CriticalFlushRetries int // Higher retries for critical flushes (shutdown, runtimeDone)
	EnableGzip           bool
	CompressionThreshold int // Only compress if payload > this size (bytes)

	// Custom labels
	Labels map[string]string

	// Buffer
	BufferSize int

	// Message limits
	MaxLineSize int // Max bytes per log line (0 = no limit)

	// Request ID
	ExtractRequestID bool // Extract and embed request_id into log message content
}

func Load() (*Config, error) {
	cfg := &Config{
		LokiEndpoint:         os.Getenv("LOKI_URL"),
		LokiUsername:         os.Getenv("LOKI_USERNAME"),
		LokiPassword:         os.Getenv("LOKI_PASSWORD"),
		LokiAPIKey:           os.Getenv("LOKI_API_KEY"),
		LokiTenantID:         os.Getenv("LOKI_TENANT_ID"),
		BatchSize:            getEnvInt("LOKI_BATCH_SIZE", 100),
		MaxBatchSizeBytes:    getEnvInt("LOKI_MAX_BATCH_SIZE_BYTES", 5*1024*1024), // 5MB default
		FlushIntervalMs:      getEnvInt("LOKI_FLUSH_INTERVAL_MS", 1000),
		IdleFlushMultiplier:  getEnvInt("LOKI_IDLE_FLUSH_MULTIPLIER", 3),
		MaxRetries:           getEnvInt("LOKI_MAX_RETRIES", 3),
		CriticalFlushRetries: getEnvInt("LOKI_CRITICAL_FLUSH_RETRIES", 5),
		EnableGzip:           getEnvBool("LOKI_ENABLE_GZIP", true),
		CompressionThreshold: getEnvInt("LOKI_COMPRESSION_THRESHOLD", 1024), // 1KB default
		BufferSize:           getEnvInt("BUFFER_SIZE", 10000),
		MaxLineSize:          getEnvInt("LOKI_MAX_LINE_SIZE", 204800), // 200KB default
		ExtractRequestID:     getEnvBool("LOKI_EXTRACT_REQUEST_ID", true),
		Labels:               make(map[string]string),
	}

	// Parse custom labels from JSON
	if labelsJSON := os.Getenv("LOKI_LABELS"); labelsJSON != "" {
		if err := json.Unmarshal([]byte(labelsJSON), &cfg.Labels); err != nil {
			return nil, err
		}
	}

	// Add service_name from SERVICE_NAME env var
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		cfg.Labels["service_name"] = serviceName
	}

	return cfg, nil
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
