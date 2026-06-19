---
name: brainstorm
description: Use when exploring approaches for a Casbin feature, refactor, bug strategy, or architectural decision before planning or implementation, especially when module ownership, route strata, contract impact, or risky boundaries need comparison first.
---

# Brainstorm

## Overview

Brainstorm in this repo means compare implementation options against live architecture, not generic ideation.

Goal: pick smallest safe path that fits current router, middleware, module, and frontend ownership reality.

## When To Use

Use this skill when:

- request can be solved in multiple valid ways
- route owner or module owner is not obvious
- auth, tenant, Casbin, API-key, worker, upload, or frontend proxy boundary may be involved
- user wants options before planning or coding

Do not use this skill when:

- owner and patch path are already obvious
- direct bugfix only needs root-cause proof; use `systematic-debugging`

## Read Order

1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/workflows/*`
4. closest live code precedent
5. `internal/router/router.go` if route or middleware is involved
6. `internal/config/app.go` if wiring or worker dependencies matter

## Brainstorm Workflow

### Step 1 — Frame Problem

State:

- target outcome
- owning surface candidates
- risky boundaries
- explicit constraints from user or repo rules

### Step 2 — Build Option Set

Produce 2-3 real options only.

For each option include:

- likely files or modules touched
- route or app owner
- benefits
- risks
- verification needed

### Step 3 — Compare Against Repo Reality

Check each option against:

- existing route strata and middleware layering
- module boundaries already used in repo
- frontend proxy or shared type impact
- transaction or side-effect risk

### Step 4 — Recommend One Path

Recommend option with:

- smallest safe change
- strongest fit to live wiring
- lowest blast radius
- clearest verification path

## Common Mistakes

- proposing abstract option with no file or module owner
- ignoring frontend consumer sync on backend contract changes
- treating access-control risk as implementation detail to decide later

## Stop Conditions

- stop if live code contradicts prior assumptions about owner boundary
- stop if route ownership, tenant boundary, or auth stratum is unclear
- stop if useful comparison would require unverified external assumptions

## Completion Output

Report:

- options compared
- recommended path and why
- main risks rejected options carried
- next skill to use
