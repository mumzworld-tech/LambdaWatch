<p align="center">
  <img src="https://raw.githubusercontent.com/Sami-AlEsh/lambdawatch/main/assets/logo.png" alt="LambdaWatch Logo" width="200"/>
</p>

<h1 align="center">LambdaWatch</h1>

<p align="center">
  <strong>High-performance AWS Lambda Extension for shipping logs to Grafana Loki</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
  <img src="https://img.shields.io/badge/AWS-Lambda-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white" alt="AWS Lambda"/>
  <img src="https://img.shields.io/badge/Grafana-Loki-F46800?style=for-the-badge&logo=grafana&logoColor=white" alt="Grafana Loki"/>
  <img src="https://img.shields.io/github/license/Sami-AlEsh/lambdawatch?style=for-the-badge" alt="License"/>
</p>

---

## What is LambdaWatch?

**LambdaWatch** is a lightweight, zero-dependency AWS Lambda Extension written in Go that automatically captures logs from your Lambda functions and ships them to [Grafana Loki](https://grafana.com/oss/loki/) in real-time.

No code changes required. Just add the layer and configure your Loki endpoint.

```
┌─────────────────────────────────────────────────────────────┐
│                     AWS Lambda                               │
│  ┌─────────────┐         ┌─────────────────────────────┐   │
│  │   Your      │         │      LambdaWatch            │   │
│  │  Function   │  logs   │  ┌─────┐ ┌─────┐ ┌──────┐  │   │
│  │             │ ──────► │  │Batch│→│Gzip │→│ Push │  │   │
│  │  console.log│         │  └─────┘ └─────┘ └──┬───┘  │   │
│  └─────────────┘         └─────────────────────┼──────┘   │
└────────────────────────────────────────────────┼──────────┘
                                                 │ HTTPS
                                                 ▼
                                        ┌──────────────┐
                                        │ Grafana Loki │
                                        └──────────────┘
```

## Features

### Core
- **Zero code changes** — Works as a Lambda Layer, no SDK required
- **Automatic batching** — Efficiently groups logs to minimize API calls
- **Gzip compression** — Reduces payload size by ~80%
- **Guaranteed delivery** — Critical flush on invocation end ensures no logs are lost

### Reliability
- **Two-tier retry system** — 5 retries for critical flushes, 3 for regular
- **Exponential backoff** — Intelligent retry delays on failures
- **Graceful shutdown** — Drains all logs before container termination
- **Bounded buffer** — Prevents memory overflow under high load

### Performance
- **Adaptive flush intervals** — 3x longer intervals when idle (cost optimization)
- **Byte-size limits** — Prevents oversized payloads to Loki
- **Compression threshold** — Skips gzip for small payloads (<1KB)
- **~6MB binary** — Minimal cold start impact

### Observability
- **Request ID extraction** — Automatic `request_id` label for request tracing
- **Auto-labeling** — Adds `function_name`, `function_version`, `region`
- **Custom labels** — Add your own labels via JSON config
- **Long message splitting** — Handles logs exceeding Loki's line limit

---

## Installation

### Option 1: Pre-built Layer (Recommended)

Download the latest release and publish as a Lambda Layer:

```bash
# Download the layer zip
curl -LO https://github.com/Sami-AlEsh/lambdawatch/releases/latest/download/lambdawatch-layer-arm64.zip

# Publish the layer
aws lambda publish-layer-version \
  --layer-name lambdawatch \
  --zip-file fileb://lambdawatch-layer-arm64.zip \
  --compatible-architectures arm64 \
  --compatible-runtimes provided.al2023 provided.al2
```

Then attach to your function:

```bash
aws lambda update-function-configuration \
  --function-name YOUR_FUNCTION \
  --layers arn:aws:lambda:REGION:ACCOUNT:layer:lambdawatch:VERSION
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/Sami-AlEsh/lambdawatch.git
cd lambdawatch

# Build for ARM64 (Graviton)
make build-arm64

# Or for x86_64
make build-amd64

# Package as Lambda Layer
make package
```

---

## Configuration

Configure via environment variables on your Lambda function:

### Required

| Variable | Description |
|----------|-------------|
| `LOKI_ENDPOINT` | Loki push URL (e.g., `https://loki.example.com/loki/api/v1/push`) |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `LOKI_USERNAME` | — | Basic auth username |
| `LOKI_PASSWORD` | — | Basic auth password |
| `LOKI_API_KEY` | — | Bearer token (alternative to basic auth) |
| `LOKI_TENANT_ID` | — | Multi-tenant org ID (`X-Scope-OrgID` header) |

### Batching & Performance

| Variable | Default | Description |
|----------|---------|-------------|
| `LOKI_BATCH_SIZE` | `100` | Max logs per batch |
| `LOKI_MAX_BATCH_SIZE_BYTES` | `5242880` | Max batch size (5MB) |
| `LOKI_FLUSH_INTERVAL_MS` | `1000` | Flush interval in ms |
| `LOKI_IDLE_FLUSH_MULTIPLIER` | `3` | Interval multiplier when idle |

### Reliability

| Variable | Default | Description |
|----------|---------|-------------|
| `LOKI_MAX_RETRIES` | `3` | Retry attempts for regular flushes |
| `LOKI_CRITICAL_FLUSH_RETRIES` | `5` | Retry attempts for critical flushes |
| `LOKI_ENABLE_GZIP` | `true` | Enable gzip compression |
| `LOKI_COMPRESSION_THRESHOLD` | `1024` | Compress only if > 1KB |

### Labels & Processing

| Variable | Default | Description |
|----------|---------|-------------|
| `LOKI_LABELS` | `{}` | Custom labels as JSON (e.g., `{"env":"prod"}`) |
| `LOKI_EXTRACT_REQUEST_ID` | `true` | Extract and add `request_id` label |
| `LOKI_MAX_LINE_SIZE` | `204800` | Max line size before splitting (200KB) |
| `BUFFER_SIZE` | `10000` | Max logs in memory buffer |

### Example Configuration

```bash
aws lambda update-function-configuration \
  --function-name my-function \
  --environment "Variables={
    LOKI_ENDPOINT=https://loki.example.com/loki/api/v1/push,
    LOKI_USERNAME=myuser,
    LOKI_PASSWORD=mypassword,
    LOKI_LABELS={\"env\":\"production\",\"team\":\"backend\"}
  }"
```

---

## Querying Logs in Grafana

### Automatic Labels

Every log entry includes these labels automatically:

| Label | Description | Source |
|-------|-------------|--------|
| `function_name` | Lambda function name | Extensions API |
| `function_version` | Function version ($LATEST, 1, 2, etc.) | Extensions API |
| `region` | AWS region (us-east-1, etc.) | AWS_REGION env |
| `request_id` | Invocation request ID | Extracted from logs (if enabled) |
| `source` | Always `lambda` | Hardcoded |
| `service_name` | Service identifier for grouping functions | SERVICE_NAME env (optional) |

### Example Queries

```logql
# All logs from a function
{function_name="my-function"}

# Filter by request ID (great for debugging specific invocations)
{function_name="my-function", request_id="abc-123-def-456"}

# Filter by region
{function_name="my-function", region="us-east-1"}

# All logs from a service (multiple functions)
{service_name="payment-service"}

# Search for errors
{function_name="my-function"} |= "ERROR"

# JSON parsing (if your logs are JSON)
{function_name="my-function"} | json | level="error"
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Lambda Execution Environment                      │
│                                                                      │
│  ┌──────────────┐         ┌──────────────────────────────────────┐  │
│  │   Lambda     │         │           LambdaWatch                 │  │
│  │   Function   │         │                                       │  │
│  └──────┬───────┘         │  ┌─────────┐    ┌──────────────────┐ │  │
│         │                 │  │Telemetry│    │    Lifecycle     │ │  │
│         │                 │  │ Server  │───►│    Manager       │ │  │
│         ▼                 │  │ :8080   │    │                  │ │  │
│  ┌──────────────┐         │  └────┬────┘    │ • State Machine  │ │  │
│  │   Lambda     │ ──────► │       │         │ • Adaptive Timer │ │  │
│  │ Telemetry API│  POST   │       ▼         └────────┬─────────┘ │  │
│  └──────────────┘         │  ┌─────────┐             │           │  │
│         │                 │  │ Buffer  │◄────────────┘           │  │
│         │ runtimeDone     │  │ • Byte  │  Critical Flush         │  │
│         └─────────────────┼─►│   Track │  on runtimeDone         │  │
│                           │  └────┬────┘                         │  │
│                           │       │                              │  │
│                           │       ▼                              │  │
│                           │  ┌─────────────┐                     │  │
│                           │  │ Loki Client │                     │  │
│                           │  │ • Batching  │                     │  │
│                           │  │ • Gzip      │                     │  │
│                           │  │ • 2-tier    │                     │  │
│                           │  │   Retries   │                     │  │
│                           │  └──────┬──────┘                     │  │
│                           └─────────┼────────────────────────────┘  │
└─────────────────────────────────────┼───────────────────────────────┘
                                      │ HTTPS
                                      ▼
                            ┌──────────────────┐
                            │   Grafana Loki   │
                            └──────────────────┘
```

### State Machine

LambdaWatch uses adaptive flush intervals based on invocation state:

```
     INVOKE
        │
        ▼
    ┌───────┐
    │ACTIVE │ ◄─── Normal flush interval (1s)
    └───┬───┘
        │ runtimeDone
        ▼
  ┌──────────┐
  │ FLUSHING │ ◄─── Critical flush + extended interval
  └────┬─────┘
       │ complete
       ▼
   ┌──────┐
   │ IDLE │ ◄─── Extended interval (3x) — saves costs!
   └──────┘
```

---

## Performance

| Metric | Value |
|--------|-------|
| Binary size | ~6MB |
| Memory overhead | ~10MB |
| Cold start impact | <50ms |
| Compression ratio | ~80% reduction |

---

## Comparison with Alternatives

| Feature | LambdaWatch | CloudWatch | Datadog | Other Extensions |
|---------|:-----------:|:----------:|:-------:|:----------------:|
| Self-hosted | ✅ | ❌ | ❌ | Varies |
| Zero vendor lock-in | ✅ | ❌ | ❌ | ❌ |
| No code changes | ✅ | ✅ | ❌ | ✅ |
| Request ID tracking | ✅ | ❌ | ✅ | ❌ |
| Adaptive intervals | ✅ | N/A | ❌ | ❌ |
| Critical flush | ✅ | N/A | ✅ | ❌ |
| Cost | Free + Loki | $$$$ | $$$$ | Varies |

---

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint
make lint

# Build for local testing
make build
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Grafana Loki](https://grafana.com/oss/loki/) for the amazing log aggregation system
- [AWS Lambda Extensions](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html) for making this possible

---

<p align="center">
  <sub>Built with ❤️ for the cloud-native community</sub>
</p>
