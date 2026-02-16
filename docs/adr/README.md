# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records for the InsightEngine AI platform.

## What is an ADR?

An Architecture Decision Record (ADR) documents a significant architectural or technical decision made during the development of this project. Each record captures the context, the decision itself, and the consequences of that decision.

## ADR Index

| # | Title | Status | Date |
|---|-------|--------|------|
| 001 | [Go + Next.js Dual-Runtime Architecture](0001-go-nextjs-dual-runtime.md) | Accepted | 2026-01-25 |
| 002 | [Pure Go Export Generation (No External Binaries)](0002-pure-go-export-generation.md) | Accepted | 2026-02-16 |
| 003 | [NextAuth.js for Authentication with Go JWT Backend](0003-nextauth-go-jwt-hybrid.md) | Accepted | 2026-01-27 |
| 004 | [WebSocket Hub for Real-Time Collaboration](0004-websocket-hub-realtime.md) | Accepted | 2026-02-01 |
| 005 | [Monorepo with Separate Build Pipelines](0005-monorepo-separate-builds.md) | Accepted | 2026-01-25 |

## Format

Each ADR follows this template:

```
# ADR-NNN: Title

## Status
Accepted | Superseded | Deprecated

## Context
What is the issue that we're seeing that motivates this decision?

## Decision
What is the change we're proposing?

## Consequences
What are the positive and negative outcomes of this decision?
```

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Michael Nygard's original post](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
