export async function getGitHubStars(): Promise<number | null> {
  try {
    const response = await fetch(
      "https://api.github.com/repos/mumzworld-tech/lambdawatch",
      { next: { revalidate: 3600 } } // ISR: revalidate every hour
    );
    if (!response.ok) return null;
    const data = await response.json();
    return data.stargazers_count ?? null;
  } catch {
    return null;
  }
}
