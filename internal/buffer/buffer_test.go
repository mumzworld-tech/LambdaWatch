package buffer

import (
	"sync"
	"testing"
	"time"
)

// TC-2.1.1: Add Single Entry
func TestBuffer_AddSingleEntry(t *testing.T) {
	buf := New(100)
	entry := LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Message:   "test message",
		Type:      "function",
		RequestID: "req-123",
	}

	atCapacity := buf.Add(entry)

	if atCapacity {
		t.Error("Add() returned true, expected false (not at capacity)")
	}
	if buf.Len() != 1 {
		t.Errorf("Len() = %d, want 1", buf.Len())
	}
}

// TC-2.1.2: Add Multiple Entries
func TestBuffer_AddMultipleEntries(t *testing.T) {
	buf := New(100)

	for i := 0; i < 50; i++ {
		buf.Add(LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Message:   "message",
			Type:      "function",
		})
	}

	if buf.Len() != 50 {
		t.Errorf("Len() = %d, want 50", buf.Len())
	}
}

// TC-2.1.3: Add At Capacity
func TestBuffer_AddAtCapacity(t *testing.T) {
	buf := New(10)

	for i := 0; i < 9; i++ {
		buf.Add(LogEntry{Message: "message"})
	}

	atCapacity := buf.Add(LogEntry{Message: "final"})

	if !atCapacity {
		t.Error("Add() returned false at capacity, expected true")
	}
	if buf.Len() != 10 {
		t.Errorf("Len() = %d, want 10", buf.Len())
	}
}

// TC-2.1.4: Add Over Capacity (drops oldest)
func TestBuffer_AddOverCapacity(t *testing.T) {
	buf := New(5)

	// Add 5 entries
	for i := 0; i < 5; i++ {
		buf.Add(LogEntry{Message: "msg-" + string(rune('A'+i))})
	}

	// Add one more (should drop oldest)
	buf.Add(LogEntry{Message: "msg-new"})

	if buf.Len() != 5 {
		t.Errorf("Len() = %d, want 5", buf.Len())
	}

	// Flush and check oldest was dropped
	entries := buf.Flush(5)
	if entries[0].Message != "msg-B" {
		t.Errorf("First entry = %s, want msg-B (oldest should be dropped)", entries[0].Message)
	}
}

// TC-2.2.1: AddBatch
func TestBuffer_AddBatch(t *testing.T) {
	buf := New(100)

	entries := []LogEntry{
		{Message: "msg1"},
		{Message: "msg2"},
		{Message: "msg3"},
	}
	buf.AddBatch(entries)

	if buf.Len() != 3 {
		t.Errorf("Len() = %d, want 3", buf.Len())
	}
}

// TC-2.2.2: AddBatch Signals Ready
func TestBuffer_AddBatchSignalsReady(t *testing.T) {
	buf := New(100)

	entries := []LogEntry{{Message: "msg1"}}

	// Start goroutine to wait for signal
	done := make(chan bool, 1)
	go func() {
		select {
		case <-buf.Ready():
			done <- true
		case <-time.After(100 * time.Millisecond):
			done <- false
		}
	}()

	time.Sleep(10 * time.Millisecond) // Give goroutine time to start
	buf.AddBatch(entries)

	if !<-done {
		t.Error("AddBatch did not signal ready channel")
	}
}

// TC-2.2.3: AddBatch Over Capacity
func TestBuffer_AddBatchOverCapacity(t *testing.T) {
	buf := New(5)

	entries := make([]LogEntry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = LogEntry{Message: "msg"}
	}
	buf.AddBatch(entries)

	if buf.Len() != 5 {
		t.Errorf("Len() = %d, want 5 (capped at max)", buf.Len())
	}
}

// TC-2.3.1: Flush Partial
func TestBuffer_FlushPartial(t *testing.T) {
	buf := New(100)

	for i := 0; i < 50; i++ {
		buf.Add(LogEntry{Message: "msg"})
	}

	entries := buf.Flush(20)

	if len(entries) != 20 {
		t.Errorf("Flush returned %d entries, want 20", len(entries))
	}
	if buf.Len() != 30 {
		t.Errorf("Len() = %d, want 30", buf.Len())
	}
}

// TC-2.3.2: Flush All
func TestBuffer_FlushAll(t *testing.T) {
	buf := New(100)

	for i := 0; i < 50; i++ {
		buf.Add(LogEntry{Message: "msg"})
	}

	entries := buf.Flush(100) // Request more than available

	if len(entries) != 50 {
		t.Errorf("Flush returned %d entries, want 50", len(entries))
	}
	if buf.Len() != 0 {
		t.Errorf("Len() = %d, want 0", buf.Len())
	}
}

