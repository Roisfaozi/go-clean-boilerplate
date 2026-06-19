---
name: systematic-debugging
description: Use when encountering a bug, failing test, suspicious behavior, integration failure, auth/tenant issue, or runtime mismatch in the Casbin repo.
---

# Systematic Debugging: Root Cause Before Fix

**Announce at start:** "I'm using systematic-debugging; root cause before patch."

## Iron Rule

No fix before reproducing or tracing the exact failing path.

## Four Phases

### Phase 1 — Evidence

- capture exact error/test/log/request
- identify changed layer
- read relevant cache and workflow

### Phase 2 — Trace

Trace through live code:

- router/middleware for HTTP behavior
- controller/usecase/repository for module behavior
- worker/distributor/handler for async behavior
- proxy/client/component for frontend behavior

### Phase 3 — Hypothesis

Write 1-3 hypotheses and expected evidence.

### Phase 4 — Patch

- patch minimal root cause
- add regression test if nearby pattern exists
- verify narrow path first

## Stop Signals

- cannot reproduce or trace
- multiple possible root causes remain
- environment blocker hides result
- fix requires product/security decision
