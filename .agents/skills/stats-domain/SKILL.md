---
name: stats-domain
description: Use when changing dashboard stats endpoints, stats aggregation logic, or realtime metrics broadcasting behavior in this Casbin repo.
---

# Stats Domain

## Overview

Stats domain spans HTTP dashboard endpoints and periodic realtime broadcasting.

Changes that look like harmless query tweaks can affect background load or websocket payload behavior.

## When To Use

Use this skill when:

- changing summary, activity, or insights endpoints
- changing stats usecase aggregation
- changing metrics broadcaster inputs or payload expectations

## Read Order

1. `AGENTS.md`
2. `llm/cache/stats-system.md`
3. `llm/cache/realtime-system.md`
4. `internal/router/router.go`
5. `internal/config/app.go`
6. `internal/modules/stats/module.go`
7. `internal/modules/stats/usecase/stats_usecase.go`
8. `internal/modules/stats/delivery/http/stats_controller.go`

## Runtime Truth To Preserve

- stats routes live under `authenticated` group in `internal/router/router.go`
- exposed endpoints are `/api/v1/stats/summary`, `/activity`, and `/insights`
- `internal/config/app.go` uses stats usecase in realtime metrics broadcasting path

## Workflow

### Step 1 — Classify Surface

State whether change affects:

- HTTP contract
- aggregation query logic
- broadcaster cadence or payload shape
- both HTTP and realtime consumers

### Step 2 — Review Cost And Scope

Check for:

- expensive queries in periodic path
- coupling between HTTP response model and websocket metrics payload
- authenticated-route assumptions

### Step 3 — Patch Carefully

- preserve authenticated route boundary
- avoid unbounded query growth
- keep controller thin and aggregation logic in usecase

### Step 4 — Verify

Start with:

- `internal/modules/stats/test/stats_usecase_test.go`

Add broader verification if broadcaster behavior in `internal/config/app.go` changes.

## Common Mistakes

- tuning dashboard query without considering periodic broadcaster load
- assuming stats affects only HTTP routes
- moving expensive computation into controller path

## Stop Conditions

- stop if change affects broadcaster semantics but websocket consumer expectations are not reviewed
- stop if query cost could materially increase periodic workload and no validation path exists

## Completion Output

Report:

- stats surface changed
- realtime or load implications
- files changed
- verification run and exact result
