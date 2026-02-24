"use client";

import { useMotionValue, useSpring, type MotionValue } from "motion/react";
import { useCallback, useEffect } from "react";

interface MousePosition {
  x: MotionValue<number>;
  y: MotionValue<number>;
  rotateX: MotionValue<number>;
  rotateY: MotionValue<number>;
}

export function useMousePosition(
  ref: React.RefObject<HTMLElement | null>,
  options: { maxRotation?: number; springConfig?: { stiffness?: number; damping?: number } } = {}
): MousePosition {
  const { maxRotation = 5, springConfig = { stiffness: 150, damping: 20 } } = options;

  const x = useMotionValue(0);
  const y = useMotionValue(0);

  const rotateX = useSpring(0, springConfig);
  const rotateY = useSpring(0, springConfig);

  const handleMouseMove = useCallback(
    (event: MouseEvent) => {
      const element = ref.current;
      if (!element) return;

      const rect = element.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;

      const deltaX = (event.clientX - centerX) / (rect.width / 2);
      const deltaY = (event.clientY - centerY) / (rect.height / 2);

      x.set(deltaX);
      y.set(deltaY);
      rotateX.set(-deltaY * maxRotation);
      rotateY.set(deltaX * maxRotation);
    },
    [ref, x, y, rotateX, rotateY, maxRotation]
  );

  const handleMouseLeave = useCallback(() => {
    x.set(0);
    y.set(0);
    rotateX.set(0);
    rotateY.set(0);
  }, [x, y, rotateX, rotateY]);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    element.addEventListener("mousemove", handleMouseMove);
    element.addEventListener("mouseleave", handleMouseLeave);

    return () => {
      element.removeEventListener("mousemove", handleMouseMove);
      element.removeEventListener("mouseleave", handleMouseLeave);
    };
  }, [ref, handleMouseMove, handleMouseLeave]);

  return { x, y, rotateX, rotateY };
}
