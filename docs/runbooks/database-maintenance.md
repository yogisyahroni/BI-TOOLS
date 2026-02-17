# Database Maintenance Runbook

## Overview

Procedures for maintaining the PostgreSQL database, including backups, restores, and vacuuming.

## Backup Procedure

**Frequency:** Daily (Automated), On-Demand (Pre-deployment).

### Command

```bash
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME -F c -b -v -f "/backups/db_backup_$(date +%Y%m%d%H%M%S).dump"
```

## Restore Procedure

**WARNING:** This will overwrite the existing database.

1. **Stop Backend Services:**
    Prevent new writes during restore.

2. **Run Restore:**

    ```bash
    pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME -v "/backups/target_backup.dump"
    ```

3. **Verify:**
    Check row counts of critical tables (`users`, `dashboard_cards`).

## Routine Maintenance

### Vacuuming

PostgreSQL requires periodic vacuuming to reclaim storage and update statistics.

```sql
-- Standard Vacuum (Online)
VACUUM ANALYZE;

-- Full Vacuum (Locks tables - Maintenance Window Only)
VACUUM FULL;
```

### Index Reindexing

If query performance degrades, indexes might be bloated.

```sql
REINDEX DATABASE insight_engine;
```
