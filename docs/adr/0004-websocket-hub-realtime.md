# ADR-004: WebSocket Hub for Real-Time Collaboration

## Status

Accepted

## Date

2026-02-01

## Context

InsightEngine dashboards support real-time features: live query result streaming, collaborative editing indicators, alert notifications pushed to connected clients, and dashboard refresh triggers. The options considered were:

1. **Server-Sent Events (SSE)**: Unidirectional server-to-client streaming. Simple, HTTP/2 compatible, but no client-to-server messaging.
2. **WebSocket (gorilla/websocket)**: Bidirectional, persistent TCP connections. Full duplex communication.
3. **gRPC Streaming**: Bidirectional streaming with Protocol Buffers. Requires gRPC client in browser (grpc-web proxy).
4. **Polling**: Client repeatedly fetches updates via HTTP. Simple but high latency and bandwidth waste.

## Decision

Use **WebSocket** via `gorilla/websocket` with a centralized **Hub pattern**:

- `WebSocketHub` — Central registry of connected clients, runs as a background goroutine.
- `Client` — Per-connection struct with read/write goroutines and channel-based message passing.
- **Room-based broadcasting**: Clients subscribe to rooms (e.g., `dashboard:{id}`) for targeted message delivery.
- **Auth integration**: JWT token validated on WebSocket upgrade handshake.
- **Graceful shutdown**: Hub drains connections on server shutdown signal.

The Fiber framework's WebSocket adapter (`gofiber/websocket/v2`) wraps gorilla/websocket and integrates with Fiber's middleware chain.

## Consequences

### Positive

- **Low latency**: Sub-100ms message delivery for notifications and alerts.
- **Bidirectional**: Clients can send heartbeats, typing indicators, and room subscription changes.
- **Resource efficient**: Single TCP connection per client vs. multiple HTTP requests for polling.
- **Room isolation**: Dashboard-specific rooms prevent cross-user data leakage.

### Negative

- **Connection state**: WebSocket connections are stateful. Horizontal scaling requires sticky sessions or a shared pub/sub backend (Redis).
- **Reconnection logic**: Frontend must handle disconnections, exponential backoff, and message replay.
- **Load balancer configuration**: Requires WebSocket-aware load balancers (support for `Upgrade` header, long-lived connections).
- **Memory per connection**: Each connection holds goroutines and channel buffers (~8KB per client).
