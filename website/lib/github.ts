import { GITHUB_REPO } from "@/lib/constants";

export async function getGitHubStars(): Promise<number | null> {
  try {
    const response = await fetch(
      `https://api.github.com/repos/${GITHUB_REPO}`,
      { next: { revalidate: 3600 } } // ISR: revalidate every hour
    );
    if (!response.ok) return null;
    const data = await response.json();
    return data.stargazers_count ?? null;
  } catch {
    return null;
  }
}

export interface GitHubRelease {
  tagName: string;
  name: string;
  publishedAt: string;
  htmlUrl: string;
}

export async function getLatestRelease(): Promise<GitHubRelease | null> {
  try {
    const response = await fetch(
      `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`,
      { next: { revalidate: 3600 } } // ISR: revalidate every hour
    );
    if (!response.ok) return null;
    const data = await response.json();
    return {
      tagName: data.tag_name ?? null,
      name: data.name ?? null,
      publishedAt: data.published_at ?? null,
      htmlUrl: data.html_url ?? null,
    };
  } catch {
    return null;
  }
}
