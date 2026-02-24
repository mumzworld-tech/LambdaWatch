import { cn } from "@/lib/utils";

interface GradientTextProps {
  children: React.ReactNode;
  className?: string;
  from?: string;
  via?: string;
  to?: string;
  gradient?: string;
}

export function GradientText({
  children,
  className,
  from = "from-white",
  via,
  to = "to-brand",
  gradient,
}: GradientTextProps) {
  return (
    <span
      className={cn(
        "bg-gradient-to-r bg-clip-text text-transparent",
        gradient ? gradient : [from, via, to].filter(Boolean).join(" "),
        className
      )}
    >
      {children}
    </span>
  );
}
