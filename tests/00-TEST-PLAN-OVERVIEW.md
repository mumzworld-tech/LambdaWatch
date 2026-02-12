# LambdaWatch Test Plan Overview

## Test Categories

| # | Category | File | Description |
|---|----------|------|-------------|
| 1 | Configuration | `01-config-tests.md` | Environment variables, defaults, validation |
| 2 | Buffer | `02-buffer-tests.md` | Bounded buffer operations, overflow, thread safety |
| 3 | Flush Behavior | `03-flush-tests.md` | Intervals, batch sizes, critical flush |
| 4 | State Machine | `04-state-tests.md` | State transitions, interval adjustments |
| 5 | Loki Client | `05-loki-client-tests.md` | Push, retries, compression, auth |
| 6 | Telemetry Server | `06-telemetry-tests.md` | Event parsing, formatting, filtering |
| 7 | Lifecycle | `07-lifecycle-tests.md` | Registration, event loop, shutdown |
| 8 | End-to-End | `08-e2e-tests.md` | Full Lambda deployment tests |

## Test Types

### Unit Tests (Go)
- Test individual functions/methods in isolation
- Mock dependencies
- Run with `go test ./...`

### Integration Tests (Go)
- Test component interactions
- Use test servers (mock Loki, mock Lambda APIs)
- Run with `go test -tags=integration ./...`

### E2E Tests (Manual + Scripts)
- Deploy actual Lambda with extension
- Configure different settings
- Verify logs in Grafana Loki
- Documented as test scenarios with expected results

## Test Environment

### Local Testing
- Go test framework
- Mock HTTP servers for Loki/Lambda APIs

### AWS Testing
- Test Lambda functions: `<your-function-1>`, `<your-function-2>`
- Grafana Cloud Loki for log verification
- Different layer versions for A/B testing

## Success Criteria

1. All unit tests pass
2. All integration tests pass
3. E2E scenarios produce expected logs in Loki
4. No memory leaks or goroutine leaks
5. Extension doesn't impact Lambda function performance significantly
