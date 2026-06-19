---
name: ui
description: Use when building, reviewing, or polishing frontend UI in `apps/web`, `apps/client`, or shared packages in this Casbin monorepo, especially when ownership, shared component reuse, and state completeness matter.
---

# UI

## Overview

UI work here must respect owning app, proxy/auth context, and shared package boundaries.

Do not treat frontend polish as isolated from loading, error, auth-expired, or tenant-sensitive behavior.

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/conventions/typescript.md`
5. `frontend-surface` skill when ownership is not already proven
6. `.agents/skills/vercel-react-best-practices/SKILL.md` when component performance or client/server boundary matters
7. `.agents/skills/vercel-composition-patterns/SKILL.md` when composition API or shared components matter

## Workflow

### Step 1 — Confirm Owner

Confirm whether UI belongs to:

- `apps/web`
- `apps/client`
- shared `packages/ui` or other package

### Step 2 — Reuse Before Inventing

- reuse `packages/ui`, existing hooks, and existing app components first
- create new shared primitive only when second real reuse exists or design is clearly stable

### Step 3 — Preserve Full State Set

Check relevant subset:

- loading
- empty
- error
- success
- auth-expired
- tenant-switch or org-sensitive states

### Step 4 — Respect Backend Boundary

- do not duplicate API client or proxy logic across apps
- do not treat hidden button or hidden page as authorization
- if payload/contract changes, also load `forms`, `frontend-surface`, or `api-endpoint`

### Step 5 — Verify

- app-specific typecheck first
- build or browser flow when visual/runtime path changed

## Common Mistakes

- editing wrong app surface
- inventing new primitive before checking shared library
- polishing happy path only and skipping error/auth-expired states
- implementing authorization as UI visibility only

## Stop Conditions

- stop if owner surface is unclear
- stop if UI change actually depends on unresolved contract or auth-boundary change

## Completion Output

Report:

- owner surface
- reused vs new components
- states checked
- files changed
- verification run and exact result
