import {
  Layers,
  Timer,
  Shrink,
  ShieldCheck,
  RefreshCw,
  Activity,
  Fingerprint,
  Tag,
  Settings2,
} from "lucide-react";

// Navigation
export const NAV_LINKS = [
  { label: "Features", href: "#features" },
  { label: "Architecture", href: "#architecture" },
  { label: "Performance", href: "#performance" },
  { label: "Compare", href: "#comparison" },
  { label: "FAQ", href: "#faq" },
] as const;

export const GITHUB_REPO = "mumzworld-tech/lambdawatch";
export const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;
export const RELEASES_URL = `${GITHUB_URL}/releases`;

// Hero
export const HERO = {
  badgeFallback: "Open Source Lambda Extension",
  headlineWhite: "Fastest Way to Ship",
  headlineMid: "Lambda Logs to",
  headlineGradient: "Grafana Loki",
  subtitle: "Zero code changes. Zero vendor lock-in. Just add the layer.",
  downloadCommand: `curl -sL https://github.com/${GITHUB_REPO}/releases/latest/download/extension-arm64.zip -o lambdawatch.zip`,
} as const;

// Features (9 items for 3x3 grid)
export const FEATURES = [
  {
    icon: Layers,
    title: "Zero Code Changes",
    description:
      "Runs as an external Lambda Extension (Layer). No SDK, no wrapper, no code modifications required.",
  },
  {
    icon: Timer,
    title: "Automatic Batching",
    description:
      "Intelligent batch sizing with configurable max entries and byte limits per flush cycle.",
  },
  {
    icon: Shrink,
    title: "Gzip Compression",
    description:
      "Automatic gzip compression for payloads exceeding the threshold. ~80% reduction in bandwidth.",
  },
  {
    icon: ShieldCheck,
    title: "Guaranteed Delivery",
    description:
      "Critical flush on runtime completion ensures no logs are lost, even during Lambda freezes.",
  },
  {
    icon: RefreshCw,
    title: "Two-Tier Retry",
    description:
      "Regular flush retries 3x, critical flush retries 5x with exponential backoff. Handles transient failures.",
  },
  {
    icon: Activity,
    title: "Adaptive Intervals",
    description:
      "1s flush during active invocations, 3s during idle. Optimizes cost without sacrificing latency.",
  },
  {
    icon: Fingerprint,
    title: "Request ID Tracking",
    description:
      "Automatically extracts and injects request IDs into log content for easy correlation in Loki.",
  },
  {
    icon: Tag,
    title: "Auto-Labeling",
    description:
      "Automatically labels streams with function name, version, and region. Zero configuration needed.",
  },
  {
    icon: Settings2,
    title: "Custom Labels",
    description:
      "Add custom Loki labels via LOKI_LABELS JSON environment variable for advanced filtering.",
  },
] as const;

// Architecture nodes
export const ARCHITECTURE_NODES = [
  { id: "lambda", label: "Lambda Function", icon: "function" },
  { id: "telemetry", label: "Telemetry API", icon: "radio" },
  { id: "server", label: "Server :8080", icon: "server" },
  { id: "buffer", label: "Buffer", icon: "database" },
  { id: "client", label: "Loki Client", icon: "send" },
  { id: "loki", label: "Grafana Loki", icon: "bar-chart" },
] as const;

export const STATE_MACHINE = [
  { state: "ACTIVE", interval: "1s flush", color: "text-brand-green" },
  { state: "FLUSHING", interval: "Critical flush", color: "text-brand" },
  { state: "IDLE", interval: "3s flush", color: "text-text-secondary" },
] as const;

// Performance metrics
export const PERFORMANCE_METRICS = [
  { value: 6, suffix: " MB", label: "Binary Size", description: "Compiled Go binary with zero external dependencies" },
  { value: 50, prefix: "<", suffix: " ms", label: "Cold Start Impact", description: "Minimal overhead on Lambda cold start latency" },
  { value: 80, suffix: "%", prefix: "~", label: "Compression Ratio", description: "Gzip compression reduces payload bandwidth" },
  { value: 10, suffix: " MB", prefix: "~", label: "Memory Overhead", description: "Lightweight footprint alongside your function" },
] as const;

