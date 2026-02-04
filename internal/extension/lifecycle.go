package extension

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
	"github.com/Sami-AlEsh/lambdawatch/internal/config"
	"github.com/Sami-AlEsh/lambdawatch/internal/loki"
	"github.com/Sami-AlEsh/lambdawatch/internal/telemetryapi"
)

const telemetryServerPort = 8080

// State represents the extension's current operational state
type State int32

const (
	StateIdle     State = iota // No active invocation, longer flush intervals
	StateActive                // Invocation in progress, normal flush intervals
	StateFlushing              // Critical flush in progress
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateActive:
		return "ACTIVE"
	case StateFlushing:
		return "FLUSHING"
	default:
		return "UNKNOWN"
	}
}

// Manager orchestrates the extension lifecycle
type Manager struct {
	cfg             *config.Config
	extClient       *Client
	telemetryClient *telemetryapi.Client
	telemetryServer *telemetryapi.Server
	lokiClient      *loki.Client
	buffer          *buffer.Buffer
	labels          map[string]string
	stopFlush       chan struct{}

	// State management for adaptive intervals
	state atomic.Int32

	// Critical flush synchronization
	criticalFlushMu sync.Mutex
	criticalFlushWg sync.WaitGroup

	// Channel to signal interval changes
	intervalChange chan struct{}
}

// NewManager creates a new lifecycle manager
func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		cfg:            cfg,
		buffer:         buffer.New(cfg.BufferSize),
		stopFlush:      make(chan struct{}),
		intervalChange: make(chan struct{}, 1),
	}
	m.state.Store(int32(StateIdle))
	return m
}

// Run runs the extension lifecycle
func (m *Manager) Run(ctx context.Context) error {
	// Initialize components
	if err := m.init(ctx); err != nil {
		return err
	}

	// Start background flush goroutine
	go m.flushLoop(ctx)

	// Main event loop
	return m.eventLoop(ctx)
}

func (m *Manager) init(ctx context.Context) error {
	// Register with Lambda Extensions API
	m.extClient = NewClient()
	regResp, err := m.extClient.Register(ctx)
	if err != nil {
		return err
	}
	log.Printf("Registered extension for function: %s", regResp.FunctionName)

	// Build labels from config and Lambda environment
	m.labels = m.buildLabels(regResp)

	// Create Loki client
	m.lokiClient = loki.NewClient(m.cfg)

	// Start HTTP server to receive telemetry with runtimeDone handler
	m.telemetryServer = telemetryapi.NewServer(
		m.buffer,
		telemetryServerPort,
		m.cfg.MaxLineSize,
		m.cfg.ExtractRequestID,
		m.onRuntimeDone,
	)
	if err := m.telemetryServer.Start(); err != nil {
		return err
	}

	// Subscribe to Telemetry API
	m.telemetryClient = telemetryapi.NewClient(m.extClient.GetExtensionID())
	if err := m.telemetryClient.Subscribe(ctx, m.telemetryServer.ListenerURI()); err != nil {
		return err
	}
	log.Printf("Subscribed to Telemetry API")

	return nil
}

func (m *Manager) buildLabels(regResp *RegisterResponse) map[string]string {
	labels := make(map[string]string)

	// Add configured labels
	for k, v := range m.cfg.Labels {
		labels[k] = v
	}

	// Add Lambda-specific labels
	labels["function_name"] = regResp.FunctionName
	labels["function_version"] = regResp.FunctionVersion

	if region := os.Getenv("AWS_REGION"); region != "" {
		labels["region"] = region
	}

	// Add source label
	labels["source"] = "lambda"

	return labels
}

func (m *Manager) eventLoop(ctx context.Context) error {
	for {
		// Wait for any pending critical flushes before blocking on NextEvent
		m.criticalFlushWg.Wait()

		event, err := m.extClient.NextEvent(ctx)
		if err != nil {
			return err
		}

		switch event.EventType {
		case Invoke:
			m.setState(StateActive)
			log.Printf("Received INVOKE event for request: %s (state: ACTIVE)", event.RequestID)

		case Shutdown:
			log.Printf("Received SHUTDOWN event, reason: %s", event.ShutdownReason)
			return m.shutdown(ctx)
		}
	}
}

// setState updates the state and signals the flush loop to adjust interval
func (m *Manager) setState(newState State) {
	oldState := State(m.state.Swap(int32(newState)))
	if oldState != newState {
		log.Printf("State transition: %s -> %s", oldState, newState)
		// Signal flush loop to recalculate interval
		select {
		case m.intervalChange <- struct{}{}:
		default:
		}
	}
}

// getState returns the current state
func (m *Manager) getState() State {
	return State(m.state.Load())
}

// getFlushInterval returns the appropriate flush interval based on current state
func (m *Manager) getFlushInterval() time.Duration {
	baseInterval := time.Duration(m.cfg.FlushIntervalMs) * time.Millisecond

	switch m.getState() {
	case StateActive:
		// Normal interval during active invocation
		return baseInterval
	case StateIdle:
		// Longer interval when idle (default 3x)
		return baseInterval * time.Duration(m.cfg.IdleFlushMultiplier)
	case StateFlushing:
		// Slightly longer during critical flush to avoid conflicts
		return baseInterval * 3 / 2
	default:
		return baseInterval
	}
}