// TC-2.3.3: Flush Empty Buffer
func TestBuffer_FlushEmpty(t *testing.T) {
	buf := New(100)

	entries := buf.Flush(10)

	if entries != nil {
		t.Errorf("Flush on empty buffer = %v, want nil", entries)
	}
}

// TC-2.3.4: Flush Preserves Order
func TestBuffer_FlushPreservesOrder(t *testing.T) {
	buf := New(100)

	for i := 0; i < 5; i++ {
		buf.Add(LogEntry{Message: string(rune('A' + i))})
	}

	entries := buf.Flush(5)

	for i, e := range entries {
		expected := string(rune('A' + i))
		if e.Message != expected {
			t.Errorf("Entry[%d].Message = %s, want %s", i, e.Message, expected)
		}
	}
}

// TC-2.4.1: FlushBySize Within Limits
func TestBuffer_FlushBySizeWithinLimits(t *testing.T) {
	buf := New(100)

	// Add entries ~50 bytes each
	for i := 0; i < 10; i++ {
		buf.Add(LogEntry{
			Message:   "short message here",
			Type:      "function",
			RequestID: "req-123",
		})
	}

	// Flush with large byte limit
	entries := buf.FlushBySize(5, 10000)

	if len(entries) != 5 {
		t.Errorf("FlushBySize returned %d entries, want 5", len(entries))
	}
}

// TC-2.4.2: FlushBySize Byte Limited
func TestBuffer_FlushBySizeByteLimited(t *testing.T) {
	buf := New(100)

	// Add entries with known size
	for i := 0; i < 10; i++ {
		entry := LogEntry{
			Message:   "x", // 1 byte message
			Type:      "f",
			RequestID: "r",
		}
		buf.Add(entry)
	}

	// Entry size is ~11 bytes (1+1+1+8)
	// Limit to 30 bytes should return ~2-3 entries
	entries := buf.FlushBySize(100, 30)

	if len(entries) > 3 {
		t.Errorf("FlushBySize returned %d entries, expected <= 3 for 30 byte limit", len(entries))
	}
}

// TC-2.4.3: FlushBySize Single Large Entry
func TestBuffer_FlushBySizeSingleLargeEntry(t *testing.T) {
	buf := New(100)

	// Add one large entry that exceeds byte limit
	buf.Add(LogEntry{
		Message: "this is a very long message that exceeds our byte limit",
	})

	// Even with small byte limit, should return at least 1 entry
	entries := buf.FlushBySize(10, 10)

	if len(entries) != 1 {
		t.Errorf("FlushBySize returned %d entries, want 1 (should return at least one)", len(entries))
	}
}

// TC-2.5.1: Drain Returns All
func TestBuffer_DrainReturnsAll(t *testing.T) {
	buf := New(100)

	for i := 0; i < 50; i++ {
		buf.Add(LogEntry{Message: "msg"})
	}

	entries := buf.Drain()

	if len(entries) != 50 {
		t.Errorf("Drain returned %d entries, want 50", len(entries))
	}
	if buf.Len() != 0 {
		t.Errorf("Len() after Drain = %d, want 0", buf.Len())
	}
}

// TC-2.5.2: Drain Closes Buffer
func TestBuffer_DrainClosesBuffer(t *testing.T) {
	buf := New(100)
	buf.Add(LogEntry{Message: "before drain"})

	buf.Drain()

	// Add after drain should be ignored
	atCapacity := buf.Add(LogEntry{Message: "after drain"})

	if atCapacity {
		t.Error("Add after Drain should return false")
	}
	if buf.Len() != 0 {
		t.Errorf("Len() = %d, want 0 (closed buffer)", buf.Len())
	}
}

// TC-2.5.3: Drain Empty Buffer
func TestBuffer_DrainEmpty(t *testing.T) {
	buf := New(100)

	entries := buf.Drain()

	if len(entries) != 0 {
		t.Errorf("Drain on empty = %d entries, want 0", len(entries))
	}
}

// TC-2.6.1: Len Accuracy
func TestBuffer_LenAccuracy(t *testing.T) {
	buf := New(100)

	for i := 0; i < 25; i++ {
		buf.Add(LogEntry{Message: "msg"})
	}

	if buf.Len() != 25 {
		t.Errorf("Len() = %d, want 25", buf.Len())
	}

	buf.Flush(10)

	if buf.Len() != 15 {
		t.Errorf("Len() after flush = %d, want 15", buf.Len())
	}
}