// Performance chart data
export const PERFORMANCE_CHART_DATA = [
  { name: "LambdaWatch", size: 6, color: "#FF9900" },
  { name: "Datadog Extension", size: 7, color: "#71717A" },
  { name: "CloudWatch (built-in)", size: 0, color: "#71717A" },
  { name: "Other Extensions", size: 45, color: "#71717A" },
] as const;

// Comparison table
export const COMPARISON_FEATURES = [
  "Self-hosted / Open Source",
  "Zero vendor lock-in",
  "No code changes required",
  "Request ID tracking",
  "Adaptive flush intervals",
  "Critical flush guarantee",
  "Cost",
] as const;

export const COMPARISON_PRODUCTS = [
  {
    name: "LambdaWatch",
    highlighted: true,
    values: [true, true, true, true, true, true, "Free"],
  },
  {
    name: "CloudWatch",
    highlighted: false,
    values: [false, false, true, false, false, false, "Pay per GB"],
  },
  {
    name: "Datadog",
    highlighted: false,
    values: [false, false, false, true, false, true, "$$/host"],
  },
  {
    name: "Other Extensions",
    highlighted: false,
    values: [false, false, true, false, false, false, "$$/GB"],
  },
] as const;

// FAQ
export const FAQ_ITEMS = [
  {
    question: "Which Lambda runtimes are supported?",
    answer:
      "LambdaWatch works with all Lambda runtimes (Node.js, Python, Go, Java, .NET, Ruby, and custom runtimes). It runs as an external extension via the Telemetry API, so it's completely runtime-agnostic.",
  },
  {
    question: "What's the cold start impact?",
    answer:
      "Less than 50ms. The extension binary is ~6MB and initializes quickly. It registers with the Extensions API and starts the Telemetry API listener before your function code runs.",
  },
  {
    question: "What happens during Lambda shutdown?",
    answer:
      "LambdaWatch performs a critical flush with deadline-bounded context (Lambda's DeadlineMs - 500ms). It retries up to 5 times with exponential backoff to ensure all buffered logs are delivered before the execution environment is frozen.",
  },
  {
    question: "Does it work with Grafana Cloud?",
    answer:
      "Yes. Set LOKI_URL to your Grafana Cloud Loki endpoint, LOKI_USERNAME to your instance ID, and LOKI_PASSWORD to your API key. Basic auth and bearer token auth are both supported.",
  },
  {
    question: "Why inject request_id into content instead of labels?",
    answer:
      "Loki labels should be low-cardinality. Request IDs are high-cardinality (unique per invocation) and would create millions of separate streams, degrading Loki performance. Instead, we inject the request_id into the log message content and query with: {function_name=\"x\"} | json | request_id=\"abc\"",
  },
  {
    question: "How do I add custom labels?",
    answer:
      "Set the LOKI_LABELS environment variable to a JSON string: LOKI_LABELS='{\"team\":\"platform\",\"env\":\"prod\"}'. These are merged with the automatic labels (source, function_name, region).",
  },
  {
    question: "What does it cost?",
    answer:
      "LambdaWatch itself is free and open source (MIT license). You only pay for your Loki infrastructure (self-hosted or Grafana Cloud) and minimal Lambda overhead (~10MB memory, <50ms cold start).",
  },
  {
    question: "What if Loki is unreachable?",
    answer:
      "Logs are buffered in a circular buffer (default 10,000 entries). Push failures retry with exponential backoff (3x regular, 5x critical). If the buffer fills, oldest entries are dropped to prevent memory issues. The extension never crashes your Lambda function.",
  },
] as const;

// Footer
export const FOOTER_LINKS = {
  resources: [
    { label: "Documentation", href: GITHUB_URL + "#readme" },
    { label: "Releases", href: RELEASES_URL },
    { label: "Architecture", href: GITHUB_URL + "/blob/main/.claude/claude-md-refs/architecture.md" },
    { label: "Contributing", href: GITHUB_URL + "/blob/main/CONTRIBUTING.md" },
  ],
  community: [
    { label: "GitHub Issues", href: GITHUB_URL + "/issues" },
    { label: "Discussions", href: GITHUB_URL + "/discussions" },
    { label: "Security Policy", href: GITHUB_URL + "/blob/main/SECURITY.md" },
  ],
} as const;
