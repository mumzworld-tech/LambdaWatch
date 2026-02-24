import { cn } from "@/lib/utils";

interface GlowEffectProps {
  className?: string;
  size?: "sm" | "md" | "lg";
}

export function GlowEffect({ className, size = "md" }: GlowEffectProps) {
  const sizes = {
    sm: "h-[300px] w-[300px]",
    md: "h-[500px] w-[500px]",
    lg: "h-[700px] w-[700px]",
  };

  return (
    <div
      className={cn(
        "pointer-events-none absolute rounded-full blur-[120px] animate-glow-pulse",
        "bg-brand/15",
        sizes[size],
        className
      )}
      aria-hidden="true"
    />
  );
}
