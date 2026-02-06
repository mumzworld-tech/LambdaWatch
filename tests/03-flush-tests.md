# Flush Behavior Tests

## 3.1 Timer-Based Flush

### TC-3.1.1: Flush at Interval (ACTIVE State)
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, state=ACTIVE
- **Action**: Add logs, wait 1.5 seconds
- **Expected**: Flush triggered once
- **Verify**: Log message "Pushing X log entries to Loki"

### TC-3.1.2: Flush at Idle Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, `LOKI_IDLE_FLUSH_MULTIPLIER=3`, state=IDLE
- **Action**: Add logs, wait 4 seconds
- **Expected**: Flush triggered once (at 3s mark)
- **Verify**: Log "Flush interval adjusted to: 3s (state: IDLE)"

### TC-3.1.3: Interval Change on State Transition
- **Setup**: Start in IDLE (3s interval)
- **Action**: Receive INVOKE event (transition to ACTIVE)
- **Expected**:
  - Log "State transition: IDLE -> ACTIVE"
  - Log "Flush interval adjusted to: 1s (state: ACTIVE)"

### TC-3.1.4: No Flush When Buffer Empty
- **Setup**: Empty buffer
- **Action**: Wait for flush interval
- **Expected**: No "Pushing X log entries" message

---

## 3.2 Ready-Signal Flush

### TC-3.2.1: Flush on Buffer Ready
- **Setup**: `LOKI_BATCH_SIZE=10`
- **Action**: `AddBatch()` with 15 entries
- **Expected**: Flush triggered immediately (not waiting for timer)

### TC-3.2.2: No Flush Below Batch Size
- **Setup**: `LOKI_BATCH_SIZE=100`
- **Action**: Add 50 entries, signal ready
- **Expected**: No immediate flush (wait for timer)

### TC-3.2.3: Flush on Byte Size Threshold
- **Setup**: `LOKI_MAX_BATCH_SIZE_BYTES=1024`
- **Action**: Add entries totaling 2KB
- **Expected**: Flush triggered when threshold reached

---

## 3.3 Critical Flush

### TC-3.3.1: Critical Flush on runtimeDone
- **Setup**: Buffer has 10 entries
- **Action**: Receive `platform.runtimeDone` event
- **Expected**:
  - Log "Received PLATFORM_RUNTIME_DONE event for request: xxx"
  - Log "State transition: ACTIVE -> FLUSHING"
  - Log "Critical flush: 10 entries"
  - All entries flushed before "Invocation complete"

### TC-3.3.2: Critical Flush Higher Retries
- **Setup**: `LOKI_CRITICAL_FLUSH_RETRIES=5`, Loki returns 500 error
- **Action**: Trigger critical flush
- **Expected**: 5 retry attempts (vs 3 for regular)

### TC-3.3.3: Critical Flush Timeout
- **Setup**: Loki very slow (>10s per request)
- **Action**: Trigger critical flush
- **Expected**: Flush aborted after 10s timeout

### TC-3.3.4: Critical Flush Doesn't Loop Infinitely
- **Setup**: Buffer with entries
- **Action**: Trigger critical flush
- **Expected**:
  - Flushes original entries
  - Logs generated during flush NOT flushed in same cycle
  - No infinite loop

### TC-3.3.5: Multiple runtimeDone Events
- **Setup**: Rapid invocations
- **Action**: Multiple runtimeDone events in quick succession
- **Expected**: Each processed correctly, no race conditions

---

## 3.4 Batch Size Limits

### TC-3.4.1: Batch Size Limit
- **Setup**: `LOKI_BATCH_SIZE=50`, buffer has 120 entries
- **Action**: Single flush
- **Expected**: Only 50 entries in request

### TC-3.4.2: Multiple Batches in Critical Flush
- **Setup**: `LOKI_BATCH_SIZE=50`, buffer has 120 entries
- **Action**: Critical flush
- **Expected**: 3 separate push requests (50+50+20)

### TC-3.4.3: Byte Size Limit
- **Setup**: `LOKI_MAX_BATCH_SIZE_BYTES=1024`, entries totaling 3KB
- **Action**: Flush
- **Expected**: Batch split at ~1KB boundaries

### TC-3.4.4: Count vs Byte Limit (Count First)
- **Setup**: `LOKI_BATCH_SIZE=10`, `LOKI_MAX_BATCH_SIZE_BYTES=1MB`
- **Action**: Add 20 small entries
- **Expected**: Batch limited to 10 (count limit reached first)

### TC-3.4.5: Count vs Byte Limit (Bytes First)
- **Setup**: `LOKI_BATCH_SIZE=100`, `LOKI_MAX_BATCH_SIZE_BYTES=500`
- **Action**: Add 20 entries of 100 bytes each
- **Expected**: Batch limited to ~5 entries (byte limit reached first)

---

## 3.5 Shutdown Flush

### TC-3.5.1: Drain on Shutdown
- **Setup**: Buffer has entries
- **Action**: Receive SHUTDOWN event
- **Expected**:
  - Log "Received SHUTDOWN event, reason: xxx"
  - Log "Draining buffer..."
  - All remaining entries flushed

### TC-3.5.2: Shutdown Waits for Critical Flush
- **Setup**: Critical flush in progress
- **Action**: Receive SHUTDOWN event
- **Expected**: Shutdown waits for critical flush to complete

### TC-3.5.3: Final Logs Captured
- **Setup**: Function completes, generates final logs
- **Action**: SHUTDOWN received
- **Expected**:
  - Sleep 100ms for final telemetry delivery
  - Then drain and flush

---

## 3.6 Flush Mutex

### TC-3.6.1: No Concurrent Flushes
- **Action**: Trigger timer flush and critical flush simultaneously
- **Expected**: One waits for other (mutex protection)

### TC-3.6.2: FlushBatch Thread Safety
- **Action**: Multiple goroutines calling flushBatch
- **Expected**: No race conditions, no duplicate entries flushed
