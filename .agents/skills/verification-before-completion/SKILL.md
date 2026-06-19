---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing in the Casbin repo.
---

# Verification Before Completion

**Announce at start:** "I'm using verification-before-completion; no evidence, no success claim."

## Iron Law

Do not claim success without exact verification evidence or exact blocker.

## Gate Function

Before final response, answer:

1. What files changed?
2. Which layers changed: backend, frontend, DB, worker, upload, realtime, docs, skills?
3. What is narrowest meaningful check?
4. Did contract producer and consumer both get checked?
5. What could not run and why?

## Command Matrix

- Go backend: targeted `go test ./path/...` or `pnpm go:test`.
- Integration: `pnpm go:test-integration` for DB/Redis/Casbin/worker/tenant/upload.
- E2E: `pnpm go:test-e2e` for route lifecycle/cookies/tokens/proxies.
- Frontend web: `pnpm --filter casbin-web typecheck`, lint/build as relevant.
- Frontend client: `pnpm --filter casbin-client typecheck`, build/E2E as relevant; lint placeholder is not strong verification.
- Docs/API: `pnpm go:docs` when Swagger changes.

## Red Flags

- claiming pass after only reading code
- hiding Docker/Snap/network blockers
- treating unrelated failing tests as fixed
- skipping frontend consumer checks after API contract changes
- skipping integration when tenant/Casbin/API-key behavior changed

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
