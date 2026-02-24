"use client";

import { BlurFade } from "@/components/ui/blur-fade";
import { MagicCard } from "@/components/ui/magic-card";
import { SectionWrapper, SectionHeading, IconBox } from "@/components/common";
import { FEATURES } from "@/lib/constants";

export function Features() {
  return (
    <SectionWrapper id="features">
      <SectionHeading
        title="Built for Production"
        subtitle="Everything you need to ship Lambda logs reliably, with zero configuration overhead."
      />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {FEATURES.map((feature, i) => (
          <BlurFade key={feature.title} delay={0.1 + i * 0.05} inView>
            <MagicCard
              className="h-full rounded-xl border border-border-subtle"
              gradientColor="rgba(255, 153, 0, 0.08)"
              gradientFrom="#FF9900"
              gradientTo="#CC7A00"
              gradientOpacity={0.6}
              gradientSize={250}
            >
              <div className="p-6">
                <IconBox icon={feature.icon} className="mb-4" />
                <h3 className="text-lg font-semibold text-text-primary mb-2">
                  {feature.title}
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed">
                  {feature.description}
                </p>
              </div>
            </MagicCard>
          </BlurFade>
        ))}
      </div>
    </SectionWrapper>
  );
}
