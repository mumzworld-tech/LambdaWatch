<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/mumzworld-tech/lambdawatch/main/assets/logo-dark.png">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/mumzworld-tech/lambdawatch/main/assets/logo-light.png">
    <img src="https://raw.githubusercontent.com/mumzworld-tech/lambdawatch/main/assets/logo.png" alt="LambdaWatch Logo" width="300"/>
  </picture>
</p>

<p align="center">
  <strong>High-performance AWS Lambda Extension for shipping logs to Grafana Loki</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#contributing">Contributing</a>•
    <a href="#security">Security</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
  <img src="https://img.shields.io/badge/AWS-Lambda-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white" alt="AWS Lambda"/>
  <img src="https://img.shields.io/badge/Grafana-Loki-F46800?style=for-the-badge&logo=grafana&logoColor=white" alt="Grafana Loki"/>
  <img src="https://img.shields.io/github/license/mumzworld-tech/lambdawatch?style=for-the-badge" alt="License"/>
</p>

---

## What is LambdaWatch?

**LambdaWatch** is a lightweight, zero-dependency AWS Lambda Extension written in Go that automatically captures logs from your Lambda functions and ships them to [Grafana Loki](https://grafana.com/oss/loki/) in real-time.

No code changes required. Just add the layer and configure your Loki endpoint.

```
┌───────────────────────────────────────────────────────────┐
│                     AWS Lambda                            │
│  ┌─────────────┐         ┌────────────────────────────┐   │
│  │   Your      │         │      LambdaWatch           │   │
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
- **Clean JSON extraction** — Strips Lambda log prefixes, sends pure JSON to Loki

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

- **Structured extension logs** — Extension logs use same JSON format as your application
- **Request ID extraction** — Automatic `request_id` injection into log messages for request tracing
- **Auto-labeling** — Adds `function_name`, `function_version`, `region`
- **Custom labels** — Add your own labels via JSON config
- **Long message splitting** — Handles logs exceeding Loki's line limit

---

## Installation

### Step 1: Build the Extension

Clone the repository and build for your Lambda architecture:

```bash
git clone https://github.com/mumzworld-tech/lambdawatch.git
cd lambdawatch
```

**For ARM64 (Graviton) - Recommended for cost savings:**

```bash
make package
```

**For x86_64 (Intel/AMD):**

```bash
make package-amd64
```

This creates a zip file in the `build/` directory.

### Step 2: Publish as Lambda Layer

```bash
# For ARM64
aws lambda publish-layer-version \
  --layer-name lambdawatch \
  --zip-file fileb://build/lambdawatch-layer-arm64.zip \
  --compatible-architectures arm64 \
  --compatible-runtimes provided.al2023 provided.al2 nodejs20.x nodejs18.x python3.12 python3.11

# For x86_64
aws lambda publish-layer-version \
  --layer-name lambdawatch \
  --zip-file fileb://build/lambdawatch-layer-amd64.zip \
  --compatible-architectures x86_64 \
  --compatible-runtimes provided.al2023 provided.al2 nodejs20.x nodejs18.x python3.12 python3.11
```

Note the layer ARN from the output (e.g., `arn:aws:lambda:us-east-1:123456789:layer:lambdawatch:1`).

### Step 3: Attach Layer to Your Lambda Function

```bash
aws lambda update-function-configuration \
  --function-name YOUR_FUNCTION_NAME \
  --layers arn:aws:lambda:REGION:ACCOUNT_ID:layer:lambdawatch:VERSION
```

### Step 4: Configure Environment Variables

Set the required environment variables on your Lambda function:

```bash
aws lambda update-function-configuration \
  --function-name YOUR_FUNCTION_NAME \
  --environment "Variables={
    LOKI_URL=https://your-loki-instance.com/loki/api/v1/push,
    LOKI_USERNAME=your-username,
    LOKI_PASSWORD=your-password,
    SERVICE_NAME=your-service-name
  }"
```

**Required Variables:**
| Variable | Description |
|----------|-------------|
| `LOKI_URL` | Your Loki push endpoint URL |

**Recommended Variables:**
| Variable | Description |
|----------|-------------|
| `SERVICE_NAME` | Service identifier for grouping logs from multiple functions |
| `LOKI_USERNAME` | Basic auth username (if your Loki requires auth) |
| `LOKI_PASSWORD` | Basic auth password |

**Or use Bearer token auth:**
| Variable | Description |
|----------|-------------|
| `LOKI_API_KEY` | Bearer token for authentication |
| `LOKI_TENANT_ID` | Multi-tenant org ID (for Grafana Cloud) |

### Quick Start Example

Complete setup for a function called `my-api`:

```bash
# 1. Build and package
make package

