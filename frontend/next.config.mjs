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
      'pptxgenjs': 'pptxgenjs/dist/pptxgen.min.js',
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
    return [
      {
        source: '/api/go/:path*',
        destination: `${process.env.GO_BACKEND_URL || 'http://localhost:8080'}/api/:path*`,
      },
      // Additional rewrite for WebSocket connections if needed
      {
        source: '/api/v1/ws',
        destination: `${process.env.GO_BACKEND_URL || 'http://localhost:8080'}/api/v1/ws`,
      },
    ]
  },
}

export default nextConfig;
