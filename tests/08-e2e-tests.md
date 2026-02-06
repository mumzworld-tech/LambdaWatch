# End-to-End Tests

## Test Environment

### Lambda Functions
- `<your-function-1>`
- `<your-function-2>`

### Grafana Loki
- Endpoint: `https://<your-loki-endpoint>`
- Query UI: Grafana Cloud dashboard

### Test Commands
```bash
# Invoke function
aws lambda invoke --function-name <name> --payload '{}' /dev/stdout

# Query logs
curl -u "${USER}:${PASS}" "${LOKI}/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="<your-service-name>"}'

# Delete logs (for clean slate)
curl -X POST -u "${USER}:${PASS}" "${LOKI}/loki/api/v1/delete" \
  --data-urlencode 'query={...}' --data-urlencode 'start=...' --data-urlencode 'end=...'
```

---

## E2E-1: Basic Log Shipping

### TC-E2E-1.1: Cold Start Logs
- **Action**: Invoke function (cold start)
- **Verify in Loki**:
  - [ ] Extension startup logs (LambdaWatch context)
  - [ ] `START RequestId: xxx Version: $LATEST`
  - [ ] Function logs
  - [ ] `REPORT RequestId: xxx Duration: xxx ms...`

### TC-E2E-1.2: Warm Start Logs
- **Action**: Invoke function again immediately
- **Verify in Loki**:
  - [ ] No extension startup logs
  - [ ] START, function logs, REPORT present
  - [ ] Different RequestId from cold start

### TC-E2E-1.3: All Logs Present
- **Action**: Function that generates 10 log messages
- **Verify in Loki**:
  - [ ] All 10 function logs present
  - [ ] Correct order preserved
  - [ ] Timestamps accurate

---

## E2E-2: Request ID Grouping

### TC-E2E-2.1: Logs Grouped by Request ID
- **Action**: Invoke function
- **Verify in Loki**:
  - [ ] Query `{request_id="xxx"}` returns all logs for that invocation
  - [ ] Function logs have request_id label
  - [ ] Platform events have request_id label

### TC-E2E-2.2: Extension Logs No Request ID
- **Action**: Check LambdaWatch context logs
- **Verify in Loki**:
  - [ ] Extension logs do NOT have request_id label
  - [ ] They're in a separate stream

### TC-E2E-2.3: Multiple Invocations Separated
- **Action**: Invoke function 3 times
- **Verify in Loki**:
  - [ ] 3 different request_id values
  - [ ] Each invocation's logs correctly grouped

---

## E2E-3: Labels

### TC-E2E-3.1: Auto Labels Present
- **Action**: Check any log entry
- **Verify labels**:
  - [ ] `function_name` = function name
  - [ ] `function_version` = $LATEST or version
  - [ ] `region` = ap-south-1
  - [ ] `source` = lambda
  - [ ] `service_name` = configured value

### TC-E2E-3.2: Custom Labels
- **Setup**: Set `LOKI_LABELS={"team":"platform","env":"preprod"}`
- **Action**: Invoke function
- **Verify labels**:
  - [ ] `team` = platform
  - [ ] `env` = preprod

---

## E2E-4: Flush Intervals

### TC-E2E-4.1: Active State Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=2000` (2 seconds)
- **Action**: Invoke function that runs for 5 seconds
- **Verify in Loki**:
  - [ ] Multiple "Pushing X log entries to Loki" messages
  - [ ] Approximately 2-3 flushes during execution

### TC-E2E-4.2: Idle State Interval
- **Setup**: `LOKI_FLUSH_INTERVAL_MS=1000`, `LOKI_IDLE_FLUSH_MULTIPLIER=5`
- **Action**: Check logs after invocation completes
- **Verify**:
  - [ ] "Flush interval adjusted to: 5s (state: IDLE)"
  - [ ] Flushes every 5 seconds in idle

### TC-E2E-4.3: Interval Change Logged
- **Action**: Invoke function
- **Verify logs**:
  - [ ] "State transition: IDLE -> ACTIVE"
  - [ ] "Flush interval adjusted to: 1s (state: ACTIVE)"
  - [ ] "State transition: ACTIVE -> FLUSHING"
  - [ ] "State transition: FLUSHING -> IDLE"
  - [ ] "Flush interval adjusted to: 3s (state: IDLE)"

---

## E2E-5: Critical Flush

### TC-E2E-5.1: All Logs Flushed at Invocation End
- **Action**: Function generates logs at the very end before returning
- **Verify in Loki**:
  - [ ] All logs present including final ones
  - [ ] "Critical flush: X entries" message
  - [ ] "Invocation complete, ready for next event"

