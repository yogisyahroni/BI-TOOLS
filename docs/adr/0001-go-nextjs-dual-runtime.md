# ADR-001: Go + Next.js Dual-Runtime Architecture

## Status

Accepted

## Date

2026-01-25

## Context

InsightEngine AI requires both a high-performance backend capable of handling concurrent database connections, query execution, and real-time WebSocket communication, as well as a modern, dynamic frontend with SSR/SSG capabilities, complex state management, and rich interactivity (drag-and-drop query builders, real-time dashboards, chart rendering).

The options considered were:

1. **Monolithic Next.js** — API Routes + frontend in a single Node.js runtime.
2. **Go backend + React SPA** — Compiled Go backend with a separate client-only React app.
3. **Go backend + Next.js frontend** — Compiled Go backend with Next.js providing SSR, routing, and React ecosystem.

## Decision

We chose **Option 3: Go (Fiber) backend + Next.js (App Router) frontend**.

- **Backend**: Go with Fiber framework, GORM ORM, running on port 8080.
- **Frontend**: Next.js 14+ with App Router, running on port 3000, proxying `/api/go/*` to the Go backend.
- **Communication**: REST API over HTTP, WebSocket for real-time updates.

## Consequences

### Positive

- **Performance**: Go handles compute-heavy query execution, concurrent database connections, and export generation with lower memory than Node.js.
- **Type safety end-to-end**: Go's strict typing + TypeScript frontend prevents a large class of runtime errors.
- **SSR/SEO**: Next.js provides server-side rendering for dashboards and embeds.
- **Ecosystem**: Access to Go's database driver ecosystem (pgx, mysql, mssql) and React's UI component ecosystem (shadcn/ui, recharts, dnd-kit).

### Negative

- **Operational complexity**: Two processes to manage, two build pipelines, two deployment artifacts.
- **API contract drift**: Without shared type generation, frontend and backend types can diverge. Mitigated by OpenAPI spec (TASK-172).
- **Development setup**: Developers must have both Go and Node.js toolchains installed.
