---
trigger: always_on
---

# PART 17: LEGACY MIGRATION STRATEGY (THE STRANGLER FIG)

**STANDARD:** MARTIN FOWLER'S STRANGLER PATTERN

## 17.1. THE DECOMPOSITION STRATEGY

- **Identify the Seam:** Locate a specific domain capability (e.g., "User Profile") in the Monolith that can be isolated.
- **The Proxy Interception:** Place an API Gateway (Kong/Nginx) in front of the Monolith.
  - *Phase 1:* Route `/users/*` to the Monolith (Business as usual).
  - *Phase 2:* Route `/users/new-feature` to the **New Microservice**.
  - *Phase 3:* Gradually shift `/users/*` traffic to the New Service using **Canary Releases** (1% -&gt; 10% -&gt; 100%).

## 17.2. DATA SYNCHRONIZATION (DUAL WRITE / CDC)

- **The Dual Write Problem:** When migrating, data must exist in both the Old DB and New DB.
- **Anti-Corruption Layer (ACL):** Implement an ACL to translate the Monolith's messy data model into the New Service's clean domain model.
- **Change Data Capture (CDC):** Use **Debezium** or **Kafka Connect** to listen to the Monolith's Database Transaction Log (WAL) and replay changes to the New Database asynchronously. This decouples the systems.

## 17.3. THE "KILL SWITCH" (FEATURE FLAGGING)

- **Safety Net:** Every migrated feature MUST be wrapped in a **Feature Flag** (LaunchDarkly / Unleash).
- **Rollback Protocol:** If the New Service error rate exceeds 1% (Error Budget), the system must AUTOMATICALLY flip the flag to route traffic back to the Legacy Monolith.
