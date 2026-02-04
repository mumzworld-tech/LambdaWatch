package buffer

import (
	"sync"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp int64
	Message   string
	Type      string
	RequestID string // AWS Lambda request ID for grouping
}

// Size returns the approximate byte size of the entry
func (e *LogEntry) Size() int {
	return len(e.Message) + len(e.Type) + len(e.RequestID) + 8 // 8 bytes for timestamp
}

// Buffer is a thread-safe bounded buffer for log entries
type Buffer struct {
	mu        sync.Mutex
	entries   []LogEntry
	maxSize   int
	byteSize  int // Current total byte size
	ready     chan struct{}
	closed    bool
}

// New creates a new buffer with the specified max size
func New(maxSize int) *Buffer {
	return &Buffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
		ready:   make(chan struct{}, 1),
	}
}

// Add adds a log entry to the buffer
// Returns true if the buffer is at capacity
func (b *Buffer) Add(entry LogEntry) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return false
	}

	// If at capacity, drop oldest entry
	if len(b.entries) >= b.maxSize {
		b.byteSize -= b.entries[0].Size()
		b.entries = b.entries[1:]
	}

	b.entries = append(b.entries, entry)
	b.byteSize += entry.Size()
	return len(b.entries) >= b.maxSize
}

// AddBatch adds multiple log entries to the buffer
func (b *Buffer) AddBatch(entries []LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	for _, entry := range entries {
		if len(b.entries) >= b.maxSize {
			b.byteSize -= b.entries[0].Size()
			b.entries = b.entries[1:]
		}
		b.entries = append(b.entries, entry)
		b.byteSize += entry.Size()
	}

	// Signal that batch is ready
	select {
	case b.ready <- struct{}{}:
	default:
	}
}

// Flush returns and clears up to batchSize entries from the buffer
func (b *Buffer) Flush(batchSize int) []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return nil
	}

	count := batchSize
	if count > len(b.entries) {
		count = len(b.entries)
	}

	batch := make([]LogEntry, count)
	copy(batch, b.entries[:count])

	// Update byte size
	for i := 0; i < count; i++ {
		b.byteSize -= b.entries[i].Size()
	}

	b.entries = b.entries[count:]

	return batch
}

// FlushBySize returns entries up to maxBytes or batchSize, whichever comes first
func (b *Buffer) FlushBySize(batchSize int, maxBytes int) []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return nil
	}

	count := 0
	bytes := 0
	for i := 0; i < len(b.entries) && count < batchSize; i++ {
		entrySize := b.entries[i].Size()
		if bytes+entrySize > maxBytes && count > 0 {
			break
		}
		bytes += entrySize
		count++
	}

	if count == 0 {
		return nil
	}

	batch := make([]LogEntry, count)
	copy(batch, b.entries[:count])

	// Update byte size
	b.byteSize -= bytes
	b.entries = b.entries[count:]

	return batch
}

// Drain returns all remaining entries and closes the buffer
func (b *Buffer) Drain() []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	entries := b.entries
	b.entries = nil
	b.byteSize = 0

	return entries
}

// Len returns the current number of entries in the buffer
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// ByteSize returns the current total byte size of entries in the buffer
func (b *Buffer) ByteSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.byteSize
}

// Ready returns a channel that signals when logs are ready
func (b *Buffer) Ready() <-chan struct{} {
	return b.ready
}

// SignalReady manually signals that logs are ready for processing
func (b *Buffer) SignalReady() {
	select {
	case b.ready <- struct{}{}:
	default:
	}
}
