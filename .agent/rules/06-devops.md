---
trigger: always_on
---

# PART 10: DEVOPS, CI/CD, AND GIT HYGIENE

**STACK:** DOCKER / GITHUB ACTIONS / KUBERNETES

## 10.1. CONTAINERIZATION STANDARDS

- **Dockerfile:** Always provide a multi-stage `Dockerfile` optimized for production (Distroless or Alpine base).
- **Non-Root User:** Run the application as a non-root user (`USER node` or `USER app`) inside the container to mitigate container breakout attacks.
- **Docker Compose:** Provide a `docker-compose.yml` for local development that spins up DB, Redis, and Mailhog.

## 10.2. CI/CD PIPELINE AUTOMATION

- **Mandatory Workflow:** Every project must include a `.github/workflows/main.yml` or equivalent.
- **Pipeline Stages:**
  1. **Lint:** Check code style (ESLint, Prettier, Black, Gofmt).
  2. **Test:** Run Unit and Integration tests.
  3. **Build:** Verify compilation/transpilation.
  4. **Security Audit:** Run `npm audit` or `trivy` to check for vulnerable dependencies.

## 10.3. GIT HYGIENE AND VERSION CONTROL

- **Branching Strategy:** Use Feature Branches (`feat/`, `fix/`, `chore/`). Direct pushes to `main` or `master` are blocked.
- **Commit Messages:** Use **Conventional Commits** standard.
  - *Format:* `type(scope): description`.
  - *Example:* `feat(auth): implement jwt refresh token rotation`.
- **Pull Requests:** Code must be reviewed (even by yourself via diff check) before merging.

---

# PART 13: ADVANCED CLOUD ARCHITECTURE (AWS / GCP / AZURE)

**STANDARD:** WELL-ARCHITECTED FRAMEWORK / CLOUD NATIVE

## 13.1. INFRASTRUCTURE AS CODE (TERRAFORM / PULUMI)

- **The Immutable Mandate:** Never configure cloud resources manually via the Console. All infrastructure must be defined in code.
- **State Management:**
  - Store state files in a remote backend (S3/GCS) with **State Locking** (DynamoDB) enabled to prevent race conditions.
  - Encrypt state files at rest using KMS/Cloud KMS.
- **Module Structure:**
  - Decouple resources into reusable modules (`modules/vpc`, `modules/rds`, `modules/k8s`).
  - Enforce tagging policies (`Environment`, `CostCenter`, `Owner`) on all resources for FinOps.

## 13.2. SERVERLESS ARCHITECTURE (LAMBDA / CLOUD FUNCTIONS)

- **Cold Start Mitigation:**
  - Use **Provisioned Concurrency** for critical paths (e.g., Checkout, Login).
  - Keep bundle sizes small (&lt; 50MB) by using `esbuild` and tree-shaking layers.
- **Event-Driven Patterns:**
  - Use **SQS/SNS** (AWS) or **Pub/Sub** (GCP) for decoupling services.
  - Implement **Dead Letter Queues (DLQ)** for every asynchronous function to catch failed events. Never let an event vanish.

## 13.3. KUBERNETES (K8S) ENGINEERING

- **Manifest Management:** Use **Helm Charts** or **Kustomize** for environment-specific configurations.
- **Pod Security Context:**
  - `runAsNonRoot: true` (Must run as User ID &gt; 1000).
  - `readOnlyRootFilesystem: true` (Prevent runtime modification).
  - `allowPrivilegeEscalation: false`.
- **Resource Quotas:** STRICTLY define `requests` and `limits` for CPU and Memory to prevent "Noisy Neighbor" issues and OOMKills.
- **Probes:** Define `livenessProbe` (restart if dead) and `readinessProbe` (traffic if ready) for every deployment.

---

# PART 25: OBSERVABILITY & SRE (SITE RELIABILITY ENGINEERING)

**STANDARD:** THE FOUR GOLDEN SIGNALS

## 25.1. METRICS INSTRUMENTATION

- **The Golden Signals:** You must instrument every service to emit:
  1. **Latency:** Time taken to serve a request.
  2. **Traffic:** Demand (req/sec).
  3. **Errors:** Rate of requests that fail (5xx).
  4. **Saturation:** How "full" is the service (CPU/Memory/IO).
- **Cardinality:** Avoid high-cardinality labels (e.g., UserID) in Prometheus metrics. This explodes memory usage.

## 25.2. DISTRIBUTED TRACING (OPENTELEMETRY)

- **Context Propagation:** Every incoming request must generate a `TraceID`. This ID must be passed to DB queries, downstream APIs, and Message Queues headers (`traceparent`).
- **Sampling:** Use Head-Based Sampling (e.g., 1%) in production to save costs, but 100% on errors.
