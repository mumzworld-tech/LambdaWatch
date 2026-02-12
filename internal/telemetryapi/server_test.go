package telemetryapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
)

func newTestServer(maxLineSize int, extractRequestID bool, onRuntimeDone RuntimeDoneHandler) *Server {
	buf := buffer.New(1000)
	return NewServer(buf, 0, maxLineSize, extractRequestID, onRuntimeDone)
}

func postEvents(s *Server, events []TelemetryEvent) *httptest.ResponseRecorder {
	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTelemetry(w, req)
	return w
}

// --- 6.1 HTTP Server ---

func TestServer_PostMethodOnly(t *testing.T) {
	s := newTestServer(0, true, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.handleTelemetry(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestServer_InvalidJSONBody(t *testing.T) {
	s := newTestServer(0, true, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	s.handleTelemetry(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServer_EmptyEventArray(t *testing.T) {
	s := newTestServer(0, true, nil)
	w := postEvents(s, []TelemetryEvent{})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if s.buffer.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", s.buffer.Len())
	}
}

// --- 6.2 Platform Events ---

func TestServer_PlatformStart(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{{
		Type: EventTypePlatformStart,
		Time: "2026-02-05T21:34:18.205Z",
		Record: map[string]interface{}{
			"requestId": "abc-123",
			"version":   "$LATEST",
		},
	}}
	w := postEvents(s, events)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if s.currentRequestID != "abc-123" {
		t.Errorf("expected currentRequestID=abc-123, got %s", s.currentRequestID)
	}
	if s.buffer.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", s.buffer.Len())
	}
	entries := s.buffer.Flush(1)
	if !strings.Contains(entries[0].Message, "START RequestId: abc-123 Version: $LATEST") {
		t.Errorf("unexpected message: %s", entries[0].Message)
	}
}

func TestServer_PlatformRuntimeDone(t *testing.T) {
	var calledWith string
	handler := func(reqID string) { calledWith = reqID }
	s := newTestServer(0, true, handler)
	events := []TelemetryEvent{{
		Type: EventTypePlatformRuntimeDone,
		Time: "2026-02-05T21:34:19.572Z",
		Record: map[string]interface{}{
			"requestId": "abc-123",
			"status":    "success",
		},
	}}
	postEvents(s, events)
	if calledWith != "abc-123" {
		t.Errorf("expected onRuntimeDone called with abc-123, got %s", calledWith)
	}
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", s.buffer.Len())
	}
}

func TestServer_PlatformReport(t *testing.T) {
	s := newTestServer(0, true, nil)
	s.currentRequestID = "abc-123"
	events := []TelemetryEvent{{
		Type: EventTypePlatformReport,
		Time: "2026-02-05T21:34:20.458Z",
		Record: map[string]interface{}{
			"requestId": "abc-123",
			"metrics": map[string]interface{}{
				"durationMs":       2251.86,
				"billedDurationMs": 3114.0,
				"memorySizeMB":     1024.0,
				"maxMemoryUsedMB":  184.0,
				"initDurationMs":   861.71,
			},
		},
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	msg := entries[0].Message
	if !strings.Contains(msg, "REPORT RequestId: abc-123") {
		t.Errorf("missing REPORT prefix: %s", msg)
	}
	if !strings.Contains(msg, "Duration: 2251.86 ms") {
		t.Errorf("missing duration: %s", msg)
	}
	if !strings.Contains(msg, "Init Duration: 861.71 ms") {
		t.Errorf("missing init duration: %s", msg)
	}
}

func TestServer_PlatformReportWithoutInitDuration(t *testing.T) {
	s := newTestServer(0, true, nil)
	s.currentRequestID = "abc-123"
	events := []TelemetryEvent{{
		Type: EventTypePlatformReport,
		Time: "2026-02-05T21:34:20.458Z",
		Record: map[string]interface{}{
			"requestId": "abc-123",
			"metrics": map[string]interface{}{
				"durationMs":       100.0,
				"billedDurationMs": 200.0,
				"memorySizeMB":     128.0,
				"maxMemoryUsedMB":  64.0,
			},
		},
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if strings.Contains(entries[0].Message, "Init Duration") {
		t.Errorf("should not contain Init Duration for warm start: %s", entries[0].Message)
	}
}

// --- 6.3 Function Logs ---

func TestServer_FunctionLog(t *testing.T) {
	s := newTestServer(0, true, nil)
	s.currentRequestID = "req-1"
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: `{"level":"info","message":"Hello"}`,
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if entries[0].Type != EventTypeFunction {
		t.Errorf("expected type function, got %s", entries[0].Type)
	}
	if entries[0].RequestID != "req-1" {
		t.Errorf("expected requestID req-1, got %s", entries[0].RequestID)
	}
}

func TestServer_FunctionLogWithLambdaPrefix(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: "2026-02-05T21:34:18.835Z\tabc-123\tINFO\t{\"message\":\"Hello\"}",
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if !strings.HasPrefix(entries[0].Message, "{") {
		t.Errorf("expected JSON after prefix strip, got: %s", entries[0].Message)
	}
}

func TestServer_NonJSONFunctionLog(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: "Plain text log message",
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if entries[0].Message != "Plain text log message" {
		t.Errorf("expected plain text preserved, got: %s", entries[0].Message)
	}
}

// --- 6.4 Extension Logs ---

func TestServer_OwnExtensionLogsFiltered(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeExtension,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: `{"context":"LambdaWatch","message":"Internal log"}`,
	}}
	postEvents(s, events)
	if s.buffer.Len() != 0 {
		t.Errorf("expected own extension logs filtered, got %d entries", s.buffer.Len())
	}
}

func TestServer_OtherExtensionLogsKept(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeExtension,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: `{"source":"other-extension","message":"Log"}`,
	}}
	postEvents(s, events)
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry for other extension, got %d", s.buffer.Len())
	}
}

// --- 6.5 Request ID Handling ---

func TestServer_RequestIDFromPlatformStart(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{
		{Type: EventTypePlatformStart, Time: "2026-02-05T21:34:18.205Z",
			Record: map[string]interface{}{"requestId": "start-req-1", "version": "$LATEST"}},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.835Z",
			Record: `{"message":"test"}`},
	}
	postEvents(s, events)
	entries := s.buffer.Flush(10)
	// Second entry is the function log
	if entries[1].RequestID != "start-req-1" {
		t.Errorf("expected requestID from platform.start, got %s", entries[1].RequestID)
	}
}

func TestServer_RequestIDExtractionFromMessage(t *testing.T) {
	s := newTestServer(0, true, nil)
	// No platform.start, so currentRequestID is empty
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: "START RequestId: 550e8400-e29b-41d4-a716-446655440000 Version: $LATEST",
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if entries[0].RequestID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected extracted requestID, got %s", entries[0].RequestID)
	}
}