# 2. Publish layer
LAYER_ARN=$(aws lambda publish-layer-version \
  --layer-name lambdawatch \
  --zip-file fileb://build/lambdawatch-layer-arm64.zip \
  --compatible-architectures arm64 \
  --query 'LayerVersionArn' \
  --output text)

# 3. Attach layer and configure
aws lambda update-function-configuration \
  --function-name my-api \
  --layers $LAYER_ARN \
  --environment "Variables={
    LOKI_URL=https://logs-prod-123.grafana.net/loki/api/v1/push,
    LOKI_USERNAME=123456,
    LOKI_PASSWORD=glc_xxxxxxxxxxxx,
    SERVICE_NAME=my-api-service
  }"
```

### Alternative: Pre-built Layer

Download the latest release from GitHub:

```bash
curl -LO https://github.com/mumzworld-tech/lambdawatch/releases/latest/download/lambdawatch-layer-arm64.zip
```

---

## Configuration

Configure via environment variables on your Lambda function:

### Required

| Variable   | Description                                                       |
| ---------- | ----------------------------------------------------------------- |
| `LOKI_URL` | Loki push URL (e.g., `https://loki.example.com/loki/api/v1/push`) |

### Authentication

| Variable         | Default | Description                                  |
| ---------------- | ------- | -------------------------------------------- |
| `LOKI_USERNAME`  | —       | Basic auth username                          |
| `LOKI_PASSWORD`  | —       | Basic auth password                          |
| `LOKI_API_KEY`   | —       | Bearer token (alternative to basic auth)     |
| `LOKI_TENANT_ID` | —       | Multi-tenant org ID (`X-Scope-OrgID` header) |

### Batching & Performance

| Variable                     | Default   | Description                   |
| ---------------------------- | --------- | ----------------------------- |
| `LOKI_BATCH_SIZE`            | `100`     | Max logs per batch            |
| `LOKI_MAX_BATCH_SIZE_BYTES`  | `5242880` | Max batch size (5MB)          |
| `LOKI_FLUSH_INTERVAL_MS`     | `1000`    | Flush interval in ms          |
| `LOKI_IDLE_FLUSH_MULTIPLIER` | `3`       | Interval multiplier when idle |

### Reliability

| Variable                      | Default | Description                         |
| ----------------------------- | ------- | ----------------------------------- |
| `LOKI_MAX_RETRIES`            | `3`     | Retry attempts for regular flushes  |
| `LOKI_CRITICAL_FLUSH_RETRIES` | `5`     | Retry attempts for critical flushes |
| `LOKI_ENABLE_GZIP`            | `true`  | Enable gzip compression             |
| `LOKI_COMPRESSION_THRESHOLD`  | `1024`  | Compress only if > 1KB              |

### Labels & Processing

| Variable                  | Default  | Description                                    |
| ------------------------- | -------- | ---------------------------------------------- |
| `LOKI_LABELS`             | `{}`     | Custom labels as JSON (e.g., `{"env":"prod"}`) |
| `LOKI_EXTRACT_REQUEST_ID` | `true`   | Extract `request_id` and inject into log messages |
| `LOKI_GROUP_BY_REQUEST_ID`| `false`  | Group logs into separate Loki streams by `request_id` |
| `LOKI_MAX_LINE_SIZE`      | `204800` | Max line size before splitting (200KB)         |
| `BUFFER_SIZE`             | `10000`  | Max logs in memory buffer                      |
| `DEBUG_MODE`              | `false`  | Enable verbose debug logging from extension    |

### Example Configuration

```bash
aws lambda update-function-configuration \
  --function-name my-function \
  --environment "Variables={
    LOKI_URL=https://loki.example.com/loki/api/v1/push,
    LOKI_USERNAME=myuser,
    LOKI_PASSWORD=mypassword,
    LOKI_LABELS={\"env\":\"production\",\"team\":\"backend\"}
  }"
```

---

## Querying Logs in Grafana

### Automatic Labels

Every log entry includes these labels automatically:

