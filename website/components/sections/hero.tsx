"use client";

import { BlurFade } from "@/components/ui/blur-fade";
import { AnimatedGridPattern } from "@/components/ui/animated-grid-pattern";
import { Particles } from "@/components/ui/particles";
import {
  ShimmerBadge,
  GradientText,
  TerminalBlock,
  GitHubStarButton,
  DownloadButtonGroup,
  GlowEffect,
} from "@/components/common";
import { HERO, RELEASES_URL } from "@/lib/constants";
import { ExternalLink } from "lucide-react";
import type { GitHubRelease } from "@/lib/github";

interface HeroProps {
  release?: GitHubRelease | null;
}

export function Hero({ release }: HeroProps) {
  const badgeContent = release?.tagName
    ? `${release.tagName} — Check what's new`
    : HERO.badgeFallback;

  const badgeHref = release?.htmlUrl ?? undefined;

  return (
    <section
      id="hero"
      className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden px-4"
    >
      {/* Background layers */}
      <AnimatedGridPattern
        className="absolute inset-0 fill-brand/10 stroke-brand/10 opacity-[0.06] [mask-image:radial-gradient(ellipse_at_center,white,transparent_80%)]"
        numSquares={30}
        maxOpacity={0.3}
        duration={5}
        repeatDelay={1}
      />
      <Particles
        className="absolute inset-0"
        quantity={40}
        color="#FF9900"
        size={0.3}
        staticity={60}
        ease={80}
      />
      <GlowEffect
        className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2"
        size="lg"
      />

      {/* Content */}
      <div className="relative z-10 flex max-w-full flex-col items-center px-4 text-center sm:max-w-4xl">
        <BlurFade delay={0} inView>
          <ShimmerBadge href={badgeHref}>{badgeContent}</ShimmerBadge>
        </BlurFade>

        <BlurFade delay={0.1} inView>
          <h1 className="mt-8 font-black text-4xl leading-[1.1] tracking-tight sm:text-5xl md:text-6xl lg:text-7xl">
            <span className="text-text-primary">{HERO.headlineWhite}</span>
            <br />
            <span className="text-text-primary">{HERO.headlineMid}</span>
            <br />
            <GradientText gradient="from-brand via-[#FF6B00] to-[#FF3D00]">
              {HERO.headlineGradient}
            </GradientText>
          </h1>
        </BlurFade>

        <BlurFade delay={0.2} inView>
          <p className="mt-6 max-w-2xl px-4 text-lg text-text-secondary md:text-xl">
            {HERO.subtitle}
          </p>
        </BlurFade>

        <BlurFade delay={0.3} inView>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row sm:flex-wrap">
            <DownloadButtonGroup />
            <GitHubStarButton />
            <a
              href={RELEASES_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 rounded-lg px-5 py-3 text-base font-medium text-text-secondary transition-colors hover:text-text-primary"
            >
              View Releases
              <ExternalLink className="h-4 w-4" />
            </a>
          </div>
        </BlurFade>

        <BlurFade delay={0.4} inView>
          <div className="mt-12 w-full max-w-[calc(100vw-2rem)] overflow-x-auto sm:max-w-2xl">
            <TerminalBlock command={HERO.downloadCommand} />
          </div>
        </BlurFade>
      </div>
    </section>
  );
}
