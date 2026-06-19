---
name: plan
description: Use when work in this Casbin repo needs staged implementation planning before editing, especially for multi-file backend changes, cross-stack work, security-sensitive changes, or any task where order, dependencies, and verification must be explicit.
---

# Plan

## Overview

This skill turns a validated design or request into executable staged work.

Good plan for this repo is not generic TODO list. It must encode ownership, sequence, verification, and repo-specific boundaries.

## When To Use

Use this skill when:

- task spans more than one module or layer
- auth, tenant, Casbin, API key, upload, worker, or transaction risk exists
- frontend and backend must stay in sync
- user asked for phased implementation, audit map, or approval gate

Do not use this skill when:

- change is tiny and single-file
- direct implementation path is obvious and low risk

## Required Read Order

1. approved user request or design note
2. `AGENTS.md`
3. relevant `llm/cache/*`
4. relevant `llm/workflows/*`
5. `internal/config/app.go` if wiring matters
6. `internal/router/router.go` if route/middleware matters
7. target module and consumer paths

## Plan Workflow

### Step 1 — Define Scope

Capture:

- target feature or bug
- owning modules and apps
- touched boundaries
- non-goals and explicit exclusions

### Step 2 — Break Into Execution Slices

Each slice should be coherent and verifiable, for example:

- runtime wiring
- route/controller contract
- usecase/repository logic
- shared API types and frontend proxy
- verification and reviewer cleanup

Avoid vague tasks like `implement feature`.

### Step 3 — Encode Dependencies

State which step depends on which:

- schema before repository logic
- repository/usecase before route exposure
- backend contract before frontend consumer sync
- implementation before broad verification

### Step 4 — Add File Ownership

For each step, name exact file groups or modules, such as:

- `internal/router/router.go`
- `internal/modules/<name>/delivery/http/*`
- `internal/modules/<name>/usecase/*`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`

### Step 5 — Add Verification Per Step

Every step should name narrow verification first:

- package tests
- `pnpm typecheck`
- `pnpm go:test`
- `make test-unit`
- integration/E2E only when boundary needs it

### Step 6 — Save Plan

Use:

- `llm/tasks/todo.md` for active short-lived task execution
- `llm/plans/` for durable multi-stage work

## Plan Quality Rules

- each step must have clear owner and outcome
- each step must be independently reviewable
- exact file paths beat generic area names
- include blockers or approval gates where needed
- mention matching domain skills to use during execution

## Common Mistakes

- planning around docs instead of runtime truth
- hiding risky auth/tenant/Casbin work inside one broad task
- forgetting frontend proxy update after backend contract change
- listing broad tests only, with no narrow package checks

## Stop Conditions

- stop if current request still lacks clear module or route owner
- stop if plan depends on unverified architecture assumption
- stop if proposed sequence hides destructive DB/schema/data step not explicitly approved
- stop if backend contract work has no identified consumer sync path

## Next Skill Handoff

- use `execute` after plan is accepted
- use `design` or `research` first if ownership or architecture still unclear

## Completion Output

Report:

- plan location
- ordered steps
- dependencies and approval gates
- per-step verification strategy
