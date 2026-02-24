# Exports Reference

## Go Extension Packages

### cmd/extension

| Export | Type | Purpose |
|--------|------|---------|
| `main()` | func | Entry point: config → validate → signals → Manager.Run() |

### internal/config

| Export | Type | Purpose |
|--------|------|---------|
| `Config` | struct | All configuration fields (Loki, batching, reliability, buffer, labels) |
| `Load()` | func | Reads env vars with defaults, parses LOKI_LABELS JSON |

**Config Fields:**

| Field | Env Var | Default | Purpose |
|-------|---------|---------|---------|
| LokiEndpoint | `LOKI_URL` | (required) | Loki push API endpoint |
| LokiUsername | `LOKI_USERNAME` | "" | Basic auth username |
| LokiPassword | `LOKI_PASSWORD` | "" | Basic auth password |
| LokiAPIKey | `LOKI_API_KEY` | "" | Bearer token auth |
| LokiTenantID | `LOKI_TENANT_ID` | "" | X-Scope-OrgID header |
| BatchSize | `LOKI_BATCH_SIZE` | 100 | Max entries per batch |
| MaxBatchSizeBytes | `LOKI_MAX_BATCH_SIZE_BYTES` | 5MB | Max batch bytes |
| FlushIntervalMs | `LOKI_FLUSH_INTERVAL_MS` | 1000 | Base flush interval |
| IdleFlushMultiplier | `LOKI_IDLE_FLUSH_MULTIPLIER` | 3 | Idle interval multiplier |
| MaxRetries | `LOKI_MAX_RETRIES` | 3 | Regular flush retries |
| CriticalFlushRetries | `LOKI_CRITICAL_FLUSH_RETRIES` | 5 | Critical flush retries |
| EnableGzip | `LOKI_ENABLE_GZIP` | true | Gzip compression |
| CompressionThreshold | `LOKI_COMPRESSION_THRESHOLD` | 1024 | Min bytes to compress |
| BufferSize | `BUFFER_SIZE` | 10000 | Circular buffer capacity |
| MaxLineSize | `LOKI_MAX_LINE_SIZE` | 204800 | Max bytes per log line |
| ExtractRequestID | `LOKI_EXTRACT_REQUEST_ID` | true | Embed request_id in content |
| Labels | `LOKI_LABELS` | {} | Custom Loki labels (JSON) |

### internal/extension

| Export | Type | Purpose |
|--------|------|---------|
| `State` | type (int32) | Extension operational state enum |
| `StateIdle` | const | No active invocation (3x flush interval) |
| `StateActive` | const | Invocation in progress (1x flush interval) |
| `StateFlushing` | const | Critical flush in progress |
| `Manager` | struct | Core orchestrator: state machine, flush loop, event loop |
| `NewManager(cfg)` | func | Creates Manager with buffer, channels, initial StateIdle |
| `Manager.Run(ctx)` | method | init() → flushLoop goroutine → eventLoop |
| `Client` | struct | Lambda Extensions API HTTP client |
| `NewClient()` | func | Creates client from AWS_LAMBDA_RUNTIME_API env |
| `Client.Register(ctx)` | method | POST /register → returns RegisterResponse + extension ID |
| `Client.NextEvent(ctx)` | method | GET /event/next → blocks until INVOKE or SHUTDOWN |
| `Client.GetExtensionID()` | method | Returns extension identifier string |
| `EventType` | type (string) | "INVOKE" or "SHUTDOWN" |
| `Invoke` | const (EventType) | `"INVOKE"` — Lambda invocation event |
| `Shutdown` | const (EventType) | `"SHUTDOWN"` — Lambda shutdown event |
| `RegisterResponse` | struct | FunctionName, FunctionVersion, Handler |
| `NextEventResponse` | struct | EventType, DeadlineMs, RequestID, ShutdownReason |
| `Tracing` | struct | X-Ray Type + Value |

### internal/buffer

