---
trigger: always_on
---

# PART 8: DATABASE INTEGRITY AND OPTIMIZATION (ACID)

**STACK:** POSTGRESQL / MYSQL / MONGODB / REDIS

## 8.1. SCHEMA MANAGEMENT AND MIGRATIONS

- **Migration First:** NEVER modify the database manually via GUI tools (e.g., PgAdmin, DBeaver). All changes must be scripted via Migrations (Prisma, TypeORM, Alembic, Goose).
- **Version Control:** Migration files must be committed to Git.
- **Idempotency:** Migrations must be reversible (`up` and `down` scripts).

## 8.2. QUERY OPTIMIZATION AND INDEXING

- **The N+1 Killer:** STRICTLY FORBIDDEN to execute database queries inside a loop.
  - *Solution:* Use Eager Loading (`.include()`, `.with()`, `JOIN FETCH`) or Batch Loading (`DataLoader`).
- **Indexing Mandate:** You MUST create an index for:
  - Every Foreign Key column.
  - Every column used in a `WHERE`, `ORDER BY`, or `GROUP BY` clause.
  - Every column used for text search (GIN/GiST index).

## 8.3. DATA SAFETY AND TRANSACTIONS

- **Atomicity:** Any operation involving multiple write steps (e.g., "Create Order" + "Deduct Inventory") MUST be wrapped in a **Database Transaction**.
- **Soft Deletes:** Use a `deletedAt` timestamp column instead of physical `DELETE` rows, unless compliance (GDPR) requires hard deletion.
- **Concurrency Control:** Use Optimistic Locking (`version` column) to prevent lost updates in high-concurrency environments.

---

# PART 15: DATABASE SCALING & RELIABILITY ENGINEERING

**STANDARD:** THE CAP THEOREM / ACID COMPLIANCE

## 15.1. SHARDING AND PARTITIONING STRATEGIES

- **Horizontal Sharding:** For tables exceeding 100GB, implement Application-Level Sharding based on `TenantID` or `UserID`.
- **Read Replicas:** Offload heavy `SELECT` queries (Reports, Analytics) to Read Replicas to preserve the Primary Writer's throughput.
- **Connection Pooling:** Use **PgBouncer** or **ProxySQL**. Never allow direct application connections to saturate the database max connection limit.

## 15.2. CACHING LAYERS (REDIS / MEMCACHED)

- **Cache-Aside Pattern:**
  1. Check Cache.
  2. If Miss -&gt; Query DB.
  3. Write to Cache.
- **Thundering Herd Protection:** Implement **Probabilistic Early Expiration** (Jitter) or **Request Coalescing** (Singleflight) to prevent database overload when a hot cache key expires.
- **Eviction Policy:** Configure `allkeys-lru` (Least Recently Used) to strictly bound memory usage.

---

# PART 18: MLOPS AND DATA ENGINEERING ARCHITECTURE

**STANDARD:** TFX / KUBEFLOW / VECTOR SEARCH

## 18.1. AI/ML MODEL SERVING (INFERENCE)

- **Latency Budget:** Inference APIs must respond within &lt; 100ms (P95).
  - *Strategy:* Use **ONNX Runtime** or **TorchScript** for optimized model execution. Avoid raw Python interpretation for heavy loops.
  - *Batching:* Implement **Dynamic Batching** (e.g., via BentoML or Ray Serve) to group incoming requests and saturate GPU utilization.
- **Model Versioning:**
  - Treat Models as Code. Use DVC (Data Version Control) or MLflow.
  - **Rollback:** If Model V2 drifts (accuracy drop), automatic rollback to V1 must occur within 30 seconds.

## 18.2. VECTOR DATABASE & RAG (RETRIEVAL AUGMENTED GENERATION)

- **Indexing Strategy:**
  - Use **HNSW** (Hierarchical Navigable Small World) index for low-latency approximate nearest neighbor search.
  - **Hybrid Search:** COMBINE Dense Vector Search (Semantic) with Sparse Keyword Search (BM25) using Reciprocal Rank Fusion (RRF).
- **Embeddings:**
  - Never re-embed static content on the fly. Cache embeddings in **Redis** or **Pinecone** with a content-hash key.

## 18.3. DATA PIPELINES (ETL/ELT)

- **Orchestration:** Use **Airflow** or **Temporal** for reliable workflow execution.
- **Idempotency:** Every data transformation step must be idempotent (re-runnable without side effects).
- **Schema Validation:** Use **Great Expectations** or **Pandera** to validate data quality *before* ingestion. Reject "dirty" data into a Dead Letter Queue.
