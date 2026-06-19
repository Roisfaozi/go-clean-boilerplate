---
name: improve
description: Use when making reviewer-facing improvements, cleanup, hardening, or maintainability upgrades in the Casbin repo without changing product intent.
---

# Improve: Casbin Hardening and Cleanup

**Announce at start:** "I'm using the improve skill to make a targeted, reviewer-facing improvement."

## Read Order

1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/conventions/*`
4. closest live code precedent

## When To Use

| Use this skill when...                | Use another skill when...                |
| ------------------------------------- | ---------------------------------------- |
| cleaning up one module or boundary    | change is new feature -> `feature`       |
| hardening security or maintainability | root cause bug -> `systematic-debugging` |
| improving docs or agent context       | API/route change -> `api-endpoint`       |

## Workflow

### Phase 1 — Pain Statement

State the exact pain, where it lives, and why it matters.

### Phase 2 — Boundary Check

Identify module, route group, or app surface owning the pain.

### Phase 3 — Improvement

Make smallest useful improvement with minimal blast radius.

### Phase 4 — Verification

Run the narrowest relevant check.

## Output

- what improved
- why it was worth doing
- any follow-up ideas to save into `llm/recommendations/`

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
