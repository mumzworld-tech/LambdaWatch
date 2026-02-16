# Loki Client Tests

## 5.1 Basic Push

### TC-5.1.1: Successful Push

- **Setup**: Mock Loki server returns 200
- **Action**: `client.Push(ctx, pushRequest)`
- **Expected**: Returns nil (success)

### TC-5.1.2: Push Empty Request

- **Action**: `client.Push(ctx, nil)`
- **Expected**: Returns nil (no-op)

### TC-5.1.3: Push Empty Streams

- **Action**: `client.Push(ctx, &PushRequest{Streams: []Stream{}})`
- **Expected**: Returns nil (no-op)

---

## 5.2 Retry Logic

### TC-5.2.1: Retry on 500 Error

- **Setup**: Mock Loki returns 500, then 200
- **Action**: `client.Push(ctx, request)`
- **Expected**:
  - First attempt fails
  - Retries after backoff
  - Second attempt succeeds

### TC-5.2.2: Retry on 429 (Rate Limited)

- **Setup**: Mock Loki returns 429, then 200
- **Action**: `client.Push(ctx, request)`
- **Expected**: Retries after backoff

### TC-5.2.3: No Retry on 400 (Bad Request)

- **Setup**: Mock Loki returns 400
- **Action**: `client.Push(ctx, request)`
- **Expected**: Returns error immediately, no retry

### TC-5.2.4: No Retry on 401 (Unauthorized)

- **Setup**: Mock Loki returns 401
- **Action**: `client.Push(ctx, request)`
- **Expected**: Returns error immediately, no retry

### TC-5.2.5: Max Retries Exhausted

- **Setup**: `LOKI_MAX_RETRIES=3`, Mock always returns 500
- **Action**: `client.Push(ctx, request)`
- **Expected**:
  - 4 total attempts (1 initial + 3 retries)
  - Returns error "push failed after 3 retries"

### TC-5.2.6: Critical Push Higher Retries

- **Setup**: `LOKI_CRITICAL_FLUSH_RETRIES=5`, Mock always returns 500
- **Action**: `client.PushCritical(ctx, request)`
- **Expected**: 6 total attempts (1 initial + 5 retries)

---

## 5.3 Exponential Backoff

### TC-5.3.1: Backoff Timing

- **Setup**: Mock always returns 500
- **Action**: Measure time between retries
- **Expected**:
  - Retry 1: ~100ms delay
  - Retry 2: ~200ms delay
  - Retry 3: ~400ms delay

### TC-5.3.2: Context Cancellation During Backoff

- **Setup**: Mock returns 500
- **Action**: Cancel context during backoff wait
- **Expected**: Returns `context.Canceled` error

---

## 5.4 Compression

### TC-5.4.1: Gzip Enabled Above Threshold

- **Setup**: `LOKI_ENABLE_GZIP=true`, `LOKI_COMPRESSION_THRESHOLD=1024`
- **Action**: Push 2KB payload
- **Expected**:
  - `Content-Encoding: gzip` header set
  - Body is gzip compressed

### TC-5.4.2: No Compression Below Threshold

- **Setup**: `LOKI_ENABLE_GZIP=true`, `LOKI_COMPRESSION_THRESHOLD=1024`
- **Action**: Push 500 byte payload
- **Expected**: No `Content-Encoding` header, body uncompressed

### TC-5.4.3: Gzip Disabled

- **Setup**: `LOKI_ENABLE_GZIP=false`
- **Action**: Push 10KB payload
- **Expected**: No compression regardless of size

### TC-5.4.4: Compression Reduces Size

- **Setup**: Gzip enabled
- **Action**: Push repetitive JSON payload
- **Expected**: Compressed size < original size

---

## 5.5 Authentication

### TC-5.5.1: Basic Auth Header

- **Setup**: `LOKI_USERNAME=user`, `LOKI_PASSWORD=pass`
- **Action**: Push logs
- **Expected**: `Authorization: Basic dXNlcjpwYXNz` header

### TC-5.5.2: Bearer Token Header

- **Setup**: `LOKI_API_KEY=my-token`
- **Action**: Push logs
- **Expected**: `Authorization: Bearer my-token` header

### TC-5.5.3: Tenant ID Header

- **Setup**: `LOKI_TENANT_ID=tenant-123`
- **Action**: Push logs
- **Expected**: `X-Scope-OrgID: tenant-123` header

### TC-5.5.4: All Auth Combined

- **Setup**: API key + Tenant ID
- **Action**: Push logs
- **Expected**: Both `Authorization` and `X-Scope-OrgID` headers

### TC-5.5.5: No Auth

- **Setup**: No auth environment variables
- **Action**: Push logs
- **Expected**: No auth headers (anonymous)

---

## 5.6 HTTP Client

### TC-5.6.1: Request Timeout

- **Setup**: Mock Loki delays 15 seconds
- **Action**: `client.Push(ctx, request)`
- **Expected**: Times out after 10 seconds

### TC-5.6.2: Content-Type Header

- **Action**: Push logs
- **Expected**: `Content-Type: application/json` header

### TC-5.6.3: Network Error

- **Setup**: Invalid endpoint URL
- **Action**: `client.Push(ctx, request)`
- **Expected**: Retryable error returned

---

## 5.7 Request Body

### TC-5.7.1: Valid JSON Body

- **Action**: Push logs
- **Expected**: Body is valid JSON matching Loki push API format

### TC-5.7.2: Body Preserved Across Retries

- **Setup**: Mock returns 500, then 200
- **Action**: Push logs
- **Expected**: Same body sent on retry (not consumed)
