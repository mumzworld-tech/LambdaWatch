package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mumzworld-tech/lambdawatch/internal/buffer"
	"github.com/mumzworld-tech/lambdawatch/internal/config"
	"github.com/mumzworld-tech/lambdawatch/internal/loki"
)

func newTestConfig() *config.Config {
	return &config.Config{
		LokiEndpoint:         "http://localhost:3100/loki/api/v1/push",
		BatchSize:            100,
		MaxBatchSizeBytes:    5 * 1024 * 1024,
		FlushIntervalMs:      1000,
		IdleFlushMultiplier:  3,
		MaxRetries:           3,
		CriticalFlushRetries: 5,
		EnableGzip:           false,
		CompressionThreshold: 1024,
		BufferSize:           10000,
		MaxLineSize:          204800,
		ExtractRequestID:     true,
		Labels:               map[string]string{},
	}
}

func newTestManager(cfg *config.Config) *Manager {
	m := &Manager{
		cfg:            cfg,
		buffer:         buffer.New(cfg.BufferSize),
		stopFlush:      make(chan struct{}),
		intervalChange: make(chan struct{}, 1),
	}
	m.state.Store(int32(StateIdle))
	return m
}

// =====================
// 4.1 State Definitions
// =====================

func TestState_InitialState(t *testing.T) {
	m := newTestManager(newTestConfig())
	if m.getState() != StateIdle {
		t.Errorf("expected initial state IDLE, got %s", m.getState())
	}
}

func TestState_StringRepresentation(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateIdle, "IDLE"},
		{StateActive, "ACTIVE"},
		{StateFlushing, "FLUSHING"},
		{State(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

// =====================
// 4.2 State Transitions
// =====================

func TestState_IdleToActive(t *testing.T) {
	m := newTestManager(newTestConfig())
	m.setState(StateActive)
	if m.getState() != StateActive {
		t.Errorf("expected ACTIVE, got %s", m.getState())
	}
}

func TestState_ActiveToFlushing(t *testing.T) {
	m := newTestManager(newTestConfig())
	m.setState(StateActive)
	m.setState(StateFlushing)
	if m.getState() != StateFlushing {
		t.Errorf("expected FLUSHING, got %s", m.getState())
	}
}

func TestState_FlushingToIdle(t *testing.T) {
	m := newTestManager(newTestConfig())
	m.setState(StateFlushing)
	m.setState(StateIdle)
	if m.getState() != StateIdle {
		t.Errorf("expected IDLE, got %s", m.getState())
	}
}

func TestState_SignalOnChange(t *testing.T) {
	m := newTestManager(newTestConfig())
	// Drain any existing signal
	select {
	case <-m.intervalChange:
	default:
	}

	m.setState(StateActive)
	select {
	case <-m.intervalChange:
		// good
	case <-time.After(100 * time.Millisecond):
		t.Error("expected signal on intervalChange channel")
	}
}

func TestState_NoSignalOnSameState(t *testing.T) {
	m := newTestManager(newTestConfig())
	m.setState(StateActive)
	// Drain signal from first transition
	select {
	case <-m.intervalChange:
	default:
	}

	m.setState(StateActive) // same state
	select {
	case <-m.intervalChange:
		t.Error("should not signal on same state")
	case <-time.After(50 * time.Millisecond):
		// good
	}
}

// =====================
// 4.3 Interval Adjustments
// =====================

func TestFlushInterval_Active(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 1000
	m := newTestManager(cfg)
	m.setState(StateActive)
	if got := m.getFlushInterval(); got != time.Second {
		t.Errorf("expected 1s, got %v", got)
	}
}

func TestFlushInterval_Idle(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 1000
	cfg.IdleFlushMultiplier = 3
	m := newTestManager(cfg)
	// Initial state is IDLE
	if got := m.getFlushInterval(); got != 3*time.Second {
		t.Errorf("expected 3s, got %v", got)
	}
}

func TestFlushInterval_Flushing(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 1000
	m := newTestManager(cfg)
	m.setState(StateFlushing)
	expected := 1500 * time.Millisecond
	if got := m.getFlushInterval(); got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestFlushInterval_CustomIdleMultiplier(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 1000
	cfg.IdleFlushMultiplier = 5
	m := newTestManager(cfg)
	if got := m.getFlushInterval(); got != 5*time.Second {
		t.Errorf("expected 5s, got %v", got)
	}
}

// =====================
// 4.5 Atomic State Operations
// =====================

func TestState_ConcurrentReadWrite(t *testing.T) {
	m := newTestManager(newTestConfig())
	var wg sync.WaitGroup
	states := []State{StateIdle, StateActive, StateFlushing}

	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(s State) {
			defer wg.Done()
			m.setState(s)
		}(states[i%3])
		go func() {
			defer wg.Done()
			_ = m.getState()
		}()
	}
	wg.Wait()
	// No race condition = pass
}

// =====================
// 3.2 shouldFlush
// =====================

func TestShouldFlush_ByCount(t *testing.T) {
	cfg := newTestConfig()
	cfg.BatchSize = 10
	m := newTestManager(cfg)
	for i := 0; i < 10; i++ {
		m.buffer.Add(buffer.LogEntry{Message: "test"})
	}
	if !m.shouldFlush() {
		t.Error("expected shouldFlush=true when buffer >= batchSize")
	}
}

func TestShouldFlush_BelowCount(t *testing.T) {
	cfg := newTestConfig()
	cfg.BatchSize = 100
	m := newTestManager(cfg)
	for i := 0; i < 5; i++ {
		m.buffer.Add(buffer.LogEntry{Message: "test"})
	}
	if m.shouldFlush() {
		t.Error("expected shouldFlush=false when buffer < batchSize")
	}
}

func TestShouldFlush_ByBytes(t *testing.T) {
	cfg := newTestConfig()
	cfg.BatchSize = 1000
	cfg.MaxBatchSizeBytes = 100
	m := newTestManager(cfg)
	// Add entries that exceed byte limit
	for i := 0; i < 5; i++ {
		m.buffer.Add(buffer.LogEntry{Message: "a]long message that takes up space in the buffer"})
	}
	if !m.shouldFlush() {
		t.Error("expected shouldFlush=true when buffer bytes >= maxBatchSizeBytes")
	}
}

// =====================
// 3.3 & 3.4 Flush / Critical Flush with mock Loki
// =====================

func startMockLoki(t *testing.T) (*httptest.Server, *int, *[][]byte) {
	t.Helper()
	pushCount := new(int)
	bodies := &[][]byte{}
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		*pushCount++
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		*bodies = append(*bodies, body)
		w.WriteHeader(http.StatusNoContent)
	}))
	return server, pushCount, bodies
}