// TC-2.6.2: ByteSize Tracking
func TestBuffer_ByteSizeTracking(t *testing.T) {
	buf := New(100)

	entry := LogEntry{
		Message:   "test",    // 4 bytes
		Type:      "fn",      // 2 bytes
		RequestID: "req",     // 3 bytes
		Timestamp: 123456789, // 8 bytes
	}
	expectedSize := entry.Size()

	buf.Add(entry)

	if buf.ByteSize() != expectedSize {
		t.Errorf("ByteSize() = %d, want %d", buf.ByteSize(), expectedSize)
	}
}

// TC-2.6.3: ByteSize After Flush
func TestBuffer_ByteSizeAfterFlush(t *testing.T) {
	buf := New(100)

	entry := LogEntry{Message: "test", Type: "fn", RequestID: "req"}
	buf.Add(entry)
	buf.Add(entry)

	initialSize := buf.ByteSize()

	buf.Flush(1)

	if buf.ByteSize() >= initialSize {
		t.Errorf("ByteSize() = %d, should be less than %d after flush", buf.ByteSize(), initialSize)
	}
}

// TC-2.7.1: Ready Channel
func TestBuffer_ReadyChannel(t *testing.T) {
	buf := New(100)

	ready := buf.Ready()
	if ready == nil {
		t.Error("Ready() returned nil channel")
	}
}

// TC-2.7.2: SignalReady
func TestBuffer_SignalReady(t *testing.T) {
	buf := New(100)

	done := make(chan bool, 1)
	go func() {
		select {
		case <-buf.Ready():
			done <- true
		case <-time.After(100 * time.Millisecond):
			done <- false
		}
	}()

	time.Sleep(10 * time.Millisecond)
	buf.SignalReady()

	if !<-done {
		t.Error("SignalReady did not signal ready channel")
	}
}

// TC-2.7.3: SignalReady Non-blocking
func TestBuffer_SignalReadyNonBlocking(t *testing.T) {
	buf := New(100)

	// Signal multiple times without reading
	for i := 0; i < 10; i++ {
		buf.SignalReady() // Should not block
	}
	// If we reach here without blocking, test passes
}

// TC-2.8.1: Concurrent Add
func TestBuffer_ConcurrentAdd(t *testing.T) {
	buf := New(1000)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				buf.Add(LogEntry{Message: "msg"})
			}
		}()
	}

	wg.Wait()

	// Should have 1000 entries (10 goroutines * 100 each)
	if buf.Len() != 1000 {
		t.Errorf("Len() = %d, want 1000", buf.Len())
	}
}

// TC-2.8.2: Concurrent Add and Flush
func TestBuffer_ConcurrentAddAndFlush(t *testing.T) {
	buf := New(1000)
	var wg sync.WaitGroup

	// Writer goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				buf.Add(LogEntry{Message: "msg"})
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Reader goroutines
	flushed := make(chan int, 150)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				entries := buf.Flush(10)
				flushed <- len(entries)
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()
	close(flushed)

	// Just verify no race conditions occurred (test passes if no panic)
	total := 0
	for n := range flushed {
		total += n
	}
	t.Logf("Flushed %d entries total", total)
}

// TC-2.8.3: Concurrent Len Calls
func TestBuffer_ConcurrentLen(t *testing.T) {
	buf := New(1000)
	var wg sync.WaitGroup

	// Concurrent adds
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			buf.Add(LogEntry{Message: "msg"})
		}
	}()

	// Concurrent len checks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = buf.Len()
			}
		}()
	}

	wg.Wait()
}

// TC-2.8.4: Concurrent ByteSize Calls
func TestBuffer_ConcurrentByteSize(t *testing.T) {
	buf := New(1000)
	var wg sync.WaitGroup

	// Concurrent adds
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			buf.Add(LogEntry{Message: "msg"})
		}
	}()

	// Concurrent bytesize checks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = buf.ByteSize()
			}
		}()
	}

	wg.Wait()
}

// Test LogEntry.Size()
func TestLogEntry_Size(t *testing.T) {
	entry := LogEntry{
		Timestamp: 1234567890,
		Message:   "hello",     // 5 bytes
		Type:      "function",  // 8 bytes
		RequestID: "req-12345", // 9 bytes
	}

	// Size = len(Message) + len(Type) + len(RequestID) + 8 (timestamp)
	expected := 5 + 8 + 9 + 8
	if entry.Size() != expected {
		t.Errorf("Size() = %d, want %d", entry.Size(), expected)
	}
}
