---
name: research
description: Use when researching technical or product approaches for this Casbin repo before design or implementation, especially when live code precedents, module ownership, security boundaries, or workflow tradeoffs are unclear.
---

# Research

## Overview

Research in this repo means evidence first.

Prefer live code precedents, existing module patterns, command surfaces, and verified docs over abstract best-practice summaries.

## When To Use

Use this skill when:

- architecture choice is unclear
- two modules seem to own same concern
- auth, tenant, Casbin, API-key, worker, upload, or realtime behavior needs precedent search
- user asks for comparison, audit, or options before code changes

Do not use this skill when:

- implementation path is already obvious from target module
- task is direct patch with known owner and bounded blast radius

## Read Order

1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/workflows/*`
4. `internal/config/app.go` when composition matters
5. `internal/router/router.go` when route or middleware matters
6. target module code and adjacent precedent modules
7. tests near affected code
8. supporting docs only after runtime code

## Research Workflow

### Step 1 — Define Questions

List exact questions such as:

- which module owns behavior?
- how is tenant resolved today?
- where is Casbin policy enforced and written?
- what request contract do frontend consumers rely on?
- what command or test surface can verify this safely?

### Step 2 — Search Local Precedent

Look for:

- same pattern in another `internal/modules/*`
- same middleware layering in `internal/router/router.go`
- same dependency wiring in `internal/config/app.go`
- same frontend proxy or type shape in `apps/web`, `apps/client`, `packages/api-types`

### Step 3 — Separate Facts From Recommendations

Research output must clearly separate:

- verified fact from live code
- inferred behavior from multiple files
- recommendation or preferred approach
- `needs confirmation` for anything not provable locally

### Step 4 — Save Durable Output

Save to:

- `llm/research/` for reusable findings
- `llm/tasks/` for active audit notes
- `llm/recommendations/` for non-urgent follow-up

Do not move unstable findings into `llm/cache/`.

## Required Output Structure

- question
- evidence
- current behavior
- options
- recommendation
- verification implications

## Common Mistakes

- mixing opinion with verified runtime truth
- treating old prompt/workflow docs as source of truth
- missing existing module precedent and inventing new pattern
- forgetting both frontend surfaces are active

## Stop Conditions

- stop if claimed behavior cannot be backed by file path or command evidence
- stop if route or module ownership remains ambiguous after code search
- stop if requested conclusion would require internet or external system confirmation not available here

## Completion Output

Report:

- research questions answered
- file paths used as evidence
- clear facts vs recommendations
- unresolved items marked `needs confirmation`
