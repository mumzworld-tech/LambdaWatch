package extension

// EventType represents Lambda extension event types
type EventType string

const (
	Invoke   EventType = "INVOKE"
	Shutdown EventType = "SHUTDOWN"
)

// RegisterResponse is the response from extension registration
type RegisterResponse struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
}

// NextEventResponse is the response from the next event API
type NextEventResponse struct {
	EventType          EventType `json:"eventType"`
	DeadlineMs         int64     `json:"deadlineMs"`
	RequestID          string    `json:"requestId"`
	InvokedFunctionArn string    `json:"invokedFunctionArn"`
	Tracing            *Tracing  `json:"tracing,omitempty"`
	ShutdownReason     string    `json:"shutdownReason,omitempty"`
}

// Tracing contains X-Ray tracing information
type Tracing struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
