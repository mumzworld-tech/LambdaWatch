# Architecture

## Project Structure

```
LambdaWatch/
├── cmd/extension/main.go          # Go entry point
├── internal/
│   ├── buffer/buffer.go           # Thread-safe circular buffer
│   ├── config/config.go           # Env var config loader
│   ├── extension/
│   │   ├── lifecycle.go           # Core orchestrator + state machine
│   │   ├── client.go              # Lambda Extensions API client
│   │   └── events.go              # Event types
│   ├── loki/
│   │   ├── client.go              # Loki HTTP client (retry/gzip)
│   │   ├── batch.go               # Batch converter (request ID injection)
│   │   └── types.go               # PushRequest, Stream types
│   ├── telemetryapi/
│   │   ├── server.go              # HTTP :8080 receiver + event parsing
│   │   ├── client.go              # Telemetry API subscriber
│   │   └── types.go               # Event types + records
│   ├── logger/logger.go           # JSON logger → stdout + buffer
│   └── logsapi/                   # Legacy (unused)
├── website/                        # Next.js marketing site
│   ├── app/
│   │   ├── layout.tsx             # Root layout (fonts, metadata)
│   │   ├── page.tsx               # Home page (server fetch → GitHubDataWrapper)
│   │   └── globals.css            # Theme + animations
│   ├── components/
│   │   ├── github-data-wrapper.tsx # Client wrapper: sections + client-side data refresh
│   │   ├── sections/              # Page sections (8)
│   │   ├── common/                # Shared components (12)
│   │   └── ui/                    # shadcn + animation components (24)
│   ├── hooks/                     # Custom React hooks (2)
│   └── lib/                       # Utils, constants, fonts, github API
├── Makefile                        # Build, test, package, deploy
└── lefthook.yml                    # Pre-commit hooks (fmt + lint)
```

## Go Extension — Data Flow

```
Lambda Function
    │
    ▼
Telemetry API (POST to :8080)
    │
    ▼
Server.handleTelemetry()
    ├── Parse TelemetryEvent[] JSON
    ├── Switch on event type:
    │   ├── platform.start     → Extract requestID, format "START RequestId: ..."
    │   ├── platform.runtimeDone → Extract requestID, trigger onRuntimeDone callback
    │   ├── platform.report    → Format "REPORT RequestId: ... Duration: ..."
    │   ├── function           → Extract timestamp from Lambda prefix, split long messages
    │   └── extension          → Same as function (skip own LambdaWatch logs)
    ├── Respond HTTP 200 immediately (non-blocking)
    └── If runtimeDone: trigger critical flush AFTER responding
    │
    ▼
Buffer (thread-safe circular, 10k entries default)
    │
    ├── Periodic flush loop (timer-based)
    │   └── flushBatch() → Batch.ToPushRequest() → Client.Push()
    │
    └── Critical flush (runtimeDone / shutdown)
        └── criticalFlush() → drain all → Client.PushCritical()
    │
    ▼
Loki Client
    ├── Serialize JSON
    ├── Optional gzip (if > 1KB threshold)
    ├── Set auth headers (Basic or Bearer)
    ├── POST with exponential backoff retries
    │   ├── Regular: 3 retries
    │   └── Critical: 5 retries
    └── Retry on 429 + 5xx only
    │
    ▼
Grafana Loki
```

## State Machine

```
         INVOKE event
              │
              ▼
┌─────────────────────────┐
│     ACTIVE              │
│  flush interval: 1x     │
│  (default 1000ms)       │
└────────────┬────────────┘
             │ platform.runtimeDone
             ▼
┌─────────────────────────┐
│     FLUSHING            │
│  critical flush active  │
│  periodic flush yields  │
│  deadline-bounded       │
└────────────┬────────────┘
             │ flush complete
             ▼
┌─────────────────────────┐
│     IDLE                │
│  flush interval: 3x     │
│  (default 3000ms)       │
└────────────┬────────────┘
             │ next INVOKE
             └──────► back to ACTIVE
```

| Current | Trigger | New State | Action |
|---------|---------|-----------|--------|
| IDLE | INVOKE event | ACTIVE | Store deadline, create invocationDone channel |
| ACTIVE | platform.runtimeDone | FLUSHING | Critical flush with deadline context |
| FLUSHING | Flush complete | IDLE | Signal invocationDone, ready for next event |
| Any | SHUTDOWN event | - | Stop flush loop → shutdown server → drain buffer → final push |

## Concurrency Model