| Export | Type | Purpose |
|--------|------|---------|
| `LogEntry` | struct | Timestamp, Message, Type, RequestID |
| `LogEntry.Size()` | method | Approximate byte size of entry |
| `Buffer` | struct | Thread-safe bounded circular buffer |
| `New(maxSize)` | func | Creates buffer with capacity |
| `Buffer.Add(entry)` | method | Add one entry, drops oldest if full |
| `Buffer.AddBatch(entries)` | method | Add multiple entries, signals Ready |
| `Buffer.Flush(batchSize)` | method | Extract up to batchSize entries |
| `Buffer.FlushBySize(batchSize, maxBytes)` | method | Extract entries bounded by count AND bytes |
| `Buffer.Drain()` | method | Return all entries, close buffer |
| `Buffer.Len()` | method | Current entry count |
| `Buffer.ByteSize()` | method | Current total byte size |
| `Buffer.Ready()` | method | Returns channel signaling logs available |
| `Buffer.SignalReady()` | method | Manually signal log readiness |

### internal/loki

| Export | Type | Purpose |
|--------|------|---------|
| `Client` | struct | Loki HTTP client with auth, gzip, retries |
| `NewClient(cfg)` | func | Creates client from Config |
| `Client.Push(ctx, req)` | method | Regular flush push (MaxRetries) |
| `Client.PushCritical(ctx, req)` | method | Critical flush push (CriticalFlushRetries) |
| `Batch` | struct | Collects LogEntries for a single push |
| `NewBatch(labels, extractRequestID)` | func | Creates batch with stream labels |
| `Batch.Add(entries)` | method | Append entries to batch |
| `Batch.Len()` | method | Entry count |
| `Batch.ToPushRequest()` | method | Convert to PushRequest (ms→ns timestamps, inject request ID) |
| `PushRequest` | struct | Loki push API body: Streams[]  |
| `Stream` | struct | Stream labels + Values ([][timestamp, message]) |
| `NewPushRequest(labels, values)` | func | Helper to create PushRequest |

### internal/telemetryapi

| Export | Type | Purpose |
|--------|------|---------|
| `RuntimeDoneHandler` | type (func) | Callback for platform.runtimeDone |
| `Server` | struct | HTTP receiver on :8080 for Telemetry API |
| `NewServer(buf, port, maxLineSize, extractRequestID, onRuntimeDone)` | func | Creates telemetry server |
| `Server.Start()` | method | ListenAndServe in goroutine |
| `Server.Shutdown(ctx)` | method | Graceful HTTP shutdown |
| `Server.ListenerURI()` | method | Returns sandbox URI for subscription |
| `Client` | struct | Telemetry API subscription client |
| `NewClient(extensionID)` | func | Creates client from AWS_LAMBDA_RUNTIME_API |
| `Client.Subscribe(ctx, listenerURI)` | method | PUT subscribe: platform+function+extension |
| `TelemetryEvent` | struct | Time, Type, Record (interface{}) |
| `PlatformStartRecord` | struct | RequestID, Version |
| `PlatformRuntimeDoneRecord` | struct | RequestID, Status, Metrics |
| `PlatformReportRecord` | struct | RequestID, Status, Metrics |
| `Metrics` | struct | DurationMs, BilledDurationMs, MemorySizeMB, MaxMemoryUsedMB, InitDurationMs |
| `SubscribeRequest` | struct | SchemaVersion, Types, Buffering, Destination |
| `BufferConfig` | struct | MaxItems, MaxBytes, TimeoutMs |
| `Destination` | struct | Protocol, URI |
| `EventTypePlatformStart` | const | `"platform.start"` — Lambda start event |
| `EventTypePlatformEnd` | const | `"platform.end"` — Lambda end event |
| `EventTypePlatformReport` | const | `"platform.report"` — Invocation metrics |
| `EventTypePlatformRuntimeDone` | const | `"platform.runtimeDone"` — Triggers critical flush |
| `EventTypePlatformFault` | const | `"platform.fault"` — Lambda fault |
| `EventTypePlatformExtension` | const | `"platform.extension"` — Extension event |
| `EventTypePlatformLogsDropped` | const | `"platform.logsDropped"` — Logs dropped |
| `EventTypeFunction` | const | `"function"` — Function logs |
| `EventTypeExtension` | const | `"extension"` — Extension logs |

### internal/logger

