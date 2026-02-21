---
trigger: always_on
---

# PART 7: BACKEND ENGINEERING ARCHITECTURE

**STACK:** NODE.JS / GO (GOLANG) / PYTHON / RUST

## 7.1. THE "END-TO-END" WIRING OATH

"Done" does not mean "The API works in Postman." "Done" means the data flows seamlessly from the User Interface to the Disk and back.

### 7.1.1. THE CONNECTION MANDATE

- **Rule:** Every UI Component (e.g., `&lt;UserProfile /&gt;`) MUST have a corresponding API Service (`UserService`).
- **Rule:** Every API Service MUST have a corresponding Controller (`UserController`) and Repository/Model (`UserModel`).
- **Constraint:** You cannot mark a feature as "Complete" until the Frontend is successfully consuming the Backend data with error handling.

## 7.2. NODE.JS / TYPESCRIPT STANDARDS (NESTJS / EXPRESS)

- **Architecture:** Adhere to **Clean Architecture** or **Domain-Driven Design (DDD)**.
  - *Layers:* Controller -&gt; Service -&gt; Repository -&gt; Entity.
- **Error Handling:** Use global Exception Filters. NEVER let the app crash on `uncaughtException` or `unhandledRejection`.
- **Logging:** Use structured JSON logging (e.g., `pino`, `winston`). `console.log` is strictly forbidden in production code.
- **Performance:** Use `sharp` for image processing (off the main thread). Use `BullMQ` or `Redis` for background jobs.

## 7.3. GO (GOLANG) STANDARDS

- **Project Structure:** Follow `cmd/`, `internal/`, `pkg/` layout.
- **Error Handling:** Handle errors explicitly (`if err != nil`). Panic is reserved for startup failures only.
- **Concurrency:** Use `goroutines` and `channels` for parallel processing, but ALWAYS implement a `WaitGroup` or `ErrGroup` to prevent zombie routines.
- **Context:** Propagate `context.Context` through every function call for timeout and cancellation control.

## 7.4. PYTHON STANDARDS (FASTAPI / DJANGO)

- **Type Safety:** Use **Pydantic V2** for all data validation and serialization.
- **Async:** Use `async def` for I/O-bound operations. Use `def` for CPU-bound operations (to leverage thread pool).
- **Dependency Injection:** Use FastAPI's `Depends()` or a container like `Dependency Injector` to manage database sessions and services.

---

# PART 19: REAL-TIME SYSTEMS & HIGH-FREQUENCY PROTOCOLS

**STANDARD:** WEBSOCKETS / GRPC / MQTT

## 19.1. WEBSOCKET ARCHITECTURE (SOCKET.IO / WS)

- **Connection Handling:**
  - **Heartbeats:** Implement strict Ping/Pong intervals (30s). If Pong misses x2, terminate and reconnect with Exponential Backoff.
  - **State Reconciliation:** On reconnect, the client must request a "State Sync" to catch up on missed events.
- **Scalability:**
  - **Pub/Sub Backplane:** Use **Redis Pub/Sub** or **NATS** to broadcast messages across multiple WebSocket server nodes. Sticky Sessions are a fragility; avoid them if possible.

## 19.2. gRPC AND PROTOBUF (MICROSERVICES)

- **Contract First:** Define `.proto` files as the Single Source of Truth.
- **Backward Compatibility:** NEVER delete fields or change field IDs in `.proto`. Mark them as `reserved` or `deprecated`.
- **Deadlines:** Every gRPC call must have a `deadline` (timeout) propagated from the edge. Infinite waits are forbidden.

## 19.3. EVENT SOURCING & CQRS

- **Command Side (Write):** Optimize for Consistency. Write to an Append-Only Log (Event Store).
- **Query Side (Read):** Optimize for Availability. Project events into Denormalized Views (Materialized Views) in SQL/NoSQL.
- **Eventual Consistency:** Accept that the Read side may lag by milliseconds. Handle this in the UI via Optimistic Updates or Loading States.

---

# PART 20: API GOVERNANCE & DOCUMENTATION

**STANDARD:** OPENAPI 3.1 / GRAPHQL FEDERATION

## 20.1. API CONTRACT ENFORCEMENT

- **Spec-Driven Development:**
  - Write the **OpenAPI (Swagger)** spec *before* writing the controller.
  - Use tools like `spectral` to lint the API spec against style guides.
- **Versioning:**
  - Use URI Versioning (`/v1/users`) or Header Versioning (`Accept: application/vnd.myapi.v1+json`).
  - Breaking Changes require a new Major Version. No exceptions.

## 20.2. GRAPHQL FEDERATION (APOLLO)

- **N+1 Protection:** STRICTLY enforce `DataLoader` patterns in all resolvers.
- **Complexity Limits:** Implement Query Complexity Analysis to reject deep nested queries (DoS protection).
- **Schema Stewardship:** Deprecate fields using `@deprecated(reason: "...")` for at least 3 months before removal.