| Label              | Description                               | Source                           |
| ------------------ | ----------------------------------------- | -------------------------------- |
| `function_name`    | Lambda function name                      | Extensions API                   |
| `function_version` | Function version ($LATEST, 1, 2, etc.)    | Extensions API                   |
| `region`           | AWS region (us-east-1, etc.)              | AWS_REGION env                   |
| `request_id`       | Invocation request ID (stream label only when `LOKI_GROUP_BY_REQUEST_ID=true`) | Extracted from logs (if enabled) |
| `source`           | Always `lambda`                           | Hardcoded                        |
| `service_name`     | Service identifier for grouping functions | SERVICE_NAME env (optional)      |

### Example Queries

```logql
# All logs from a function
{function_name="my-function"}

# Filter by request ID (injected into log message by default)
{function_name="my-function"} | json | request_id="abc-123-def-456"

# Filter by request ID (when LOKI_GROUP_BY_REQUEST_ID=true)
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

### Structured Extension Logs

LambdaWatch extension logs are output in the same JSON structure as your application logs for consistent parsing in Grafana:

```json
{
  "level": "info",
  "timestamp": "2026-02-05T08:24:18.000Z",
  "app_name": "my-service",
  "environment": "production",
  "context": "LambdaWatch",
  "message": "Registered extension for function: my-function"
}
```

The extension uses `APP_NAME` environment variable for `app_name` field, falling back to `SERVICE_NAME` if not set. The `environment` field is populated from `NODE_ENV`.

To filter extension logs vs application logs in Grafana:

```logql
# Extension logs only
{service_name="my-service"} | json | context="LambdaWatch"

# Application logs only (exclude extension)
{service_name="my-service"} | json | context!="LambdaWatch"
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Lambda Execution Environment                     │
│                                                                     │
│  ┌──────────────┐         ┌──────────────────────────────────────┐  │
│  │   Lambda     │         │           LambdaWatch                │  │
│  │   Function   │         │                                      │  │
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

| Metric            | Value          |
| ----------------- | -------------- |
| Binary size       | ~6MB           |
| Memory overhead   | ~10MB          |
| Cold start impact | <50ms          |
| Compression ratio | ~80% reduction |

---

## Comparison with Alternatives

| Feature             | LambdaWatch | CloudWatch | Datadog | Other Extensions |
| ------------------- | :---------: | :--------: | :-----: | :--------------: |
| Self-hosted         |     ✅      |     ❌     |   ❌    |      Varies      |
| Zero vendor lock-in |     ✅      |     ❌     |   ❌    |        ❌        |
| No code changes     |     ✅      |     ✅     |   ❌    |        ✅        |
| Request ID tracking |     ✅      |     ❌     |   ✅    |        ❌        |
| Adaptive intervals  |     ✅      |    N/A     |   ❌    |        ❌        |
| Critical flush      |     ✅      |    N/A     |   ✅    |        ❌        |
| Cost                | Free + Loki |    $$$$    |  $$$$   |      Varies      |

---

## Development

### Setup

```bash
# Clone and install git hooks (required once after cloning)
git clone https://github.com/mumzworld-tech/lambdawatch.git
cd lambdawatch
lefthook install
```

> **Note:** [Lefthook](https://github.com/evilmartians/lefthook) is required for git hooks. Install it via `brew install lefthook` (macOS) or `go install github.com/evilmartians/lefthook@latest`.

### Git Hooks

This project uses Lefthook to run the following checks automatically on every commit:

- **`go fmt`** — Formats code and re-stages fixed files
- **`golangci-lint`** — Runs linter to catch issues early

These are configured in [`lefthook.yml`](lefthook.yml) and shared across the team.

### Commands

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

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.MD) for details on the development workflow, how to submit pull requests, and how to report bugs.

Please make sure to follow our [Code of Conduct](CODE_OF_CONDUCT.md) when participating in this project.

## Security

If you discover a security vulnerability, please **do not** open a public issue. Instead, follow the instructions in our [Security Policy](SECURITY.md) to report it responsibly.


---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Grafana Loki](https://grafana.com/oss/loki/) for the amazing log aggregation system
- [AWS Lambda Extensions](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html) for making this possible

---

<p align="center">
  <sub>Built with ❤️ by <a href="https://github.com/Sami-AlEsh">Sami</a> for the cloud-native community</sub>
</p>
