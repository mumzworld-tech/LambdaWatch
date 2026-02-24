"use client";

import { Star } from "lucide-react";
import { cn } from "@/lib/utils";
import { GITHUB_URL } from "@/lib/constants";

interface GitHubStarButtonProps {
  stars?: number | null;
  className?: string;
}

export function GitHubStarButton({ stars, className }: GitHubStarButtonProps) {
  return (
    <a
      href={GITHUB_URL}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        "inline-flex items-center gap-2 rounded-lg border border-border-medium bg-glass px-4 py-2.5 text-sm font-medium text-text-primary backdrop-blur-md transition-all duration-300",
        "hover:border-border-strong hover:bg-glass-light hover:shadow-[0_0_20px_rgba(255,153,0,0.08)]",
        className
      )}
    >
      <Star className="h-4 w-4 text-brand" />
      <span>Star on GitHub</span>
      {stars != null && (
        <>
          <span className="h-4 w-px bg-border-medium" />
          <span className="text-text-secondary">
            {stars.toLocaleString()}
          </span>
        </>
      )}
    </a>
  );
}
