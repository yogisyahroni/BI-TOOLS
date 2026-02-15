# Disaster Recovery & Business Continuity Plan (BCP)

**Severity Level:** Critical
**RTO (Recovery Time Objective):** < 1 Hour
**RPO (Recovery Point Objective):** < 15 Minutes

## 1. Backup Strategy

The system relies on PostgreSQL as the primary data store. Backups are critical for recovering from data corruption, accidental deletion, or catastrophic infrastructure failure.

### 1.1 Automated Backups

- **Frequency:** Daily (Full), Every 15 minutes (WAL Logs/Point-in-Time).
- **Storage:** Secure S3 Bucket (or equivalent off-site storage).
- **Retention:** 30 Days.

### 1.2 Database Dump (Manual Trigger)

Use the provided `scripts/backup.ps1` (Windows) or `pg_dump` command:

```powershell
./scripts/backup.ps1
```

Or manually:

```bash
pg_dump -h localhost -U postgres -d insight_engine > backup_$(date +%Y%m%d).sql
```

## 2. Restore Procedures

### 2.1 Full Database Restore

**WARNING:** This will overwrite existing data.

1. **Stop the Backend Service:**
    Ensure no active connections are writing to the DB.

2. **Drop & Recreate Database:**

    ```sql
    DROP DATABASE insight_engine;
    CREATE DATABASE insight_engine;
    ```

3. **Restore from Dump:**

    ```bash
    psql -h localhost -U postgres -d insight_engine < backup_YYYYMMDD.sql
    ```

4. **Verify Data Integrity:**
    - Check row counts of critical tables (`users`, `data_sources`).
    - Run `SELECT count(*) FROM users;`.

5. **Restart Backend Service:**
    `go run main.go`

### 2.2 Point-in-Time Recovery (PITR)

(Requires WAL Archiving enabled in `postgresql.conf`).
Refer to cloud provider documentation (AWS RDS / GCP Cloud SQL) if using managed databases.

## 3. Failover Scenarios

### 3.1 App Server Failure

- **Symptom:** API Unreachable (503/Timeout).
- **Action:**
    1. Check logs: `docker logs backend` / console output.
    2. Restart container/process: `docker restart backend` or `go run main.go`.
    3. If hardware failure, deploy to standby server.

### 3.2 Database Failure

- **Symptom:** "Connection refused" or Circuit Breaker Open.
- **Action:**
    1. Check DB status.
    2. Promote Read Replica to Primary (if configured).
    3. Update `DB_HOST` in `.env` to point to new Primary.
    4. Restart Backend.

## 4. Circuit Breaker & Degradation

The system implements:

- **Circuit Breaker:** Automatically stops requests to DB if failure rate > 60%.
- **Graceful Degradation:**
  - If DB is down, API returns `X-System-Status: degraded`.
  - Read-only mode may be active.
  - Cached data via Redis (if enabled) is served.

## 5. Contact List

- **DevOps Lead:** [Name] - [Phone]
- **Backend Lead:** [Name] - [Phone]