### TC-E2E-5.2: No Log Loss Between Invocations
- **Action**: Rapid successive invocations
- **Verify in Loki**:
  - [ ] No logs missing
  - [ ] Each invocation complete

---

## E2E-6: Batch Sizes

### TC-E2E-6.1: Large Batch
- **Setup**: `LOKI_BATCH_SIZE=1000`
- **Action**: Function generates 50 logs
- **Verify**:
  - [ ] Single "Pushing 50 log entries" (all in one batch)

### TC-E2E-6.2: Small Batch
- **Setup**: `LOKI_BATCH_SIZE=10`
- **Action**: Function generates 50 logs
- **Verify**:
  - [ ] Multiple push messages (10 entries each)

### TC-E2E-6.3: Byte Size Limit
- **Setup**: `LOKI_MAX_BATCH_SIZE_BYTES=1024`
- **Action**: Function generates large log messages
- **Verify**:
  - [ ] Batches split by size, not count

---

## E2E-7: Compression

### TC-E2E-7.1: Gzip Enabled (Default)
- **Action**: Generate logs > 1KB
- **Verify** (via network inspection or Loki metrics):
  - [ ] Compressed payload sent

### TC-E2E-7.2: Gzip Disabled
- **Setup**: `LOKI_ENABLE_GZIP=false`
- **Action**: Generate logs
- **Verify**:
  - [ ] Uncompressed payload sent

---

## E2E-8: Large Messages

### TC-E2E-8.1: Message Splitting
- **Setup**: `LOKI_MAX_LINE_SIZE=1024`
- **Action**: Log a 5KB message
- **Verify in Loki**:
  - [ ] Message split into multiple chunks
  - [ ] `[chunk 1/5]`, `[chunk 2/5]`, etc. prefixes
  - [ ] All chunks present and in order

### TC-E2E-8.2: No Splitting Under Limit
- **Setup**: `LOKI_MAX_LINE_SIZE=10240` (10KB)
- **Action**: Log a 5KB message
- **Verify in Loki**:
  - [ ] Single log entry, no chunking

---

## E2E-9: Error Handling

### TC-E2E-9.1: Loki Temporarily Unavailable
- **Setup**: Invalid Loki endpoint temporarily
- **Action**: Invoke function
- **Verify**:
  - [ ] Retries logged
  - [ ] Function completes (extension doesn't block)
  - [ ] Logs may be lost but function works

### TC-E2E-9.2: Invalid Credentials
- **Setup**: Wrong password
- **Action**: Invoke function
- **Verify**:
  - [ ] Auth error logged
  - [ ] Function completes normally

---

## E2E-10: Shutdown

### TC-E2E-10.1: Graceful Shutdown
- **Action**: Let Lambda freeze/terminate instance
- **Verify in Loki**:
  - [ ] "Received SHUTDOWN event, reason: xxx"
  - [ ] "Draining buffer..."
  - [ ] Final logs present

### TC-E2E-10.2: Shutdown Reasons
- **Action**: Various shutdown triggers
- **Verify**:
  - [ ] `spindown` - idle timeout
  - [ ] `timeout` - function timeout
  - [ ] `failure` - extension crash

---

## E2E-11: Performance

### TC-E2E-11.1: Extension Overhead
- **Action**: Compare function duration with/without extension
- **Verify**:
  - [ ] Init Duration increase < 100ms
  - [ ] Billed Duration increase minimal

### TC-E2E-11.2: Memory Usage
- **Action**: Check Max Memory Used
- **Verify**:
  - [ ] Memory increase reasonable (< 50MB typically)

### TC-E2E-11.3: High Volume Logs
- **Action**: Function generates 1000 logs rapidly
- **Verify**:
  - [ ] All logs shipped
  - [ ] No significant delay
  - [ ] No OOM

---

## E2E-12: No Duplicate Logs

### TC-E2E-12.1: Extension Logs Not Duplicated
- **Action**: Invoke function
- **Verify in Loki**:
  - [ ] Each LambdaWatch log appears exactly once
  - [ ] Not duplicated via Telemetry API

### TC-E2E-12.2: Function Logs Not Duplicated
- **Action**: Invoke function
- **Verify in Loki**:
  - [ ] Each function log appears exactly once

---

## Test Execution Checklist

### Pre-Test
- [ ] Clean Loki logs (delete old test data)
- [ ] Note current layer version
- [ ] Verify Lambda function configuration

### During Test
- [ ] Record layer version used
- [ ] Record invocation timestamps
- [ ] Save CloudWatch logs for comparison

### Post-Test
- [ ] Compare Loki logs with CloudWatch
- [ ] Document any discrepancies
- [ ] Update test results
