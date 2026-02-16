# Runbook 01: Deployment

## Overview

This runbook covers deploying the InsightEngine AI platform (Go backend + Next.js frontend).

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go | 1.22+ |
| Node.js | 18+ (LTS) |
| Docker | 24+ |
| PostgreSQL | 15+ |

## Environment Variables

Ensure these are set before deployment:

```bash
# Backend
DATABASE_URL=postgresql://user:pass@host:5432/insightengine?sslmode=require
JWT_SECRET=<min-32-char-random>
REDIS_URL=redis://host:6379
GEMINI_API_KEY=<key>
PORT=8080

# Frontend
NEXT_PUBLIC_API_URL=https://api.insightengine.io
NEXTAUTH_SECRET=<min-32-char-random>
NEXTAUTH_URL=https://app.insightengine.io
GOOGLE_CLIENT_ID=<oauth-client-id>
GOOGLE_CLIENT_SECRET=<oauth-client-secret>
```

---

## Backend Deployment

### 1. Build

```bash
cd backend/
go build -ldflags="-s -w" -o insightengine ./main.go
```

### 2. Run Database Migrations

GORM AutoMigrate runs on startup. For manual migrations:

```bash
./insightengine --migrate-only
```

### 3. Health Check

```bash
curl -f http://localhost:8080/api/health/ready
# Expected: 200 OK, {"status": "healthy", "checks": {...}}
```

### 4. Docker Build

```bash
docker build -t insightengine-backend:latest -f Dockerfile.backend .
docker run -d --name backend \
  --env-file .env.production \
  -p 8080:8080 \
  insightengine-backend:latest
```

---

## Frontend Deployment

### 1. Build

```bash
cd frontend/
npm ci --production=false
npm run build
```

### 2. Start Production Server

```bash
npm start
# Listens on port 3000
```

### 3. Docker Build

```bash
docker build -t insightengine-frontend:latest -f Dockerfile.frontend .
docker run -d --name frontend \
  --env-file .env.production \
  -p 3000:3000 \
  insightengine-frontend:latest
```

---

## Rollback Procedure

### Backend Rollback

```bash
# 1. Stop current version
docker stop backend

# 2. Start previous version
docker run -d --name backend \
  insightengine-backend:<previous-tag>

# 3. Verify health
curl -f http://localhost:8080/api/health/ready
```

### Frontend Rollback

```bash
docker stop frontend
docker run -d --name frontend \
  insightengine-frontend:<previous-tag>
```

### Database Rollback

> [!CAUTION]
> Database rollbacks may cause data loss. Always take a snapshot first.

```sql
-- Revert last migration (check migration table for version)
SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;
-- Manual revert SQL goes here based on migration content
```

---

## Post-Deployment Verification

| Check | Command | Expected |
|-------|---------|----------|
| Backend health | `curl /api/health/ready` | `200 OK` |
| Frontend loads | `curl -I https://app.insightengine.io` | `200 OK` |
| API proxy works | `curl /api/go/health` through frontend | `200 OK` |
| WebSocket | Connect to `wss://app.insightengine.io/ws` | Upgrade success |
| Login flow | Manual test Google OAuth | Session created |