func newManagerWithMockLoki(cfg *config.Config, lokiURL string) *Manager {
	cfg.LokiEndpoint = lokiURL
	m := newTestManager(cfg)
	m.lokiClient = loki.NewClient(cfg)
	m.labels = map[string]string{"source": "lambda", "function_name": "test-fn"}
	return m
}

func TestFlush_PushesToLoki(t *testing.T) {
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()

	cfg := newTestConfig()
	m := newManagerWithMockLoki(cfg, server.URL)

	for i := 0; i < 5; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: fmt.Sprintf("log %d", i)})
	}

	m.flush(context.Background())
	if *pushCount != 1 {
		t.Errorf("expected 1 push, got %d", *pushCount)
	}
	if m.buffer.Len() != 0 {
		t.Errorf("expected buffer empty after flush, got %d", m.buffer.Len())
	}
}

func TestFlush_EmptyBufferNoPush(t *testing.T) {
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()

	m := newManagerWithMockLoki(newTestConfig(), server.URL)
	m.flush(context.Background())
	if *pushCount != 0 {
		t.Errorf("expected 0 pushes for empty buffer, got %d", *pushCount)
	}
}

func TestCriticalFlush_FlushesAll(t *testing.T) {
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()

	cfg := newTestConfig()
	cfg.BatchSize = 10
	m := newManagerWithMockLoki(cfg, server.URL)

	// Add 25 entries, batch size 10 → should need 3 pushes
	for i := 0; i < 25; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: fmt.Sprintf("log %d", i)})
	}

	m.criticalFlush(context.Background())
	if *pushCount != 3 {
		t.Errorf("expected 3 pushes (25 entries / batch 10), got %d", *pushCount)
	}
	if m.buffer.Len() != 0 {
		t.Errorf("expected buffer empty after critical flush, got %d", m.buffer.Len())
	}
}

func TestCriticalFlush_EmptyBuffer(t *testing.T) {
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()

	m := newManagerWithMockLoki(newTestConfig(), server.URL)
	m.criticalFlush(context.Background())
	if *pushCount != 0 {
		t.Errorf("expected 0 pushes for empty buffer, got %d", *pushCount)
	}
}

