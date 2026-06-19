---
name: improve
description: Use when making reviewer-facing improvements, cleanup, hardening, or maintainability upgrades in this Casbin repo without changing product intent.
---

# Improve

## Overview

This skill is for scoped improvement, not stealth feature work.

Use it to tighten clarity, safety, maintainability, or repo-agent context while preserving runtime intent.

## When To Use

Use this skill when:

- cleaning up one module or boundary
- hardening maintainability or security posture without changing product behavior
- improving docs, workflows, skills, or agent context
- reducing reviewer pain in an existing area

Do not use this skill when:

- request is new feature; use `feature`
- root cause of bug is not yet proven; use `systematic-debugging`
- route or API contract change dominates; use `api-endpoint`

## Read Order

1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/conventions/*`
4. closest live code precedent
5. current diff or reviewer comment if improvement is follow-up

## Workflow

### Step 1 — State Pain Precisely

Capture:

- exact pain
- where it lives
- why it matters to correctness, reviewability, or maintenance

### Step 2 — Confirm Owner Boundary

Identify whether pain belongs to:

- module boundary
- route protection boundary
- frontend ownership boundary
- docs or agent-context boundary

### Step 3 — Make Smallest Useful Improvement

- avoid unrelated refactors
- keep product intent stable
- prefer local hardening over broad redesign

### Step 4 — Save Durable Follow-Up Only When Needed

- `llm/recommendations/` for non-urgent follow-up
- `llm/tasks/` for active task notes

### Step 5 — Verify Narrowly

Run smallest check that proves improvement did not regress behavior.

## Common Mistakes

- turning cleanup into architecture rewrite
- changing behavior while claiming maintainability-only work
- improving docs or skills without rechecking live code truth

## Stop Conditions

- stop if improvement request is actually hidden feature work
- stop if owner boundary is unclear
- stop before destructive DB/schema/data operations not explicitly requested

## Completion Output

Report:

- what improved
- why it mattered
- files changed
- verification run and exact result
- non-urgent follow-up ideas
