"use client";

import { cn } from "@/lib/utils";

interface ShimmerBadgeProps {
  children: React.ReactNode;
  className?: string;
  href?: string;
}

export function ShimmerBadge({ children, className, href }: ShimmerBadgeProps) {
  const content = (
    <div
      className={cn(
        "inline-flex items-center gap-2.5 rounded-full border border-border-medium bg-surface-light/80 px-4 py-2 text-sm backdrop-blur-sm transition-colors",
        href && "hover:border-border-strong hover:bg-surface-lighter/80 cursor-pointer",
        className
      )}
    >
      <span className="relative flex h-2 w-2">
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-green opacity-75" />
        <span className="relative inline-flex h-2 w-2 rounded-full bg-brand-green" />
      </span>
      <span className="text-text-primary font-medium">{children}</span>
    </div>
  );

  if (href) {
    return (
      <a href={href} target="_blank" rel="noopener noreferrer">
        {content}
      </a>
    );
  }

  return content;
}