func TestFlushBatch_RespectsCountLimit(t *testing.T) {
	cfg := newTestConfig()
	cfg.BatchSize = 5
	cfg.MaxBatchSizeBytes = 0 // disable byte limit
	m := newManagerWithMockLoki(cfg, "http://unused")

	for i := 0; i < 20; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: "test"})
	}

	req, count := m.flushBatch()
	if req == nil {
		t.Fatal("expected non-nil push request")
	}
	if count != 5 {
		t.Errorf("expected 5 entries, got %d", count)
	}
	if m.buffer.Len() != 15 {
		t.Errorf("expected 15 remaining, got %d", m.buffer.Len())
	}
}

func TestFlushBatch_RespectsByteLimit(t *testing.T) {
	cfg := newTestConfig()
	cfg.BatchSize = 100
	cfg.MaxBatchSizeBytes = 100
	m := newManagerWithMockLoki(cfg, "http://unused")

	// Each entry ~50 bytes, so byte limit should cap at ~2 entries
	for i := 0; i < 10; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: "a]message that is about forty bytes long"})
	}

	_, count := m.flushBatch()
	if count >= 10 {
		t.Errorf("expected byte limit to cap entries, got %d", count)
	}
}

func TestFlushBatch_EmptyBuffer(t *testing.T) {
	m := newManagerWithMockLoki(newTestConfig(), "http://unused")
	req, count := m.flushBatch()
	if req != nil || count != 0 {
		t.Errorf("expected nil/0 for empty buffer, got %v/%d", req, count)
	}
}

// =====================
// 3.6 Flush Mutex
// =====================

func TestFlush_MutexPreventsConurrent(t *testing.T) {
	server, _, _ := startMockLoki(t)
	defer server.Close()

	m := newManagerWithMockLoki(newTestConfig(), server.URL)
	for i := 0; i < 10; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: "test"})
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		m.flush(context.Background())
	}()
	go func() {
		defer wg.Done()
		m.criticalFlush(context.Background())
	}()
	wg.Wait()
	// No race/panic = pass
}

// =====================
// 3.1 Flush Loop
// =====================

func TestFlushLoop_StopsOnStopChannel(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 50
	server, _, _ := startMockLoki(t)
	defer server.Close()
	m := newManagerWithMockLoki(cfg, server.URL)

	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		m.flushLoop(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	close(m.stopFlush)

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Error("flushLoop did not stop after stopFlush closed")
	}
}

func TestFlushLoop_StopsOnContextCancel(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 50
	server, _, _ := startMockLoki(t)
	defer server.Close()
	m := newManagerWithMockLoki(cfg, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		m.flushLoop(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Error("flushLoop did not stop after context cancel")
	}
}

func TestFlushLoop_FlushesOnTimer(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 100
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()
	m := newManagerWithMockLoki(cfg, server.URL)
	m.setState(StateActive)

	for i := 0; i < 5; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: "test"})
	}

	ctx, cancel := context.WithCancel(context.Background())
	go m.flushLoop(ctx)

	time.Sleep(250 * time.Millisecond)
	cancel()

	if *pushCount == 0 {
		t.Error("expected at least 1 push from timer-based flush")
	}
}

