---
name: frontend-surface
description: Use when deciding whether a frontend change belongs in `apps/web`, `apps/client`, or shared `packages/*`, or when auditing active frontend ownership.
---

# Frontend Surface: App Ownership Decision

**Announce at start:** "I'm using the frontend-surface skill to choose the correct active frontend surface."

## Read Order

1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `llm/conventions/typescript.md`
4. target app route registry/component tree

## Surface Map

- `apps/web`: Next.js App Router, server components/actions, backend API proxy.
- `apps/client`: React Router, feature folders, route registry, API proxy route.
- `packages/api-types`: shared contract/types.
- `packages/ui`: shared UI primitives.
- `packages/hooks`: shared hooks.
- `packages/utils`: shared utilities.

## Workflow

### Phase 1 — Ownership

- Identify URL/route/component owner.
- Check if both apps expose same domain feature.
- Check shared package candidates before duplicating code.

### Phase 2 — API Boundary

- If backend contract changes, load `cross-stack-change`.
- If auth/tenant route behavior changes, load `auth-tenant-casbin`.

### Phase 3 — UI Patch

- Preserve loading, empty, error, success, auth-expired, and tenant-switch states.
- Prefer existing UI components and variants.
- Avoid new shared abstraction until reused by both apps or clearly stable.

### Phase 4 — Verify

- `apps/web`: lint/typecheck/build as relevant.
- `apps/client`: typecheck/build/E2E as relevant; lint script is placeholder-only.

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
