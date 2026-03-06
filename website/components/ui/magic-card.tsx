"use client"

import React, { useCallback, useEffect, useRef } from "react"
import { motion, useMotionTemplate, useMotionValue, useSpring } from "motion/react"

import { cn } from "@/lib/utils"

interface MagicCardProps {
  children?: React.ReactNode
  className?: string
  gradientSize?: number
  gradientColor?: string
  gradientOpacity?: number
  gradientFrom?: string
  gradientTo?: string
  tilt?: boolean
  tiltAmount?: number
}

export function MagicCard({
  children,
  className,
  gradientSize = 200,
  gradientColor = "#262626",
  gradientOpacity = 0.8,
  gradientFrom = "#9E7AFF",
  gradientTo = "#FE8BBB",
  tilt = false,
  tiltAmount = 5,
}: MagicCardProps) {
  const mouseX = useMotionValue(-gradientSize)
  const mouseY = useMotionValue(-gradientSize)
  const cardRef = useRef<HTMLDivElement>(null)

  const rawRotateX = useMotionValue(0)
  const rawRotateY = useMotionValue(0)
  const rotateX = useSpring(rawRotateX, { stiffness: 300, damping: 30 })
  const rotateY = useSpring(rawRotateY, { stiffness: 300, damping: 30 })

  const reset = useCallback(() => {
    mouseX.set(-gradientSize)
    mouseY.set(-gradientSize)
    if (tilt) {
      rawRotateX.set(0)
      rawRotateY.set(0)
    }
  }, [gradientSize, mouseX, mouseY, tilt, rawRotateX, rawRotateY])

  const handlePointerMove = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      const rect = e.currentTarget.getBoundingClientRect()
      const x = e.clientX - rect.left
      const y = e.clientY - rect.top
      mouseX.set(x)
      mouseY.set(y)

      if (tilt) {
        const centerX = rect.width / 2
        const centerY = rect.height / 2
        const normalizedX = (x - centerX) / centerX
        const normalizedY = (y - centerY) / centerY
        rawRotateX.set(-normalizedY * tiltAmount)
        rawRotateY.set(normalizedX * tiltAmount)
      }
    },
    [mouseX, mouseY, tilt, tiltAmount, rawRotateX, rawRotateY]
  )

  useEffect(() => {
    reset()
  }, [reset])

  useEffect(() => {
    const handleGlobalPointerOut = (e: PointerEvent) => {
      if (!e.relatedTarget) {
        reset()
      }
    }

    const handleVisibility = () => {
      if (document.visibilityState !== "visible") {
        reset()
      }
    }

    window.addEventListener("pointerout", handleGlobalPointerOut)
    window.addEventListener("blur", reset)
    document.addEventListener("visibilitychange", handleVisibility)

    return () => {
      window.removeEventListener("pointerout", handleGlobalPointerOut)
      window.removeEventListener("blur", reset)
      document.removeEventListener("visibilitychange", handleVisibility)
    }
  }, [reset])

  const cardContent = (
    <>
      <motion.div
        className="bg-border pointer-events-none absolute inset-0 rounded-[inherit] duration-300 group-hover:opacity-100"
        style={{
          background: useMotionTemplate`
          radial-gradient(${gradientSize}px circle at ${mouseX}px ${mouseY}px,
          ${gradientFrom},
          ${gradientTo},
          var(--border) 100%
          )
          `,
        }}
      />
      <div className="bg-[rgba(10,10,10,0.85)] absolute inset-px rounded-[inherit]" />
      <motion.div
        className="pointer-events-none absolute inset-px rounded-[inherit] opacity-0 transition-opacity duration-300 group-hover:opacity-100"
        style={{
          background: useMotionTemplate`
            radial-gradient(${gradientSize}px circle at ${mouseX}px ${mouseY}px, ${gradientColor}, transparent 100%)
          `,
          opacity: gradientOpacity,
        }}
      />
      <div
        className="relative h-full"
        style={{
          transform: tilt ? "translateZ(30px)" : "none",
          transition: "transform 0.3s ease"
        }}
      >
        {children}
      </div>
    </>
  )

  if (tilt) {
    return (
      <div
        ref={cardRef}
        className={cn("group relative rounded-[inherit] block", className)}
        onPointerMove={handlePointerMove}
        onPointerLeave={reset}
        style={{ perspective: 1000 }}
      >
        <motion.div
          className="w-full h-full rounded-[inherit] bg-transparent"
          style={{
            rotateX,
            rotateY,
            transformStyle: "preserve-3d",
          }}
        >
          {cardContent}
        </motion.div>
      </div>
    )
  }

  return (
    <div
      className={cn("group relative rounded-[inherit]", className)}
      onPointerMove={handlePointerMove}
      onPointerLeave={reset}
    >
      {cardContent}
    </div>
  )
}
