# ADR-003: NextAuth.js for Authentication with Go JWT Backend

## Status

Accepted

## Date

2026-01-27

## Context

InsightEngine requires multi-provider authentication (email/password, Google OAuth, GitHub OAuth) with session management. Two competing approaches exist:

1. **Go-native auth**: Go backend handles all auth flows (password hashing, OAuth redirects, JWT issuance, session management). Frontend makes raw API calls.
2. **NextAuth.js + Go JWT**: NextAuth.js handles OAuth flows and session cookies on the frontend. Go backend validates JWTs and issues its own tokens for API access.

The challenge is that OAuth flows require server-side redirects, PKCE, and cookie management. Implementing this entirely in Go while the frontend runs on a separate origin (port 3000 vs 8080) creates CORS and cookie-sharing complexity.

## Decision

Use a **hybrid approach**:

- **NextAuth.js** manages OAuth provider integration (Google, GitHub), session cookies (`HttpOnly, Secure, SameSite=Lax`), and CSRF protection on the frontend server.
- **Go backend** independently validates JWT tokens, handles email/password registration and login, and issues its own JWTs with `bcrypt` password hashing.
- The frontend's `api-client.ts` attaches the Go-issued JWT to all `/api/go/*` requests via `Authorization: Bearer` header.
- NextAuth session callback syncs the Go JWT into the NextAuth session object.

## Consequences

### Positive

- **Best-of-both-worlds**: NextAuth.js's battle-tested OAuth flows + Go's high-performance JWT validation.
- **SSR-compatible**: NextAuth session is available in Server Components for auth-gated rendering.
- **Provider flexibility**: Adding new OAuth providers (Azure AD, Okta) requires only NextAuth config changes.
- **Security**: Separate token lifetimes — NextAuth session (30 days) and Go JWT (24 hours) provide defense-in-depth.

### Negative

- **Dual token management**: Two token systems (NextAuth session token + Go JWT) must be kept in sync. Session callback handles this, but adds complexity.
- **Debugging difficulty**: Auth failures can originate from either NextAuth or Go, requiring investigation in both systems.
- **Token refresh coordination**: Go JWT expiry requires re-authentication through the frontend session refresh mechanism.
