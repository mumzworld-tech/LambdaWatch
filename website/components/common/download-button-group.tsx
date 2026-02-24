"use client";

import { Download, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ShineBorder } from "@/components/ui/shine-border";
import { cn } from "@/lib/utils";
import { GITHUB_REPO } from "@/lib/constants";

interface DownloadButtonGroupProps {
  className?: string;
}

const DOWNLOAD_BASE = `https://github.com/${GITHUB_REPO}/releases/latest/download`;

export function DownloadButtonGroup({ className }: DownloadButtonGroupProps) {
  return (
    <div className={cn("relative inline-flex", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            size="lg"
            className="relative overflow-hidden bg-brand text-black font-semibold hover:bg-brand-light gap-2 px-6 rounded-lg"
          >
            <Download className="h-4 w-4" />
            Get Started
            <ChevronDown className="h-4 w-4 ml-1" />
            <ShineBorder
              shineColor={["#FFB84D", "#FF9900", "#CC7A00"]}
              borderWidth={1}
            />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="center"
          className="bg-surface-light border-border-medium"
        >
          <DropdownMenuItem asChild>
            <a
              href={`${DOWNLOAD_BASE}/extension-arm64.zip`}
              className="cursor-pointer gap-2"
            >
              <Download className="h-4 w-4" />
              ARM64 (Graviton)
              <span className="ml-auto text-xs text-text-muted">
                Recommended
              </span>
            </a>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a
              href={`${DOWNLOAD_BASE}/extension-amd64.zip`}
              className="cursor-pointer gap-2"
            >
              <Download className="h-4 w-4" />
              x86_64 (AMD64)
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
