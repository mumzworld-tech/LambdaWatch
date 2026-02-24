"use client";

import { cn } from "@/lib/utils";
import { BlurFade } from "@/components/ui/blur-fade";

interface SectionHeadingProps {
  title: string;
  subtitle?: string;
  className?: string;
  align?: "left" | "center";
}

export function SectionHeading({
  title,
  subtitle,
  className,
  align = "center",
}: SectionHeadingProps) {
  return (
    <div
      className={cn("mb-16", align === "center" && "text-center", className)}
    >
      <BlurFade delay={0.1} inView>
        <h2 className="font-bold text-3xl sm:text-4xl md:text-5xl tracking-tight text-text-primary">
          {title}
        </h2>
      </BlurFade>
      {subtitle && (
        <BlurFade delay={0.2} inView>
          <p className="mt-4 text-lg text-text-secondary max-w-2xl mx-auto">
            {subtitle}
          </p>
        </BlurFade>
      )}
    </div>
  );
}
