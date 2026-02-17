# Deployment Runbook

## Overview

This runbook details the procedures for deploying the InsightEngine AI platform to production environments.

## Prerequisites

- Access to the target server/cluster (SSH or Kubeconfig).
- Docker and Docker Compose installed (for single-node deployment).
- Database migration credentials (`DATABASE_URL`).

## Deployment Steps (Docker Compose)

1. **Pull Latest Images:**

    ```bash
    docker-compose pull
    ```

2. **Backup Database (Pre-Depoyment):**
    Run the backup script (see `database-maintenance.md`).

    ```bash
    ./scripts/backup_db.sh
    ```

3. **Apply Migrations:**
    The backend container usually runs migrations on startup, but you can trigger manually:

    ```bash
    docker-compose run --rm backend /app/main migrate
    ```

4. **Restart Services:**

    ```bash
    docker-compose up -d
    ```

5. **Health Check:**
    Verify services are healthy:

    ```bash
    curl http://localhost:8080/health
    ```

## Rollback Procedure

If a deployment fails:

1. **Revert to Previous Image Tag:**
    Edit `.env` or `docker-compose.yml` to point to the previous working version.

    ```bash
    export IMAGE_TAG=previous-version
    docker-compose up -d
    ```

2. **Rollback Migrations (If necessary):**
    **WARNING:** Data loss potential. Only do this if the migration broke the schema.

    ```bash
    docker-compose run --rm backend /app/main migrate down
    ```