| Export | Type | Purpose |
|--------|------|---------|
| `Init()` | func | Set appName, environment, debugMode from env |
| `SetBuffer(buf)` | func | Direct buffer writes (Telemetry API can't capture own logs) |
| `Info(msg)` | func | Info log |
| `Debug(msg)` | func | Debug log (requires DEBUG_MODE=true) |
| `Warn(msg)` | func | Warning log |
| `Error(msg)` | func | Error log |
| `Infof/Debugf/Warnf/Errorf` | func | Formatted variants |
| `Fatal(msg)` / `Fatalf(fmt)` | func | Fatal log + os.Exit(1) |

---

## Website (Next.js) Exports

### Routes

| Route | File | Type | Purpose |
|-------|------|------|---------|
| `/` | `website/app/page.tsx` | Server Component | Home page — renders all 8 sections |

### Layouts

| Layout | File | Scope |
|--------|------|-------|
| Root | `website/app/layout.tsx` | All pages — fonts, metadata, dark mode |

### Library Modules

| Module | File | Exports | Purpose |
|--------|------|---------|---------|
| constants | `website/lib/constants.ts` | NAV_LINKS (5), GITHUB_REPO, GITHUB_URL, RELEASES_URL, HERO, FEATURES (9), ARCHITECTURE_NODES (6), STATE_MACHINE (3), PERFORMANCE_METRICS (4), PERFORMANCE_CHART_DATA (4), COMPARISON_FEATURES (7), COMPARISON_PRODUCTS (4), FAQ_ITEMS (8), FOOTER_LINKS | All static content and configuration |
| fonts | `website/lib/fonts.ts` | calSans, inter, jetbrainsMono | Font configuration (Cal Sans local, Inter + JetBrains Mono Google) |
| github | `website/lib/github.ts` | getGitHubStars(), getLatestRelease(), GitHubRelease | Fetch GitHub stars + latest release (ISR: 3600s) |
| utils | `website/lib/utils.ts` | cn() | clsx + tailwind-merge utility |

### Hooks

| Hook | File | Parameters | Returns |
|------|------|------------|---------|
| useMousePosition | `use-mouse-position.ts` | `ref: RefObject, options?: { maxRotation?, springConfig? }` | `{ x, y, rotateX, rotateY }` (MotionValues) |
| useScrollProgress | `use-scroll-progress.ts` | `options?: { offset? }` | `{ ref, progress, opacity, y }` (ref + MotionValues) |

### Section Components (`website/components/sections/`)

| Component | File | Purpose |
|-----------|------|---------|
| Navbar | `navbar.tsx` | Floating glass morphism pill navbar, nav links, mobile menu, GitHub stars |
| Hero | `hero.tsx` | Full-height hero: badge, headline, download buttons, terminal |
| Features | `features.tsx` | 3-column grid of 9 feature cards with MagicCard |
| Architecture | `architecture.tsx` | Data flow diagram with AnimatedBeam + state machine viz |
| Performance | `performance.tsx` | Metrics grid + bar chart (horizontal desktop, vertical mobile) |
| Comparison | `comparison.tsx` | Product comparison table (4 products, 7 features) |
| FAQ | `faq.tsx` | Accordion with 8 FAQ items |
| Footer | `footer.tsx` | 3-column footer: brand, resources, community, Mumzworld branding |

### Common Components (`website/components/common/`)

| Component | File | Key Props | Purpose |
|-----------|------|-----------|---------|
| SectionWrapper | `section-wrapper.tsx` | `id?, className?, fullWidth?` | Responsive section container (max-w-7xl, py-24 md:py-32) |
| SectionHeading | `section-heading.tsx` | `title, subtitle?, align?("left"\|"center")` | Title + subtitle with BlurFade animation |
| SectionDivider | `section-divider.tsx` | `className?` | Gradient line divider between sections |
| GlowEffect | `glow-effect.tsx` | `size?("sm"\|"md"\|"lg"), className?` | Animated glow background (300-700px) |
| GlassmorphicCard | `glassmorphic-card.tsx` | `children, className?, hover?` | Glass morphism card with backdrop blur |
| GradientText | `gradient-text.tsx` | `from?, via?, to?, gradient?` | Gradient text (use `gradient` for full string or `from/via/to` for Tailwind) |
| ShimmerBadge | `shimmer-badge.tsx` | `children, href?, className?` | Badge with pulsing green dot, optional link |
| TerminalBlock | `terminal-block.tsx` | `command, className?` | Terminal-like code block with copy button |
| GitHubStarButton | `github-star-button.tsx` | `stars?, className?` | GitHub star count pill button |
| DownloadButtonGroup | `download-button-group.tsx` | `className?` | ARM64/AMD64 download dropdown |
| IconBox | `icon-box.tsx` | `icon: LucideIcon, size?("sm"\|"md"\|"lg")` | Icon container with gradient background |
| AnimatedCounter | `animated-counter.tsx` | `value, prefix?, suffix?, className?` | Wraps NumberTicker with prefix/suffix |

All common components are barrel-exported from `website/components/common/index.ts`.

### UI Components (`website/components/ui/`) — shadcn + custom

**Base (Radix UI wrappers):**

| Component | File | Key Exports |
|-----------|------|-------------|
| Accordion | `accordion.tsx` | Accordion, AccordionItem, AccordionTrigger, AccordionContent |
| Badge | `badge.tsx` | Badge, badgeVariants (default/secondary/destructive/outline/ghost/link) |
| Button | `button.tsx` | Button, buttonVariants (default/destructive/outline/secondary/ghost/link, sizes: xs/sm/default/lg/xl/icon/icon-xs) |
| Card | `card.tsx` | Card, CardHeader, CardTitle, CardDescription, CardAction, CardContent, CardFooter |
| Chart | `chart.tsx` | ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent |
| DropdownMenu | `dropdown-menu.tsx` | DropdownMenu + 15 sub-components |
| NavigationMenu | `navigation-menu.tsx` | NavigationMenu + 8 sub-components, navigationMenuTriggerStyle |
| ScrollArea | `scroll-area.tsx` | ScrollArea, ScrollBar |
| Separator | `separator.tsx` | Separator |
| Table | `table.tsx` | Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption |
| Tabs | `tabs.tsx` | Tabs, TabsList, TabsTrigger, TabsContent, tabsListVariants |
| Tooltip | `tooltip.tsx` | Tooltip, TooltipProvider, TooltipTrigger, TooltipContent |

**Animation/Effect:**

| Component | File | Props | Purpose |
|-----------|------|-------|---------|
| AnimatedBeam | `animated-beam.tsx` | containerRef, fromRef, toRef, curvature, reverse, colors | SVG beams connecting elements |
| AnimatedGridPattern | `animated-grid-pattern.tsx` | width, height, numSquares, maxOpacity, duration | Animated background grid |
| AnimatedShinyText | `animated-shiny-text.tsx` | shimmerWidth | Shimmer text effect |
| BlurFade | `blur-fade.tsx` | duration, delay, offset, direction, inView, blur | Scroll-triggered blur+fade |
| BorderBeam | `border-beam.tsx` | size, duration, delay, colors, reverse, borderWidth | Animated border effect |
| MagicCard | `magic-card.tsx` | gradientSize, gradientColor, gradientFrom, gradientTo, gradientOpacity, tilt, tiltAmount | Mouse-tracking gradient card with optional 3D tilt |
| Marquee | `marquee.tsx` | reverse, pauseOnHover, vertical, repeat | Scrolling content |
| NumberTicker | `number-ticker.tsx` | value, startValue, direction, delay, decimalPlaces | Spring-animated counter |
| Particles | `particles.tsx` | quantity, staticity, ease, size, color, vx, vy | Canvas particle system |
| ScriptCopyBtn | `script-copy-btn.tsx` | text, showText | Copy-to-clipboard button |
| ShineBorder | `shine-border.tsx` | borderWidth, duration, shineColor | Animated shine border |
| TextAnimate | `text-animate.tsx` | by (char/word/line/text), variants (fadeIn/blurIn/slideUp/etc.) | Text entry animation |

### Import Patterns

```typescript
// Components
import { Navbar } from "@/components/sections/navbar";
import { SectionWrapper, SectionHeading, SectionDivider } from "@/components/common";
import { Button } from "@/components/ui/button";
import { MagicCard } from "@/components/ui/magic-card";

// Hooks
import { useMousePosition } from "@/hooks/use-mouse-position";
import { useScrollProgress } from "@/hooks/use-scroll-progress";

// Lib
import { FEATURES, HERO, NAV_LINKS } from "@/lib/constants";
import { calSans, inter, jetbrainsMono } from "@/lib/fonts";
import { getGitHubStars, getLatestRelease } from "@/lib/github";
import { cn } from "@/lib/utils";

// External
import { motion, useInView, useMotionValue } from "motion/react";
import { Check, X, ChevronDown } from "lucide-react";
```
