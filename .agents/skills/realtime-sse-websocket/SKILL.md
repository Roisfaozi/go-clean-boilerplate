---
name: realtime-sse-websocket
description: Use when changing SSE, WebSocket ticket flow, origin checks, Redis presence, distributed realtime behavior, or frontend realtime consumers.
---

# Realtime SSE WebSocket: Ticket and Presence Boundary

**Announce at start:** "I'm using the realtime-sse-websocket skill because realtime auth and origin checks are security-sensitive."

## Read Order

1. `llm/cache/domain-rules.md`
2. `llm/cache/backend-map.md`
3. `internal/router/router.go`
4. `pkg/sse`
5. `pkg/ws`
6. relevant frontend consumers
7. realtime tests

## Boundary Map

- SSE route requires auth token.
- WebSocket route uses short-lived Redis ticket flow.
- Presence/distributed behavior uses Redis-backed managers.
- Origin validation is security-sensitive.

## Workflow

### Phase 1 — Trace Connection

- ticket issuance
- ticket storage/expiry
- WS upgrade route
- origin check
- presence registration
- message lifecycle

### Phase 2 — Patch Rules

- Do not accept raw access token at WS route unless live code intentionally supports it.
- Preserve ticket one-time/expiry semantics if present.
- Preserve distributed Redis behavior.

### Phase 3 — Verify

- unit tests for ticket validation/origin behavior if present
- integration/realtime tests where available
- frontend smoke flow if UI depends on realtime updates

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
