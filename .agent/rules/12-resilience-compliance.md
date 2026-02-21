---
trigger: always_on
---

# PART 28: CHAOS ENGINEERING & RESILIENCE TESTING

**STANDARD:** PRINCIPLES OF CHAOS (NETFLIX SIMIAN ARMY)

## 28.1. FAULT INJECTION PROTOCOLS

- **The Blast Radius:** Start chaos experiments in Staging with a strictly defined blast radius (e.g., "1% of non-critical users").
- **Latency Injection:** Don't just kill services. Inject 2000ms latency into DB calls. The system must degrade gracefully (Circuit Breaker Open), not hang.
- **Dependency Failure:** Block access to S3/Redis. The app must switch to Read-Only mode or serve stale cache, not crash.

## 28.2. CIRCUIT BREAKERS AND BULKHEADS

- **Circuit Breaker:** Wrap every external call (HTTP/gRPC/DB) in a Circuit Breaker (Resilience4j / Polly).
  - *Config:* Open after 50% failure rate. Wait 30s. Half-Open to test.
- **Bulkhead Pattern:** Isolate thread pools. If the "Image Processing" service is stuck, it must not exhaust the threads for the "User Login" service.

---

# PART 29: ADVANCED DATA GOVERNANCE & COMPLIANCE

**STANDARD:** GDPR / HIPAA / SOC2 TYPE II

## 29.1. PRIVACY ENGINEERING (GDPR/CCPA)

- **Right to be Forgotten:**
  - **Crypto-Shredding:** Encrypt PII (Personally Identifiable Information) with a unique per-user key. To "delete" the user, destroy the key. The data becomes ciphertext garbage.
  - **Hard Deletion:** If deleting rows, ensure backups are also scrubbed (or aged out within 30 days).
- **Data Residency:** Respect strict locality rules. EU data must not leave `eu-central-1`. Tag resources with `Region: EU`.

## 29.2. AUDIT LOGGING IMMUTABILITY

- **WORM Storage:** Write Once, Read Many. Store critical audit logs in S3 Object Lock (Compliance Mode) to prevent tampering by anyone (even Root).
- **Chain of Custody:** Logs must contain cryptographic hashes linking to the previous log entry (Blockchain style) to detect deletion.

## 29.3. HIPAA SPECIFICS (HEALTHCARE)

- **PHI Isolation:** Protected Health Information (PHI) must be stored in a separate, isolated database or schema from generic user data.
- **Access Logs:** EVERY access to a PHI record (Read/Write) must be logged with the viewer's ID and justification.

---

# PART 30: LEGAL ENGINEERING & OPEN SOURCE COMPLIANCE

**STANDARD:** SPDX / APACHE 2.0 / MIT / AGPL

## 30.1. LICENSE COMPATIBILITY CHECK

- **The Viral Infection:** STRICTLY FORBIDDEN to use GPL/AGPL libraries in a closed-source/proprietary project (unless dual-licensed).
- **Dependency Audit:**
  - Before installing a package (`npm install`), scan its `package.json` license field.
  - *Safe:* MIT, Apache-2.0, BSD-3-Clause, ISC.
  - *Risk:* GPL-3.0, AGPL-3.0, CC-BY-SA.
- **Attribution:** Automatically generate a `THIRD-PARTY-NOTICES.txt` file listing all dependencies and their licenses for legal compliance.

## 30.2. DATA SOVEREIGNTY & GDPR/CCPA

- **Data Residency:** If the user specifies "Region: EU", you MUST NOT use services (like generic OpenAI endpoints) that process data in the US without a DPA (Data Processing Agreement).
- **Right to Audit:** Architecture must support "Black Box Recording". Every decision made by the AI/Algo must be traceable to a specific dataset or logic path.

---

# PART 31: DISASTER RECOVERY & BUSINESS CONTINUITY (BCP)

**STANDARD:** ISO 22301 / RTO &lt; 15 MIN / RPO &lt; 5 MIN

## 31.1. MULTI-REGION FAILOVER (ACTIVE-PASSIVE)

- **Database Replication:**
  - Enable Cross-Region Read Replicas (e.g., Primary: `us-east-1`, Replica: `eu-west-1`).
  - In case of Region Failure, promote Replica to Primary immediately.
- **DNS Failover:** Use Route53/Cloudflare Health Checks. If Primary is unhealthy (5xx &gt; 10%), automatically update DNS A-Records to the Secondary Region.

## 31.2. BACKUP STRATEGY (3-2-1 RULE)

- **3 Copies:** Production Data, On-Site Backup, Off-Site Backup.
- **2 Media Types:** Disk (Fast Restore) + Object Storage/Tape (Archival).
- **1 Off-Site:** Immutable S3 Bucket (Object Lock) in a different continent.
- **Drill:** You must verify backups by performing a "Dry Run Restore" once a month.

---

# PART 32: ETHICAL AI & ALGORITHMIC BIAS

**STANDARD:** EU AI ACT / NIST AI RMF

## 32.1. BIAS DETECTION

- **Demographic Parity:** When building ranking algorithms (Hiring, Lending), you must test for disparate impact across protected groups (Gender, Race).
- **Explainability (XAI):** For high-stakes decisions (Medical/Financial), use models like Decision Trees or SHAP values. Avoid "Black Box" Neural Networks where explanation is required by law.

## 32.2. SAFETY GUARDRAILS

- **Hallucination Check:**
  - If the output contains factual claims (Medical, Legal), you must verify against a Trusted Knowledge Base (RAG) before displaying.
  - *Fallback:* If uncertain, output: "Confidence Low. Human Verification Required."
