"use client";

import { Terminal, ChevronDown, Cpu } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { ShineBorder } from "@/components/ui/shine-border";
import { cn } from "@/lib/utils";
import { GITHUB_REPO } from "@/lib/constants";
import { Badge } from "@/components/ui/badge";

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
            size="xl"
            className="relative overflow-hidden bg-brand text-black font-semibold hover:bg-brand-light gap-2 rounded-lg"
          >
            <Terminal className="h-5 w-5" />
            Download Layer
            <ChevronDown className="h-4 w-4 ml-1" />
            <ShineBorder
              shineColor={["#FFB84D", "#FF9900", "#CC7A00"]}
              borderWidth={1}
            />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="center"
          className="w-72 p-2 bg-surface-light border-border-medium"
        >
          <DropdownMenuItem asChild>
            <a
              href={`${DOWNLOAD_BASE}/extension-arm64.zip`}
              className="cursor-pointer gap-3 py-3 px-4 text-base rounded-lg"
            >
              <Cpu className="h-5 w-5 text-brand" />
              <div className="flex flex-col">
                <span className="font-medium text-text-primary">ARM64 (Graviton)</span>
              </div>
              <Badge
                variant="secondary"
                className="ml-auto bg-brand-green/10 text-brand-green border-brand-green/20 text-xs font-semibold"
              >
                Recommended
              </Badge>
            </a>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a
              href={`${DOWNLOAD_BASE}/extension-amd64.zip`}
              className="cursor-pointer gap-3 py-3 px-4 text-base rounded-lg"
            >
              <Cpu className="h-5 w-5 text-text-muted" />
              <div className="flex flex-col">
                <span className="font-medium text-text-primary">x86_64 (AMD64)</span>
              </div>
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
