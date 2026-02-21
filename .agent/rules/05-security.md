---
trigger: always_on
---

# PART 9: SECURITY HARDENING (GRADE S++ / OWASP)

**PROTOCOL:** ZERO TRUST ARCHITECTURE

## 9.1. AUTHENTICATION AND SESSION MANAGEMENT

- **Stateless Auth:** Use **JWT (JSON Web Tokens)** with short expiration (15 min) + Refresh Tokens (7 days).
- **Cookie Security:** Store tokens in **HttpOnly, Secure, SameSite=Strict** cookies.
  - *Forbidden:* Storing sensitive tokens in `localStorage` (XSS Vulnerable).
- **Rate Limiting:** Implement strict Rate Limiting (Redis-backed) on all public endpoints (Login, Register, Reset Password) to prevent Brute Force.

## 9.2. INPUT VALIDATION AND SANITIZATION

- **Trust No One:** Treat all input (Body, Params, Headers, Cookies) as malicious payloads.
- **Validation Layer:** Validate strictly against a schema (Zod/Pydantic) before business logic execution.
- **Sanitization:** Strip HTML tags from string inputs to prevent **Stored XSS**.
- **SQL Injection:** ALWAYS use Parameterized Queries or an ORM. Raw string concatenation in SQL is a firing offense.

## 9.3. MASS ASSIGNMENT PROTECTION (BOPLA)

- **The Risk:** A user sending `{"isAdmin": true}` in a profile update request.
- **The Defense:** NEVER pass `req.body` directly to the ORM update method.
  - *Correct:* `User.update({ name: body.name, email: body.email })`.
  - *Alternative:* Use DTOs (Data Transfer Objects) to whitelist allowed fields.

## 9.4. INFRASTRUCTURE SECURITY

- **Secrets Management:** API Keys, DB Passwords, and Encryption Keys must be injected via Environment Variables.
  - *Forbidden:* Hardcoding secrets in source code.
- **Network Segmentation:** Database and Internal Services must not be exposed to the public internet (Use VPC/Private Subnet).
- **Headers:** Enforce security headers: `Helmet` (Node), `Content-Security-Policy`, `X-Frame-Options: DENY`.

---

# PART 14: ADVANCED CYBER SECURITY OPERATIONS (SEC-OPS)

**STANDARD:** NIST 800-53 / SOC2 TYPE II COMPLIANCE

## 14.1. CRYPTOGRAPHIC STANDARDS

- **Data at Rest:**
  - Use **AES-256-GCM** for database encryption.
  - Use **Argon2id** (min configuration: m=65536, t=3, p=4) for password hashing. *Bcrypt* is acceptable but deprecated for high-security.
- **Data in Transit:**
  - Enforce **TLS 1.3** exclusively. Disable TLS 1.0/1.1 support.
  - Implement **HSTS (HTTP Strict Transport Security)** with `max-age=63072000; includeSubDomains; preload`.

## 14.2. IDENTITY AND ACCESS MANAGEMENT (IAM)

- **Principle of Least Privilege:**
  - Grant permissions only for the specific resources needed (e.g., `s3:GetObject` on `bucket-x`, NOT `s3:*`).
- **Service Accounts:** Rotate Service Account Keys every 90 days automatically.
- **MFA Enforcement:** Enforce Multi-Factor Authentication for all console access and VPN users.

## 14.3. PENETRATION TESTING & VULNERABILITY MANAGEMENT

- **Automated Scanning:**
  - Run **SAST (Static Application Security Testing)** via SonarQube/CodeQL on every commit.
  - Run **DAST (Dynamic Application Security Testing)** via OWASP ZAP on staging builds.
- **Dependency Auditing:**
  - Block the build if `npm audit` reveals High/Critical vulnerabilities.
  - Use **Dependabot** or **Renovate** to keep libraries patched.

## 14.4. LOGGING & SIEM INTEGRATION

- **Audit Trails:** Log every distinct "Write" operation (Create, Update, Delete) with `ActorID`, `ResourceID`, `Action`, `Timestamp`, and `IP`.
- **Redaction:** AUTOMATICALLY strip PII (Personally Identifiable Information) like Emails, Phones, and Credit Cards from logs before ingestion.
- **Centralization:** Ship logs to ELK Stack, Splunk, or CloudWatch Logs immediately. Do not store logs locally on ephemeral instances.
