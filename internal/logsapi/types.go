package logsapi

// LogMessage represents a log message from the Lambda Logs API
type LogMessage struct {
	Time   string      `json:"time"`
	Type   string      `json:"type"`
	Record interface{} `json:"record"`
}

// SubscribeRequest is the request body for subscribing to the Logs API
type SubscribeRequest struct {
	SchemaVersion string       `json:"schemaVersion"`
	Types         []string     `json:"types"`
	Buffering     BufferConfig `json:"buffering"`
	Destination   Destination  `json:"destination"`
}

// BufferConfig configures log buffering
type BufferConfig struct {
	MaxItems  int `json:"maxItems"`
	MaxBytes  int `json:"maxBytes"`
	TimeoutMs int `json:"timeoutMs"`
}

// Destination configures where logs are sent
type Destination struct {
	Protocol string `json:"protocol"`
	URI      string `json:"URI"`
}

// LogTypes to subscribe to
const (
	LogTypePlatform  = "platform"
	LogTypeFunction  = "function"
	LogTypeExtension = "extension"
)