```
┌──────────────────────────────────────────────────────┐
│ Main Goroutine                                       │
│  eventLoop(): NextEvent() blocks → handle INVOKE     │
│  waits on invocationDone channel between invocations │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│ Flush Goroutine                                      │
│  flushLoop(): timer tick → flush() or shouldFlush()  │
│  yields when state == StateFlushing                  │
│  adjusts interval on intervalChange signal           │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│ Telemetry Server Goroutine                           │
│  HTTP :8080 → handleTelemetry()                      │
│  Responds 200 BEFORE triggering onRuntimeDone        │
│  onRuntimeDone runs synchronously after response     │
└──────────────────────────────────────────────────────┘

Synchronization:
  - buffer.mu (Mutex)          → protects buffer entries
  - criticalFlushMu (Mutex)    → prevents concurrent critical flushes
  - invocationMu (Mutex)       → protects invocationDone channel
  - state (atomic.Int32)       → lock-free state reads
  - invocationDeadline (atomic) → lock-free deadline storage
  - intervalChange (chan)       → signal flush loop to adjust
  - invocationDone (chan)       → block event loop during flush
  - stopFlush (chan)            → shutdown signal to flush loop
```

## Package Dependency Graph

```
cmd/extension/main.go
├── internal/config       ← Load()
├── internal/extension    ← NewManager(), Manager.Run()
└── internal/logger       ← Init(), Fatalf()

internal/extension/lifecycle.go
├── internal/buffer       ← Buffer operations
├── internal/config       ← Config struct
├── internal/logger       ← Info/Debug/Error logging
├── internal/loki         ← Client, Batch, PushRequest
└── internal/telemetryapi ← Server, Client

internal/telemetryapi/server.go
├── internal/buffer       ← LogEntry, AddBatch()
└── internal/logger       ← Debug logging

internal/loki/client.go
└── internal/config       ← Config struct

internal/loki/batch.go
└── internal/buffer       ← LogEntry

internal/logger/logger.go
└── internal/buffer       ← Add(), SignalReady()
```

No external dependencies — pure Go standard library throughout.

## Key Design Decisions

1. **Request ID as content, not label**: Injected into message body (`{"request_id":"..."}` for JSON, `[request_id=...] ` for text) to avoid high-cardinality Loki labels. Query: `{function_name="x"} | json | request_id="abc"`.

2. **Non-blocking telemetry response**: HTTP 200 sent and flushed *before* critical flush starts. Prevents Telemetry API delivery stalls that would cause log loss.

3. **Deadline-bounded critical flush**: Uses Lambda's `DeadlineMs - 500ms` safety margin, not arbitrary timeouts. Ensures flush completes before Lambda kills the process.

4. **Invocation synchronization**: Event loop blocks on `invocationDone` until critical flush completes. Prevents calling NextEvent (which signals readiness) while still flushing.

5. **Circular buffer with byte tracking**: Drops oldest entries when full (never blocks producers). Tracks byte size for batch size limits.

6. **Dual-path logging**: Extension's own logs go to both stdout (CloudWatch) and buffer (Loki) directly, since Telemetry API doesn't capture logs from the same extension process.

---

## Website — Architecture

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | Next.js 16 (App Router, static export) |
| UI | React 19, Tailwind CSS 4, shadcn/ui (new-york) |
| Animation | motion/react 12 (scroll-triggered, spring, gesture) |
| Icons | lucide-react |
| Charts | recharts |
| Variants | class-variance-authority (CVA) |
| Package Manager | pnpm |

### Component Architecture

```
Server Components (SSR/Build Time)
├── layout.tsx              → Font loading, metadata, favicon, HTML structure
└── page.tsx                → Data fetching (getGitHubStars + getLatestRelease)
                               └── passes initialStars/initialRelease to GitHubDataWrapper

Client Components ("use client")
├── github-data-wrapper.tsx → Receives server data as props, refreshes via client-side fetch
│   └── Composes all sections: Navbar, Hero, Features, ..., Footer
├── sections/               → Interactive page sections (animations, scroll, mouse tracking)
├── common/                 → Reusable building blocks (wrappers, cards, badges)
└── ui/                     → Primitive components (buttons, accordion, effects)
```

**Design pattern**: Page is a Server Component that fetches initial GitHub data and passes it to `GitHubDataWrapper` (Client Component). The wrapper hydrates with server data, then refreshes via client-side GitHub API fetch on mount. All interactivity (animation, scroll tracking, mouse position) happens client-side.

### Rendering Strategy

- **Static Export**: `output: "export"` — all pages pre-rendered at build time
- **Dynamic basePath**: `NEXT_PUBLIC_BASE_PATH` env var (used for GitHub Pages deployment)
- **No SSR/ISR at runtime**: Pure static HTML + JS bundles
- **GitHub data (dual-fetch)**: Initial data fetched server-side at build time via `Promise.all`; client-side refresh on mount via `GitHubDataWrapper` useEffect
- **No API routes**: Static site, no server-side endpoints

### Style System

- **Dark-only theme**: oklch() color space in CSS custom properties
- **Glass morphism**: Backdrop blur + semi-transparent backgrounds (navbar, feature cards)
- **Gradient effects**: Warm spectrum (brand → #FF6B00 → #FF3D00) on hero text, brand gradients on borders
- **3D effects**: Mouse-tracking tilt on feature cards via motion/react spring values
- **Animation layers**: Background particles + grid pattern + glow effects
