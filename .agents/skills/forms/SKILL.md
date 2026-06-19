---
name: forms
description: Use when building or changing forms, validation, payload mapping, error rendering, or submit flows across `apps/web`, `apps/client`, and Go backend validation.
---

# Forms: Frontend Validation and Backend Contract

**Announce at start:** "I'm using the forms skill to align UI form state with backend validation and contract behavior."

## Read Order

1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-proxy-system.md`
4. target backend request struct/validation tags
5. closest frontend form precedent

## Workflow

### Phase 1 — Field Ownership

Classify each field:

- UI-only
- API payload
- persisted model
- derived/display-only

### Phase 2 — Validation

- Match frontend validation with backend validation tags and usecase rules.
- Preserve server error display.
- Handle auth/session expiry and tenant errors.

### Phase 3 — States

Cover:

- loading
- empty/default
- dirty/disabled
- validation error
- submit error
- success

### Phase 4 — Verify

- submitted payload shape
- backend response/error shape
- affected app typecheck/build
- browser/E2E flow when user-critical

## Project examples

- `apps/web/src/components/auth/login-form.tsx`
- `apps/web/src/app/actions/auth.ts`
- `apps/client/app/features/shared/crud-form-dialog.tsx`

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
