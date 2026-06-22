# Realtime System

## Purpose

Durable map for realtime request/connection surfaces in this repo:

- authenticated SSE stream
- ticket-based WebSocket handshake
- presence tracking
- distributed broadcast plumbing
- stats-driven metrics broadcasting
- frontend consumers that depend on payload and auth semantics

Use this file before changing `pkg/ws`, `pkg/sse`, auth ticket flow, metrics broadcasting, presence behavior, or frontend code that consumes live events.

## Primary source of truth

Read these first, then verify target code:

1. `internal/config/app.go`
2. `internal/router/router.go`
3. `pkg/ws/ws_controller.go`
4. `pkg/ws/ticket_manager.go`
5. `pkg/ws/ws_manager.go`
6. `pkg/sse/manager.go`
7. `internal/modules/auth/*` for `/auth/ticket`
8. active frontend consumers under `apps/web` or `apps/client`

## Runtime ownership

### Router ownership

- `internal/router/router.go` registers authenticated `/api/v1/auth/ticket` under `authenticated` routes.
- `internal/router/router.go` registers `/api/v1/upload/*` separately from realtime; do not confuse upload long-lived requests with realtime channels.
- `internal/router/router.go` wires `/api/v1/events` for SSE and `/api/v1/ws` for WebSocket behavior through shared app wiring.

### App wiring ownership

- `internal/config/app.go` creates `presenceManager`, `ticketManager`, `wsManager`, `sseManager`, and `wsController`.
- same file starts `wsManager.Run()` in goroutine.
- same file launches periodic metrics broadcaster that pushes payloads to WebSocket channel `system:metrics` every 2 seconds.
- same file also prunes stale presence users and emits leave updates after prune.

### Package ownership

- `pkg/sse` owns SSE manager, client registration, and event fanout.
- `pkg/ws` owns WebSocket controller, connection lifecycle, channel subscriptions, presence broadcasting, and ticket storage/validation.
- auth module owns issuing one-time tickets; WS route only consumes validated ticket-derived identity.

## Connection model

### SSE path

- SSE is token-protected, not public.
- `pkg/sse/manager.go` exposes HTTP handler that keeps long-lived response open.
- manager owns client register/unregister channels and broadcast channel.
- payload shape comes from `pkg/sse.Event` and should be treated as frontend contract.

### WebSocket path

- WebSocket handshake does not trust raw bearer token at upgrade boundary.
- expected model is: authenticated HTTP caller requests `/auth/ticket`, then client connects to `/ws` with one-time ticket.
- `pkg/ws/ticket_manager.go` stores ticket in Redis key `ws:ticket:<ticket>`.
- `ValidateTicket` is one-time use and deletes ticket immediately after successful validation.
- `pkg/ws/ws_controller.go` upgrades request, derives `user_id` and `organization_id` from context, loads user profile for presence payload, and asks enforcer for user role in org.

### Presence path

- presence is organization-aware, not global-only.
- controller builds `PresenceUser` with `UserID`, `Name`, `AvatarURL`, `Role`, and online status.
- stale-presence pruning in `internal/config/app.go` emits leave updates by organization.
- changes to membership, tenant identity, or role lookup can alter presence payload semantics even if ws code untouched.

## Auth and trust boundaries

- ticket issuance is protected by authenticated route group in `internal/router/router.go`.
- WebSocket auth is short-lived Redis ticket flow; do not replace with unsigned query params or raw user IDs.
- `pkg/ws/ws_controller.go` origin checks compare `Origin` header against configured allowed origins.
- if allowed origins list is empty, upgrader check becomes nil; treat this as security-sensitive config behavior, not harmless convenience.
- API key auth and realtime are not interchangeable by default; check live auth controller and middleware before assuming API keys can open realtime channels.

## Payload and channel semantics

Current durable semantics proven from wiring:

- stats broadcaster publishes JSON payload with top-level `type: metrics_update`
- metrics data includes `rps`, `active_users`, `total_users`, `avg_latency`, `error_rate`, `uptime`, `cpu_usage`, `memory_usage`, `active_threads`
- broadcaster channel name is `system:metrics`
- presence updates are org-scoped and emitted through ws manager helpers, not ad hoc handler JSON responses

If changing any of these, audit frontend consumers and docs together.

## Coupling map

Realtime changes can silently break these neighbors:

- `internal/modules/auth` for ticket issuance
- `internal/modules/stats` because metrics broadcaster reads dashboard summary and system insights
- `internal/modules/organization` because org context and presence semantics intersect
- `llm/cache/authentication-system.md` for session and ticket truth
- `llm/cache/stats-system.md` for metrics payload origin
- `llm/cache/frontend-proxy-system.md` if frontend live-update surfaces proxy or hydrate data around realtime state

## Known implementation details that matter

- `pkg/ws/ticket_manager.go` uses Redis TTL-backed one-time tickets.
- `pkg/ws/ws_controller.go` falls back org to `global` when no org context present.
- role shown in presence payload comes from `GetRolesForUser(userID, orgID)` and uses first returned role.
- SSE manager and WS manager are separate systems; feature parity between them must not be assumed.
- stats metrics broadcaster is bootstrapped in app wiring, not stats module itself.

## Change checklist

Before editing realtime-related code, prove all relevant answers:

1. Is change on SSE only, WS only, or both?
2. Does change alter auth source: JWT/session, ticket, origin, org context, or role lookup?
3. Does change alter frontend payload fields or channel names?
4. Does change alter Redis key lifetime or one-time ticket invalidation?
5. Does change alter distributed or stale-presence pruning behavior?
6. Does change touch stats broadcaster assumptions?

If any answer is unclear, stop and inspect live code first.

## Verification paths

Narrow evidence first:

- `pkg/sse/manager_test.go`
- `pkg/ws/ws_controller_test.go`
- `pkg/ws/ws_manager_test.go`
- `pkg/ws/ticket_manager.go`
- `internal/router/router.go`
- `internal/config/app.go`

If frontend consumer changed too:

- owning app typecheck/build
- targeted browser or E2E flow for live updates

## Hard rules

- Do not weaken one-time ticket semantics.
- Do not swap origin checks for permissive wildcard behavior without explicit reason and config audit.
- Do not change channel names or payload keys without checking frontend consumers.
- Do not assume JWT validation alone is enough for WS route.
- Do not move realtime auth truth into frontend-only guards.
