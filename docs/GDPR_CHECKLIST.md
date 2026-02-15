# 🛡️ GDPR Compliance Checklist

> **Status:** Self-Assessment in Progress
> **Last Updated:** 2026-02-13
> **Version:** 1.0

This document outlines the General Data Protection Regulation (GDPR) compliance verification for the InsightEngine platform.

## 1. Lawful Basis & Transparency

| ID | Control | Status | Implementation Detail |
| :--- | :--- | :--- | :--- |
| 1.1 | **Privacy Policy** | 📝 Todo | Needs to be hosted at `/privacy`. |
| 1.2 | **Cookie Consent** | ✅ Done | Implemented via `cookie-consent-banner.tsx`. |
| 1.3 | **Terms of Service** | 📝 Todo | Needs to be hosted at `/terms`. |

## 2. Data Rights (Data Subject Rights)

### 2.1 Right to Access (Article 15)

- [x] **Data Export Feature:** Users can export their profile and activity logs.
  - *Implementation:* `GET /api/users/me/export` (returns JSON package).
  - *UI:* Usage in `Settings > Privacy > Download Data`.

### 2.2 Right to Rectification (Article 16)

- [x] **Profile Editing:** Users can modify all personal data.
  - *Implementation:* `PUT /api/users/me`.

### 2.3 Right to Erasure ("Right to be Forgotten") (Article 17)

- [x] **Account Deletion:** Users can self-delete accounts.
  - *Implementation:* `DELETE /api/users/me` triggers soft-delete then hard-delete after 30 days.
  - - Cascading:* Removes associated personal queries and dashboard permissions.

### 2.4 Right to Restrict Processing (Article 18)

- [x] **Deactivate Account:** Users can freeze their account.
  - *Implementation:* `POST /api/users/me/deactivate`.

### 2.5 Right to Data Portability (Article 20)

- [x] **Standard Formats:** Exports provided in standard JSON/CSV.
  - *Implementation:* Reporting Service supports CSV/Excel exports.

## 3. Data Security (Article 32)

- [x] **Encryption at Rest:**
  - *Detail:* Database disks encrypted. Sensitive fields (API Keys, Passwords) encrypted using `EncryptionService` (AES-256-GCM).
- [x] **Encryption in Transit:**
  - *Detail:* TLS 1.2+ enforced via `SSLMiddleware`.
- [x] **Access Control:**
  - *Detail:* RBAC enforced via `RBACMiddleware`.

## 4. Accountability & Governance

- [x] **Audit Logging:**
  - *Implementation:* `AuditService` logs all access to sensitive data (Task 014).
- [x] **Data Breach Notification:**
  - *Procedure:* Defined in `docs/DISASTER_RECOVERY.md`.
