"use client";

import { motion, useInView } from "motion/react";
import { useRef } from "react";
import { BlurFade } from "@/components/ui/blur-fade";
import {
  SectionWrapper,
  SectionHeading,
  AnimatedCounter,
  GlassmorphicCard,
} from "@/components/common";
import {
  PERFORMANCE_METRICS,
  PERFORMANCE_CHART_DATA,
} from "@/lib/constants";

export function Performance() {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInView = useInView(chartRef, { once: true, margin: "-100px" });

  const maxSize = Math.max(
    ...PERFORMANCE_CHART_DATA.filter((d) => d.size > 0).map((d) => d.size)
  );

  return (
    <SectionWrapper id="performance">
      <SectionHeading
        title="Lightweight by Design"
        subtitle="Pure Go. Zero dependencies. Minimal overhead."
      />

      {/* Metrics Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
        {PERFORMANCE_METRICS.map((metric, i) => (
          <BlurFade key={metric.label} delay={0.1 + i * 0.1} inView>
            <GlassmorphicCard className="text-center">
              <div className="font-bold text-3xl md:text-4xl text-brand mb-2">
                <AnimatedCounter
                  value={metric.value}
                  prefix={"prefix" in metric ? metric.prefix : undefined}
                  suffix={metric.suffix}
                />
              </div>
              <div className="text-sm font-medium text-text-primary mb-1">
                {metric.label}
              </div>
              <div className="text-xs text-text-muted">
                {metric.description}
              </div>
            </GlassmorphicCard>
          </BlurFade>
        ))}
      </div>

      {/* Bar Chart */}
      <BlurFade delay={0.3} inView>
        <GlassmorphicCard>
          <h3 className="text-lg font-semibold text-text-primary mb-6">
            Binary Size Comparison
          </h3>
          <div ref={chartRef}>
            {/* Desktop: horizontal bars */}
            <div className="hidden md:block space-y-4">
              {PERFORMANCE_CHART_DATA.map((item) => (
                <div key={item.name} className="flex items-center gap-4">
                  <span className="w-40 shrink-0 text-sm text-text-secondary truncate">
                    {item.name}
                  </span>
                  <div className="flex-1 h-8 bg-surface-lighter rounded-full overflow-hidden">
                    <motion.div
                      className="h-full rounded-full"
                      style={{ backgroundColor: item.color }}
                      initial={{ width: 0 }}
                      animate={
                        chartInView
                          ? {
                              width:
                                item.size === 0
                                  ? "2%"
                                  : `${(item.size / maxSize) * 100}%`,
                            }
                          : { width: 0 }
                      }
                      transition={{
                        duration: 0.8,
                        delay: 0.2,
                        ease: "easeOut",
                      }}
                    />
                  </div>
                  <span className="w-16 text-right text-sm font-mono text-text-muted">
                    {item.size === 0 ? "N/A" : `${item.size} MB`}
                  </span>
                </div>
              ))}
            </div>

            {/* Mobile: vertical bars */}
            <div className="md:hidden">
              <div className="grid grid-cols-4 gap-3">
                {PERFORMANCE_CHART_DATA.map((item) => (
                  <div
                    key={item.name}
                    className="flex flex-col items-center gap-2"
                  >
                    <span className="text-xs font-mono text-text-muted">
                      {item.size === 0 ? "N/A" : `${item.size} MB`}
                    </span>
                    <div className="w-full h-40 bg-surface-lighter rounded-lg overflow-hidden flex items-end">
                      <motion.div
                        className="w-full rounded-lg"
                        style={{ backgroundColor: item.color }}
                        initial={{ height: 0 }}
                        animate={
                          chartInView
                            ? {
                                height:
                                  item.size === 0
                                    ? "5%"
                                    : `${(item.size / maxSize) * 100}%`,
                              }
                            : { height: 0 }
                        }
                        transition={{
                          duration: 0.8,
                          delay: 0.2,
                          ease: "easeOut",
                        }}
                      />
                    </div>
                    <span className="text-xs text-text-secondary text-center leading-tight">
                      {item.name}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </GlassmorphicCard>
      </BlurFade>
    </SectionWrapper>
  );
}
