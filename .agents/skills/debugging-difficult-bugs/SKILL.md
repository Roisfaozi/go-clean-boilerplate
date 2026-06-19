---
name: debugging-difficult-bugs
description: Use early when debugging medium or hard bugs involving runtime state, ordering, persistence, concurrency, auth, tenant, or external services.
---

# Debugging Difficult Bugs

## Read First

- relevant cache and exact runtime path

## Workflow

1. Reproduce with tests or narrow code trace before editing.
2. Capture evidence from logs, payloads, DB state, or middleware layering.
3. Prove root cause before patching.
4. Remove temporary instrumentation unless intentionally kept.