func TestServer_RequestIDExtractionDisabled(t *testing.T) {
	s := newTestServer(0, false, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: "START RequestId: some-id Version: $LATEST",
	}}
	postEvents(s, events)
	entries := s.buffer.Flush(1)
	if entries[0].RequestID != "" {
		t.Errorf("expected empty requestID when extraction disabled, got %s", entries[0].RequestID)
	}
}

func TestServer_RequestIDPersistsAcrossBatches(t *testing.T) {
	s := newTestServer(0, true, nil)
	// First batch sets requestID
	postEvents(s, []TelemetryEvent{{
		Type: EventTypePlatformStart, Time: "2026-02-05T21:34:18.205Z",
		Record: map[string]interface{}{"requestId": "persist-id", "version": "$LATEST"},
	}})
	s.buffer.Flush(10) // clear

	// Second batch should still use the requestID
	postEvents(s, []TelemetryEvent{{
		Type: EventTypeFunction, Time: "2026-02-05T21:34:19.000Z",
		Record: `{"message":"later log"}`,
	}})
	entries := s.buffer.Flush(1)
	if entries[0].RequestID != "persist-id" {
		t.Errorf("expected persisted requestID, got %s", entries[0].RequestID)
	}
}

// --- 6.6 Message Processing ---

func TestServer_LargeMessageSplit(t *testing.T) {
	s := newTestServer(100, true, nil)
	// Create a message larger than maxLineSize
	bigMsg := strings.Repeat("x", 350)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: bigMsg,
	}}
	postEvents(s, events)
	count := s.buffer.Len()
	if count < 2 {
		t.Errorf("expected message to be split into multiple chunks, got %d entries", count)
	}
	entries := s.buffer.Flush(count)
	if !strings.Contains(entries[0].Message, "[chunk 1/") {
		t.Errorf("expected chunk prefix, got: %s", entries[0].Message)
	}
}

func TestServer_MessageUnderLimit(t *testing.T) {
	s := newTestServer(1000, true, nil)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: "short message",
	}}
	postEvents(s, events)
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry (no split), got %d", s.buffer.Len())
	}
}

func TestServer_MaxLineSizeZeroNoLimit(t *testing.T) {
	s := newTestServer(0, true, nil)
	bigMsg := strings.Repeat("x", 500000)
	events := []TelemetryEvent{{
		Type:   EventTypeFunction,
		Time:   "2026-02-05T21:34:18.835Z",
		Record: bigMsg,
	}}
	postEvents(s, events)
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry (no split when maxLineSize=0), got %d", s.buffer.Len())
	}
}

// --- 6.7 Timestamp Parsing ---

func TestParseTimestamp_RFC3339Nano(t *testing.T) {
	ts := parseTimestamp("2026-02-05T21:34:18.205123456Z")
	expected := time.Date(2026, 2, 5, 21, 34, 18, 205123456, time.UTC).UnixMilli()
	if ts != expected {
		t.Errorf("expected %d, got %d", expected, ts)
	}
}

func TestParseTimestamp_Invalid(t *testing.T) {
	before := time.Now().UnixMilli()
	ts := parseTimestamp("invalid")
	after := time.Now().UnixMilli()
	if ts < before || ts > after {
		t.Errorf("expected fallback to time.Now(), got %d (range %d-%d)", ts, before, after)
	}
}

