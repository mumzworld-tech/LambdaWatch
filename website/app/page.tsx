import { Navbar } from "@/components/sections/navbar";
import { Hero } from "@/components/sections/hero";
import { Features } from "@/components/sections/features";
import { Architecture } from "@/components/sections/architecture";
import { Performance } from "@/components/sections/performance";
import { Comparison } from "@/components/sections/comparison";
import { FAQ } from "@/components/sections/faq";
import { Footer } from "@/components/sections/footer";
import { SectionDivider } from "@/components/common";
import { getGitHubStars, getLatestRelease } from "@/lib/github";

export default async function Home() {
  const [stars, release] = await Promise.all([
    getGitHubStars(),
    getLatestRelease(),
  ]);

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
