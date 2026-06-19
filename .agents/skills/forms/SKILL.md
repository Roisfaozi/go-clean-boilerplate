---
name: forms
description: Use when building or changing forms, validation, payload mapping, error rendering, or submit flows across `apps/web`, `apps/client`, and Go backend validation in this Casbin monorepo.
---

# Forms

## Overview

Forms in this repo cross frontend state, proxy transport, backend validation, and error-shape handling.

Treat form work as contract work, not only UI work.

## When To Use

Use this skill when:

- creating or changing submit forms
- changing field validation or payload mapping
- changing frontend handling of backend validation errors
- changing auth, tenant, or multipart form behavior

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/cache/api-contracts.md`
5. target frontend form files
6. target backend request structs and controller validation path
7. proxy path in `apps/web` or `apps/client` if request transport matters

## Workflow

### Step 1 — Identify Owner And Contract

State:

- owning app: `apps/web` or `apps/client`
- exact backend endpoint
- payload shape
- validation source on frontend and backend

### Step 2 — Audit Form States

Preserve or add:

- initial state
- loading state
- success state
- backend validation error state
- network failure state
- auth-expired or tenant-related failure state when relevant

### Step 3 — Patch Mapping Carefully

- keep field names aligned with backend request structs
- keep frontend-to-backend transformations explicit
- if file upload or avatar form is involved, also load `tus-upload-storage`

### Step 4 — Check Proxy And Shared Types

If request or response shape changed, audit:

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`

### Step 5 — Verify

Run:

- closest app typecheck
- closest form or page test if present
- backend controller/usecase test if validation contract changed

Use `e2e-test` when submit flow crosses multiple layers and failure states matter.

## Common Mistakes

- changing backend validation but not frontend error rendering
- renaming payload field in UI only
- ignoring auth-expired or tenant-switch states for protected forms
- treating upload form as normal JSON submit path

## Stop Conditions

- stop if owning app or backend endpoint is unclear
- stop if form contract changes but proxy/shared type path is not traced

## Completion Output

Report:

- form surface changed
- contract and validation implications
- files changed
- verification run and exact result