// --- 6.8 Batch Processing ---

func TestServer_MultipleEventsInBatch(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.000Z", Record: "log1"},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.001Z", Record: "log2"},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.002Z", Record: "log3"},
	}
	postEvents(s, events)
	if s.buffer.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", s.buffer.Len())
	}
}

func TestServer_MixedEventTypes(t *testing.T) {
	var runtimeDoneCalled bool
	handler := func(reqID string) { runtimeDoneCalled = true }
	s := newTestServer(0, true, handler)

	events := []TelemetryEvent{
		{Type: EventTypePlatformStart, Time: "2026-02-05T21:34:18.000Z",
			Record: map[string]interface{}{"requestId": "mix-req", "version": "$LATEST"}},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.100Z", Record: "func log 1"},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.200Z", Record: "func log 2"},
		{Type: EventTypePlatformRuntimeDone, Time: "2026-02-05T21:34:18.300Z",
			Record: map[string]interface{}{"requestId": "mix-req", "status": "success"}},
	}
	postEvents(s, events)

	if !runtimeDoneCalled {
		t.Error("expected onRuntimeDone to be called")
	}
	// platform.start + 2 function logs + runtimeDone = 4 entries
	if s.buffer.Len() != 4 {
		t.Errorf("expected 4 entries, got %d", s.buffer.Len())
	}
}

func TestServer_RuntimeDoneAfterBufferAdd(t *testing.T) {
	var bufLenAtCallback int
	handler := func(reqID string) {
		// At callback time, entries should already be in buffer
		// We can't access s.buffer here directly, so we capture via closure
	}
	s := newTestServer(0, true, nil)
	// Override handler to check buffer state
	s.onRuntimeDone = func(reqID string) {
		bufLenAtCallback = s.buffer.Len()
	}

	events := []TelemetryEvent{
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.100Z", Record: "log before done"},
		{Type: EventTypePlatformRuntimeDone, Time: "2026-02-05T21:34:18.300Z",
			Record: map[string]interface{}{"requestId": "order-req", "status": "success"}},
	}
	_ = handler
	postEvents(s, events)

	// Both entries (function + runtimeDone) should be in buffer before callback
	if bufLenAtCallback < 2 {
		t.Errorf("expected entries in buffer before onRuntimeDone callback, got %d", bufLenAtCallback)
	}
}

func TestServer_EventOrderPreserved(t *testing.T) {
	s := newTestServer(0, true, nil)
	events := []TelemetryEvent{
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.001Z", Record: "first"},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.002Z", Record: "second"},
		{Type: EventTypeFunction, Time: "2026-02-05T21:34:18.003Z", Record: "third"},
	}
	postEvents(s, events)
	entries := s.buffer.Flush(3)
	if entries[0].Message != "first" || entries[1].Message != "second" || entries[2].Message != "third" {
		t.Errorf("order not preserved: %s, %s, %s", entries[0].Message, entries[1].Message, entries[2].Message)
	}
}

// --- Helper function tests ---

func TestExtractRequestID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"START RequestId: abc-123-def-456 Version: $LATEST", "abc-123-def-456"},
		{"RequestId: 550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
		{"no request id here", ""},
	}
	for _, tt := range tests {
		got := extractRequestID(tt.input)
		if got != tt.expected {
			t.Errorf("extractRequestID(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSplitMessage(t *testing.T) {
	// Message under limit
	chunks := splitMessage("short", 100)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}

	// Message over limit
	msg := strings.Repeat("a", 300)
	chunks = splitMessage(msg, 100)
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if !strings.Contains(c, "[chunk") {
			t.Errorf("chunk %d missing prefix: %s", i, c)
		}
	}
}

func TestFormatPlatformStart(t *testing.T) {
	record := map[string]interface{}{"requestId": "req-1", "version": "$LATEST"}
	msg := formatPlatformStart(record)
	if msg != "START RequestId: req-1 Version: $LATEST" {
		t.Errorf("unexpected: %s", msg)
	}
}

func TestFormatPlatformReport(t *testing.T) {
	record := map[string]interface{}{
		"requestId": "req-1",
		"metrics": map[string]interface{}{
			"durationMs":       100.5,
			"billedDurationMs": 200.0,
			"memorySizeMB":     128.0,
			"maxMemoryUsedMB":  64.0,
		},
	}
	msg := formatPlatformReport(record)
	if !strings.Contains(msg, "REPORT RequestId: req-1") {
		t.Errorf("missing REPORT: %s", msg)
	}
	if strings.Contains(msg, "Init Duration") {
		t.Errorf("should not have Init Duration: %s", msg)
	}
}

func TestListenerURI(t *testing.T) {
	s := newTestServer(0, true, nil)
	s.port = 8080
	uri := s.ListenerURI()
	if uri != "http://sandbox.localdomain:8080" {
		t.Errorf("unexpected URI: %s", uri)
	}
}
