"use client";

import React, { useRef } from "react";
import { motion } from "motion/react";
import { AnimatedBeam } from "@/components/ui/animated-beam";
import { BorderBeam } from "@/components/ui/border-beam";
import { BlurFade } from "@/components/ui/blur-fade";
import {
  SectionWrapper,
  SectionHeading,
  GlassmorphicCard,
} from "@/components/common";
import { ARCHITECTURE_NODES, STATE_MACHINE } from "@/lib/constants";
import { useMousePosition } from "@/hooks/use-mouse-position";
import { cn } from "@/lib/utils";
import {
  FunctionSquare,
  Radio,
  Server,
  Database,
  Send,
  BarChart3,
  ArrowRight,
  type LucideIcon,
} from "lucide-react";

const iconMap: Record<string, LucideIcon> = {
  function: FunctionSquare,
  radio: Radio,
  server: Server,
  database: Database,
  send: Send,
  "bar-chart": BarChart3,
};

function ArchitectureNode({
  label,
  icon,
  nodeRef,
}: {
  label: string;
  icon: string;
  nodeRef: React.RefObject<HTMLDivElement | null>;
}) {
  const Icon = iconMap[icon] ?? FunctionSquare;

  return (
    <div
      ref={nodeRef}
      className={cn(
        "relative z-10 flex flex-col items-center gap-2",
        "rounded-xl border border-border-medium bg-glass backdrop-blur-md",
        "px-3 py-3 sm:px-4 sm:py-4",
        "transition-colors duration-300 hover:border-border-strong hover:bg-glass-light"
      )}
    >
      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-surface-lighter border border-border-subtle">
        <Icon className="h-5 w-5 text-brand" />
      </div>
      <span className="text-[11px] sm:text-xs font-medium text-text-secondary text-center leading-tight whitespace-nowrap">
        {label}
      </span>
    </div>
  );
}

