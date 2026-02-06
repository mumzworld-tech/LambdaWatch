package loki

import (
	"testing"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
)

func TestBatch_NewBatch(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, true)
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

func TestBatch_SingleStream(t *testing.T) {
	labels := map[string]string{"source": "lambda"}
	b := NewBatch(labels, false)
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
	b := NewBatch(labels, true)
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
	b := NewBatch(map[string]string{}, true)
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
	b := NewBatch(map[string]string{}, false)
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
	b := NewBatch(map[string]string{}, true)
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
