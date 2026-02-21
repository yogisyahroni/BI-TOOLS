import React from "react";
import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Analytics } from "@vercel/analytics/next";

import { Providers } from "./providers";
import "./globals.css";
import "leaflet/dist/leaflet.css";
import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";

const geist = Geist({ subsets: ["latin"] });
const _geistMono = Geist_Mono({ subsets: ["latin"] });

export const viewport: Viewport = {
  themeColor: "#0f172a",
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
};

export const metadata: Metadata = {
  title: "InsightEngine AI - Business Intelligence Platform",
  description:
    "Hybrid Business Intelligence platform combining SQL precision with AI intuition for data analysis",
  generator: "v0.app",
  icons: {
    icon: [
      {
        url: "/icon-light-32x32.png",
        media: "(prefers-color-scheme: light)",
      },
      {
        url: "/icon-dark-32x32.png",
        media: "(prefers-color-scheme: dark)",
      },
      {
        url: "/icon.svg",
        type: "image/svg+xml",
      },
    ],
    apple: "/apple-icon.png",
  },
};

import { ThemeProvider } from "@/components/theme-provider";
import { WebSocketProvider } from "@/components/providers/websocket-provider";
import { ServiceWorkerReset } from "@/components/ServiceWorkerReset";
import { TracingProvider } from "@/components/tracing-provider";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geist.className} font-sans antialiased bg-background text-foreground transition-colors duration-300`}
      >
        <ServiceWorkerReset />
        <Providers>
          <TracingProvider>
            <ThemeProvider
              attribute="class"
              defaultTheme="system"
              enableSystem
              disableTransitionOnChange
            >
              <WebSocketProvider>{children}</WebSocketProvider>
            </ThemeProvider>
          </TracingProvider>
        </Providers>
        {/* Only enable Analytics on Vercel deployments */}
        {process.env.VERCEL && <Analytics />}
      </body>
    </html>
  );
}
