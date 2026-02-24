import { cn } from "@/lib/utils";

interface GlassmorphicCardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}

export function GlassmorphicCard({
  children,
  className,
  hover = false,
}: GlassmorphicCardProps) {
  return (
    <div
      className={cn(
        "rounded-xl border border-border-medium bg-glass backdrop-blur-md p-6",
        hover &&
          "transition-all duration-300 hover:border-border-strong hover:bg-glass-light hover:shadow-[0_0_30px_rgba(255,153,0,0.05)]",
        className
      )}
    >
      {children}
    </div>
  );
}
