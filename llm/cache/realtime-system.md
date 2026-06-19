# Realtime System

## Purpose

Durable map for SSE, WebSocket ticket flow, presence, distributed realtime behavior, and frontend-facing realtime consumers.

## Runtime truth

- `internal/router/router.go` registers `/api/v1/events` and `/api/v1/ws`.
- `/events` uses token validation; `/ws` uses WebSocket ticket validation path.
- `pkg/ws` owns ticket manager, websocket serving helpers, and presence-related logic.
- `pkg/sse` owns SSE serving helpers.
- `internal/config/app.go` wires realtime managers, ticket manager, and distributed/runtime behavior.

## Behavior surfaces

- SSE authenticated event stream
- WebSocket ticket issuance and validation
- presence tracking and pruning
- distributed broadcast behavior when enabled
- frontend consumer event handling in active apps

## Auth and ticket semantics

- WebSocket route does not use raw access token directly by default.
- Ticket flow is part of auth plus realtime contract.
- Event and socket consumers depend on payload and connection semantics staying stable.

## Coupling to other systems

- auth module issues ticket path for realtime use
- stats broadcaster can publish metrics over websocket channels
- Redis or distributed presence behavior may affect runtime semantics

## Hard rules

- Do not weaken auth or ticket validation during realtime changes.
- Do not change event payload or channel semantics without checking frontend consumers.
- Treat distributed presence or broadcast changes as infrastructure-sensitive behavior.

## Verification and evidence paths

- `internal/router/router.go`
- `internal/config/app.go`
- `pkg/ws/*`
- `pkg/sse/*`
- frontend consumer paths under `apps/web` and `apps/client`
