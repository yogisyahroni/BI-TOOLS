# Runbook 04: Monitoring & Alerting

## Overview

Prometheus metrics, Grafana dashboards, and alerting configuration for InsightEngine AI.

---

## Metrics Endpoints

| Service | URL | Format |
|---------|-----|--------|
| Backend | `http://localhost:8080/metrics` | Prometheus text |
| Frontend | Next.js build-time metrics only | N/A |
| PostgreSQL | `pg_exporter:9187/metrics` | Prometheus text |

---

## Key Metrics to Monitor

### Application Health

| Metric | Alert Threshold | Severity |
|--------|----------------|----------|
| `http_requests_total{status="5xx"}` | > 10/min | P1 |
| `http_request_duration_seconds{quantile="0.99"}` | > 5s | P2 |
| `go_goroutines` | > 10,000 | P2 |
| `process_resident_memory_bytes` | > 2GB | P2 |
| `websocket_connections_active` | > 1,000 | P3 |

### Database Health

| Metric | Alert Threshold | Severity |
|--------|----------------|----------|
| `connection_pool_active_connections` | > 20 (of 25 max) | P2 |
| `query_execution_duration_seconds{quantile="0.95"}` | > 10s | P2 |
| `pg_stat_activity_count{state="active"}` | > 50 | P1 |

### Business Metrics

| Metric | Purpose |
|--------|---------|
| `dashboard_views_total` | Usage tracking |
| `ai_requests_total` | AI feature adoption |
| `alert_evaluations_total` | Alert system health |
| `auth_attempts_total{result="failure"}` | Brute force detection |

---

## Prometheus Alert Rules

File: `prometheus/alerts.yml`

```yaml
groups:
  - name: insightengine
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High 5xx error rate ({{ $value | humanizePercentage }})"

      - alert: SlowQueries
        expr: histogram_quantile(0.95, rate(query_execution_duration_seconds_bucket[5m])) > 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "P95 query latency above 10s"

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes > 2e9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Backend memory usage above 2GB"

      - alert: ConnectionPoolNearCapacity
        expr: connection_pool_active_connections > 20
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool near capacity ({{ $value }}/25)"

      - alert: BruteForceDetected
        expr: rate(auth_attempts_total{result="failure"}[5m]) > 1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Potential brute force attack: {{ $value }} failures/sec"
```

---

## Grafana Dashboard Setup

### Recommended Panels

1. **Request Rate** — `rate(http_requests_total[5m])` by status code
2. **Latency Heatmap** — `http_request_duration_seconds_bucket`
3. **Error Rate** — `rate(http_requests_total{status=~"5.."}[5m])`
4. **Active Connections** — `connection_pool_active_connections` vs `connection_pool_idle_connections`
5. **Goroutine Count** — `go_goroutines` over time
6. **Memory Usage** — `process_resident_memory_bytes`
7. **WebSocket Connections** — `websocket_connections_active`
8. **AI Request Duration** — `ai_request_duration_seconds{quantile="0.95"}`

---

## Health Check Endpoints

| Endpoint | Purpose | Expected |
|----------|---------|----------|
| `/api/health/ready` | K8s readiness — checks DB + cache | 200 or 503 |
| `/api/health/live` | K8s liveness — lightweight heartbeat | Always 200 |
| `/metrics` | Prometheus scrape target | 200 + text |

### Readiness Probe Integration (Kubernetes)

```yaml
readinessProbe:
  httpGet:
    path: /api/health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 15
  failureThreshold: 3

livenessProbe:
  httpGet:
    path: /api/health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 30
  failureThreshold: 5
```
