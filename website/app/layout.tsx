import type { Metadata } from "next";
import { calSans, inter, jetbrainsMono } from "@/lib/fonts";
import "./globals.css";

const basePath = process.env.NEXT_PUBLIC_BASE_PATH || "";

export const metadata: Metadata = {
  title: "LambdaWatch - Ship Lambda Logs to Grafana Loki",
  description:
    "Zero-dependency AWS Lambda Extension that ships function logs to Grafana Loki in real-time. Zero code changes. Zero vendor lock-in. Just add the layer.",
  keywords: [
    "AWS Lambda",
    "Grafana Loki",
    "Lambda Extension",
    "observability",
    "logging",
    "serverless",
    "Lambda logs",
    "Loki",
    "Go",
    "Lambda Layer",
  ],
  authors: [{ name: "Mumzworld Tech" }],
  icons: { icon: `${basePath}/icon.svg`, apple: `${basePath}/icon.svg`, shortcut: `${basePath}/icon.svg` },
  openGraph: {
    title: "LambdaWatch - Ship Lambda Logs to Grafana Loki",
    description:
      "Zero-dependency AWS Lambda Extension that ships function logs to Grafana Loki in real-time. Zero code changes. Zero vendor lock-in.",
    url: "https://github.com/mumzworld-tech/LambdaWatch",
    siteName: "LambdaWatch",
    type: "website",
    images: [{ url: `${basePath}/thumbnail.png`, alt: "LambdaWatch" }],
  },
  twitter: {
    card: "summary_large_image",
    title: "LambdaWatch - Ship Lambda Logs to Grafana Loki",
    description: "Zero-dependency AWS Lambda Extension. Zero code changes. Zero vendor lock-in.",
    images: [`${basePath}/thumbnail.png`],
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${calSans.variable} ${inter.variable} ${jetbrainsMono.variable} dark`}
      suppressHydrationWarning
    >
      <body className="min-h-screen bg-black antialiased">{children}</body>
    </html>
  );
}
