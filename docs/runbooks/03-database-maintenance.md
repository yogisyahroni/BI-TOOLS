# Runbook 03: Database Maintenance

## Overview

PostgreSQL maintenance procedures for the InsightEngine AI database.

---

## Scheduled Maintenance

### Weekly: VACUUM ANALYZE

Run weekly during low-traffic window (Sunday 02:00 UTC):

```sql
-- Analyze all tables for query planner statistics
ANALYZE;

-- VACUUM to reclaim dead tuple space
VACUUM (VERBOSE);
```

### Monthly: Index Health Check

```sql
-- Check index usage stats
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan AS index_scans,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- Identify unused indexes (0 scans)
SELECT indexrelid::regclass AS index_name,
       relid::regclass AS table_name,
       pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Monthly: Table Bloat Assessment

```sql
-- Check table sizes and dead tuples
SELECT
    relname AS table_name,
    n_live_tup AS live_rows,
    n_dead_tup AS dead_rows,
    ROUND(n_dead_tup::numeric / GREATEST(n_live_tup, 1) * 100, 2) AS dead_pct,
    last_vacuum,
    last_autovacuum,
    pg_size_pretty(pg_total_relation_size(relid)) AS total_size
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC
LIMIT 20;
```

---

## Emergency: Connection Pool Exhaustion

### Symptoms

- Backend logs: `too many connections` or `connection pool exhausted`
- Health check fails: `/api/health/ready` returns 503

### Resolution

```sql
-- 1. Check current connections
SELECT count(*), state, usename, application_name
FROM pg_stat_activity
GROUP BY state, usename, application_name
ORDER BY count DESC;

-- 2. Terminate idle connections older than 5 minutes
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND query_start < now() - interval '5 minutes'
  AND pid != pg_backend_pid();

-- 3. Check max connections setting
SHOW max_connections;
```

### Prevention

Ensure GORM connection pool is configured in `database/connection.go`:

```go
sqlDB.SetMaxOpenConns(25)       // Max open connections
sqlDB.SetMaxIdleConns(10)       // Max idle connections
sqlDB.SetConnMaxLifetime(5 * time.Minute) // Connection max lifetime
```

---

## Emergency: Long-Running Query Kill

### Detection

```sql
-- Find queries running > 60 seconds
SELECT
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state,
    usename
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '60 seconds'
  AND state = 'active'
ORDER BY duration DESC;
```

### Kill

```sql
-- Graceful cancel (sends SIGINT)
SELECT pg_cancel_backend(<pid>);

-- Force terminate (sends SIGTERM, use if cancel fails)
SELECT pg_terminate_backend(<pid>);
```

---

## Backup & Restore

### Full Backup (pg_dump)

```bash
pg_dump -h <host> -U <user> -d insightengine \
  --format=custom \
  --compress=9 \
  --file=backup_$(date +%Y%m%d_%H%M%S).dump
```

### Restore

```bash
pg_restore -h <host> -U <user> -d insightengine \
  --clean --if-exists \
  backup_YYYYMMDD_HHMMSS.dump
```

### Point-in-Time Recovery

> [!IMPORTANT]
> Requires WAL archiving to be enabled in `postgresql.conf`:
> `archive_mode = on`, `archive_command = '...'`

```bash
# Restore to specific timestamp
pg_restore --target-time="2026-02-16 10:00:00+00" ...
```

---

## Migration Safety Checklist

Before running any schema migration in production:

```
[ ] Migration tested in staging environment
[ ] Migration is reversible (has down migration)
[ ] Migration does not hold exclusive locks for > 5 seconds
[ ] Large table ALTERs use CONCURRENTLY flag
[ ] Backup taken before migration
[ ] Migration runs within a transaction
[ ] Post-migration ANALYZE run on affected tables
```
