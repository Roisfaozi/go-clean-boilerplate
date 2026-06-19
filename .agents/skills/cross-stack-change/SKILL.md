---
name: cross-stack-change
description: Use when a change spans Go backend plus `apps/web`, `apps/client`, frontend proxies, shared API types, or shared packages in this Casbin monorepo.
---

# Cross Stack Change

## Overview

Cross-stack changes fail most often at consumer sync points.

This skill keeps backend contract, proxies, shared types, and frontend behavior aligned.

## Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/frontend-proxy-system.md`
5. `llm/workflows/cross-stack-change.md`
6. backend route/controller/usecase files
7. `apps/web/src/app/api/v1/[...path]/route.ts`
8. `apps/client/app/routes/api-proxy.ts`
9. `packages/api-types/*`

## Workflow

### Step 1 — Name Producer And Consumers

State:

- backend producer endpoint or payload owner
- `apps/web` consumer or proxy path
- `apps/client` consumer or proxy path
- shared type owner if any

### Step 2 — Define Contract Delta

Write exact change in:

- request params or body
- response shape
- error shape
- auth or tenant expectations

### Step 3 — Patch In Correct Order

Prefer order:

1. backend contract owner
2. shared types
3. proxies
4. frontend consumers

### Step 4 — Verify Producer And Consumer

- backend tests or route checks
- app typecheck or focused consumer checks
- browser/E2E only when needed for flow confidence

## Common Mistakes

- changing backend payload and forgetting one active app
- updating proxy but not shared type
- claiming frontend unaffected without checking both active surfaces

## Stop Conditions

- stop if one consumer path cannot be identified
- stop if contract changed but producer/consumer verification pair is missing

## Completion Output

Report:

- producer and consumers touched
- contract delta
- files changed
- verification run and exact result
