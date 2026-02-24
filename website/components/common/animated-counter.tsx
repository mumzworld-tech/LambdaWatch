"use client";

import { cn } from "@/lib/utils";
import { NumberTicker } from "@/components/ui/number-ticker";

interface AnimatedCounterProps {
  value: number;
  prefix?: string;
  suffix?: string;
  className?: string;
}

export function AnimatedCounter({
  value,
  prefix,
  suffix,
  className,
}: AnimatedCounterProps) {
  return (
    <span className={cn("tabular-nums", className)}>
      {prefix}
      <NumberTicker value={value} />
      {suffix}
    </span>
  );
}
