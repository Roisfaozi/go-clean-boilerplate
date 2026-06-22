---
name: writing-skills
description: Use when creating, rewriting, or auditing agent skills for this Casbin repo, especially when existing skills are too generic, too short, missing repo-specific paths or commands, or not strong enough to change agent behavior reliably.
---

# Writing Skills

## Overview

This repo needs skills that change agent behavior, not thin placeholders.

Good Casbin skills are short enough to scan but detailed enough to force correct routing, file reads, decision points, verification, and final output.

Do not treat `llm/cache/*` as substitute for skill content. Cache stores facts. Skill must encode how future agents should act on those facts.

## What Good Looks Like In This Repo

Every rewritten skill should usually contain:

- clear trigger in frontmatter `description`
- short overview with repo-specific purpose
- when-to-use and when-not-to-use boundaries
- exact read order with real repo paths
- step-by-step workflow tied to Casbin architecture
- decision rules or checklists for risky branches
- exact verification command surfaces that exist in repo
- stop conditions when architecture truth is unclear
- completion output contract for handoff

If those are missing, skill is probably still placeholder-grade.

## Required Read Order Before Editing Skills

1. `AGENTS.md`
2. `llm/cache/project-overview.md`
3. `llm/cache/environment.md`
4. `llm/cache/architecture.md`
5. relevant `llm/cache/*` domain files for target skill
6. relevant `llm/workflows/*`
7. live code entrypoints and module paths referenced by skill

## Skill Writing Workflow

### Step 1 — Audit Current Skill

Check whether current `SKILL.md` has only any of these weak patterns:

- generic 3-step workflow
- no repo paths
- no verification section
- no stop conditions
- no distinction between similar skills
- no output contract
- no mention of `apps/web`, `apps/client`, `internal/router/router.go`, `internal/config/app.go`, or target module paths when they matter

If yes, rewrite instead of lightly editing.

### Step 2 — Prove Trigger Boundaries

Frontmatter `description` must say when skill should be used.

It must not summarize the whole workflow.

Good:

- `Use when changing Go backend module logic, usecases, repositories, module constructors, worker-owned backend behavior, or dependency wiring in the Casbin backend.`

Bad:

- `Use for backend changes by reading cache, editing usecases, testing packages, then reporting output.`

### Step 3 — Encode Repo-Specific Behavior

Put operational rules in skill body, not only in cache:

- real composition roots: `cmd/api/main.go`, `internal/config/app.go`, `internal/router/router.go`
- active frontend surfaces: `apps/web`, `apps/client`
- shared API boundary: `packages/api-types`
- high-risk middleware boundaries: auth, tenant, API key, Casbin, upload, worker
- confirmed commands from `AGENTS.md`, `package.json`, and `Makefile`

Future agent should know what to read, what to compare, and when to stop.

### Step 4 — Add Decision Structure

For most Casbin skills, include at least one of:

- when-to-use vs when-not-to-use table
- route or layer decision matrix
- review checklist
- common mistakes list
- stop conditions

This is what turns reference into behavior.

### Step 5 — Make Verification Real

Only list commands that exist in this repo.

Prefer narrow-to-broad verification:

- package or module tests first
- `pnpm go:test` or `make test-unit` for backend package changes
- `pnpm typecheck`, app-specific typecheck, or `pnpm build` when frontend/shared types move
- integration/E2E only when boundary crossing really changed

If Docker, Snap Go, browser auth, or infrastructure is blocker, say exact blocker.

### Step 6 — Make Completion Output Explicit

Each skill should end with required handoff output such as:

- files changed
- commands run and result
- skipped verification and blocker
- residual risk or next step

## Casbin Skill Quality Checklist

- [ ] trigger is specific and discovery-friendly
- [ ] body distinguishes this skill from neighboring skills
- [ ] read order uses real repo files only
- [ ] workflow reflects live code architecture, not template stack
- [ ] verification names real commands and realistic scope
- [ ] no placeholder text like TBD, appropriate file, similar to X
- [ ] no foreign stack assumptions from Carbon or other repos
- [ ] output contract tells future agent how to report work

## Common Mistakes

- copying Carbon structure without swapping in Casbin runtime truth
- keeping skill tiny because cache exists
- naming files but not saying why each file matters
- listing commands without when-to-run guidance
- omitting stop conditions for auth, tenant, Casbin, API-key, or transaction risk
- writing domain skill without module-specific paths

## When To Split Supporting Files

Keep single-file skill when guidance fits and stays scannable.

Split into helper docs only when one of these is true:

- reference examples become large
- repeated command matrices deserve separate maintenance
- domain rules are stable and reused by many skills

Even then, `SKILL.md` must still stand on its own as operational entrypoint.

## Completion Output

Report:

- files audited or rewritten
- why old skill was inadequate
- repo-specific behavior added
- remaining weak skills for next batch
