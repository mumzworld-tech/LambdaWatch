package logsapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
)

func newTestServer(maxLineSize int) *Server {
	buf := buffer.New(1000)
	return NewServer(buf, 0, maxLineSize)
}

func postLogs(s *Server, msgs []LogMessage) *httptest.ResponseRecorder {
	body, _ := json.Marshal(msgs)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleLogs(w, req)
	return w
}

// --- HTTP Server ---

func TestServer_PostMethodOnly(t *testing.T) {
	s := newTestServer(0)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.handleLogs(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestServer_InvalidJSON(t *testing.T) {
	s := newTestServer(0)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	s.handleLogs(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServer_EmptyArray(t *testing.T) {
	s := newTestServer(0)
	w := postLogs(s, []LogMessage{})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if s.buffer.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", s.buffer.Len())
	}
}

// --- Log Processing ---

func TestServer_FunctionLog(t *testing.T) {
	s := newTestServer(0)
	msgs := []LogMessage{{
		Time:   "2026-02-05T21:34:18.835Z",
		Type:   "function",
		Record: `{"level":"info","message":"Hello"}`,
	}}
	postLogs(s, msgs)
	if s.buffer.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", s.buffer.Len())
	}
	entries := s.buffer.Flush(1)
	if entries[0].Type != "function" {
		t.Errorf("expected type function, got %s", entries[0].Type)
	}
}

func TestServer_PlainTextRecord(t *testing.T) {
	s := newTestServer(0)
	msgs := []LogMessage{{
		Time:   "2026-02-05T21:34:18.835Z",
		Type:   "function",
		Record: "plain text log",
	}}
	postLogs(s, msgs)
	entries := s.buffer.Flush(1)
	if entries[0].Message != "plain text log" {
		t.Errorf("expected plain text, got: %s", entries[0].Message)
	}
}

func TestServer_MultipleLogs(t *testing.T) {
	s := newTestServer(0)
	msgs := []LogMessage{
		{Time: "2026-02-05T21:34:18.001Z", Type: "function", Record: "log1"},
		{Time: "2026-02-05T21:34:18.002Z", Type: "function", Record: "log2"},
		{Time: "2026-02-05T21:34:18.003Z", Type: "platform", Record: "log3"},
	}
	postLogs(s, msgs)
	if s.buffer.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", s.buffer.Len())
	}
}

func TestServer_OrderPreserved(t *testing.T) {
	s := newTestServer(0)
	msgs := []LogMessage{
		{Time: "2026-02-05T21:34:18.001Z", Type: "function", Record: "first"},
		{Time: "2026-02-05T21:34:18.002Z", Type: "function", Record: "second"},
		{Time: "2026-02-05T21:34:18.003Z", Type: "function", Record: "third"},
	}
	postLogs(s, msgs)
	entries := s.buffer.Flush(3)
	if entries[0].Message != "first" || entries[1].Message != "second" || entries[2].Message != "third" {
		t.Error("order not preserved")
	}
}

// --- Message Splitting ---

func TestServer_LargeMessageSplit(t *testing.T) {
	s := newTestServer(100)
	msgs := []LogMessage{{
		Time:   "2026-02-05T21:34:18.835Z",
		Type:   "function",
		Record: strings.Repeat("x", 350),
	}}
	postLogs(s, msgs)
	if s.buffer.Len() < 2 {
		t.Errorf("expected split into multiple chunks, got %d", s.buffer.Len())
	}
	entries := s.buffer.Flush(s.buffer.Len())
	if !strings.Contains(entries[0].Message, "[chunk 1/") {
		t.Errorf("expected chunk prefix: %s", entries[0].Message)
	}
}

func TestServer_MessageUnderLimit(t *testing.T) {
	s := newTestServer(1000)
	msgs := []LogMessage{{
		Time:   "2026-02-05T21:34:18.835Z",
		Type:   "function",
		Record: "short",
	}}
	postLogs(s, msgs)
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", s.buffer.Len())
	}
}

func TestServer_NoSplitWhenZeroLimit(t *testing.T) {
	s := newTestServer(0)
	msgs := []LogMessage{{
		Time:   "2026-02-05T21:34:18.835Z",
		Type:   "function",
		Record: strings.Repeat("x", 500000),
	}}
	postLogs(s, msgs)
	if s.buffer.Len() != 1 {
		t.Errorf("expected 1 entry when maxLineSize=0, got %d", s.buffer.Len())
	}
}

// --- Timestamp Parsing ---

func TestParseTimestamp_Valid(t *testing.T) {
	ts := parseTimestamp("2026-02-05T21:34:18.205123456Z")
	expected := time.Date(2026, 2, 5, 21, 34, 18, 205123456, time.UTC).UnixNano()
	if ts != expected {
		t.Errorf("expected %d, got %d", expected, ts)
	}
}

func TestParseTimestamp_Invalid(t *testing.T) {
	before := time.Now().UnixNano()
	ts := parseTimestamp("invalid")
	after := time.Now().UnixNano()
	if ts < before || ts > after {
		t.Errorf("expected fallback to time.Now()")
	}
}

// --- Helper Functions ---

func TestFormatRecord_String(t *testing.T) {
	got := formatRecord("hello")
	if got != "hello" {
		t.Errorf("expected hello, got %s", got)
	}
}

func TestFormatRecord_NonString(t *testing.T) {
	got := formatRecord(map[string]interface{}{"key": "val"})
	if got != `{"key":"val"}` {
		t.Errorf("expected JSON, got %s", got)
	}
}

func TestSplitMessage_UnderLimit(t *testing.T) {
	chunks := splitMessage("short", 100)
	if len(chunks) != 1 || chunks[0] != "short" {
		t.Errorf("expected single chunk, got %v", chunks)
	}
}

func TestSplitMessage_OverLimit(t *testing.T) {
	chunks := splitMessage(strings.Repeat("a", 300), 100)
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if !strings.Contains(c, "[chunk") {
			t.Errorf("chunk %d missing prefix: %s", i, c)
		}
	}
}

func TestListenerURI(t *testing.T) {
	s := newTestServer(0)
	s.port = 9090
	if uri := s.ListenerURI(); uri != "http://sandbox.localdomain:9090" {
		t.Errorf("unexpected URI: %s", uri)
	}
}
