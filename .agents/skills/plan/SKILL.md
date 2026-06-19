---
name: plan
description: Use when work needs a staged implementation plan, scoped checklist, or approval gate before editing the Casbin repo.
---

# Plan: Casbin Implementation Planning

**Announce at start:** "I'm using the plan skill to create a staged, verifiable implementation plan."

## Read Order

1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/workflows/*`
4. live code owner paths

## Plan Shape

Every plan must include:

- scope and non-scope
- file list
- route/module owner
- risk boundaries
- stage order
- verification per stage
- stop conditions

## Task Granularity

Each task should be:

- one coherent behavior change
- independently verifiable
- small enough to review

## Writing Rules

- Do not write vague steps like "fix backend".
- Prefer exact file paths and exact commands.
- Add review notes or blocker notes when uncertain.
- Save active state to `llm/tasks/todo.md`.

## Where To Save

- active work: `llm/tasks/todo.md`
- durable staged plan: `llm/plans/`
- future improvement: `llm/recommendations/`

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
