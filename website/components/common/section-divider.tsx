import { cn } from "@/lib/utils";

interface SectionDividerProps {
  className?: string;
}

export function SectionDivider({ className }: SectionDividerProps) {
  return (
    <div
      className={cn(
        "h-px w-full bg-gradient-to-r from-transparent via-brand/20 to-transparent",
        className
      )}
    />
  );
}
