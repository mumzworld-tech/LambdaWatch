# Lifecycle Tests

## 7.1 Registration

### TC-7.1.1: Successful Registration

- **Setup**: Mock Lambda Extensions API
- **Action**: `extClient.Register(ctx)`
- **Expected**:
  - POST to `/2020-01-01/extension/register`
  - `Lambda-Extension-Name` header set
  - Extension ID stored from response header

### TC-7.1.2: Registration Response Parsed

- **Setup**: Mock returns function info
- **Action**: Register
- **Expected**:
  - `FunctionName` parsed
  - `FunctionVersion` parsed
  - Used in labels

### TC-7.1.3: Registration Failure

- **Setup**: Mock returns 500
- **Action**: Register
- **Expected**: Error returned, extension exits

---

## 7.2 Telemetry Subscription

### TC-7.2.1: Successful Subscription

- **Setup**: Mock Telemetry API
- **Action**: `telemetryClient.Subscribe(ctx, uri)`
- **Expected**:
  - PUT to `/2022-07-01/telemetry`
  - Body contains types: platform, function, extension
  - Extension ID header set

### TC-7.2.2: Subscription Buffering Config

- **Action**: Check subscription request body
- **Expected**:
  - MaxItems: 1000
  - MaxBytes: 262144
  - TimeoutMs: 100

### TC-7.2.3: Subscription Failure

- **Setup**: Mock returns error
- **Action**: Subscribe
- **Expected**: Error logged, extension may continue or exit

---

## 7.3 Event Loop

### TC-7.3.1: INVOKE Event Handling

- **Setup**: Mock NextEvent returns INVOKE
- **Action**: Event loop iteration
- **Expected**:
  - State set to ACTIVE
  - Log "Received INVOKE event for request: xxx"
  - Waits for invocationDone

### TC-7.3.2: SHUTDOWN Event Handling

- **Setup**: Mock NextEvent returns SHUTDOWN
- **Action**: Event loop iteration
- **Expected**:
  - Log "Received SHUTDOWN event, reason: xxx"
  - `shutdown()` called
  - Event loop exits

### TC-7.3.3: NextEvent Error

- **Setup**: Mock NextEvent returns error
- **Action**: Event loop iteration
- **Expected**: Error returned, extension exits

### TC-7.3.4: Context Cancellation

- **Setup**: Cancel context while waiting
- **Action**: Event loop
- **Expected**: Returns context error

---

## 7.4 Invocation Synchronization

### TC-7.4.1: Wait for runtimeDone

- **Setup**: Receive INVOKE
- **Action**: Check event loop behavior
- **Expected**: Blocks on `invocationDone` channel until runtimeDone processed

### TC-7.4.2: runtimeDone Unblocks Loop

- **Setup**: Event loop waiting on invocationDone
- **Action**: onRuntimeDone completes critical flush
- **Expected**:
  - Channel closed
  - Event loop continues to NextEvent

### TC-7.4.3: Fresh Channel Per Invocation

- **Setup**: Multiple invocations
- **Action**: Check channel creation
- **Expected**: New channel created for each INVOKE

### TC-7.4.4: Context Cancel While Waiting

- **Setup**: Waiting on invocationDone
- **Action**: Cancel context
- **Expected**: Returns immediately with context error

---

## 7.5 Shutdown Sequence

### TC-7.5.1: Stop Flush Loop

- **Action**: Shutdown
- **Expected**: `stopFlush` channel closed, flush loop exits

### TC-7.5.2: Telemetry Server Shutdown

- **Action**: Shutdown
- **Expected**:
  - Server shutdown with 2s timeout
  - Graceful connection close

### TC-7.5.3: Final Telemetry Delivery Wait

- **Action**: Shutdown
- **Expected**: 100ms sleep for final telemetry delivery

### TC-7.5.4: Buffer Drain

- **Action**: Shutdown with entries in buffer
- **Expected**:
  - Log "Draining buffer..."
  - All entries flushed with critical retries

### TC-7.5.5: Shutdown Timeout

- **Setup**: Loki very slow
- **Action**: Shutdown
- **Expected**: Completes within reasonable time, doesn't hang

---

## 7.6 Label Building

### TC-7.6.1: Function Labels

- **Setup**: Register with function info
- **Action**: Build labels
- **Expected**:
  - `function_name` from registration
  - `function_version` from registration

### TC-7.6.2: Region Label

- **Setup**: `AWS_REGION=ap-south-1`
- **Action**: Build labels
- **Expected**: `region=ap-south-1`

### TC-7.6.3: Source Label

- **Action**: Build labels
- **Expected**: `source=lambda`

### TC-7.6.4: Custom Labels Merged

- **Setup**: `LOKI_LABELS={"custom":"value"}`
- **Action**: Build labels
- **Expected**: Custom labels included with auto labels

### TC-7.6.5: Label Override

- **Setup**: Custom label with same key as auto label
- **Action**: Build labels
- **Expected**: Auto label wins (function_name, etc.)

---

## 7.7 Component Initialization

### TC-7.7.1: Init Order

- **Action**: `manager.init(ctx)`
- **Expected**:
  1. Register with Extensions API
  2. Build labels
  3. Create Loki client
  4. Start telemetry server
  5. Subscribe to Telemetry API

### TC-7.7.2: Init Failure Handling

- **Setup**: Registration fails
- **Action**: init()
- **Expected**: Error returned, no partial initialization

### TC-7.7.3: Logger Buffer Set

- **Action**: NewManager()
- **Expected**: Logger buffer set for direct extension log capture

---

## 7.8 Flush Loop

### TC-7.8.1: Loop Starts

- **Action**: `manager.Run(ctx)`
- **Expected**: Flush loop goroutine started

### TC-7.8.2: Loop Stops on Context Cancel

- **Action**: Cancel context
- **Expected**: Flush loop exits cleanly

### TC-7.8.3: Loop Stops on stopFlush

- **Action**: Close stopFlush channel
- **Expected**: Flush loop exits cleanly

### TC-7.8.4: Multiple Stop Signals

- **Action**: Context cancel AND stopFlush close
- **Expected**: No panic, exits once
