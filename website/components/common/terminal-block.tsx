"use client";

import { cn } from "@/lib/utils";
import { ScriptCopyBtn } from "@/components/ui/script-copy-btn";

interface TerminalBlockProps {
  command: string;
  className?: string;
}

export function TerminalBlock({ command, className }: TerminalBlockProps) {
  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-lg border border-border-medium bg-surface p-4",
        className
      )}
    >
      {/* Terminal header dots */}
      <div className="mb-3 flex items-center gap-1.5">
        <div className="h-3 w-3 rounded-full bg-brand-red/70" />
        <div className="h-3 w-3 rounded-full bg-yellow-500/70" />
        <div className="h-3 w-3 rounded-full bg-brand-green/70" />
      </div>
      {/* Command line */}
      <div className="flex items-center gap-2">
        <span className="select-none text-brand font-mono text-sm">$</span>
        <code className="flex-1 overflow-x-auto text-sm font-mono text-text-secondary whitespace-nowrap">
          {command}
        </code>
        <ScriptCopyBtn text={command} showText={false} className="shrink-0" />
      </div>
    </div>
  );
}