export function Architecture() {
  const containerRef = useRef<HTMLDivElement>(null);
  const tiltRef = useRef<HTMLDivElement>(null);
  const { rotateX, rotateY } = useMousePosition(tiltRef, {
    maxRotation: 3,
    springConfig: { stiffness: 100, damping: 25 },
  });

  // Create individual refs for each architecture node
  const lambdaRef = useRef<HTMLDivElement>(null);
  const telemetryRef = useRef<HTMLDivElement>(null);
  const serverRef = useRef<HTMLDivElement>(null);
  const bufferRef = useRef<HTMLDivElement>(null);
  const clientRef = useRef<HTMLDivElement>(null);
  const lokiRef = useRef<HTMLDivElement>(null);

  const nodeRefs = [
    lambdaRef,
    telemetryRef,
    serverRef,
    bufferRef,
    clientRef,
    lokiRef,
  ];

  return (
    <SectionWrapper id="architecture">
      <SectionHeading
        title="How It Works"
        subtitle="Data flows from Lambda through a high-performance pipeline to Grafana Loki."
      />

      {/* Data Flow Diagram with 3D tilt */}
      <BlurFade delay={0.2} inView>
        <div style={{ perspective: "1000px" }}>
          <motion.div
            ref={tiltRef}
            style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
          >
            <div
              ref={containerRef}
              className="relative rounded-2xl border border-border-medium bg-glass backdrop-blur-md p-6 sm:p-8 md:p-12 overflow-hidden"
            >
              <BorderBeam
                size={80}
                duration={8}
                colorFrom="#FF9900"
                colorTo="#CC7A00"
                borderWidth={1}
              />

              {/* Diagram label */}
              <div className="mb-8 flex items-center gap-2">
                <div className="h-px flex-1 bg-gradient-to-r from-border-medium to-transparent" />
                <span className="text-xs font-mono text-text-muted uppercase tracking-wider">
                  Data Flow Pipeline
                </span>
                <div className="h-px flex-1 bg-gradient-to-l from-border-medium to-transparent" />
              </div>

              {/* Nodes: horizontal on desktop, 2-col grid on mobile */}
              <div className="flex flex-col items-center gap-6 md:flex-row md:flex-nowrap md:justify-between">
                {ARCHITECTURE_NODES.map((node, i) => (
                  <React.Fragment key={node.id}>
                    <ArchitectureNode
                      label={node.label}
                      icon={node.icon}
                      nodeRef={nodeRefs[i]}
                    />
                    {i < ARCHITECTURE_NODES.length - 1 && (
                      <div className="h-6 w-px bg-border-medium md:hidden" />
                    )}
                  </React.Fragment>
                ))}
              </div>

              {/* Animated beams connecting sequential nodes - desktop only */}
              <div className="hidden md:block">
                {nodeRefs.slice(0, -1).map((fromRef, i) => (
                  <AnimatedBeam
                    key={`beam-${i}`}
                    containerRef={containerRef}
                    fromRef={fromRef}
                    toRef={nodeRefs[i + 1]}
                    pathColor="rgba(255, 153, 0, 0.15)"
                    pathWidth={2}
                    pathOpacity={0.3}
                    gradientStartColor="#FF9900"
                    gradientStopColor="#CC7A00"
                    duration={3}
                    delay={i * 0.6}
                    curvature={0}
                  />
                ))}
              </div>
            </div>
          </motion.div>
        </div>
      </BlurFade>

      {/* State Machine */}
      <BlurFade delay={0.3} inView>
        <div className="mt-12">
          {/* State machine label */}
          <div className="mb-6 flex items-center gap-2 justify-center">
            <div className="h-px w-12 bg-gradient-to-r from-transparent to-border-medium" />
            <span className="text-xs font-mono text-text-muted uppercase tracking-wider">
              Extension State Machine
            </span>
            <div className="h-px w-12 bg-gradient-to-l from-transparent to-border-medium" />
          </div>

          {/* State pills with transitions */}
          <div className="flex flex-wrap items-center justify-center gap-3 sm:gap-4">
            {STATE_MACHINE.map((item, i) => (
              <div key={item.state} className="flex items-center gap-3 sm:gap-4">
                <GlassmorphicCard
                  className={cn(
                    "!rounded-full !px-5 !py-2.5 !border flex items-center gap-3",
                    item.state === "ACTIVE" &&
                      "!border-brand-green/30 !bg-brand-green/5",
                    item.state === "FLUSHING" &&
                      "!border-brand/30 !bg-brand/5",
                    item.state === "IDLE" &&
                      "!border-text-muted/20 !bg-text-muted/5"
                  )}
                >
                  {/* Status dot */}
                  <span
                    className={cn(
                      "inline-block h-2 w-2 rounded-full",
                      item.state === "ACTIVE" && "bg-brand-green animate-pulse",
                      item.state === "FLUSHING" && "bg-brand animate-pulse",
                      item.state === "IDLE" && "bg-text-muted"
                    )}
                  />
                  <span
                    className={cn(
                      "text-sm font-mono font-semibold",
                      item.state === "ACTIVE" && "text-brand-green",
                      item.state === "FLUSHING" && "text-brand",
                      item.state === "IDLE" && "text-text-muted"
                    )}
                  >
                    {item.state}
                  </span>
                  <span
                    className={cn(
                      "text-xs font-mono opacity-60",
                      item.state === "ACTIVE" && "text-brand-green",
                      item.state === "FLUSHING" && "text-brand",
                      item.state === "IDLE" && "text-text-muted"
                    )}
                  >
                    {item.interval}
                  </span>
                </GlassmorphicCard>

                {/* Arrow between states */}
                {i < STATE_MACHINE.length - 1 && (
                  <div className="flex items-center">
                    <div className="hidden sm:flex items-center gap-1">
                      <div className="h-px w-6 bg-gradient-to-r from-text-muted/40 to-text-muted/20" />
                      <ArrowRight className="h-3.5 w-3.5 text-text-muted/60" />
                    </div>
                    <ArrowRight className="h-3.5 w-3.5 text-text-muted/60 sm:hidden" />
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Transition labels */}
          <div className="mt-4 flex flex-wrap items-center justify-center gap-6 text-[11px] font-mono text-text-muted/50">
            <span>platform.runtimeDone</span>
            <span>critical flush complete</span>
            <span>next INVOKE</span>
          </div>
        </div>
      </BlurFade>
    </SectionWrapper>
  );
}
