package logsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
	"github.com/Sami-AlEsh/lambdawatch/internal/logger"
)

// Server is an HTTP server that receives logs from Lambda
type Server struct {
	server      *http.Server
	buffer      *buffer.Buffer
	port        int
	maxLineSize int
}

// NewServer creates a new log receiver server
func NewServer(buf *buffer.Buffer, port int, maxLineSize int) *Server {
	s := &Server{
		buffer:      buf,
		port:        port,
		maxLineSize: maxLineSize,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleLogs)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.Infof("Starting log receiver on port %d", s.port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Infof("Log server error: %v", err)
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ListenerURI returns the URI for the Logs API subscription
func (s *Server) ListenerURI() string {
	return fmt.Sprintf("http://sandbox.localdomain:%d", s.port)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Infof("Failed to read log body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var messages []LogMessage
	if err := json.Unmarshal(body, &messages); err != nil {
		logger.Infof("Failed to parse log messages: %v", err)
		http.Error(w, "Failed to parse logs", http.StatusBadRequest)
		return
	}

	entries := make([]buffer.LogEntry, 0, len(messages))
	for _, msg := range messages {
		ts := parseTimestamp(msg.Time)
		message := formatRecord(msg.Record)
		msgType := msg.Type

		// Split long messages if maxLineSize is configured
		if s.maxLineSize > 0 && len(message) > s.maxLineSize {
			chunks := splitMessage(message, s.maxLineSize)
			for i, chunk := range chunks {
				entry := buffer.LogEntry{
					Timestamp: ts + int64(i), // Increment timestamp slightly to preserve order
					Message:   chunk,
					Type:      msgType,
				}
				entries = append(entries, entry)
			}
		} else {
			entry := buffer.LogEntry{
				Timestamp: ts,
				Message:   message,
				Type:      msgType,
			}
			entries = append(entries, entry)
		}
	}

	s.buffer.AddBatch(entries)
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
	switch v := record.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

// splitMessage splits a message into chunks of maxSize bytes
// It adds chunk markers to help reassemble the message if needed
func splitMessage(message string, maxSize int) []string {
	if len(message) <= maxSize {
		return []string{message}
	}

	// Reserve space for chunk markers like "[chunk 1/3] "
	// Max marker size: "[chunk 999/999] " = 16 bytes
	markerReserve := 20
	effectiveSize := maxSize - markerReserve
	if effectiveSize < 100 {
		effectiveSize = 100 // Minimum chunk size
	}

	// Calculate number of chunks needed
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
