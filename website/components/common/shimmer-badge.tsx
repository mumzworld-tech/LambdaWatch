"use client";

import { cn } from "@/lib/utils";
import { AnimatedShinyText } from "@/components/ui/animated-shiny-text";

interface ShimmerBadgeProps {
  children: React.ReactNode;
  className?: string;
}

export function ShimmerBadge({ children, className }: ShimmerBadgeProps) {
  return (
    <div
      className={cn(
        "inline-flex items-center rounded-full border border-border-medium bg-surface-light/50 px-4 py-1.5 text-sm backdrop-blur-sm",
        className
      )}
    >
      <AnimatedShinyText className="text-text-secondary" shimmerWidth={200}>
        {children}
      </AnimatedShinyText>
    </div>
  );
}
