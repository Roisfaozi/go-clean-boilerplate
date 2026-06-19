---
name: feature
description: Use when building new backend, frontend, or cross-stack capability from scratch in this Casbin monorepo, especially when research, design, planning, implementation, and verification must be orchestrated end to end.
---

# Feature

## Overview

This is orchestration skill for new capability, not direct code pattern.

Use it when request spans multiple phases and multiple boundaries. It coordinates other repo-specific skills in order.

## Pipeline

`research -> brainstorm -> design -> plan -> execute -> verification-before-completion -> self-review`

## When To Use

Use this skill when:

- feature spans multiple files or layers
- backend/frontend/API contract may all move
- auth, tenant, Casbin, API-key, upload, worker, or realtime boundaries may be touched
- design or planning does not yet exist

Do not use this skill when:

- change is one bounded bugfix
- design and plan already exist; use `execute`

## Read Order

1. `AGENTS.md`
2. `llm/cache/project-overview.md`
3. relevant `llm/cache/*`
4. relevant `llm/workflows/*`
5. live code ownership paths

## Feature Workflow

### Phase 1 — Research Current Shape

- identify owner module and app surface
- inspect live code before trusting docs
- save durable investigation to `llm/research/` only when useful

### Phase 2 — Compare Approaches

- use `brainstorm`
- compare options against live architecture
- choose smallest safe path

### Phase 3 — Lock Design

- use `design` if architecture or contract choices matter
- record route, contract, dependency, and side-effect implications

### Phase 4 — Write Execution Plan

- use `plan`
- write staged work to `llm/tasks/todo.md` or `llm/plans/`
- include per-step verification and ownership

### Phase 5 — Execute By Boundary

Use exact boundary skills as needed:

- backend logic: `go-service`
- route or contract: `api-endpoint`
- auth or tenant: `auth-tenant-casbin`
- API-key: `api-key-scope`
- cross-stack contract: `frontend-surface` or `cross-stack-change`
- storage, worker, webhook, stats, user, role, access, permission, project: matching domain skill

### Phase 6 — Verify And Review

- use `verification-before-completion`
- use `self-review`
- report exact commands, exact result, and exact coverage gaps

## Artifact Rules

- active task state: `llm/tasks/todo.md`
- reusable research: `llm/research/`
- durable plans: `llm/plans/`
- non-urgent follow-up: `llm/recommendations/`

## Common Mistakes

- jumping to implementation before clarifying owner boundary
- mixing feature delivery with unrelated cleanup
- changing backend contract without frontend proxy/type audit

## Stop Conditions

- stop if route ownership, tenant boundary, or auth stratum is unclear
- stop if live code contradicts cache/docs in way that changes feature design
- stop before destructive DB/schema/data operations not explicitly requested

## Completion Output

Report:

- feature phase reached
- artifacts created or updated
- skills used by phase
- verification state
- next step or blocker
