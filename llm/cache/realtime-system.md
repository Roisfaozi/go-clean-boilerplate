# Realtime System

## Purpose

Durable map for SSE, WebSocket ticket auth, presence, and distributed realtime behavior.

## Runtime truth

- `/api/v1/events` uses `authMiddleware.ValidateToken()` and `sseManager.ServeHTTP()`.
- `/api/v1/ws` uses `authMiddleware.ValidateWebSocketToken()` and `wsController.HandleWebSocket`.
- `pkg/sse` owns server-sent event manager and event stream handling.
- `pkg/ws` owns ticket manager, WebSocket controller/client/manager, and presence manager.
- `internal/config/app.go` wires realtime managers and a metrics broadcaster.

## SSE behavior

- SSE route requires auth token.
- SSE manager registers/unregisters clients and broadcasts named events as JSON data.

## WebSocket ticket behavior

- Ticket creation stores user context in Redis under `ws:ticket:<ticket>` with TTL.
- Ticket validation uses Redis `GetDel` to consume ticket one time.
- WebSocket auth sets user/session/role/username context and organization ID when ticket contains it.

## Presence/distributed behavior

- Presence and distributed WebSocket behavior use Redis-backed managers where configured.
- `internal/config/app.go` also prunes stale presence in metrics broadcaster loop.

## Hard rules

- Do not accept raw access token at WS route unless intentionally redesigning auth boundary.
- Preserve one-time ticket and expiry semantics.
- WebSocket origin validation is security-sensitive.
- Realtime changes need tests or focused integration/manual evidence.

## Tests and evidence paths

- `pkg/ws/*_test.go`
- `pkg/sse/manager_test.go`
- `internal/router/router.go`
- `internal/config/app.go`
