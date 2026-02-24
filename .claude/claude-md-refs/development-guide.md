# Development Guide

## Go Extension

### Adding a New Config Field

1. Add field to `Config` struct in `internal/config/config.go`:
```go
type Config struct {
    // ... existing fields
    NewField int
}
```

2. Load from env var in `Load()`:
```go
cfg := &Config{
    // ... existing
    NewField: getEnvInt("LOKI_NEW_FIELD", 42),
}
```

3. Use in consuming package (e.g., `internal/extension/lifecycle.go` or `internal/loki/client.go`)

4. Add test in `internal/config/config_test.go`

5. Document in `exports-reference.md` Config Fields table

### Adding a New Telemetry Event Handler

1. Add event type constant in `internal/telemetryapi/types.go`:
```go
const EventTypePlatformNewEvent = "platform.newEvent"
```

2. Add record struct if needed:
```go
type PlatformNewEventRecord struct {
    RequestID string `json:"requestId"`
    // fields...
}
```

3. Add case in `Server.handleTelemetry()` switch in `internal/telemetryapi/server.go`:
```go
case EventTypePlatformNewEvent:
    ts := parseTimestamp(event.Time)
    entry := buffer.LogEntry{
        Timestamp: ts,
        Message:   formatNewEvent(event.Record),
        Type:      event.Type,
        RequestID: currentReqID,
    }
    entries = append(entries, entry)
```

4. Add formatter function:
```go
func formatNewEvent(record interface{}) string {
    // Format as human-readable or JSON
}
```

5. Add tests in `internal/telemetryapi/server_test.go`

### Adding a New State Transition

States live in `internal/extension/lifecycle.go`. The state machine is:

```
INVOKE → ACTIVE (1x flush) → platform.runtimeDone → FLUSHING (critical) → IDLE (3x flush)
```

1. Add state constant:
```go
const (
    StateIdle     State = iota
    StateActive
    StateFlushing
    StateNewState  // Add here
)
```

2. Update `State.String()` method

3. Add interval logic in `getFlushInterval()`

4. Add transition trigger in appropriate handler (eventLoop, onRuntimeDone, etc.)

5. Add test in `internal/extension/lifecycle_test.go`

### Running Tests

```bash
# All tests
make test

# Single package
go test -v ./internal/buffer/
go test -v ./internal/loki/

# Single test
go test -v -run TestCriticalFlush ./internal/extension/

# Coverage
make test-coverage  # Opens coverage.html
```

### Building & Deploying

```bash
make build-arm64    # Graviton (recommended)
make package        # Creates extension.zip Lambda Layer
make deploy         # Publishes to AWS
```

---

## Website (Next.js)

### Project Setup

```bash
cd website
pnpm install        # Install dependencies
pnpm dev            # Start dev server (localhost:3000)
pnpm build          # Static export build
pnpm lint           # ESLint check
```

**Key config:** Static export (`output: "export"` in next.config.ts), unoptimized images, Tailwind CSS 4 with @tailwindcss/postcss.

### Adding a New Page Section

1. Create section component in `website/components/sections/`:
```tsx
"use client";

import { SectionWrapper } from "@/components/common";
import { SectionHeading } from "@/components/common";
import { BlurFade } from "@/components/ui/blur-fade";

export function NewSection() {
  return (
    <SectionWrapper id="new-section">
      <SectionHeading
        title="Section Title"
        subtitle="Description text"
      />
      <BlurFade delay={0.3}>
        {/* Section content */}
      </BlurFade>
    </SectionWrapper>
  );
}
```

2. Add to page in `website/app/page.tsx`:
```tsx
import { NewSection } from "@/components/sections/new-section";

export default async function Home() {
  return (
    <main>
      {/* existing sections */}
      <SectionDivider />
      <NewSection />
      {/* more sections */}
    </main>
  );
}
```

3. Add nav link in `website/lib/constants.ts` (NAV_LINKS array):
```typescript
export const NAV_LINKS = [
  // existing...
  { label: "New Section", href: "#new-section" },
];
```

### Adding a New Common Component

1. Create in `website/components/common/`:
```tsx
interface NewComponentProps {
  children: React.ReactNode;
  className?: string;
}

export function NewComponent({ children, className }: NewComponentProps) {
  return (
    <div className={cn("base-styles", className)}>
      {children}
    </div>
  );
}
```

2. Export from `website/components/common/index.ts`:
```typescript
export { NewComponent } from "./new-component";
```

### Adding Content to Constants

All static content lives in `website/lib/constants.ts`. Pattern:

```typescript
// Feature items (used by Features section)
export const FEATURES: FeatureItem[] = [
  {
    icon: Zap,          // Lucide icon
    title: "Feature Name",
    description: "Description text",
  },
];

// FAQ items (used by FAQ section)
export const FAQ_ITEMS = [
  {
    question: "Question?",
    answer: "Answer text",
  },
];
```

### Adding a shadcn UI Component

```bash
cd website
pnpm dlx shadcn@latest add [component-name]
```

Components install to `website/components/ui/`. Config in `website/components.json`:
- Style: new-york
- Base color: neutral
- Icons: lucide
- Path alias: `@/`

### Design System

**Colors** (defined in `website/app/globals.css` via `@theme`):
- Brand: `--color-brand` (green-400 oklch)
- Surface: `--color-surface-*` (dark grays)
- Border: `--color-border-*` (subtle/default)
- Glass: `--color-glass-*` (backdrop effects)
- Text: `--color-text-*` (primary/secondary/muted)

**Fonts:**
- `font-display` → Cal Sans SemiBold (headings)
- `font-sans` → Inter (body)
- `font-mono` → JetBrains Mono (code)

**Animations** (CSS):
- `fade-in`, `slide-up`, `glow-pulse`
- `shiny-text`, `shine`, `marquee`, `reverse-marquee`
- `border-beam`, `border-beam-reverse`

**Component Patterns:**
- All interactive components use `"use client"` directive
- Animation via `motion/react` (not framer-motion)
- Icons from `lucide-react`
- Variants via `class-variance-authority` (CVA)
- Class merging via `cn()` = clsx + tailwind-merge

### Component Dependency Tree

```
page.tsx (Server Component)
├── Navbar → GitHubStarButton
├── Hero → ShimmerBadge, GradientText, TerminalBlock, DownloadButtonGroup, AnimatedGridPattern, Particles, GlowEffect
├── Features → SectionWrapper, SectionHeading, MagicCard, IconBox
├── Architecture → SectionWrapper, SectionHeading, AnimatedBeam, BorderBeam, GlassmorphicCard, useMousePosition
├── Performance → SectionWrapper, SectionHeading, AnimatedCounter, GlassmorphicCard
├── Comparison → SectionWrapper, SectionHeading, Table, ScrollArea, GlassmorphicCard, ShineBorder
├── FAQ → SectionWrapper, SectionHeading, Accordion
├── SectionDivider (between each section)
└── Footer → SectionDivider, Badge, GitHubStarButton
```

### GitHub Stars (Server-Side Data)

Stars are fetched server-side in `page.tsx` via `getGitHubStars()` and passed as props:

```tsx
// website/app/page.tsx
const stars = await getGitHubStars();
// Passed to: <Navbar stars={stars} />, <Hero stars={stars} />, <Footer stars={stars} />
```

`getGitHubStars()` uses Next.js ISR with `revalidate: 3600` (1 hour).
