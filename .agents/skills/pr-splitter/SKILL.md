---
name: pr-splitter
description: Use when current Casbin repo branch or diff should be split into smaller focused change groups, especially when docs, cache, skills, backend runtime, frontend, or verification updates should be reviewed independently.
---

# PR Splitter

## Overview

Split by independent reviewer concern, not by arbitrary file count.

Good split makes runtime risk, docs drift, and verification easier to review.

## Read Order

1. current diff
2. commit history on branch if present
3. user constraints about ordering or exclusions

## Workflow

### Step 1 — Find Independent Concerns

Common split axes in this repo:

- runtime backend changes
- frontend or shared type sync
- cache or workflow documentation
- skill or prompt system updates
- tests and verification scaffolding

### Step 2 — Check Dependencies

Only split when concern can stand on its own without hiding required context.

Call out if one group depends on another.

### Step 3 — Propose Ordered Sequence

For each proposed PR or commit group, provide:

- purpose
- file set
- dependency on earlier group

### Step 4 — Avoid Unsafe Git Advice

Do not suggest destructive history rewrite or cleanup without explicit approval.

## Common Mistakes

- splitting files that must be reviewed together
- separating contract change from consumer update
- mixing generated docs and runtime fix in same reviewer narrative when they can stand alone

## Stop Conditions

- stop if proposed split would hide required context for correctness review
- stop if safe split depends on destructive git history operations not approved by user

## Completion Output

Report:

- proposed groups in order
- why each group is independent
- dependencies or blockers for safe split
