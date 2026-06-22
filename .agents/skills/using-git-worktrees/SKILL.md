---
name: using-git-worktrees
description: Use when work in this Casbin repo needs isolation from current workspace state, especially if multiple task streams, risky refactors, or user-owned uncommitted changes make direct editing in current tree unsafe.
---

# Using Git Worktrees

## Overview

Isolation is useful when current workspace has unrelated changes or when risky work should not contaminate active tree.

In this repo, never use worktree idea as excuse to hide file ownership or discard user changes.

## When To Use

Use this skill when:

- current workspace has unrelated uncommitted changes
- user wants parallel feature or audit stream
- risky refactor should be isolated from stable branch work

Do not use this skill when:

- change is tiny and current workspace is already clean enough
- isolation would add more overhead than safety

## Read Order

1. `git status --short`
2. current task scope
3. branch and commit context if relevant

## Workflow

### Step 1 — Audit Current State

Classify current workspace:

- clean
- contains user-owned unrelated changes
- contains agent-generated in-progress work

### Step 2 — Decide Isolation Need

Recommend worktree only if it protects one of:

- user changes from accidental edits
- focused patch reviewability
- parallel implementation streams

### Step 3 — Explain Tradeoff

State:

- why current tree is risky
- what isolation would protect
- what commands or branch flow would be used

### Step 4 — Preserve User State

Never suggest destructive cleanup without explicit approval.

If worktree is not feasible in current harness or permissions, say so and propose safer same-tree alternative.

## Common Mistakes

- recommending worktree before checking actual workspace state
- treating worktree as default for every task
- forgetting user may have uncommitted files that must remain untouched

## Stop Conditions

- stop before any cleanup, checkout, reset, or deletion not explicitly approved
- stop if current harness or filesystem restrictions make worktree guidance unreliable
- stop if simpler same-tree isolation already solves risk with less operational cost

## Completion Output

Report:

- current workspace risk
- whether worktree is recommended
- safer fallback if not using worktree
