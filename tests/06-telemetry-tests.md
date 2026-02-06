# Telemetry Server Tests

## 6.1 HTTP Server

### TC-6.1.1: Server Starts on Port 8080
- **Action**: Start server
- **Expected**: Listening on `sandbox.localdomain:8080`

### TC-6.1.2: POST Method Only
- **Action**: Send GET request to `/`
- **Expected**: 405 Method Not Allowed

### TC-6.1.3: Invalid JSON Body
- **Action**: POST non-JSON body
- **Expected**: 400 Bad Request

### TC-6.1.4: Empty Event Array
- **Action**: POST `[]`
- **Expected**: 200 OK, no entries added to buffer

### TC-6.1.5: Graceful Shutdown
- **Action**: Call `server.Shutdown(ctx)`
- **Expected**: Server stops accepting connections

---

## 6.2 Platform Events

### TC-6.2.1: platform.start Event
- **Input**:
  ```json
  {
    "type": "platform.start",
    "time": "2026-02-05T21:34:18.205Z",
    "record": {"requestId": "abc-123", "version": "$LATEST"}
  }
  ```
- **Expected**:
  - `currentRequestID` set to "abc-123"
  - Log entry: `START RequestId: abc-123 Version: $LATEST`

### TC-6.2.2: platform.runtimeDone Event
- **Input**:
  ```json
  {
    "type": "platform.runtimeDone",
    "time": "2026-02-05T21:34:19.572Z",
    "record": {"requestId": "abc-123", "status": "success", "metrics": {...}}
  }
  ```
- **Expected**:
  - `onRuntimeDone` callback triggered with requestId
  - Log entry contains raw JSON record

### TC-6.2.3: platform.report Event
- **Input**:
  ```json
  {
    "type": "platform.report",
    "time": "2026-02-05T21:34:20.458Z",
    "record": {
      "requestId": "abc-123",
      "metrics": {
        "durationMs": 2251.86,
        "billedDurationMs": 3114,
        "memorySizeMB": 1024,
        "maxMemoryUsedMB": 184,
        "initDurationMs": 861.71
      }
    }
  }
  ```
- **Expected**: Log entry formatted as:
  ```
  REPORT RequestId: abc-123  Duration: 2251.86 ms  Billed Duration: 3114 ms  Memory Size: 1024 MB  Max Memory Used: 184 MB  Init Duration: 861.71 ms
  ```

### TC-6.2.4: platform.report Without Init Duration
- **Input**: Report without `initDurationMs`
- **Expected**: No "Init Duration" in output (warm start)

---

## 6.3 Function Logs

### TC-6.3.1: Function Log Event
- **Input**:
  ```json
  {
    "type": "function",
    "time": "2026-02-05T21:34:18.835Z",
    "record": "{\"level\":\"info\",\"message\":\"Hello\"}"
  }
  ```
- **Expected**:
  - Entry added to buffer
  - Type = "function"
  - RequestID from `currentRequestID`

### TC-6.3.2: Function Log with Lambda Prefix
- **Input**:
  ```json
  {
    "type": "function",
    "time": "2026-02-05T21:34:18.835Z",
    "record": "2026-02-05T21:34:18.835Z\tabc-123\tINFO\t{\"message\":\"Hello\"}"
  }
  ```
- **Expected**:
  - Prefix stripped, only JSON remains
  - Timestamp extracted from prefix

### TC-6.3.3: Non-JSON Function Log
- **Input**: `"record": "Plain text log message"`
- **Expected**: Stored as-is

---

## 6.4 Extension Logs

### TC-6.4.1: Other Extension Logs
- **Input**:
  ```json
  {
    "type": "extension",
    "time": "2026-02-05T21:34:18.835Z",
    "record": "{\"source\":\"other-extension\",\"message\":\"Log\"}"
  }
  ```
- **Expected**: Added to buffer (not our extension)

### TC-6.4.2: Own Extension Logs Filtered
- **Input**:
  ```json
  {
    "type": "extension",
    "record": "{\"context\":\"LambdaWatch\",\"message\":\"Internal log\"}"
  }
  ```
- **Expected**: NOT added to buffer (filtered out)

### TC-6.4.3: Own Extension Detection
- **Input**: Various log formats with "context":"LambdaWatch"
- **Expected**: All filtered regardless of other fields

---

## 6.5 Request ID Handling

### TC-6.5.1: RequestID from platform.start
- **Setup**: Receive platform.start with requestId
- **Action**: Receive function logs
- **Expected**: All subsequent logs have that requestId

### TC-6.5.2: RequestID Extraction from Message
- **Setup**: No platform.start yet, `LOKI_EXTRACT_REQUEST_ID=true`
- **Input**: Log containing `RequestId: abc-123`
- **Expected**: RequestID extracted from message

### TC-6.5.3: RequestID Extraction Disabled
- **Setup**: `LOKI_EXTRACT_REQUEST_ID=false`
- **Input**: Log containing RequestId
- **Expected**: RequestID not extracted, empty

### TC-6.5.4: RequestID Persists Across Batches
- **Setup**: platform.start sets requestId
- **Action**: Multiple telemetry batches
- **Expected**: All batches use same requestId until next platform.start

---

## 6.6 Message Processing

### TC-6.6.1: Large Message Split
- **Setup**: `LOKI_MAX_LINE_SIZE=100`
- **Input**: 350-byte message
- **Expected**:
  - Split into 4 chunks
  - Each chunk prefixed with `[chunk N/4]`
  - Timestamps incremented (ts, ts+1, ts+2, ts+3)

### TC-6.6.2: Message Under Limit
- **Setup**: `LOKI_MAX_LINE_SIZE=1000`
- **Input**: 500-byte message
- **Expected**: Single entry, no splitting

### TC-6.6.3: Max Line Size Zero (No Limit)
- **Setup**: `LOKI_MAX_LINE_SIZE=0`
- **Input**: Very large message
- **Expected**: No splitting

---

## 6.7 Timestamp Parsing

### TC-6.7.1: RFC3339Nano Timestamp
- **Input**: `"time": "2026-02-05T21:34:18.205123456Z"`
- **Expected**: Parsed to milliseconds correctly

### TC-6.7.2: Invalid Timestamp
- **Input**: `"time": "invalid"`
- **Expected**: Falls back to `time.Now().UnixMilli()`

### TC-6.7.3: Timestamp from Lambda Log Prefix
- **Input**: Record with `2026-02-05T21:34:18.835Z\t...` prefix
- **Expected**: Timestamp extracted from prefix

---

## 6.8 Batch Processing

### TC-6.8.1: Multiple Events in Batch
- **Input**: Array with 5 different events
- **Expected**: All processed, correct types assigned

### TC-6.8.2: Event Order Preserved
- **Input**: Events in specific order
- **Expected**: Buffer entries in same order

### TC-6.8.3: Mixed Event Types
- **Input**: platform.start, function, function, platform.runtimeDone
- **Expected**:
  - RequestID set by platform.start
  - Function logs use that requestId
  - runtimeDone triggers callback

### TC-6.8.4: runtimeDone After Buffer Add
- **Input**: Batch with function logs and runtimeDone
- **Expected**: Function logs added to buffer BEFORE onRuntimeDone called
