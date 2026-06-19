---
name: receiving-code-review
description: Use when addressing reviewer comments on an existing Casbin repo patch or branch, especially when comments question architecture fit, security boundaries, missing verification, or over-broad changes.
---

# Receiving Code Review

## Overview

Review response in this repo should be smallest correct patch, not defensive rewrite.

Treat reviewer comments as signal about runtime risk, missing proof, or scope drift.

## Read Order

1. reviewer comment exactly as written
2. current diff or commit range
3. affected files in live code
4. prior verification output

## Workflow

### Step 1 — Classify Comment

Classify as one of:

- correctness bug
- architecture mismatch
- security or boundary risk
- missing verification
- style or readability cleanup

### Step 2 — Re-check Live Code

Do not trust reviewer or author memory alone.

Re-open exact router, middleware, module, proxy, or test path that comment references.

### Step 3 — Make Smallest Fix

- preserve product intent
- avoid unrelated cleanup
- add or tighten verification if comment reveals missing proof

### Step 4 — Respond With Evidence

State:

- what changed
- why this addresses comment
- what verification now covers

## Common Mistakes

- over-correcting beyond reviewer ask
- fixing style while leaving root concern open
- replying without rerunning affected verification

## Stop Conditions

- stop if reviewer concern depends on behavior not yet reproven in live code
- stop if requested fix would broaden scope beyond safe targeted response

## Completion Output

Report:

- comment classification
- files changed
- verification rerun
- anything intentionally left unchanged