func (m *Manager) flushLoop(ctx context.Context) {
	interval := m.getFlushInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Flush loop started with interval: %v (state: %s)", interval, m.getState())

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopFlush:
			return
		case <-m.intervalChange:
			// State changed, adjust the ticker interval
			newInterval := m.getFlushInterval()
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
				log.Printf("Flush interval adjusted to: %v (state: %s)", interval, m.getState())
			}
		case <-ticker.C:
			m.flush(ctx, false)
		case <-m.buffer.Ready():
			// Check if we have enough for a batch (by count or bytes)
			if m.shouldFlush() {
				m.flush(ctx, false)
			}
		}
	}
}

// shouldFlush returns true if buffer has enough data to flush
func (m *Manager) shouldFlush() bool {
	if m.buffer.Len() >= m.cfg.BatchSize {
		return true
	}
	if m.cfg.MaxBatchSizeBytes > 0 && m.buffer.ByteSize() >= m.cfg.MaxBatchSizeBytes {
		return true
	}
	return false
}

// onRuntimeDone is called when platform.runtimeDone is received
// This triggers a critical flush to ensure all logs are shipped at invocation end
func (m *Manager) onRuntimeDone(requestID string) {
	log.Printf("Received runtimeDone for request: %s, triggering critical flush (buffer: %d entries)", requestID, m.buffer.Len())

	// Transition to flushing state
	m.setState(StateFlushing)

	// Run in goroutine to not block telemetry API response
	m.criticalFlushWg.Add(1)
	go func() {
		defer m.criticalFlushWg.Done()
		// Use a timeout context for critical flush
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		m.criticalFlush(ctx)
		m.setState(StateIdle)
	}()
}

// flush performs a regular flush with standard retries
func (m *Manager) flush(ctx context.Context, isCritical bool) {
	m.criticalFlushMu.Lock()
	defer m.criticalFlushMu.Unlock()

	var entries []buffer.LogEntry
	if m.cfg.MaxBatchSizeBytes > 0 {
		entries = m.buffer.FlushBySize(m.cfg.BatchSize, m.cfg.MaxBatchSizeBytes)
	} else {
		entries = m.buffer.Flush(m.cfg.BatchSize)
	}

	if len(entries) == 0 {
		return
	}

	batch := loki.NewBatch(m.labels, m.cfg.ExtractRequestID)
	batch.Add(entries)

	pushReq := batch.ToPushRequest()

	var err error
	if isCritical {
		err = m.lokiClient.PushCritical(ctx, pushReq)
	} else {
		err = m.lokiClient.Push(ctx, pushReq)
	}

	if err != nil {
		log.Printf("Failed to push logs to Loki: %v", err)
	} else {
		log.Printf("Pushed %d log entries to Loki", len(entries))
	}
}

// criticalFlush flushes all buffered logs with higher retry count
func (m *Manager) criticalFlush(ctx context.Context) {
	log.Printf("Critical flush starting (buffer: %d entries)", m.buffer.Len())

	// Flush all remaining entries
	for m.buffer.Len() > 0 {
		log.Printf("Critical flush loop iteration (buffer: %d)", m.buffer.Len())
		var entries []buffer.LogEntry
		if m.cfg.MaxBatchSizeBytes > 0 {
			entries = m.buffer.FlushBySize(m.cfg.BatchSize, m.cfg.MaxBatchSizeBytes)
		} else {
			entries = m.buffer.Flush(m.cfg.BatchSize)
		}

		log.Printf("Got %d entries from buffer", len(entries))

		if len(entries) == 0 {
			break
		}

		batch := loki.NewBatch(m.labels, m.cfg.ExtractRequestID)
		batch.Add(entries)

		pushReq := batch.ToPushRequest()
		jsonBody, _ := json.Marshal(pushReq)
		log.Printf("Pushing to Loki: %s", string(jsonBody[:min(500, len(jsonBody))]))
		if err := m.lokiClient.PushCritical(ctx, pushReq); err != nil {
			log.Printf("Critical flush failed: %v", err)
		} else {
			log.Printf("Critical flush pushed %d entries to Loki", len(entries))
		}
	}
	log.Printf("Critical flush complete")
}

func (m *Manager) shutdown(ctx context.Context) error {
	// Stop the flush loop
	close(m.stopFlush)

	// Wait for any pending critical flushes to complete
	log.Printf("Waiting for pending critical flushes...")
	m.criticalFlushWg.Wait()

	// Shutdown telemetry server
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := m.telemetryServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down telemetry server: %v", err)
	}

	// Drain and flush all remaining logs with critical retries
	log.Printf("Draining buffer...")
	entries := m.buffer.Drain()

	if len(entries) > 0 {
		log.Printf("Flushing %d remaining log entries with critical retries", len(entries))
		batch := loki.NewBatch(m.labels, m.cfg.ExtractRequestID)
		batch.Add(entries)

		pushReq := batch.ToPushRequest()
		if err := m.lokiClient.PushCritical(ctx, pushReq); err != nil {
			log.Printf("Failed to push final logs to Loki: %v", err)
			// Continue shutdown even on error
		}
	}

	log.Printf("Shutdown complete")
	return nil
}
