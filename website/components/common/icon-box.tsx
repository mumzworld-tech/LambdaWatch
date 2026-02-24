import { cn } from "@/lib/utils";
import { type LucideIcon } from "lucide-react";

interface IconBoxProps {
  icon: LucideIcon;
  className?: string;
  size?: "sm" | "md" | "lg";
}

export function IconBox({ icon: Icon, className, size = "md" }: IconBoxProps) {
  const sizes = {
    sm: "h-10 w-10 [&_svg]:h-5 [&_svg]:w-5",
    md: "h-12 w-12 [&_svg]:h-6 [&_svg]:w-6",
    lg: "h-14 w-14 [&_svg]:h-7 [&_svg]:w-7",
  };

  return (
    <div
      className={cn(
        "inline-flex items-center justify-center rounded-xl",
        "bg-gradient-to-br from-surface-lighter to-surface",
        "border border-border-medium",
        "shadow-[0_2px_8px_rgba(0,0,0,0.3),inset_0_1px_0_rgba(255,255,255,0.05)]",
        "text-brand",
        sizes[size],
        className
      )}
    >
      <Icon />
    </div>
  );
}
