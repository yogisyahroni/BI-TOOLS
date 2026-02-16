# ADR-005: Monorepo with Separate Build Pipelines

## Status

Accepted

## Date

2026-01-25

## Context

The InsightEngine codebase consists of two primary components:

- **`backend/`**: Go application (Fiber, GORM, WebSocket hub)
- **`frontend/`**: Next.js application (React, TypeScript, shadcn/ui)

The options for repository structure were:

1. **Polyrepo**: Separate Git repositories for backend and frontend. Independent versioning and CI/CD.
2. **Monorepo (unified build)**: Single repository with a unified build system (Bazel, Nx, Turborepo).
3. **Monorepo (separate builds)**: Single repository, but each component has its own build toolchain (`go build` vs `npm run build`).

## Decision

Use **Option 3: Monorepo with separate build pipelines**.

```
insight-engine-ai-ui/
├── backend/           # Go module (go.mod)
│   ├── handlers/
│   ├── services/
│   ├── models/
│   └── main.go
├── frontend/          # Node.js package (package.json)
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── next.config.mjs
├── docs/              # Shared documentation
└── .github/workflows/ # CI/CD for both
```

- Each component builds independently: `go build ./...` and `npm run build`.
- Shared documentation lives at the repository root (`docs/`, `ROADMAP_100_PERCENT_PARITY.md`).
- CI/CD workflow triggers component-specific builds based on changed paths.

## Consequences

### Positive

- **Atomic changes**: Cross-cutting changes (API contract changes affecting both frontend and backend) are a single commit/PR.
- **Shared docs**: Architecture docs, ADRs, and runbooks live alongside code. No cross-repo linking.
- **Simplified environment**: One `git clone` gives developers the full system.
- **PR context**: Reviewers see the full impact of a change across both components.

### Negative

- **Build isolation**: A broken frontend shouldn't block backend deployments. Requires path-based CI triggers.
- **Toolchain requirements**: Developers need both Go and Node.js toolchains installed locally.
- **No shared build cache**: Go and Node.js build caches are independent. No cross-language build graph optimization.
