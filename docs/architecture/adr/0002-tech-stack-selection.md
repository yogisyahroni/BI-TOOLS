# 2. Technology Stack Selection

Date: 2026-02-17

## Status

Accepted

## Context

We need to select a technology stack for the InsightEngine AI platform that ensures high performance, scalability, and developer productivity for an enterprise-grade BI tool.

## Decision

We have selected the following stack:

### Backend: Go (Golang) with Fiber

- **Why Go?** Static typing, high concurrency (goroutines), and compilation to single binary make it ideal for high-throughput data processing and distribution.
- **Why Fiber?** Use of `fasthttp` under the hood provides superior performance compared to `net/http` wrappers like Gin/Echo. Its Express-like syntax lowers the barrier for Node.js developers.

### Frontend: Next.js (App Router) with TypeScript

- **Why Next.js?** Server-Side Rendering (SSR) and React Server Components (RSC) offer better initial load performance and SEO than pure SPAs. The App Router provides a robust routing framework.
- **Why TypeScript?** Critical for maintainability in a complex UI with heavy data manipulation.

### Database: PostgreSQL

- **Why?** Reliable, ACID-compliant, and supports JSONB for semi-structured data (dashboards, configurations). Excellent for complex analytical queries (OLAP) when tuned.

### Caching: Redis

- **Why?** Essential for caching query results and session management to reduce database load and improve user latency.

## Consequences

- **Positive:** High performance, strong type safety across the stack, and a modern, maintainable codebase.
- **Negative:** Go's ecosystem for some niche BI libraries might be smaller than Python's, requiring custom implementation (e.g., Formula Engine).
- **Mitigation:** We implemented a custom Formula Engine and Export Service in pure Go to avoid external dependencies.