func TestFlushLoop_IntervalChangesOnStateTransition(t *testing.T) {
	cfg := newTestConfig()
	cfg.FlushIntervalMs = 100
	cfg.IdleFlushMultiplier = 3
	server, _, _ := startMockLoki(t)
	defer server.Close()
	m := newManagerWithMockLoki(cfg, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	go m.flushLoop(ctx)

	// Transition from IDLE to ACTIVE
	time.Sleep(50 * time.Millisecond)
	m.setState(StateActive)
	time.Sleep(50 * time.Millisecond)

	cancel()
	// No panic/deadlock = pass. Interval change was processed.
}

// =====================
// 7.4 onRuntimeDone
// =====================

func TestOnRuntimeDone_TriggersFlushAndSignals(t *testing.T) {
	server, pushCount, _ := startMockLoki(t)
	defer server.Close()

	cfg := newTestConfig()
	m := newManagerWithMockLoki(cfg, server.URL)

	// Simulate what eventLoop does on INVOKE: store Lambda's deadline
	m.invocationDeadline.Store(time.Now().Add(10 * time.Second).UnixMilli())

	// Set up invocationDone channel like eventLoop would
	m.invocationMu.Lock()
	m.invocationDone = make(chan struct{})
	m.invocationMu.Unlock()

	m.setState(StateActive)
	for i := 0; i < 5; i++ {
		m.buffer.Add(buffer.LogEntry{Timestamp: time.Now().UnixMilli(), Message: "test"})
	}

	done := make(chan struct{})
	go func() {
		m.onRuntimeDone("req-123")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("onRuntimeDone did not complete")
	}

	if *pushCount == 0 {
		t.Error("expected push during critical flush")
	}
	if m.getState() != StateIdle {
		t.Errorf("expected state IDLE after onRuntimeDone, got %s", m.getState())
	}

	// invocationDone should be closed
	m.invocationMu.Lock()
	ch := m.invocationDone
	m.invocationMu.Unlock()
	if ch != nil {
		t.Error("expected invocationDone to be nil after onRuntimeDone")
	}
}

// =====================
// 7.6 Label Building
// =====================

func TestBuildLabels_FunctionLabels(t *testing.T) {
	m := newTestManager(newTestConfig())
	labels := m.buildLabels(&RegisterResponse{
		FunctionName:    "my-func",
		FunctionVersion: "$LATEST",
	})
	if labels["function_name"] != "my-func" {
		t.Errorf("expected function_name=my-func, got %s", labels["function_name"])
	}
	if labels["function_version"] != "$LATEST" {
		t.Errorf("expected function_version=$LATEST, got %s", labels["function_version"])
	}
}

func TestBuildLabels_SourceLabel(t *testing.T) {
	m := newTestManager(newTestConfig())
	labels := m.buildLabels(&RegisterResponse{FunctionName: "f", FunctionVersion: "1"})
	if labels["source"] != "lambda" {
		t.Errorf("expected source=lambda, got %s", labels["source"])
	}
}

func TestBuildLabels_CustomLabelsMerged(t *testing.T) {
	cfg := newTestConfig()
	cfg.Labels = map[string]string{"custom": "value", "env": "prod"}
	m := newTestManager(cfg)
	labels := m.buildLabels(&RegisterResponse{FunctionName: "f", FunctionVersion: "1"})
	if labels["custom"] != "value" {
		t.Errorf("expected custom=value, got %s", labels["custom"])
	}
	if labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", labels["env"])
	}
}

func TestBuildLabels_AutoLabelsOverrideCustom(t *testing.T) {
	cfg := newTestConfig()
	cfg.Labels = map[string]string{"function_name": "should-be-overridden"}
	m := newTestManager(cfg)
	labels := m.buildLabels(&RegisterResponse{FunctionName: "real-name", FunctionVersion: "1"})
	if labels["function_name"] != "real-name" {
		t.Errorf("expected auto label to win, got %s", labels["function_name"])
	}
}

// =====================
// 7.1 Registration (mock Extensions API)
// =====================

func TestClient_Register(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2020-01-01/extension/register" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set(extensionIDHeader, "test-ext-id")
		_ = json.NewEncoder(w).Encode(RegisterResponse{
			FunctionName:    "test-func",
			FunctionVersion: "$LATEST",
		})
	}))
	defer server.Close()

	c := &Client{
		baseURL:       server.URL + "/2020-01-01/extension",
		httpClient:    &http.Client{},
		extensionName: "lambdawatch",
	}

	resp, err := c.Register(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FunctionName != "test-func" {
		t.Errorf("expected test-func, got %s", resp.FunctionName)
	}
	if c.GetExtensionID() != "test-ext-id" {
		t.Errorf("expected test-ext-id, got %s", c.GetExtensionID())
	}
}

func TestClient_Register_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := &Client{
		baseURL:       server.URL + "/2020-01-01/extension",
		httpClient:    &http.Client{},
		extensionName: "lambdawatch",
	}

	_, err := c.Register(context.Background())
	if err == nil {
		t.Error("expected error on 500 response")
	}
}

// =====================
// 7.3 NextEvent
// =====================

func TestClient_NextEvent_Invoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(NextEventResponse{
			EventType: Invoke,
			RequestID: "req-abc",
		})
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL + "/2020-01-01/extension",
		httpClient:  &http.Client{},
		extensionID: "ext-id",
	}

	event, err := c.NextEvent(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.EventType != Invoke {
		t.Errorf("expected INVOKE, got %s", event.EventType)
	}
	if event.RequestID != "req-abc" {
		t.Errorf("expected req-abc, got %s", event.RequestID)
	}
}

func TestClient_NextEvent_Shutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(NextEventResponse{
			EventType:      Shutdown,
			ShutdownReason: "spindown",
		})
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL + "/2020-01-01/extension",
		httpClient:  &http.Client{},
		extensionID: "ext-id",
	}

	event, err := c.NextEvent(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.EventType != Shutdown {
		t.Errorf("expected SHUTDOWN, got %s", event.EventType)
	}
	if event.ShutdownReason != "spindown" {
		t.Errorf("expected spindown, got %s", event.ShutdownReason)
	}
}
