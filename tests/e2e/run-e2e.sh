#!/bin/bash
# LambdaWatch E2E Test Runner
# TEST ONLY - verifies logs arrive in Grafana Loki correctly
set -uo pipefail

FUNCTION_NAME="lambda-watch-test"
REGION="ap-south-1"
LOKI_URL="https://logs-prod-033.grafana.net"
LOKI_USER=""
LOKI_PASS=""
SERVICE_NAME="lambda-watch-test"
WAIT_SECONDS=15

PASS=0
FAIL=0

now_ns() { python3 -c "import time; print(int(time.time()*1e9))"; }
ago_ns() { python3 -c "import time; print(int((time.time()-$1)*1e9))"; }

invoke_fn() {
  local payload="$1"
  local log_result
  log_result=$(aws lambda invoke --function-name "$FUNCTION_NAME" --region "$REGION" \
    --payload "$payload" --cli-binary-format raw-in-base64-out \
    --log-type Tail --query 'LogResult' --output text /tmp/e2e-out.json 2>/dev/null)
  echo "$log_result" | base64 -d 2>/dev/null | grep -oE 'RequestId: [a-f0-9-]+' | head -1 | awk '{print $2}'
}

query_loki() {
  local query="$1" start="$2" end="$3"
  curl -s -u "${LOKI_USER}:${LOKI_PASS}" \
    -G "${LOKI_URL}/loki/api/v1/query_range" \
    --data-urlencode "query=${query}" \
    --data-urlencode "start=${start}" \
    --data-urlencode "end=${end}" \
    --data-urlencode 'limit=500'
}

count_logs() {
  echo "$1" | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(sum(len(s.get('values',[])) for s in d.get('data',{}).get('result',[])))
" 2>/dev/null || echo "0"
}

has_text() {
  echo "$1" | python3 -c "
import json,sys
pat=sys.argv[1]
d=json.load(sys.stdin)
for s in d.get('data',{}).get('result',[]):
    for _,m in s.get('values',[]):
        if pat in m:
            print('yes'); sys.exit(0)
print('no')
" "$2" 2>/dev/null
}

get_labels() {
  echo "$1" | python3 -c "
import json,sys
d=json.load(sys.stdin)
r=d.get('data',{}).get('result',[])
if r:
    for k,v in sorted(r[0].get('stream',{}).items()):
        if k!='detected_level': print(f'    {k}={v}')
" 2>/dev/null
}

ok()   { echo "  ✅ $1"; PASS=$((PASS+1)); }
fail() { echo "  ❌ $1"; FAIL=$((FAIL+1)); }
assert_gte() { [ "$3" -ge "$2" ] 2>/dev/null && ok "$1 (got=$3)" || fail "$1 (expected>=$2, got=$3)"; }
assert_has() { [ "$(has_text "$2" "$3")" = "yes" ] && ok "$1" || fail "$1 (missing '$3')"; }

# Query helper: invoke, wait, query with wide window
run_test() {
  local payload="$1"
  local start_ns
  start_ns=$(ago_ns 60)
  RID=$(invoke_fn "$payload")
  echo "  request_id=$RID — waiting ${WAIT_SECONDS}s..."
  sleep $WAIT_SECONDS
  local end_ns
  end_ns=$(python3 -c "import time; print(int((time.time()+60)*1e9))")
  # Return results via global vars
  TEST_RID="$RID"
  TEST_START="$start_ns"
  TEST_END="$end_ns"
  TEST_RESULT=$(query_loki "{service_name=\"${SERVICE_NAME}\",request_id=\"${RID}\"}" "$start_ns" "$end_ns")
  TEST_ALL=$(query_loki "{service_name=\"${SERVICE_NAME}\"}" "$start_ns" "$end_ns")
  TEST_COUNT=$(count_logs "$TEST_RESULT")
}

echo "============================================"
echo "  LambdaWatch E2E Test Suite"
echo "============================================"
echo ""

# ==========================================
echo "📋 Test 1: Basic Logs (5 structured logs)"
echo "-------------------------------------------"
run_test '{"testType":"basic"}'
assert_gte "Function logs received (>=5)" 5 "$TEST_COUNT"
assert_has "Has 'Starting processing'" "$TEST_RESULT" "Starting processing"
assert_has "Has 'Processing complete'" "$TEST_RESULT" "Processing complete"
assert_has "Has 'item-001'" "$TEST_RESULT" "item-001"
assert_has "Has 'item-002'" "$TEST_RESULT" "item-002"

