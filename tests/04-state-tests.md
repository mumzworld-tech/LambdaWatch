# State Machine Tests

## 4.1 State Definitions

### TC-4.1.1: Initial State
- **Action**: Create new Manager
- **Expected**: State is IDLE

### TC-4.1.2: State String Representation
- **Action**: Call `State.String()` for each state
- **Expected**:
  - StateIdle → "IDLE"
  - StateActive → "ACTIVE"
  - StateFlushing → "FLUSHING"

---

## 4.2 State Transitions

### TC-4.2.1: IDLE → ACTIVE (on INVOKE)
- **Setup**: State is IDLE
- **Action**: Receive INVOKE event
- **Expected**:
  - Log "State transition: IDLE -> ACTIVE"
  - `getState() == StateActive`

### TC-4.2.2: ACTIVE → FLUSHING (on runtimeDone)
- **Setup**: State is ACTIVE
- **Action**: Receive `platform.runtimeDone`
- **Expected**:
  - Log "State transition: ACTIVE -> FLUSHING"
  - `getState() == StateFlushing`

### TC-4.2.3: FLUSHING → IDLE (after critical flush)
- **Setup**: State is FLUSHING
- **Action**: Critical flush completes
- **Expected**:
  - Log "State transition: FLUSHING -> IDLE"
  - `getState() == StateIdle`

### TC-4.2.4: No Log on Same State
- **Setup**: State is ACTIVE
- **Action**: Call `setState(StateActive)`
- **Expected**: No "State transition" log (state unchanged)

---

## 4.3 Interval Adjustments

### TC-4.3.1: ACTIVE Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, state=ACTIVE
- **Action**: `getFlushInterval()`
- **Expected**: 1000ms

### TC-4.3.2: IDLE Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, `LOKI_IDLE_FLUSH_MULTIPLIER=3`, state=IDLE
- **Action**: `getFlushInterval()`
- **Expected**: 3000ms

### TC-4.3.3: FLUSHING Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, state=FLUSHING
- **Action**: `getFlushInterval()`
- **Expected**: 1500ms (1.5x base)

### TC-4.3.4: Ticker Reset on State Change
- **Setup**: Running in IDLE (3s interval)
- **Action**: Transition to ACTIVE
- **Expected**:
  - Log "Flush interval adjusted to: 1s (state: ACTIVE)"
  - Next flush at 1s, not remainder of 3s

---

## 4.4 Interval Signal Channel

### TC-4.4.1: Signal Sent on State Change
- **Setup**: Flush loop running
- **Action**: Change state
- **Expected**: `intervalChange` channel receives signal

### TC-4.4.2: Non-Blocking Signal
- **Action**: Rapid state changes
- **Expected**: No blocking (channel buffered)

### TC-4.4.3: No Signal on Same State
- **Action**: `setState(currentState)`
- **Expected**: No signal sent

---

## 4.5 Atomic State Operations

### TC-4.5.1: Concurrent State Read
- **Action**: Multiple goroutines calling `getState()`
- **Expected**: No race conditions

### TC-4.5.2: Concurrent State Write
- **Action**: Multiple goroutines calling `setState()`
- **Expected**: No race conditions, state consistent

### TC-4.5.3: State Swap Returns Old Value
- **Setup**: State is ACTIVE
- **Action**: `setState(StateFlushing)`
- **Expected**: Old state (ACTIVE) available for comparison
