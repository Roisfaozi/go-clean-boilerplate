---
name: frontend-surface
description: Use when deciding whether a frontend change belongs in `apps/web`, `apps/client`, or shared `packages/*`, or when auditing active frontend ownership in this Casbin monorepo.
---

# Frontend Surface

## Overview

This repo has two active frontend apps plus shared packages.

Wrong ownership choice causes duplicated logic, broken proxies, or types drifting between apps.

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/cache/api-contracts.md`
5. `llm/conventions/typescript.md`
6. target app route registry, component tree, or feature folder

## Surface Map

- `apps/web`
  - Next.js App Router
  - server components and route handlers
  - backend proxy: `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client`
  - React Router app and feature folders
  - backend proxy: `apps/client/app/routes/api-proxy.ts`
- `packages/api-types`
  - shared contract typing
- `packages/ui`, `packages/hooks`, `packages/utils`
  - shared primitives only when reuse is real

## Ownership Workflow

### Step 1 — Find Real Owner

Decide based on:

- actual route owner
- actual page or feature owner
- whether both apps expose same feature domain

### Step 2 — Decide Shared Vs Local

Promote to shared package only when:

- both apps use same abstraction
- abstraction is stable enough to reuse
- transport or route ownership does not differ meaningfully

Otherwise keep local.

### Step 3 — Audit Backend Boundary

If backend contract changes, inspect:

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`

If auth or tenant semantics change, also load `auth-tenant-casbin`.

### Step 4 — Preserve Frontend States

For UI changes, preserve:

- loading
- empty
- error
- success
- auth-expired
- tenant-switch state when relevant

### Step 5 — Verify

- `apps/web`: typecheck/build/lint as relevant
- `apps/client`: typecheck/build/E2E as relevant
- remember `apps/client` lint is placeholder-only

## Common Mistakes

- editing wrong app because feature names look similar
- creating shared abstraction before second real use
- changing backend contract without updating proxy or shared types

## Stop Conditions

- stop if route or feature owner is unclear between two apps
- stop if proposed shared package extraction would hide app-specific transport or auth behavior

## Completion Output

Report:

- chosen owner surface
- why local vs shared decision was made
- files changed
- verification run and exact result
