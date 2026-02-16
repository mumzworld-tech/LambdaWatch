package loki

import (
	"strconv"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
)

// Batch collects log entries for sending to Loki
type Batch struct {
	entries          []buffer.LogEntry
	labels           map[string]string
	groupByRequestID bool
}

// NewBatch creates a new batch with the given labels
func NewBatch(labels map[string]string, groupByRequestID bool) *Batch {
	return &Batch{
		entries:          make([]buffer.LogEntry, 0),
		labels:           labels,
		groupByRequestID: groupByRequestID,
	}
}

// Add adds entries to the batch
func (b *Batch) Add(entries []buffer.LogEntry) {
	b.entries = append(b.entries, entries...)
}

// Len returns the number of entries in the batch
func (b *Batch) Len() int {
	return len(b.entries)
}

// ToPushRequest converts the batch to a Loki push request
// If groupByRequestID is true, entries are grouped into separate streams by request ID
func (b *Batch) ToPushRequest() *PushRequest {
	if len(b.entries) == 0 {
		return nil
	}

	if !b.groupByRequestID {
		return b.toSingleStreamRequest()
	}

	return b.toGroupedStreamRequest()
}

// toSingleStreamRequest creates a push request with all entries in one stream
func (b *Batch) toSingleStreamRequest() *PushRequest {
	values := make([][]string, len(b.entries))
	for i, entry := range b.entries {
		// Convert milliseconds to nanoseconds for Loki
		tsNano := entry.Timestamp * 1000000
		ts := strconv.FormatInt(tsNano, 10)
		values[i] = []string{ts, entry.Message}
	}

	return NewPushRequest(b.labels, values)
}

// toGroupedStreamRequest creates a push request with entries grouped by request ID
func (b *Batch) toGroupedStreamRequest() *PushRequest {
	// Group entries by request ID
	groups := make(map[string][]buffer.LogEntry)
	var order []string // Preserve order of first appearance

	for _, entry := range b.entries {
		reqID := entry.RequestID
		if reqID == "" {
			reqID = "unknown"
		}

		if _, exists := groups[reqID]; !exists {
			order = append(order, reqID)
		}
		groups[reqID] = append(groups[reqID], entry)
	}

	// Create streams for each group
	streams := make([]Stream, 0, len(groups))

	for _, reqID := range order {
		entries := groups[reqID]

		// Copy base labels and add request_id
		streamLabels := make(map[string]string, len(b.labels)+1)
		for k, v := range b.labels {
			streamLabels[k] = v
		}
		if reqID != "unknown" {
			streamLabels["request_id"] = reqID
		}

		// Build values for this stream
		values := make([][]string, len(entries))
		for i, entry := range entries {
			// Convert milliseconds to nanoseconds for Loki
			tsNano := entry.Timestamp * 1000000
			ts := strconv.FormatInt(tsNano, 10)
			values[i] = []string{ts, entry.Message}
		}

		streams = append(streams, Stream{
			Stream: streamLabels,
			Values: values,
		})
	}

	return &PushRequest{Streams: streams}
}
