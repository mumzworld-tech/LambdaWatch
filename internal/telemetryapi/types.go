package telemetryapi

// Event types from Lambda Telemetry API
const (
	// Platform events
	EventTypePlatformStart       = "platform.start"
	EventTypePlatformEnd         = "platform.end"
	EventTypePlatformReport      = "platform.report"
	EventTypePlatformRuntimeDone = "platform.runtimeDone"
	EventTypePlatformFault       = "platform.fault"
	EventTypePlatformExtension   = "platform.extension"
	EventTypePlatformLogsDropped = "platform.logsDropped"

	// Function logs
	EventTypeFunction = "function"

	// Extension logs
	EventTypeExtension = "extension"
)

// TelemetryEvent represents a single telemetry event from Lambda
type TelemetryEvent struct {
	Time   string      `json:"time"`
	Type   string      `json:"type"`
	Record interface{} `json:"record"`
}

// PlatformStartRecord is the record for platform.start events
type PlatformStartRecord struct {
	RequestID string `json:"requestId"`
	Version   string `json:"version,omitempty"`
}

// PlatformRuntimeDoneRecord is the record for platform.runtimeDone events
type PlatformRuntimeDoneRecord struct {
	RequestID string  `json:"requestId"`
	Status    string  `json:"status"`
	Metrics   Metrics `json:"metrics,omitempty"`
}

// PlatformReportRecord is the record for platform.report events
type PlatformReportRecord struct {
	RequestID string  `json:"requestId"`
	Status    string  `json:"status"`
	Metrics   Metrics `json:"metrics,omitempty"`
}

// Metrics contains invocation metrics
type Metrics struct {
	DurationMs       float64 `json:"durationMs"`
	BilledDurationMs int     `json:"billedDurationMs"`
	MemorySizeMB     int     `json:"memorySizeMB"`
	MaxMemoryUsedMB  int     `json:"maxMemoryUsedMB"`
	InitDurationMs   float64 `json:"initDurationMs,omitempty"`
}

// SubscribeRequest is the request body for subscribing to the Telemetry API
type SubscribeRequest struct {
	SchemaVersion string       `json:"schemaVersion"`
	Types         []string     `json:"types"`
	Buffering     BufferConfig `json:"buffering"`
	Destination   Destination  `json:"destination"`
}

// BufferConfig configures telemetry buffering
type BufferConfig struct {
	MaxItems  int `json:"maxItems"`
	MaxBytes  int `json:"maxBytes"`
	TimeoutMs int `json:"timeoutMs"`
}

// Destination configures where telemetry is sent
type Destination struct {
	Protocol string `json:"protocol"`
	URI      string `json:"URI"`
}
