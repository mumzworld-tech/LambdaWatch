# Configuration Tests

## 1.1 Required Configuration

### TC-1.1.1: Missing LOKI_URL

- **Setup**: Unset `LOKI_URL`
- **Action**: Start extension
- **Expected**: Extension exits with error "LOKI_URL environment variable is required"

### TC-1.1.2: Valid LOKI_URL

- **Setup**: Set `LOKI_URL=https://<your-loki-endpoint>/loki/api/v1/push`
- **Action**: Start extension
- **Expected**: Extension starts successfully

---

## 1.2 Authentication Configuration

### TC-1.2.1: Basic Auth

- **Setup**:
  ```
  LOKI_USERNAME=<your-username>
  LOKI_PASSWORD=glc_xxx
  ```
- **Action**: Push logs
- **Expected**: Request includes `Authorization: Basic <base64>` header

### TC-1.2.2: Bearer Token Auth

- **Setup**:
  ```
  LOKI_API_KEY=glc_xxx
  ```
- **Action**: Push logs
- **Expected**: Request includes `Authorization: Bearer glc_xxx` header

### TC-1.2.3: Tenant ID

- **Setup**:
  ```
  LOKI_TENANT_ID=my-tenant
  ```
- **Action**: Push logs
- **Expected**: Request includes `X-Scope-OrgID: my-tenant` header

### TC-1.2.4: Auth Priority (API Key over Basic)

- **Setup**:
  ```
  LOKI_USERNAME=user
  LOKI_PASSWORD=pass
  LOKI_API_KEY=glc_xxx
  ```
- **Action**: Push logs
- **Expected**: Bearer token used, not Basic auth

---

## 1.3 Batching Configuration

### TC-1.3.1: Default Batch Size

- **Setup**: Don't set `LOKI_BATCH_SIZE`
- **Action**: Add 150 logs, trigger flush
- **Expected**: First flush contains 100 entries (default)

### TC-1.3.2: Custom Batch Size

- **Setup**: `LOKI_BATCH_SIZE=50`
- **Action**: Add 150 logs, trigger flush
- **Expected**: First flush contains 50 entries

### TC-1.3.3: Default Max Batch Bytes

- **Setup**: Don't set `LOKI_MAX_BATCH_SIZE_BYTES`
- **Action**: Check config
- **Expected**: Default is 5MB (5242880 bytes)

### TC-1.3.4: Custom Max Batch Bytes

- **Setup**: `LOKI_MAX_BATCH_SIZE_BYTES=1048576` (1MB)
- **Action**: Add logs totaling 2MB
- **Expected**: Batch split at 1MB boundary

---

## 1.4 Interval Configuration

### TC-1.4.1: Default Flush Interval

- **Setup**: Don't set `LOKI_FLUSH_INTERVAL_MS`
- **Action**: Check flush timing in ACTIVE state
- **Expected**: Flush every 1000ms (1 second)

### TC-1.4.2: Custom Flush Interval

- **Setup**: `LOKI_FLUSH_INTERVAL_MS=500`
- **Action**: Check flush timing
- **Expected**: Flush every 500ms

### TC-1.4.3: Default Idle Multiplier

- **Setup**: Don't set `LOKI_IDLE_FLUSH_MULTIPLIER`
- **Action**: Check flush timing in IDLE state
- **Expected**: Flush every 3000ms (3x default)

### TC-1.4.4: Custom Idle Multiplier

- **Setup**:
  ```
  LOKI_FLUSH_INTERVAL_MS=1000
  LOKI_IDLE_FLUSH_MULTIPLIER=5
  ```
- **Action**: Check flush timing in IDLE state
- **Expected**: Flush every 5000ms

---

## 1.5 Retry Configuration

### TC-1.5.1: Default Max Retries

- **Setup**: Don't set `LOKI_MAX_RETRIES`
- **Action**: Simulate Loki failure
- **Expected**: 3 retry attempts for regular flush

### TC-1.5.2: Custom Max Retries

- **Setup**: `LOKI_MAX_RETRIES=5`
- **Action**: Simulate Loki failure
- **Expected**: 5 retry attempts

### TC-1.5.3: Default Critical Retries

- **Setup**: Don't set `LOKI_CRITICAL_FLUSH_RETRIES`
- **Action**: Simulate Loki failure during critical flush
- **Expected**: 5 retry attempts

### TC-1.5.4: Custom Critical Retries

- **Setup**: `LOKI_CRITICAL_FLUSH_RETRIES=10`
- **Action**: Simulate Loki failure during critical flush
- **Expected**: 10 retry attempts

---

## 1.6 Compression Configuration

### TC-1.6.1: Gzip Enabled (Default)

- **Setup**: Don't set `LOKI_ENABLE_GZIP`
- **Action**: Push logs > 1KB
- **Expected**: Request includes `Content-Encoding: gzip`

### TC-1.6.2: Gzip Disabled

- **Setup**: `LOKI_ENABLE_GZIP=false`
- **Action**: Push logs
- **Expected**: No `Content-Encoding` header

### TC-1.6.3: Compression Threshold

- **Setup**: `LOKI_COMPRESSION_THRESHOLD=2048`
- **Action**: Push 1KB payload, then 3KB payload
- **Expected**: 1KB not compressed, 3KB compressed

---

## 1.7 Labels Configuration

### TC-1.7.1: Custom Labels JSON

- **Setup**: `LOKI_LABELS={"env":"prod","team":"platform"}`
- **Action**: Push logs
- **Expected**: Logs have `env=prod` and `team=platform` labels

### TC-1.7.2: Invalid Labels JSON

- **Setup**: `LOKI_LABELS=invalid-json`
- **Action**: Start extension
- **Expected**: Extension fails with JSON parse error

### TC-1.7.3: SERVICE_NAME Label

- **Setup**: `SERVICE_NAME=my-service`
- **Action**: Push logs
- **Expected**: Logs have `service_name=my-service` label

### TC-1.7.4: Auto Labels (function_name, region, source)

- **Setup**: Deploy to Lambda
- **Action**: Push logs
- **Expected**: Logs have `function_name`, `function_version`, `region`, `source=lambda`

---

## 1.8 Buffer Configuration

### TC-1.8.1: Default Buffer Size

- **Setup**: Don't set `BUFFER_SIZE`
- **Action**: Check buffer capacity
- **Expected**: 10000 entries max

### TC-1.8.2: Custom Buffer Size

- **Setup**: `BUFFER_SIZE=5000`
- **Action**: Add 6000 entries
- **Expected**: Buffer contains 5000 entries, oldest dropped

---

## 1.9 Message Configuration

### TC-1.9.1: Default Max Line Size

- **Setup**: Don't set `LOKI_MAX_LINE_SIZE`
- **Action**: Log 300KB message
- **Expected**: Message split into chunks of ~200KB

### TC-1.9.2: Custom Max Line Size

- **Setup**: `LOKI_MAX_LINE_SIZE=1024`
- **Action**: Log 3KB message
- **Expected**: Message split into 3 chunks

### TC-1.9.3: Request ID Extraction Enabled

- **Setup**: `LOKI_EXTRACT_REQUEST_ID=true` (default)
- **Action**: Push function logs
- **Expected**: Logs grouped by request_id label

### TC-1.9.4: Request ID Extraction Disabled

- **Setup**: `LOKI_EXTRACT_REQUEST_ID=false`
- **Action**: Push function logs
- **Expected**: All logs in single stream, no request_id label
