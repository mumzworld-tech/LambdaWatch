"use client";

import { useScroll, useTransform, type MotionValue } from "motion/react";
import { useRef } from "react";

interface ScrollProgress {
  ref: React.RefObject<HTMLElement | null>;
  progress: MotionValue<number>;
  opacity: MotionValue<number>;
  y: MotionValue<number>;
}

export function useScrollProgress(options: {
  offset?: [string, string];
} = {}): ScrollProgress {
  const { offset = ["start end", "end start"] } = options;
  const ref = useRef<HTMLElement>(null);

  const { scrollYProgress } = useScroll({
    target: ref,
    offset: offset as any,
  });

  const opacity = useTransform(scrollYProgress, [0, 0.2, 0.8, 1], [0, 1, 1, 0]);
  const y = useTransform(scrollYProgress, [0, 0.2], [50, 0]);

  return { ref, progress: scrollYProgress, opacity, y };
}
