# 🛡️ Security Hardening Guide

> **Status:** Production Ready
> **Last Updated:** 2026-02-13
> **Version:** 1.0

This guide outlines the security hardening procedures required for deploying InsightEngine to production environments.

## 1. Operating System Hardening (Linux/Ubuntu)

- [ ] **Minimal Install:** Install only necessary packages.
- [ ] **SSH Configuration:**
  - Disable root login (`PermitRootLogin no`).
  - Use key-based authentication only (`PasswordAuthentication no`).
  - Change default SSH port (optional, e.g., 2222).
- [ ] **Firewall (UFW):**
  - Deny all incoming by default (`ufw default deny incoming`).
  - Allow SSH (`ufw allow 2222/tcp`).
  - Allow HTTP/HTTPS (`ufw allow 80/tcp`, `ufw allow 443/tcp`).
- [ ] **Updates:** Enable automatic security updates (`unattended-upgrades`).

## 2. Container Security (Docker)

- [ ] **Non-Root User:** Run application containers as non-root user.
- [ ] **Image Vuln Scanning:** Scan images using Trivy or Clair before deployment.
- [ ] **Resource Limits:** Set CPU and Memory limits in `docker-compose.yml`.
- [ ] **Read-Only Filesystem:** Mount root filesystem as read-only where possible.

## 3. Database Security (PostgreSQL)

- [ ] **Network Isolation:** Database should not be publicly accessible (bind to localhost or private network).
- [ ] **Strong Passwords:** Use complex generated passwords for DB users.
- [ ] **Least Privilege:** Application user should only have necessary permissions (CRUD), not superuser.
- [ ] **SSL Connections:** Enforce SSL for all connections (`sslmode=require`).
- [ ] **Encryption at Rest:** Ensure underlying storage volume is encrypted (e.g., LUKS, EBS Encryption).

## 4. Application Security (InsightEngine)

### 4.1 Configuration

- [ ] **Environment Variables:**
  - `APP_ENV=production`
  - `SSL_ENABLED=true`
  - `COOKIE_SECURE=true`
- [ ] **Secret Management:**
  - Rotate API Keys and JWT Secrets regularly.
  - Store secrets in a vault (e.g., HashiCorp Vault, AWS Secrets Manager) if possible.

### 4.2 Network Security

- [ ] **Reverse Proxy (Nginx/Traefik):**
  - Configure TLS 1.2/1.3 only.
  - Enable HSTS (`Strict-Transport-Security`).
  - Set secure headers (`X-Frame-Options`, `X-Content-Type-Options`, `Content-Security-Policy`).
- [ ] **WAF:** Deploy a Web Application Firewall (e.g., Cloudflare, AWS WAF) in front of the application.
- [ ] **DDoS Protection:** Enable rate limiting and volumetric attack protection.

### 4.3 Monitoring & Audit

- [ ] **Centralized Logging:** Ship logs to a central server (ELK, Splunk) for tamper-proof storage.
- [ ] **Alerting:** Configure alerts for failed login attempts, high error rates, or suspicious database queries.
- [ ] **Regular Audits:** Schedule quarterly vulnerability assessments and penetration tests.

## 5. Security Checklist - Pre-Flight

1. [ ] Are all default credentials changed?
2. [ ] Is debug mode disabled?
3. [ ] Are backups scheduled and verified?
4. [ ] Is monitoring active?
5. [ ] Have you reviewed the latest security advisories for dependencies?
