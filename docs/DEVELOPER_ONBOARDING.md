# Developer Onboarding Guide

Welcome to the InsightEngine AI team! This guide will get you set up and productive in < 1 day.

## 1. Environment Setup

### Prerequisites

- **Docker & Docker Compose:** Required for running the full stack locally.
- **Go 1.22+:** For backend development.
- **Node.js 20+ & npm:** For frontend development.

### Setup Steps

1. **Clone the Repository:**

    ```bash
    git clone https://github.com/your-org/insight-engine-ai-ui.git
    cd insight-engine-ai-ui
    ```

2. **Configure Environment Variables:**
    Copy `.env.example` to `.env` in both `backend/` and `frontend/`.

    ```bash
    cp backend/.env.example backend/.env
    cp frontend/.env.example frontend/.env.local
    ```

3. **Start Infrastructure (DB, Redis):**

    ```bash
    docker-compose up -d postgres redis
    ```

4. **Run Backend:**

    ```bash
    cd backend
    go mod download
    go run main.go
    ```

    API will be available at `http://localhost:8080`.

5. **Run Frontend:**

    ```bash
    cd frontend
    npm install
    npm run dev
    ```

    UI will be available at `http://localhost:3000`.

## 2. Code Structure

- `backend/`: Go (Fiber) application.
  - `handlers/`: HTTP Request handlers.
  - `services/`: Business logic.
  - `models/`: Database structs.
  - `docs/`: ADRs and Runbooks.
- `frontend/`: Next.js application.
  - `components/`: Reusable UI components.
  - `app/`: App Router pages.

## 3. Workflow

- **Branching:** Use feature branches (`feat/my-feature`, `fix/bug-id`).
- **Commits:** Use conventional commits (e.g., `feat(auth): add login endpoint`).
- **PRs:** Require 1 approval + CI checks passing.

## 4. Testing

- **Backend:** `go test ./...`
- **Frontend:** `npm test`

## 5. Troubleshooting

- **DB Connection Refused:** Ensure Docker container is running (`docker ps`). Check `.env` DB_HOST.
- **CORS Error:** Check `allowed_origins` in `backend/main.go`.
