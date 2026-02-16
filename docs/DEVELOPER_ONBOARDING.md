# InsightEngine AI — Developer Onboarding Guide

Welcome to the InsightEngine AI platform. This guide will get you from zero to running in under 30 minutes.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Repository Setup](#repository-setup)
3. [Backend Setup](#backend-setup)
4. [Frontend Setup](#frontend-setup)
5. [Architecture Overview](#architecture-overview)
6. [Code Structure](#code-structure)
7. [Development Workflow](#development-workflow)
8. [Key Conventions](#key-conventions)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 18+ LTS | [nodejs.org](https://nodejs.org/) |
| PostgreSQL | 15+ | [postgresql.org](https://www.postgresql.org/download/) |
| Git | 2.40+ | [git-scm.com](https://git-scm.com/) |
| Redis | 7+ (optional) | [redis.io](https://redis.io/) |

## Repository Setup

```bash
git clone https://github.com/your-org/insight-engine-ai-ui.git
cd insight-engine-ai-ui
```

### Project Structure

```
insight-engine-ai-ui/
├── backend/                # Go (Fiber) API server
│   ├── bootstrap/          # App initialization, DI container
│   ├── config/             # Configuration loading
│   ├── controllers/        # Business logic controllers
│   ├── database/           # GORM connection + migrations
│   ├── handlers/           # HTTP request handlers (Fiber)
│   ├── middleware/          # HTTP middleware (auth, CORS, security)
│   ├── models/             # GORM models (DB entities)
│   ├── routes/             # Route registration
│   ├── services/           # Business logic services
│   └── main.go             # Entrypoint
├── frontend/               # Next.js 14+ (App Router)
│   ├── app/                # Next.js app router pages
│   ├── components/         # React components (by feature)
│   ├── lib/                # Utilities, API clients, hooks
│   ├── public/             # Static assets
│   └── next.config.mjs     # Next.js configuration
├── docs/                   # Documentation
│   ├── adr/                # Architecture Decision Records
│   └── runbooks/           # Operational runbooks
└── .github/workflows/      # CI/CD pipelines
```

---

## Backend Setup

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Configure Environment

Create `backend/.env`:

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/insightengine?sslmode=disable
JWT_SECRET=your-development-secret-min-32-chars
PORT=8080
GEMINI_API_KEY=your-gemini-api-key
ENVIRONMENT=development
```

### 3. Create Database

```bash
createdb insightengine
# Or using psql:
# psql -c "CREATE DATABASE insightengine;"
```

### 4. Run

```bash
go run main.go
# Server starts on http://localhost:8080
```

### 5. Verify

```bash
curl http://localhost:8080/api/health/live
# Expected: {"status": "alive"}
```

---

## Frontend Setup

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Configure Environment

Create `frontend/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_GO_API_URL=http://localhost:8080
NEXTAUTH_SECRET=your-nextauth-secret
NEXTAUTH_URL=http://localhost:3000
```

### 3. Run

```bash
npm run dev
# App starts on http://localhost:3000
```

### 4. Verify

Open `http://localhost:3000` in your browser. You should see the login page.

---

## Architecture Overview

```
┌────────────────┐     ┌────────────────┐
│   Browser       │     │   Next.js       │
│   (React SPA)   │────▶│   Server        │
│                 │     │   (Port 3000)   │
└────────────────┘     └───────┬─────────┘
                               │ /api/go/* proxy
                               ▼
                       ┌────────────────┐
                       │   Go Backend    │
                       │   (Fiber)       │
                       │   (Port 8080)   │
                       └───────┬─────────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
              ┌──────────┐ ┌──────┐ ┌──────────┐
              │PostgreSQL│ │Redis │ │ Gemini   │
              │  (GORM)  │ │(Cache)│ │  AI API  │
              └──────────┘ └──────┘ └──────────┘
```

**Key patterns:**

- Frontend proxies `/api/go/*` requests to Go backend (configured in `next.config.mjs`)
- WebSocket connection: `ws://localhost:8080/ws`
- Authentication: NextAuth.js (frontend) + Go JWT (backend)
- State management: TanStack Query (server) + Zustand (client)

---

## Code Structure

### Backend Layers

```
Handler (HTTP) → Service (Business Logic) → Repository/Model (Database)
```

- **Handlers** parse requests, validate input, call services, return responses.
- **Services** contain business logic, orchestrate operations, call other services.
- **Models** define GORM entities and database schemas.

### Frontend Layers

```
Page (App Router) → Component (UI) → Hook (Logic) → API Client (Network)
```

- **Pages** (`app/`) define routes and layouts.
- **Components** (`components/`) are organized by feature domain.
- **API Clients** (`lib/api/`) wrap fetch calls to the backend.

---

## Development Workflow

### Making Changes

1. **Backend**: Edit Go files → the server must be restarted manually.
2. **Frontend**: Edit TSX/TS files → Next.js hot-reloads automatically.

### Running Tests

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npx vitest run
```

### Building for Production

```bash
# Backend
cd backend && go build -o insightengine ./main.go

# Frontend
cd frontend && npm run build
```

### Pre-commit Hooks

Husky + lint-staged is configured. On each commit:

- Go files: `golangci-lint run`
- TS/TSX files: `eslint --fix` + `prettier --write`
- Test files: `vitest related`

---

## Key Conventions

| Area | Convention |
|------|-----------|
| **Git branches** | `feat/`, `fix/`, `chore/` + Conventional Commits |
| **API routes** | `kebab-case`: `/api/dashboard-cards`, not `/api/dashboardCards` |
| **Go files** | `snake_case.go` |
| **Go structs** | `PascalCase` |
| **TS files** | `kebab-case.ts` or `kebab-case.tsx` |
| **React components** | `PascalCase` directories + files |
| **DB columns** | `snake_case` (GORM convention) |
| **Error handling** | Go: explicit `if err != nil`. TS: try/catch with typed errors. |
| **Soft deletes** | Use `gorm.DeletedAt` (GORM built-in) |

---

## Troubleshooting

### Port Already in Use

```bash
# Windows
netstat -ano | findstr :8080
taskkill /F /PID <pid>

# macOS/Linux
lsof -i :8080
kill -9 <pid>
```

### Database Connection Refused

- Check PostgreSQL is running: `pg_isready`
- Verify `DATABASE_URL` in `.env`
- Check firewall rules

### Frontend Proxy Errors

- Ensure Go backend is running on port 8080
- Check `next.config.mjs` rewrites configuration
- Look for CORS errors in browser console

### Go Module Issues

```bash
cd backend
go mod tidy
go mod verify
```

### Node Module Issues

```bash
cd frontend
rm -rf node_modules .next
npm install
npm run dev
```

---

## Further Reading

- [Architecture Decision Records](adr/README.md)
- [Operational Runbooks](runbooks/README.md)
- [API Documentation (OpenAPI)](../backend/docs/openapi.yaml)
- [Security Hardening](SECURITY_HARDENING.md)
