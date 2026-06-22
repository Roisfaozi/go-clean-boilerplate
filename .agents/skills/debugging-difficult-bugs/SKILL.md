---
name: debugging-difficult-bugs
description: Use early when debugging medium or hard bugs in this Casbin repo involving runtime state, ordering, persistence, middleware layering, auth, tenant, Casbin, API keys, workers, uploads, or other behavior that is hard to prove from a quick code skim.
---

# Debugging Difficult Bugs

## Overview

Hard bugs in this repo usually live at boundaries: router to middleware, middleware to usecase, transaction to side effect, or backend contract to frontend proxy.

Do not patch from intuition. Reproduce, trace, prove, then fix.

## Read Order

1. exact bug report, failing test, or reproduction steps
2. `AGENTS.md`
3. relevant `llm/cache/*`
4. `internal/config/app.go` if wiring may matter
5. `internal/router/router.go` if request path or middleware may matter
6. target module and nearest tests

## Workflow

### Step 1 — Pin Down Symptom

Capture exact symptom:

- request path and method
- actor type: public, authenticated, tenant user, API key, admin
- environment trigger
- expected vs actual behavior

### Step 2 — Reproduce Narrowly

Prefer one of:

- failing package test
- narrow manual request
- minimal browser path
- temporary trace through exact request path

### Step 3 — Trace Runtime Path

Trace full path through:

- router group and middleware stack
- controller input parsing
- usecase decision path
- repository or external dependency
- worker, audit, webhook, or upload side effect if present

### Step 4 — Capture Evidence

Use concrete evidence only:

- failing assertion
- exact payload or response
- log line
- DB state
- Redis/session state
- middleware branch mismatch

### Step 5 — Prove Root Cause

State one root cause that explains symptom.

If evidence only supports multiple possibilities, keep investigating. Do not patch all possibilities at once.

### Step 6 — Patch And Re-verify

- make smallest root-cause fix
- remove temporary instrumentation unless intentionally kept
- rerun narrow reproduction first
- run adjacent verification only after reproduction is fixed

## Common Bug Zones In This Repo

- Redis-backed session validity vs JWT-only assumptions
- tenant organization resolution and membership cache
- Casbin enforcement mismatch between route group and domain logic
- API-key scope drift after route additions
- transaction writes with side effects emitted too early or too late
- frontend proxy mismatch after backend contract changes

## Stop Conditions

- stop if bug cannot be reproduced and no runtime evidence exists
- stop if live code path differs from reported path and report mismatch
- stop before broad speculative patch across multiple modules

## Completion Output

Report:

- reproduction method
- root cause proven
- files changed
- commands run and exact result
- remaining risk or follow-up verification
