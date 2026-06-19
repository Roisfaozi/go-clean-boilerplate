---
name: e2e-test
description: Use when testing a feature flow end-to-end through browser or full request lifecycle in the Casbin repo.
---

# E2E Test: Casbin Flow Verification

**Announce at start:** "I'm using the e2e-test skill to verify the user flow end-to-end."

## Read Order

1. changed diff
2. target workflow cache
3. `llm/cache/frontend-map.md`
4. `llm/cache/api-contracts.md`
5. `llm/cache/frontend-proxy-system.md`
6. `llm/cache/authentication-system.md` if auth is involved

## Workflow

### Phase 1 — Select Flow

Choose the exact user journey and owning app.

### Phase 2 — Setup

Use login/setup pattern if auth is required.

### Phase 3 — Drive Flow

Exercise the actual browser/request lifecycle.

### Phase 4 — Evidence

Capture failure evidence, console/server errors, or screenshots when useful.

### Phase 5 — Save

If stable and reusable, save the steps to `llm/test-playbooks/`.

## Stop Conditions

- stop if backend/app is not running
- stop if the flow needs a data seed or fixture you cannot verify
- stop if auth/tenant boundary is ambiguous

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