LABELS=$(get_labels "$TEST_ALL")
echo "  Labels:"
echo "$LABELS"
echo "$LABELS" | grep -q "function_name=lambda-watch-test" && ok "function_name label" || fail "function_name label"
echo "$LABELS" | grep -q "region=ap-south-1" && ok "region label" || fail "region label"
echo "$LABELS" | grep -q "source=lambda" && ok "source label" || fail "source label"
echo ""

# ==========================================
echo "📋 Test 2: Volume Logs (50 logs)"
echo "-------------------------------------------"
run_test '{"testType":"volume","count":50}'
assert_gte "50 logs received" 50 "$TEST_COUNT"
assert_has "Has first log (1/50)" "$TEST_RESULT" "Log entry 1/50"
assert_has "Has last log (50/50)" "$TEST_RESULT" "Log entry 50/50"
echo ""

# ==========================================
echo "📋 Test 3: Large Message (~5KB)"
echo "-------------------------------------------"
run_test '{"testType":"large","sizeKB":5}'
assert_gte "Large log received" 1 "$TEST_COUNT"
assert_has "Has 'Large log entry'" "$TEST_RESULT" "Large log entry"
echo ""

# ==========================================
echo "📋 Test 4: Structured JSON Logs"
echo "-------------------------------------------"
run_test '{"testType":"json"}'
assert_gte "JSON logs received (>=3)" 3 "$TEST_COUNT"
assert_has "Has 'User action'" "$TEST_RESULT" "User action"
assert_has "Has 'Rate limit'" "$TEST_RESULT" "Rate limit"
assert_has "Has 'Validation failed'" "$TEST_RESULT" "Validation failed"
echo ""

# ==========================================
echo "📋 Test 5: Error Logs"
echo "-------------------------------------------"
run_test '{"testType":"error"}'
assert_gte "Error logs received (>=3)" 3 "$TEST_COUNT"
assert_has "Has 'Simulated error'" "$TEST_RESULT" "Simulated error"
assert_has "Has stack trace" "$TEST_RESULT" "stack"
echo ""

# ==========================================
echo "📋 Test 6: Slow Logs (spread over 3s)"
echo "-------------------------------------------"
run_test '{"testType":"slow","durationMs":3000}'
assert_gte "Slow logs received (>=5)" 5 "$TEST_COUNT"
assert_has "Has 'Slow log 1/5'" "$TEST_RESULT" "Slow log 1/5"
assert_has "Has 'Slow log 5/5'" "$TEST_RESULT" "Slow log 5/5"
echo ""

# ==========================================
echo "📋 Test 7: Request ID Isolation"
echo "-------------------------------------------"
T7_START=$(ago_ns 60)
RID_A=$(invoke_fn '{"testType":"basic"}')
sleep 2
RID_B=$(invoke_fn '{"testType":"basic"}')
echo "  A=$RID_A  B=$RID_B — waiting ${WAIT_SECONDS}s..."
sleep $WAIT_SECONDS
T7_END=$(python3 -c "import time; print(int((time.time()+60)*1e9))")

RA=$(query_loki "{service_name=\"${SERVICE_NAME}\",request_id=\"${RID_A}\"}" "$T7_START" "$T7_END")
RB=$(query_loki "{service_name=\"${SERVICE_NAME}\",request_id=\"${RID_B}\"}" "$T7_START" "$T7_END")
CA=$(count_logs "$RA")
CB=$(count_logs "$RB")
assert_gte "Invocation A has logs" 5 "$CA"
assert_gte "Invocation B has logs" 5 "$CB"
[ "$RID_A" != "$RID_B" ] && ok "Different request IDs" || fail "Same request ID for both"
echo ""

# ==========================================
echo "📋 Test 8: Extension Logs (LambdaWatch)"
echo "-------------------------------------------"
EXT=$(query_loki "{service_name=\"${SERVICE_NAME}\"}" "$T7_START" "$T7_END")
assert_has "Has state transition" "$EXT" "State transition"
assert_has "Has critical flush" "$EXT" "Critical flush"
assert_has "Has LambdaWatch context" "$EXT" "LambdaWatch"
echo ""

# ==========================================
echo "📋 Test 9: Platform Events"
echo "-------------------------------------------"
assert_has "Has START message" "$EXT" "START RequestId"
assert_has "Has REPORT message" "$EXT" "REPORT RequestId"
echo ""

# ==========================================
echo "============================================"
echo "  RESULTS: $PASS passed, $FAIL failed"
echo "============================================"
[ "$FAIL" -gt 0 ] && exit 1 || exit 0
