# Runbook 02: Incident Response

## Overview

Structured response process for production incidents affecting the InsightEngine AI platform.

---

## Severity Classification

| Level | Definition | Response Time | Examples |
|-------|-----------|---------------|----------|
| **P1** | System-wide outage, data loss risk | 15 min | Database down, API 5xx > 50%, auth broken |
| **P2** | Degraded functionality | 1 hour | Slow queries, export failures, WebSocket drops |
| **P3** | Minor issue, workaround exists | 4 hours | UI glitch, non-critical feature broken |
| **P4** | Cosmetic, no user impact | Next sprint | Typo, minor styling issue |

---

## Incident Response Steps

### 1. Detect & Acknowledge

```
[ ] Alert received (Prometheus/PagerDuty/manual report)
[ ] Acknowledge incident in PagerDuty/Slack
[ ] Assign Incident Commander (IC)
[ ] Create incident channel: #incident-YYYY-MM-DD-title
```

### 2. Assess & Classify

```
[ ] Determine severity (P1-P4)
[ ] Identify affected services (backend/frontend/database/WebSocket)
[ ] Check dashboards:
    - Prometheus: http://localhost:9090/graph
    - Grafana: http://localhost:3001/dashboards
    - Backend logs: docker logs backend --tail 200
```

### 3. Investigate

#### Backend Crashes

```bash
# Check backend logs
docker logs backend --tail 500 | grep -i "error\|panic\|fatal"

# Check Go runtime metrics
curl http://localhost:8080/metrics | grep go_goroutines
curl http://localhost:8080/metrics | grep process_resident_memory

# Check database connectivity
curl http://localhost:8080/api/health/ready
```

#### High Latency

```bash
# Check query performance
curl http://localhost:8080/metrics | grep query_execution_duration

# Check database connection pool
curl http://localhost:8080/metrics | grep connection_pool

# Check active WebSocket connections
curl http://localhost:8080/metrics | grep websocket_connections
```

#### Authentication Failures

```bash
# Check auth endpoint
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"test"}'

# Check JWT validation
docker logs backend | grep -i "jwt\|token\|auth"

# Check NextAuth session
docker logs frontend | grep -i "session\|auth"
```

### 4. Mitigate

```
[ ] Apply immediate fix (config change, restart, rollback)
[ ] Communicate status to stakeholders
[ ] Monitor fix effectiveness (5-min intervals)
```

### 5. Resolve

```
[ ] Confirm normal operation for 30 minutes
[ ] Update incident channel with resolution
[ ] Close PagerDuty incident
```

### 6. Post-Mortem (within 48 hours)

```
[ ] Write incident report:
    - Timeline of events
    - Root cause analysis (5 Whys)
    - Impact assessment (users affected, duration)
    - Action items to prevent recurrence
[ ] File follow-up tickets for action items
[ ] Share report in #engineering
```

---

## Quick Reference: Common Issues

### Backend OOM (Out of Memory)

```bash
# Check memory usage
docker stats backend --no-stream

# Restart with increased memory limit
docker update --memory=2g backend
docker restart backend
```

### Database Connection Pool Exhaustion

```bash
# Check active connections
docker exec -it postgres psql -U user -d insightengine \
  -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"

# Kill idle connections
docker exec -it postgres psql -U user -d insightengine \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND query_start < now() - interval '10 minutes';"
```

### Port Conflict (EADDRINUSE)

```bash
# Find process on port
netstat -ano | findstr :8080   # Windows
lsof -i :8080                 # Linux/macOS

# Kill process
taskkill /F /PID <pid>         # Windows
kill -9 <pid>                  # Linux/macOS

# Restart service
docker restart backend
```
