# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LambdaWatch is an AWS Lambda Extension written in Go 1.21+ that captures Lambda function logs and ships them to Grafana Loki in real-time. It runs as an external extension (Lambda Layer) requiring zero code changes to the monitored function. The Go extension has **no external dependencies** — pure Go standard library.

The repository also contains a **Next.js 16 marketing website** (`website/`) — a static site showcasing features, architecture, and performance comparisons. Built with React 19, Tailwind CSS 4, motion/react, and shadcn/ui.

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
make package-amd64      # Build + package as AMD64 Lambda Layer (.zip)
make deploy             # Publish ARM64 layer to AWS
make clean              # Remove build artifacts
```

Run a single package's tests:

```bash
go test -v ./internal/buffer/
go test -v -run TestSpecificName ./internal/config/
```

### Git Hooks

Lefthook runs `go fmt` and `golangci-lint` on pre-commit (configured in `lefthook.yml`). Install once after cloning: `lefthook install`.

### CI/CD

Push to `main` triggers `.github/workflows/release.yml` which runs tests, lints, builds both architectures, and creates a GitHub release with layer zip artifacts.

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
INVOKE → ACTIVE (1s flush) → platform.runtimeDone → FLUSHING (critical flush) → IDLE (3s flush)
```

### Key Packages

| Package | File(s) | Purpose |
|---------|---------|---------|
| `cmd/extension` | `main.go` | Entry point: config, signals, Manager.Run() |
| `internal/extension` | `lifecycle.go` | Core orchestrator: state machine, flush loop, event loop |
| `internal/extension` | `client.go`, `events.go` | Lambda Extensions API client + event types |
| `internal/buffer` | `buffer.go` | Thread-safe circular buffer with byte tracking |
| `internal/telemetryapi` | `server.go`, `client.go`, `types.go` | Telemetry API receiver + subscriber |
| `internal/loki` | `client.go`, `batch.go`, `types.go` | Loki HTTP client (retry/gzip) + batch converter |
| `internal/config` | `config.go` | Env var loading with defaults |
| `internal/logger` | `logger.go` | JSON logger → stdout + buffer |
| `internal/logsapi` | `*.go` | Legacy Logs API (unused, kept for reference) |

### Key Design Decisions

- **Request ID as content, not label**: Injected into message content to avoid high-cardinality Loki labels. Query: `{function_name="x"} | json | request_id="abc"`.
- **Non-blocking telemetry response**: HTTP 200 sent *before* critical flush to prevent Telemetry API delivery stalls.
- **Deadline-bounded critical flush**: Uses Lambda's `DeadlineMs - 500ms`, not arbitrary timeout.
- **Invocation synchronization**: Event loop blocks on `invocationDone` channel until critical flush completes.

### Concurrency Model

- **Main goroutine:** Extensions API event loop (blocking NextEvent)
- **Flush goroutine:** Timer-based periodic flush, yields during FLUSHING state
- **Telemetry server:** HTTP on :8080, triggers `onRuntimeDone` synchronously after responding
- **Shutdown:** SIGTERM/SIGINT → context cancel → server shutdown → buffer drain

### Configuration

Only required env var: `LOKI_URL`. Auth via `LOKI_USERNAME`/`LOKI_PASSWORD` (basic) or `LOKI_API_KEY` (bearer). Custom labels via `LOKI_LABELS` JSON string. See exports-reference.md for all fields and defaults.

## Website Commands

```bash
cd website
pnpm install        # Install dependencies
pnpm dev            # Dev server (localhost:3000)
pnpm build          # Static export build
pnpm lint           # ESLint
```

### Website Key Structure

| Directory | Purpose |
|-----------|---------|
| `website/app/` | Next.js App Router (single page: `/`) |
| `website/components/sections/` | 8 page sections: Navbar (floating glass pill), Hero, Features, Architecture, Performance, Comparison, FAQ, Footer |
| `website/components/common/` | 12 shared components: SectionWrapper, SectionHeading, GlassmorphicCard, etc. |
| `website/components/ui/` | 24 shadcn + animation components |
| `website/lib/constants.ts` | All static content (features, FAQ, metrics, comparisons) |
| `website/hooks/` | useMousePosition, useScrollProgress |

## Detailed Documentation

@.claude/claude-md-refs/architecture.md
@.claude/claude-md-refs/development-guide.md
@.claude/claude-md-refs/exports-reference.md

| Need Help With | See File |
|----------------|----------|
| Adding features, config, event handlers, sections | development-guide.md |
| Data flow, state machine, component tree, rendering | architecture.md |
| Finding types, functions, constants, components | exports-reference.md |

## Test Plans

Detailed test scenarios live in `tests/*.md` covering config, buffer, flush behavior, state machine, Loki client, telemetry, lifecycle, and e2e deployment.
