---
name: systematic-debugging
description: Use when encountering a bug, failing test, suspicious behavior, integration failure, auth or tenant issue, async side-effect mismatch, or runtime contradiction in this Casbin repo.
---

# Systematic Debugging

## Overview

This skill is for disciplined root-cause isolation.

Use it when bug surface may cross router, middleware, module layers, proxy, worker, upload, or querybuilder boundaries and a fast guess would be risky.

## Iron Rule

No fix before reproducing, tracing, or otherwise proving exact failing path.

## When To Use

Use this skill when:

- test fails but failure source is unclear
- request lifecycle behaves differently than expected
- auth, tenant, Casbin, API-key, upload, worker, or webhook behavior seems wrong
- docs or cache say one thing and runtime says another

If bug is already localized and proven, use narrower domain skill for patching.

## Read Order

1. exact bug report, failing test, or reproduction steps
2. `AGENTS.md`
3. relevant `llm/cache/*`
4. relevant `llm/workflows/*`
5. `internal/router/router.go` for HTTP path issues
6. `internal/config/app.go` when wiring, async workers, or broadcasters may matter
7. target module, proxy, or package tests

## Debug Workflow

### Phase 1 — Evidence Capture

Collect exact evidence:

- failing command or test
- exact request path and actor
- exact log or response mismatch
- whether issue is backend, frontend, proxy, worker, or storage-facing

### Phase 2 — Trace Live Path

Trace through only relevant layers:

- router and middleware for HTTP behavior
- controller, usecase, repository for module behavior
- proxy, loader, component, or hook for frontend behavior
- worker distributor, task, processor, handler for async behavior
- TUS/storage path for upload behavior

### Phase 3 — Write Hypotheses

Write 1-3 plausible root-cause hypotheses.

For each hypothesis, state what evidence would confirm or reject it.

### Phase 4 — Narrow Reproduction

Prefer smallest proof:

- package test
- focused request replay
- one browser path
- one worker handler path
- one proxy request path

### Phase 5 — Patch Minimal Root Cause

- fix only proven root cause
- remove temporary instrumentation unless intentionally retained
- add regression test when practical seam exists

### Phase 6 — Verify Narrow Then Adjacent

- rerun exact failing reproduction first
- only then widen to adjacent checks

## Repo-Specific Hot Zones

- JWT parse success vs Redis-backed session validity
- tenant resolution and membership cache behavior
- API-key scope or user-session mismatch on protected routes
- Casbin route group mismatch
- worker or webhook side effect timing
- querybuilder field restrictions
- upload metadata and completion hooks
- frontend proxy preserving cookies and auth headers

## Common Mistakes

- patching from intuition before exact trace
- broad speculative fixes across multiple modules
- stopping at first plausible explanation without proof
- treating environment blocker as proof of code bug

## Stop Conditions

- stop if bug cannot be reproduced or traced enough to isolate likely owner
- stop if multiple root causes remain and no new evidence is available
- stop if environment blocker hides result and cannot be separated from code issue
- stop if fix requires product or security decision beyond technical correction

## Completion Output

Report:

- reproduction method
- traced runtime path
- root cause proven
- files changed
- commands run and exact result
- residual uncertainty or follow-up
