"use client";

import { useEffect, useState } from "react";
import { GITHUB_REPO } from "@/lib/constants";
import type { GitHubRelease } from "@/lib/github";
import { Navbar } from "@/components/sections/navbar";
import { Hero } from "@/components/sections/hero";
import { Features } from "@/components/sections/features";
import { Architecture } from "@/components/sections/architecture";
import { Performance } from "@/components/sections/performance";
import { Comparison } from "@/components/sections/comparison";
import { FAQ } from "@/components/sections/faq";
import { Footer } from "@/components/sections/footer";
import { SectionDivider } from "@/components/common";

interface GitHubDataWrapperProps {
  initialStars: number | null;
  initialRelease: GitHubRelease | null;
}

export function GitHubDataWrapper({
  initialStars,
  initialRelease,
}: GitHubDataWrapperProps) {
  const [stars, setStars] = useState(initialStars);
  const [release, setRelease] = useState(initialRelease);

  useEffect(() => {
    async function fetchGitHubData() {
      try {
        const [repoRes, releaseRes] = await Promise.all([
          fetch(`https://api.github.com/repos/${GITHUB_REPO}`),
          fetch(
            `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`
          ),
        ]);

        if (repoRes.ok) {
          const repoData = await repoRes.json();
          if (repoData.stargazers_count != null) {
            setStars(repoData.stargazers_count);
          }
        }

        if (releaseRes.ok) {
          const releaseData = await releaseRes.json();
          setRelease({
            tagName: releaseData.tag_name ?? null,
            name: releaseData.name ?? null,
            publishedAt: releaseData.published_at ?? null,
            htmlUrl: releaseData.html_url ?? null,
          });
        }
      } catch {
        // Keep initial server-side values on error
      }
    }

    fetchGitHubData();
  }, []);

  return (
    <>
      <Navbar stars={stars} />
      <main>
        <Hero release={release} />
        <SectionDivider />
        <Features />
        <SectionDivider />
        <Architecture />
        <SectionDivider />
        <Performance />
        <SectionDivider />
        <Comparison />
        <SectionDivider />
        <FAQ />
      </main>
      <Footer stars={stars} />
    </>
  );
}
