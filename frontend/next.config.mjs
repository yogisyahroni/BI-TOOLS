// TEMPORARILY DISABLED PWA for build troubleshooting
// import withPWAInit from "@ducanh2912/next-pwa";
// 
// const withPWA = withPWAInit({
//   dest: "public",
//   cacheOnFrontEndNav: true,
//   aggressiveFrontEndNavCaching: true,
//   reloadOnOnline: true,
//   swcMinify: true,
//   disable: process.env.NODE_ENV === "development",
//   workboxOptions: {
//     disableDevLogs: true,
//     runtimeCaching: [
//       {
//         urlPattern: /^https:\/\/fonts\.(?:gstatic|googleapis)\.com\/.*/,
//         handler: "CacheFirst",
//         options: {
//           cacheName: "google-fonts",
//           expiration: { maxEntries: 30, maxAgeSeconds: 31536000 },
//         },
//       },
//       {
//         urlPattern: /^\/api\/dashboards(\/.*)?$/,
//         handler: "StaleWhileRevalidate",
//         options: {
//           cacheName: "api-dashboards",
//           expiration: { maxEntries: 50, maxAgeSeconds: 86400 },
//         },
//       },
//       {
//         urlPattern: /^\/api\/queries\/execute$/,
//         handler: "NetworkFirst",
//         options: {
//           cacheName: "api-queries",
//           expiration: { maxEntries: 200, maxAgeSeconds: 3600 },
//           networkTimeoutSeconds: 10,
//         },
//       },
//       {
//         urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp|ico)$/,
//         handler: "CacheFirst",
//         options: {
//           cacheName: "images",
//           expiration: { maxEntries: 60, maxAgeSeconds: 86400 * 30 },
//         },
//       },
//       {
//         urlPattern: /^\/_next\/static\/.*/,
//         handler: "CacheFirst",
//         options: {
//           cacheName: "next-static",
//           expiration: { maxEntries: 200, maxAgeSeconds: 86400 * 365 },
//         },
//       }
//     ]
//   },
// });

/** @type {import('next').NextConfig} */
const nextConfig = {
  // TEMPORARILY DISABLED STANDALONE OUTPUT for development
  // 'standalone' output is for Docker production builds
  // For development, comment this out to use 'npm start'
  // output: 'standalone',

  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      // pptxgenjs alias removed due to package exports conflict
    };
    config.resolve.fallback = {
      ...config.resolve.fallback,
      fs: false,
      net: false,
      tls: false,
      https: false,
      http: false,
      url: false,
      zlib: false,
      path: false,
      child_process: false,
      'node:https': false,
      'node:http': false,
      'node:fs': false,
      'node:path': false,
      'node:url': false,
      'node:zlib': false,
      'node:stream': false,
      'node:util': false,
    };
    return config;
  },

  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  async rewrites() {
    const backendUrl = process.env.GO_BACKEND_URL || 'http://localhost:8080';
    return [
      // GO Backend APIs
      {
        source: '/api/go/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
      // V1 APIs (scheduler, etc)
      {
        source: '/api/v1/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
      // WebSocket
      {
        source: '/api/v1/ws',
        destination: `${backendUrl}/api/v1/ws`,
      },
      // IMPORTANT: Do NOT proxy /api/auth/* to backend - these must be handled by NextAuth
      // Only proxy non-auth API routes to Go backend
      {
        source: '/api/connections/:path*',
        destination: `${backendUrl}/api/connections/:path*`,
      },
      {
        source: '/api/dashboards/:path*',
        destination: `${backendUrl}/api/dashboards/:path*`,
      },
      {
        source: '/api/queries/:path*',
        destination: `${backendUrl}/api/queries/:path*`,
      },
      {
        source: '/api/notifications/:path*',
        destination: `${backendUrl}/api/notifications/:path*`,
      },
      {
        source: '/api/workspaces/:path*',
        destination: `${backendUrl}/api/workspaces/:path*`,
      },
      {
        source: '/api/governance/:path*',
        destination: `${backendUrl}/api/governance/:path*`,
      },
      {
        source: '/api/pipelines/:path*',
        destination: `${backendUrl}/api/pipelines/:path*`,
      },
      {
        source: '/api/scheduler/:path*',
        destination: `${backendUrl}/api/scheduler/:path*`,
      },
      {
        source: '/api/export/:path*',
        destination: `${backendUrl}/api/export/:path*`,
      },
      {
        source: '/api/rls/:path*',
        destination: `${backendUrl}/api/rls/:path*`,
      },
      {
        source: '/api/semantic/:path*',
        destination: `${backendUrl}/api/semantic/:path*`,
      },
      {
        source: '/api/widgets/:path*',
        destination: `${backendUrl}/api/widgets/:path*`,
      },
      {
        source: '/api/comments/:path*',
        destination: `${backendUrl}/api/comments/:path*`,
      },
      {
        source: '/api/audit/:path*',
        destination: `${backendUrl}/api/audit/:path*`,
      },
      {
        source: '/api/users/:path*',
        destination: `${backendUrl}/api/users/:path*`,
      },
      {
        source: '/api/upload/:path*',
        destination: `${backendUrl}/api/upload/:path*`,
      },
    ]
  },
}

export default nextConfig;
