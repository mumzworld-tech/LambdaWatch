package telemetryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
	"github.com/Sami-AlEsh/lambdawatch/internal/logger"
)

var requestIDRegex = regexp.MustCompile(`(?i)RequestId:\s*([a-f0-9-]+)`)

// RuntimeDoneHandler is called when platform.runtimeDone is received
type RuntimeDoneHandler func(requestID string)

// Server is an HTTP server that receives telemetry from Lambda
type Server struct {
	server            *http.Server
	buffer            *buffer.Buffer
	port              int
	maxLineSize       int
	extractRequestID  bool
	onRuntimeDone     RuntimeDoneHandler
	currentRequestID  string
}

// NewServer creates a new telemetry receiver server
func NewServer(buf *buffer.Buffer, port int, maxLineSize int, extractRequestID bool, onRuntimeDone RuntimeDoneHandler) *Server {
	s := &Server{
		buffer:           buf,
		port:             port,
		maxLineSize:      maxLineSize,
		extractRequestID: extractRequestID,
		onRuntimeDone:    onRuntimeDone,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleTelemetry)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.Infof("Starting telemetry receiver on port %d", s.port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Infof("Telemetry server error: %v", err)
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ListenerURI returns the URI for the Telemetry API subscription
func (s *Server) ListenerURI() string {
	return fmt.Sprintf("http://sandbox.localdomain:%d", s.port)
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Infof("Failed to read telemetry body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var events []TelemetryEvent
	if err := json.Unmarshal(body, &events); err != nil {
		logger.Infof("Failed to parse telemetry events: %v", err)
		http.Error(w, "Failed to parse events", http.StatusBadRequest)
		return
	}

	entries := make([]buffer.LogEntry, 0, len(events))
	var runtimeDoneRequestID string

	for _, event := range events {
		switch event.Type {
		case EventTypePlatformStart:
			// Extract request ID from platform.start
			if record, ok := event.Record.(map[string]interface{}); ok {
				if reqID, ok := record["requestId"].(string); ok {
					s.currentRequestID = reqID
				}
			}
			// Ship platform.start log
			ts := parseTimestamp(event.Time)
			entry := buffer.LogEntry{
				Timestamp: ts,
				Message:   formatRecord(event.Record),
				Type:      event.Type,
				RequestID: s.currentRequestID,
			}
			entries = append(entries, entry)

		case EventTypePlatformRuntimeDone:
			// Extract request ID and ship log
			if record, ok := event.Record.(map[string]interface{}); ok {
				if id, ok := record["requestId"].(string); ok {
					runtimeDoneRequestID = id
				}
			}
			ts := parseTimestamp(event.Time)
			entry := buffer.LogEntry{
				Timestamp: ts,
				Message:   formatRecord(event.Record),
				Type:      event.Type,
				RequestID: s.currentRequestID,
			}
			entries = append(entries, entry)

		case EventTypeFunction, EventTypeExtension:
			// Process function and extension logs
			ts := parseTimestamp(event.Time)
			message := formatRecord(event.Record)

			// Extract request ID from message if enabled
			requestID := s.currentRequestID
			if s.extractRequestID && requestID == "" {
				requestID = extractRequestID(message)
			}

			// Split long messages if needed
			if s.maxLineSize > 0 && len(message) > s.maxLineSize {
				chunks := splitMessage(message, s.maxLineSize)
				for i, chunk := range chunks {
					entry := buffer.LogEntry{
						Timestamp: ts + int64(i),
						Message:   chunk,
						Type:      event.Type,
						RequestID: requestID,
					}
					entries = append(entries, entry)
				}
			} else {
				entry := buffer.LogEntry{
					Timestamp: ts,
					Message:   message,
					Type:      event.Type,
					RequestID: requestID,
				}
				entries = append(entries, entry)
			}

		case EventTypePlatformReport:
			// Log platform report
			ts := parseTimestamp(event.Time)
			message := formatRecord(event.Record)
			entry := buffer.LogEntry{
				Timestamp: ts,
				Message:   message,
				Type:      event.Type,
				RequestID: s.currentRequestID,
			}
			entries = append(entries, entry)
		}
	}

	if len(entries) > 0 {
		s.buffer.AddBatch(entries)
	}

	// Trigger critical flush AFTER entries are added to buffer
	if runtimeDoneRequestID != "" && s.onRuntimeDone != nil {
		s.onRuntimeDone(runtimeDoneRequestID)
	}

	w.WriteHeader(http.StatusOK)
}

func parseTimestamp(timeStr string) int64 {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return time.Now().UnixNano()
	}
	return t.UnixNano()
}

func formatRecord(record interface{}) string {
	var msg string
	switch v := record.(type) {
	case string:
		msg = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		msg = string(b)
	}
	// Strip Lambda log prefix: "2026-02-05T08:12:42.944Z\trequestId\tINFO\t"
	if idx := findJSONStart(msg); idx > 0 {
		return msg[idx:]
	}
	return msg
}

func findJSONStart(s string) int {
	for i, c := range s {
		if c == '{' || c == '[' {
			return i
		}
	}
	return 0
}

func extractRequestID(message string) string {
	matches := requestIDRegex.FindStringSubmatch(message)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// splitMessage splits a message into chunks of maxSize bytes
func splitMessage(message string, maxSize int) []string {
	if len(message) <= maxSize {
		return []string{message}
	}

	// Reserve space for chunk markers
	markerReserve := 20
	effectiveSize := maxSize - markerReserve
	if effectiveSize < 100 {
		effectiveSize = 100
	}

	numChunks := (len(message) + effectiveSize - 1) / effectiveSize
	chunks := make([]string, 0, numChunks)

	for i := 0; i < len(message); i += effectiveSize {
		end := i + effectiveSize
		if end > len(message) {
			end = len(message)
		}

		chunkNum := len(chunks) + 1
		chunk := fmt.Sprintf("[chunk %d/%d] %s", chunkNum, numChunks, message[i:end])
		chunks = append(chunks, chunk)
	}

	return chunks
}
