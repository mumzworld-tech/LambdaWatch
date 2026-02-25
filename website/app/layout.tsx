import type { Metadata } from "next";
import { calSans, inter, jetbrainsMono } from "@/lib/fonts";
import "./globals.css";

const basePath = process.env.NEXT_PUBLIC_BASE_PATH || "";
const title = "LambdaWatch - Fastest Way to Ship Lambda Logs to Grafana Loki";
const description =
  "High-performance AWS Lambda Extension written in Go that automatically captures and ships logs to Grafana Loki. Features automatic batching, gzip compression, request ID tracking, and guaranteed delivery.";

export const metadata: Metadata = {
  metadataBase: new URL(`https://mumzworld-tech.github.io${basePath}`),
  title: title,
  description: description,
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
  icons: { icon: `${basePath}/logo.png`, apple: `${basePath}/logo.png`, shortcut: `${basePath}/logo.png` },
  openGraph: {
    type: "website",
    locale: "en_US",
    title: title,
    description: description,
    url: "https://github.com/mumzworld-tech/LambdaWatch",
    siteName: "LambdaWatch",
    images: [{ url: "/thumbnail.png", alt: "LambdaWatch" }],
  },
  twitter: {
    card: "summary_large_image",
    title: title,
    description: description,
    images: ["/thumbnail.png"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
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
