# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LambdaWatch is an AWS Lambda Extension written in Go 1.21+ that captures Lambda function logs and ships them to Grafana Loki in real-time. It runs as an external extension (Lambda Layer) requiring zero code changes to the monitored function. The project has **no external dependencies** — pure Go standard library.

## Build & Development Commands

```bash
make build              # Build for current platform (dev)
make build-arm64        # Build for ARM64 (Graviton) — recommended for Lambda
make build-amd64        # Build for x86_64
make test               # Run all unit tests (go test -v ./...)
make test-coverage      # Tests with HTML coverage report (coverage.html)
make fmt                # Format code (go fmt ./...)
make lint               # Run golangci-lint
make tidy               # go mod tidy
make package            # Build + package as ARM64 Lambda Layer (.zip)
make deploy             # Publish ARM64 layer to AWS
make clean              # Remove build artifacts
```

Run a single package's tests:
```bash
go test -v ./internal/buffer/
go test -v -run TestSpecificName ./internal/config/
```

## Architecture

### Data Flow

```
Lambda Function → Telemetry API (POST to :8080) → Server.handleTelemetry()
  → Parse events, extract request IDs, format messages → Buffer
  → [Periodic flush loop OR runtimeDone trigger] → Loki Client.Push()
  → Serialize JSON, optional gzip, POST with retries → Grafana Loki
```

### State Machine (lifecycle.go)

```
INVOKE event → ACTIVE (1s flush) → platform.runtimeDone → FLUSHING (critical flush)
  → IDLE (3x longer flush intervals for cost optimization)
```

### Key Packages

- **`cmd/extension/main.go`** — Entry point. Loads config, sets up signal handling, creates Manager, runs lifecycle.
- **`internal/extension/lifecycle.go`** — Core orchestrator. State machine managing registration, event loop, flush loop, and shutdown. This is the central coordination point.
- **`internal/extension/client.go`** — Lambda Extensions API HTTP client (register, next event).
- **`internal/buffer/buffer.go`** — Thread-safe circular buffer (mutex-protected). Tracks entry count + byte size. Channel-based ready signaling. Drops oldest on overflow.
- **`internal/telemetryapi/server.go`** — HTTP server receiving Lambda telemetry events. Handles platform.start, platform.runtimeDone, platform.report, and function logs. Deduplicates extension logs. Auto-splits long messages.
- **`internal/telemetryapi/client.go`** — Subscribes to Lambda Telemetry API.
- **`internal/loki/client.go`** — Loki HTTP client with two-tier retry system: regular flush (3 retries) vs critical flush (5 retries). Exponential backoff (100ms × 2^attempt). Supports bearer token, basic auth, and multi-tenant org ID.
- **`internal/loki/batch.go`** — Converts buffer entries to Loki PushRequest. Can group streams by request ID.
- **`internal/config/config.go`** — Loads all `LOKI_*` environment variables with defaults. Invalid values silently fall back to defaults.
- **`internal/logger/logger.go`** — Structured JSON logger. Outputs to stdout AND directly to the buffer.

### Concurrency Model

- **Main goroutine:** Extensions API event loop (waiting for INVOKE/SHUTDOWN)
- **Flush goroutine:** Background timer-based periodic flushing with adaptive intervals
- **Telemetry server:** Go net/http handler goroutine
- **Shutdown:** Signal handling (SIGTERM/SIGINT) with context cancellation, buffer drain

### Configuration

Only required env var: `LOKI_URL`. Auth via `LOKI_USERNAME`/`LOKI_PASSWORD` (basic auth) or `LOKI_API_KEY` (bearer token). See `internal/config/config.go` for all variables and defaults. Custom labels via `LOKI_LABELS` as JSON string.

## Test Plans

Detailed test scenarios live in `tests/*.md` covering config, buffer, flush behavior, state machine, Loki client, telemetry, lifecycle, and e2e deployment.
