package loki

import (
	"testing"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
)

func TestBatch_NewBatch(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, true, true)
	if b.Len() != 0 {
		t.Errorf("expected empty batch, got %d", b.Len())
	}
}

func TestBatch_Add(t *testing.T) {
	b := NewBatch(map[string]string{}, false, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1"},
		{Timestamp: 2000, Message: "log2"},
	})
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
}

func TestBatch_ToPushRequest_Empty(t *testing.T) {
	b := NewBatch(map[string]string{}, false, false)
	if b.ToPushRequest() != nil {
		t.Error("expected nil for empty batch")
	}
}

func TestBatch_SingleStream(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, false, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1"},
		{Timestamp: 2000, Message: "log2"},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(req.Streams))
	}
	if req.Streams[0].Stream["source"] != "lambda" {
		t.Error("missing source label")
	}
	if len(req.Streams[0].Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(req.Streams[0].Values))
	}
}

func TestBatch_GroupedByRequestID(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, true, true)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1", RequestID: "req-1"},
		{Timestamp: 2000, Message: "log2", RequestID: "req-2"},
		{Timestamp: 3000, Message: "log3", RequestID: "req-1"},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 2 {
		t.Fatalf("expected 2 streams (grouped by requestID), got %d", len(req.Streams))
	}
	// First stream should be req-1 with 2 entries
	if req.Streams[0].Stream["request_id"] != "req-1" {
		t.Errorf("expected req-1, got %s", req.Streams[0].Stream["request_id"])
	}
	if len(req.Streams[0].Values) != 2 {
		t.Errorf("expected 2 values for req-1, got %d", len(req.Streams[0].Values))
	}
	// Second stream should be req-2 with 1 entry
	if req.Streams[1].Stream["request_id"] != "req-2" {
		t.Errorf("expected req-2, got %s", req.Streams[1].Stream["request_id"])
	}
}

func TestBatch_EmptyRequestID_GroupedAsUnknown(t *testing.T) {
	b := NewBatch(map[string]string{}, true, true)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1", RequestID: ""},
	})
	req := b.ToPushRequest()
	// Empty requestID grouped as "unknown", no request_id label added
	if _, ok := req.Streams[0].Stream["request_id"]; ok {
		t.Error("should not add request_id label for unknown")
	}
}

func TestBatch_TimestampConvertedToNanoseconds(t *testing.T) {
	b := NewBatch(map[string]string{}, false, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: "log1"}, // 1000ms
	})
	req := b.ToPushRequest()
	// 1000ms * 1000000 = 1000000000 nanoseconds
	if req.Streams[0].Values[0][0] != "1000000000" {
		t.Errorf("expected nanosecond timestamp, got %s", req.Streams[0].Values[0][0])
	}
}

func TestBatch_PreservesOrder(t *testing.T) {
	b := NewBatch(map[string]string{}, true, true)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1, Message: "a", RequestID: "r1"},
		{Timestamp: 2, Message: "b", RequestID: "r2"},
		{Timestamp: 3, Message: "c", RequestID: "r3"},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 3 {
		t.Fatalf("expected 3 streams, got %d", len(req.Streams))
	}
	if req.Streams[0].Stream["request_id"] != "r1" {
		t.Error("order not preserved")
	}
	if req.Streams[1].Stream["request_id"] != "r2" {
		t.Error("order not preserved")
	}
	if req.Streams[2].Stream["request_id"] != "r3" {
		t.Error("order not preserved")
	}
}

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
	msg := `{}`
	result := injectRequestID(msg, "abc-123")
	expected := `{"request_id":"abc-123"}`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestInjectRequestID_EmptyRequestID(t *testing.T) {
	msg := `{"level":"info","message":"hello"}`
	result := injectRequestID(msg, "")
	if result != msg {
		t.Errorf("expected message unchanged, got %s", result)
	}
}

func TestBatch_SingleStream_InjectsRequestID(t *testing.T) {
	b := NewBatch(map[string]string{"source": "lambda"}, false, true)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: `{"level":"info","msg":"hello"}`, RequestID: "req-1"},
		{Timestamp: 2000, Message: "plain text log", RequestID: "req-2"},
		{Timestamp: 3000, Message: "no request id", RequestID: ""},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(req.Streams))
	}

	// JSON message should have request_id injected as field
	if req.Streams[0].Values[0][1] != `{"request_id":"req-1","level":"info","msg":"hello"}` {
		t.Errorf("JSON injection failed, got: %s", req.Streams[0].Values[0][1])
	}

	// Plain text should have request_id prepended
	if req.Streams[0].Values[1][1] != "[request_id=req-2] plain text log" {
		t.Errorf("plain text injection failed, got: %s", req.Streams[0].Values[1][1])
	}

	// Empty request_id should leave message unchanged
	if req.Streams[0].Values[2][1] != "no request id" {
		t.Errorf("empty request_id should not modify message, got: %s", req.Streams[0].Values[2][1])
	}
}

func TestBatch_SingleStream_NoInjectionWhenExtractDisabled(t *testing.T) {
	b := NewBatch(map[string]string{"source": "lambda"}, false, false)
	b.Add([]buffer.LogEntry{
		{Timestamp: 1000, Message: `{"level":"info","msg":"hello"}`, RequestID: "req-1"},
		{Timestamp: 2000, Message: "plain text log", RequestID: "req-2"},
	})
	req := b.ToPushRequest()
	if len(req.Streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(req.Streams))
	}

	// Messages should be unchanged when extractRequestID is false
	if req.Streams[0].Values[0][1] != `{"level":"info","msg":"hello"}` {
		t.Errorf("message should be unchanged, got: %s", req.Streams[0].Values[0][1])
	}
	if req.Streams[0].Values[1][1] != "plain text log" {
		t.Errorf("message should be unchanged, got: %s", req.Streams[0].Values[1][1])
	}
}
