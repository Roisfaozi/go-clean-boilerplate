---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing in this Casbin repo, especially after changes to auth, tenant, routes, contracts, workers, uploads, frontend proxies, or other multi-layer behavior.
---

# Verification Before Completion

## Overview

This skill blocks false-success handoff.

Before saying fixed or complete, prove what was checked, what was not checked, and why coverage is sufficient for changed layers.

## Iron Law

No evidence, no success claim.

## Read Order

1. current diff
2. changed files grouped by layer
3. prior commands and verification output
4. relevant `llm/cache/*` only when needed to confirm expected boundary behavior

## Verification Gate

Before final response, answer:

1. what files changed?
2. which layers changed?
3. what is narrowest meaningful verification?
4. were producer and consumer both checked when contract moved?
5. what could not run and why?

## Layer Matrix

- Go backend package logic
  - targeted package tests first
  - `pnpm go:test` or `make test-unit` when broader package scope needed
- auth, tenant, Casbin, API-key, upload, worker, DB integration
  - integration or focused end-to-end checks when local unit proof is insufficient
- frontend web
  - `pnpm --filter casbin-web typecheck`
  - build or lint as relevant
- frontend client
  - `pnpm --filter casbin-client typecheck`
  - build or E2E as relevant
  - do not overstate placeholder lint script
- Swagger-visible API
  - `pnpm go:docs` when docs artifacts should change

## Verification Workflow

### Step 1 — Classify Blast Radius

State whether change touched:

- backend only
- frontend only
- cross-stack contract
- auth or permission boundary
- worker or webhook side effect
- upload or realtime path
- docs or agent context only

### Step 2 — Pick Narrowest Real Check

Start smallest check that proves changed behavior.

If smallest check is blocked, state blocker and choose best available fallback.

### Step 3 — Expand Only When Needed

Widen verification when:

- producer and consumer changed
- auth or tenant path moved
- DB/Casbin/worker side effect matters
- route lifecycle or proxy behavior changed

### Step 4 — Report Limits Honestly

Say exact blocker for anything skipped:

- Docker not available
- backend not running
- fixture missing
- browser/manual path not executed
- tool missing or script placeholder

## Red Flags

- claiming pass after only reading code
- hiding Docker, Snap, network, or env blockers
- treating unrelated red tests as fixed
- skipping frontend consumer check after backend contract change
- skipping integration after tenant, Casbin, API-key, upload, or worker boundary change

## Common Mistakes

- reporting commands without saying what they proved
- running only broad suite and missing closest targeted check
- claiming consumer unaffected without checking proxy or shared type path
- hiding that verification was partial

## Stop Conditions

- stop if changed layer has no verification story at all
- stop if success claim would depend on unrun critical check without explicit disclosure

## Completion Output

Report:

- files changed
- layers changed
- commands run and exact result
- skipped verification and blocker
- residual risk tied to unverified path
