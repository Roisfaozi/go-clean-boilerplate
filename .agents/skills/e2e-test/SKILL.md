---
name: e2e-test
description: Use when testing a feature flow end to end through browser or full request lifecycle in this Casbin repo, especially when auth, tenant, proxy, route protection, or multi-layer form behavior must be proven together.
---

# E2E Test

## Overview

E2E in this repo should prove one real flow across app, proxy, backend, and auth boundaries.

Do not run broad browser motion without naming exact path, actor, and expected result.

## Read Order

1. current diff or feature slice
2. target workflow cache or domain cache
3. `llm/cache/frontend-map.md`
4. `llm/cache/frontend-proxy-system.md`
5. `llm/cache/api-contracts.md`
6. `llm/cache/authentication-system.md` when auth is involved
7. `login` skill when authenticated actor is needed

## Workflow

### Step 1 — Select Exact Flow

State:

- owning app
- actor type
- start path
- final expected outcome

### Step 2 — Prepare Preconditions

Confirm:

- app and backend are running
- required seed or fixture state exists
- login or tenant context is available if needed

### Step 3 — Drive Real Lifecycle

Exercise actual flow through:

- page or route entry
- frontend proxy if used
- backend endpoint
- auth or tenant gate
- success or failure rendering

### Step 4 — Capture Evidence

Capture useful evidence only:

- failing step
- visible error
- console or server error when relevant
- screenshot or artifact when it helps future replay

### Step 5 — Save Reusable Playbook

If flow is stable and reusable, save concise steps to `llm/test-playbooks/`.

## Common Mistakes

- treating smoke check as full E2E
- not naming actor or tenant context
- running browser flow without checking proxy or backend path ownership
- ignoring failure-state assertions

## Stop Conditions

- stop if backend or target app is not running
- stop if flow depends on seed or fixture state that cannot be verified
- stop if auth, tenant, or route ownership is ambiguous

## Completion Output

Report:

- exact flow tested
- actor used
- evidence captured
- commands run and exact result
- skipped coverage and blocker
