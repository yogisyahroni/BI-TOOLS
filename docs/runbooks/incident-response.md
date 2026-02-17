# Incident Response Runbook

## Overview

Steps to take during a production incident (SEV1/SEV2).

## Severity Levels

- **SEV1 (Critical):** System Down, Data Loss, Security Breach. Immediate response required.
- **SEV2 (High):** Major feature broken (e.g., Exports failing), Performance degradation. Response < 1 hour.
- **SEV3 (Medium):** Minor bugs, UI glitches. Business hours response.

## Immediate Actions (SEV1/SEV2)

1. **Acknowledge:** Confirm receipt of alert.
2. **Triage:** Determine impact scope (All users? Specific region? Specific feature?).
3. **Communication:** Update status page / internal stakeholder channel. "Investigating issue with [Component]."

## Common Scenarios

### High Database CPU

1. Check running queries: `SELECT * FROM pg_stat_activity WHERE state = 'active';`
2. Kill long-running queries if necessary.
3. Scale up read replicas if load is high.

### Application Crash (OOM)

1. Check logs for "Out of Memory".
2. Restart container: `docker restart [container_id]`.
3. Analyze Heap Dump post-incident.

### API 500 Errors

1. Check application logs: `docker logs backend --tail 100`.
2. Look for panic traces or DB connection errors.
3. Verify downstream dependencies (Redis, DB) are reachable.

## Post-Mortem

After resolving the incident, create a Post-Mortem document answering:

1. What happened?
2. Why did it happen? (5 Whys)
3. How was it detected?
4. How was it resolved?
5. Action items to prevent recurrence.
