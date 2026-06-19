---
name: self-review
description: Use when Casbin repo changes are complete and need maintainability, correctness, security-boundary, and reviewer-facing quality review before handoff.
---

# Self Review

## Overview

Self-review is final code and scope audit before handoff.

It should catch layering leaks, security-boundary regressions, missing consumer sync, and over-broad changes before user or reviewer does.

## Review Dimensions

### Layering

- controller only parses, validates, and responds
- usecase owns business rules
- repository owns DB/query specifics
- router and middleware own exposure and protection boundaries

### Boundary Safety

- auth/session behavior remains consistent
- tenant resolution still comes from proper middleware/usecase path
- Casbin or API-key rules were not weakened accidentally
- worker, upload, or webhook side effects still align with request behavior

### Consumer Sync

- backend contract changes audited against frontend proxies and shared types
- docs or skills updated only when live code actually supports claim

### Scope Discipline

- no unrelated cleanup mixed in
- no silent behavior changes under maintainability label
- no placeholders or vague comments added

## Read Order

1. current diff
2. changed files in live code
3. relevant `llm/cache/*`
4. verification output already collected

## Review Workflow

### Step 1 — Re-scan By Concern

Group changed files by concern:

- runtime behavior
- route or middleware
- frontend proxy or shared types
- docs or skill context

### Step 2 — Ask Reviewer Questions Early

For each concern ask:

- does owner layer look right?
- is boundary protection still explicit?
- is contract drift possible?
- was verification proportional to risk?

### Step 3 — Fix Small Review Findings

Make tiny corrective patches only.

If review reveals design flaw, stop and reframe instead of papering over it.

## Common Mistakes

- reviewing only formatting and ignoring boundary drift
- missing producer/consumer mismatch after API change
- allowing unrelated file churn to survive
- assuming tests passing means review complete

## Stop Conditions

- stop if self-review reveals architectural issue bigger than safe final polish
- stop if success claim still depends on missing critical verification

## Completion Output

Report:

- review dimensions checked
- issues found and fixed
- remaining risk
- why handoff is or is not ready
