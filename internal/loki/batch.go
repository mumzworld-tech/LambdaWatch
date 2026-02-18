package loki

import (
	"strconv"
	"strings"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
)

// Batch collects log entries for a single Loki push request.
// All entries are sent in one stream — request_id is injected into the
// message content rather than used as a label, following Loki's best
// practice of keeping label cardinality low.
type Batch struct {
	entries          []buffer.LogEntry
	labels           map[string]string
	extractRequestID bool
}

// NewBatch creates a new batch with the given stream labels.
// When extractRequestID is true, each entry's request ID is embedded
// into the log message so it remains searchable via LogQL content filters.
func NewBatch(labels map[string]string, extractRequestID bool) *Batch {
	return &Batch{
		entries:          make([]buffer.LogEntry, 0),
		labels:           labels,
		extractRequestID: extractRequestID,
	}
}

// Add appends entries to the batch.
func (b *Batch) Add(entries []buffer.LogEntry) {
	b.entries = append(b.entries, entries...)
}

// Len returns the number of entries in the batch.
func (b *Batch) Len() int {
	return len(b.entries)
}

// ToPushRequest converts the batch into a Loki PushRequest.
// Returns nil if the batch is empty.
func (b *Batch) ToPushRequest() *PushRequest {
	if len(b.entries) == 0 {
		return nil
	}

	values := make([][]string, len(b.entries))
	for i, entry := range b.entries {
		tsNano := entry.Timestamp * 1_000_000 // milliseconds → nanoseconds
		ts := strconv.FormatInt(tsNano, 10)
		msg := entry.Message
		if b.extractRequestID {
			msg = injectRequestID(msg, entry.RequestID)
		}
		values[i] = []string{ts, msg}
	}

	return NewPushRequest(b.labels, values)
}

// injectRequestID embeds the request ID into the log message so it is
// searchable via LogQL content filters without adding a high-cardinality label.
//
// For JSON messages it inserts a "request_id" field after the opening brace.
// For plain text it prepends "[request_id=<value>] ".
// If requestID is empty the message is returned unchanged.
func injectRequestID(message, requestID string) string {
	if requestID == "" {
		return message
	}

	trimmed := strings.TrimSpace(message)
	if strings.HasPrefix(trimmed, "{") {
		idx := strings.Index(message, "{")
		rest := message[idx+1:]
		// No trailing comma for an empty object body
		if strings.HasPrefix(strings.TrimSpace(rest), "}") {
			return message[:idx+1] + `"request_id":"` + requestID + `"` + rest
		}
		return message[:idx+1] + `"request_id":"` + requestID + `",` + rest
	}

	return "[request_id=" + requestID + "] " + message
}
