---
name: design
description: Use when designing a new feature or system change in this Casbin repo before implementation, especially when route ownership, module boundaries, auth or tenant impact, API contracts, or cross-stack effects must be clarified first.
---

# Design

## Overview

Use this skill before implementation when architecture choices matter more than code speed.

In this repo, design work must start from live wiring and domain boundaries, not from generic CRUD assumptions.

## When To Use

Use this skill when:

- new feature spans router, middleware, usecase, repository, and frontend consumer layers
- auth, tenant, Casbin, API-key, upload, worker, or realtime boundaries may change
- module ownership is unclear
- payload or route design will affect `apps/web`, `apps/client`, or `packages/api-types`

Do not use this skill for:

- tiny localized bugfix with obvious owner
- purely stylistic docs cleanup
- single-package refactor with no contract or boundary change

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/architecture.md`
3. `llm/cache/backend-map.md`
4. `llm/cache/module-map.md`
5. `llm/cache/domain-rules.md`
6. relevant domain cache files
7. relevant workflow file under `llm/workflows/`
8. `internal/config/app.go`
9. `internal/router/router.go`
10. target module constructor, controller, usecase, repository paths

Add these when relevant:

- `llm/cache/frontend-map.md`
- `llm/cache/api-contracts.md`
- `llm/cache/frontend-proxy-system.md`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`

## Design Workflow

### Step 1 — Frame Problem

Write down:

- user or operator goal
- owning surface: backend only, `apps/web`, `apps/client`, or cross-stack
- target module or modules
- risky boundaries: auth, tenant, Casbin, API key, DB transaction, worker side effect, upload, realtime

### Step 2 — Trace Current Runtime

Prove live flow before proposing changes:

- route registration in `internal/router/router.go`
- dependency wiring in `internal/config/app.go`
- target module constructor and delivery/usecase/repository boundaries
- frontend proxy and shared type consumers if contract may move

Do not rely on README summaries when code says otherwise.

### Step 3 — Define Design Decisions

Cover these explicitly when relevant:

- route stratum and middleware stack
- request/response contract
- data ownership and transaction boundary
- tenant/org resolution source
- Casbin or API-key enforcement path
- worker/audit/webhook side effects
- frontend owner and shared type update scope

### Step 4 — Compare Options

For each non-trivial decision:

1. state current constraint
2. give 2-3 implementation options
3. note tradeoffs against repo architecture
4. recommend one with reason tied to live code

### Step 5 — Write Design Artifact

Save durable design output to one of:

- `llm/plans/` for staged implementation planning
- `llm/research/` for evidence-heavy analysis
- `llm/tasks/` for active-task-only design notes

Use exact repo paths and exact route/module names.

## Required Sections In Design Output

- summary
- current runtime path
- proposed ownership
- contract and data model impact
- auth/tenant/Casbin/API-key implications
- verification strategy
- open questions marked `needs confirmation` only when not provable locally

## Common Mistakes

- proposing handler-level checks instead of router/middleware enforcement
- ignoring Redis-backed session validation when auth appears simple
- treating tenant as query parameter concern instead of middleware/usecase boundary
- skipping frontend proxy ownership on backend contract changes
- designing transaction-sensitive write flow without `database-transactions`

## Stop Conditions

- stop if route ownership is unclear
- stop if live wiring contradicts cache/docs and contradiction changes design choice
- stop if change implies schema or destructive data action not explicitly requested
- stop if auth, tenant, or permission boundary cannot be proven from code

## Next Skill Handoff

- use `plan` after design is accepted
- use `research` first if core product or technical approach is still unclear
- use domain skill during implementation once owner boundary is known

## Completion Output

Report:

- files read
- current runtime path proven
- chosen design and rejected alternatives
- risks, blockers, and exact next implementation skill
