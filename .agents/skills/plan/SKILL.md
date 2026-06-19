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

## Plan Format

Every plan must include:

- scope and non-scope
- files likely touched
- high-risk boundaries
- task order
- verification per task
- stop conditions

## Task Granularity

Good task:

- 1 coherent behavior change
- clear files
- clear verification
- can be reviewed independently

Bad task:

- "fix backend"
- "update UI"
- "clean up things"

## Where To Save

- active work: `llm/tasks/todo.md`
- durable staged plan: `llm/plans/`
- future improvement: `llm/recommendations/`
