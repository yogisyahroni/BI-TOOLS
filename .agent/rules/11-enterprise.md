---
trigger: always_on
---

# PART 26: ENTERPRISE INTEGRATION PATTERNS (EIP)

**STANDARD:** APACHE CAMEL / MULESOFT / KAFKA

## 26.1. MESSAGE BROKER ARCHITECTURE

- **Guaranteed Delivery:** Implement **At-Least-Once** delivery semantics. Your consumers must be Idempotent to handle duplicate messages.
- **The Claim Check Pattern:**
  - *Rule:* NEVER send large payloads (&gt; 1MB) through the Message Bus (Kafka/RabbitMQ).
  - *Action:* Upload payload to Blob Storage (S3), send the *Reference ID* (Claim Check) via the bus. Consumer downloads the blob.
- **Dead Letter Channel (DLC):**
  - Every queue must have a corresponding DLC. If a message fails processing 3 times (with backoff), move it to DLC. Do not block the queue.

## 26.2. THE ANTI-CORRUPTION LAYER (ACL)

- **Domain Isolation:** When integrating with a Legacy System (SAP, Salesforce, Mainframe), you MUST place an ACL between the new system and the legacy system.
- **Translation:** The ACL translates the legacy system's messy model into your clean Domain Model. NEVER let legacy concepts leak into your core logic.

## 26.3. EVENT-DRIVEN CONSISTENCY (SAGA PATTERN)

- **Distributed Transactions:** XA Transactions (2PC) are forbidden in microservices due to locking.
- **Orchestration vs Choreography:**
  - Use **Orchestration (Temporal/Camunda)** for complex workflows where state visibility is critical.
  - Use **Choreography (Events)** for simple fire-and-forget notifications.
- **Compensating Transactions:** Every action (e.g., "Charge Card") must have a defined undo action (e.g., "Refund Card") in case the Saga fails later.

---

# PART 27: MAINFRAME & LEGACY MODERNIZATION (COBOL/DB2)

**STANDARD:** STRANGLER FIG / CHANGE DATA CAPTURE (CDC)

## 27.1. MAINFRAME OFFLOADING

- **Read Offloading:** Replicate Mainframe DB2 data to a modern operational store (Postgres/Elasticsearch) using CDC (Change Data Capture) tools like **IBM IIDR** or **Debezium**.
- **MIPS Reduction:** Shift read-heavy traffic to the modern store to reduce Mainframe CPU costs (MIPS).
- **EBCDIC Conversion:** Handle character encoding conversion (EBCDIC to ASCII) explicitly in the integration layer.

## 27.2. LEGACY API WRAPPING

- **Screen Scraping:** If no API exists, use RPA (Robotic Process Automation) or 3270 Emulators only as a last resort.
- **File-Based Integration:**
  - If the interface is a CSV/Fixed-Width file drop: Implement **Idempotent File Processing**. Track processed file hashes to prevent double-ingestion.
