"use client";

import { cn } from "@/lib/utils";

interface SectionWrapperProps {
  id?: string;
  children: React.ReactNode;
  className?: string;
  fullWidth?: boolean;
}

export function SectionWrapper({
  id,
  children,
  className,
  fullWidth = false,
}: SectionWrapperProps) {
  return (
    <section
      id={id}
      className={cn(
        "relative py-24 md:py-32",
        !fullWidth && "mx-auto max-w-7xl px-4 sm:px-6 lg:px-8",
        className
      )}
    >
      {children}
    </section>
  );
}
