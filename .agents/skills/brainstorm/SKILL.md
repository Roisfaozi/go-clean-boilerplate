---
name: brainstorm
description: Use when exploring approaches for a Casbin feature, refactor, bug strategy, or architectural decision before planning or implementation.
---

# Brainstorm: Evidence-Based Design Options

**Announce at start:** "I'm using the brainstorm skill to compare options before choosing an implementation path."

## Read First

- relevant `llm/cache/*`
- relevant `llm/conventions/*`
- closest live code precedent

## Workflow

### Phase 1 — Frame Problem

- State goal.
- State constraints.
- State known high-risk boundaries.

### Phase 2 — Options

Produce 2-3 options. For each:

- files touched
- benefits
- risks
- verification needed
- impact on auth/tenant/Casbin/API-key/contracts if any

### Phase 3 — Recommendation

Pick one option based on:

- smallest safe change
- alignment with live wiring
- testability
- low blast radius

## Output

- recommended approach
- rejected alternatives and why
- questions/blockers if any

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
