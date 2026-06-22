---
name: pr-explainer
description: Use when explaining a Casbin repo change set to reviewer, especially when backend routing, auth boundaries, domain rules, cross-stack contract changes, or verification scope need clear reviewer-facing framing.
---

# PR Explainer

## Overview

Reviewer explanation must map diff back to runtime behavior.

Do not dump file list only. Explain problem, solution, proof, and residual risk in terms reviewer can verify quickly.

## Read Order

1. current diff
2. recent commits if change is split
3. relevant plan or task note
4. exact verification results

## Workflow

### Step 1 — Group By Runtime Concern

Group files by concern such as:

- cache/docs context
- router or middleware
- controller/usecase/repository
- frontend proxy or shared types
- verification or tests

### Step 2 — Explain Four Things

For each concern, explain:

- problem before patch
- solution introduced
- evidence or verification run
- remaining risk or explicit non-goal

### Step 3 — Keep It Reviewer-Facing

- reference exact files and boundaries
- call out auth, tenant, Casbin, API-key, worker, or transaction risk when relevant
- distinguish verified behavior from unverified assumptions

## Good Output Shape

- summary
- files/areas changed
- verification
- risk or follow-up

## Common Mistakes

- restating commit message only
- hiding skipped verification
- not mentioning affected runtime boundary

## Stop Conditions

- stop if verification result is unknown or not rerun after requested change
- stop if reviewer-facing summary would need claims not proven by diff or command output

## Completion Output

Report reviewer-ready summary with:

- problem
- solution
- verification
- residual risk
