package loki

import (
	"testing"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
)

func TestBatch_NewBatch(t *testing.T) {
	b := NewBatch(map[string]string{"source": "lambda"}, true)
	if b.Len() != 0 {
		t.Errorf("expected empty batch, got %d", b.Len())
	}
}

func TestBatch_Add(t *testing.T) {
	b := NewBatch(map[string]string{}, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1"},
		{Timestamp: 2000, Message: "log2"},
	})
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
}

func TestBatch_ToPushRequest_Empty(t *testing.T) {
	b := NewBatch(map[string]string{}, false)
	if b.ToPushRequest() != nil {
		t.Error("expected nil for empty batch")
	}
}

func TestBatch_AlwaysSingleStream(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1", RequestID: "req-1"},
		{Timestamp: 2000, Message: "log2", RequestID: "req-2"},
		{Timestamp: 3000, Message: "log3", RequestID: "req-1"},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 1 {
		t.Fatalf("expected 1 stream regardless of request IDs, got %d", len(req.Streams))
	}
	if req.Streams[0].Stream["source"] != "lambda" {
		t.Error("missing source label")
	}
	if len(req.Streams[0].Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(req.Streams[0].Values))
	}
}

func TestBatch_TimestampConvertedToNanoseconds(t *testing.T) {
	b := NewBatch(map[string]string{}, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1"}, // 1000ms
	})
	req := b.ToPushRequest()
	// 1000ms * 1_000_000 = 1_000_000_000 nanoseconds
	if req.Streams[0].Values[0][0] != "1000000000" {
		t.Errorf("expected nanosecond timestamp, got %s", req.Streams[0].Values[0][0])
	}
}

func TestBatch_PreservesEntryOrder(t *testing.T) {
	b := NewBatch(map[string]string{}, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1, Message: "first"},
		{Timestamp: 2, Message: "second"},
		{Timestamp: 3, Message: "third"},
	})
	req := b.ToPushRequest()
	values := req.Streams[0].Values
	if values[0][1] != "first" || values[1][1] != "second" || values[2][1] != "third" {
		t.Error("entry order not preserved in single stream")
	}
}

// --- injectRequestID unit tests ---

func TestInjectRequestID_JSONMessage(t *testing.T) {
	msg := `{"level":"info","message":"hello"}`
	result := injectRequestID(msg, "abc-123")
	expected := `{"request_id":"abc-123","level":"info","message":"hello"}`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestInjectRequestID_PlainTextMessage(t *testing.T) {
	msg := "START RequestId: abc-123 Version: $LATEST"
	result := injectRequestID(msg, "abc-123")
	expected := "[request_id=abc-123] START RequestId: abc-123 Version: $LATEST"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestInjectRequestID_EmptyJSONObject(t *testing.T) {
	result := injectRequestID("{}", "abc-123")
	expected := `{"request_id":"abc-123"}`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestInjectRequestID_EmptyRequestID(t *testing.T) {
	msg := `{"level":"info","message":"hello"}`
	if injectRequestID(msg, "") != msg {
		t.Error("expected message unchanged for empty request ID")
	}
}

// --- integration: batch + injection ---

func TestBatch_InjectsRequestIDWhenEnabled(t *testing.T) {
	b := NewBatch(map[string]string{"source": "lambda"}, true)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: `{"level":"info","msg":"hello"}`, RequestID: "req-1"},
		{Timestamp: 2000, Message: "plain text log", RequestID: "req-2"},
		{Timestamp: 3000, Message: "no request id", RequestID: ""},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(req.Streams))
	}
	values := req.Streams[0].Values

	if values[0][1] != `{"request_id":"req-1","level":"info","msg":"hello"}` {
		t.Errorf("JSON injection failed: %s", values[0][1])
	}
	if values[1][1] != "[request_id=req-2] plain text log" {
		t.Errorf("plain text injection failed: %s", values[1][1])
	}
	if values[2][1] != "no request id" {
		t.Errorf("empty request_id should leave message unchanged: %s", values[2][1])
	}
}

func TestBatch_LeavesMessagesUnchangedWhenDisabled(t *testing.T) {
	b := NewBatch(map[string]string{"source": "lambda"}, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: `{"level":"info","msg":"hello"}`, RequestID: "req-1"},
		{Timestamp: 2000, Message: "plain text log", RequestID: "req-2"},
	})
	req := b.ToPushRequest()
	values := req.Streams[0].Values

	if values[0][1] != `{"level":"info","msg":"hello"}` {
		t.Errorf("message should be unchanged: %s", values[0][1])
	}
	if values[1][1] != "plain text log" {
		t.Errorf("message should be unchanged: %s", values[1][1])
	}
}
