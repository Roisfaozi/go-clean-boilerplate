---
name: feature
description: End-to-end feature development orchestrator for the Casbin Go plus TypeScript monorepo. Use when building new backend/frontend/cross-stack capability from scratch.
---

# Feature: Casbin End-to-End Development Orchestrator

**Announce at start:** "I'm using the feature skill to orchestrate research, design, plan, implementation, and verification."

## Pipeline

`research -> brainstorm -> plan -> execute -> verification-before-completion -> self-review`

## When To Use

| Use this skill when...                   | Use individual skills when...         |
| ---------------------------------------- | ------------------------------------- |
| feature spans multiple files or layers   | change is one small bugfix            |
| requirements need design before code     | design/plan already exists            |
| backend/frontend/API contract may change | pure module internals -> `go-service` |

## Phase 1 — Research

- Read `AGENTS.md` and relevant `llm/cache/*` first.
- Inspect live code before trusting docs.
- Save durable investigation to `llm/research/[feature].md` only if useful.

## Phase 2 — Brainstorm

- Identify module owner and high-risk boundaries.
- Compare 2-3 approaches.
- Choose approach with smallest safe boundary.

## Phase 3 — Plan

- Write staged plan to `llm/tasks/todo.md` for active work.
- Use `llm/plans/` for larger durable roadmap.
- Include verification per stage.

## Phase 4 — Execute

- Backend-first when data/contract changes.
- Use project-specific boundary skill when auth/tenant/Casbin/API-key/upload/worker/query/realtime touched.
- Keep changes small and reviewable.

## Phase 5 — Verify and Review

- Run `verification-before-completion`.
- Run `self-review`.
- Report exact commands and skipped checks.

## Artifacts

- active task state: `llm/tasks/todo.md`
- research: `llm/research/`
- non-urgent recommendations: `llm/recommendations/`
- durable staged plan: `llm/plans/`
