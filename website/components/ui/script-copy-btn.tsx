"use client";

import { Check, Copy } from "lucide-react";
import { useState, useCallback } from "react";
import { cn } from "@/lib/utils";

interface ScriptCopyBtnProps {
  text: string;
  className?: string;
  showText?: boolean;
}

export function ScriptCopyBtn({
  text,
  className,
  showText = true,
}: ScriptCopyBtnProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [text]);

  return (
    <div className={cn("flex items-center gap-2", className)}>
      {showText && (
        <code className="flex-1 truncate text-sm font-mono">{text}</code>
      )}
      <button
        onClick={handleCopy}
        className={cn(
          "inline-flex items-center justify-center rounded-md p-2 transition-colors",
          "hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          copied && "text-green-500"
        )}
        aria-label={copied ? "Copied" : "Copy to clipboard"}
      >
        {copied ? (
          <Check className="h-4 w-4" />
        ) : (
          <Copy className="h-4 w-4" />
        )}
      </button>
    </div>
  );
}
