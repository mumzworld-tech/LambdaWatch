import { getGitHubStars, getLatestRelease } from "@/lib/github";
import { GitHubDataWrapper } from "@/components/github-data-wrapper";

export default async function Home() {
  const [stars, release] = await Promise.all([
    getGitHubStars(),
    getLatestRelease(),
  ]);

  return (
    <GitHubDataWrapper initialStars={stars} initialRelease={release} />
  );
}
