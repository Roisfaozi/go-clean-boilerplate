---
name: realtime-sse-websocket
description: Use when changing SSE, WebSocket ticket flow, origin checks, Redis presence, distributed realtime behavior, or frontend realtime consumers in this Casbin repo.
---

# Realtime SSE WebSocket

## Overview

Realtime paths here involve auth ticket flow, WebSocket or SSE serving, presence/distribution, and frontend consumers.

Small changes can break auth, origin checks, or distributed presence behavior.

## Read Order

1. `AGENTS.md`
2. `llm/cache/realtime-system.md`
3. `llm/cache/stats-system.md` when metrics broadcasting is involved
4. `internal/router/router.go`
5. `internal/config/app.go`
6. `pkg/ws/*`
7. `pkg/sse/*`
8. relevant frontend consumer paths in `apps/web` or `apps/client`

## Workflow

### Step 1 — Classify Realtime Surface

State whether change affects:

- SSE auth path
- WebSocket ticket flow
- origin or presence checks
- distributed broadcast behavior
- frontend consumer handling

### Step 2 — Trace Auth And Distribution

Confirm:

- how ticket or token is issued and validated
- where origin checks happen
- whether Redis/distributed presence path is involved
- which frontend surface consumes events

### Step 3 — Patch Without Breaking Semantics

- preserve auth boundary on `/events` and `/ws`
- preserve distributed behavior assumptions when enabled
- keep frontend event contract in sync if payload changes

### Step 4 — Verify

- narrow package tests if present
- focused request or ticket lifecycle verification
- frontend consumer check when event payload or connection behavior changes

## Common Mistakes

- treating realtime endpoint like ordinary REST route
- changing event payload without checking consumer handling
- weakening auth or origin validation during debugging

## Stop Conditions

- stop if ticket or token validation owner path is unclear
- stop if distributed presence behavior changes but Redis/distribution path is unreviewed

## Completion Output

Report:

- realtime surface changed
- auth/distribution implications
- files changed
- verification run and exact result
