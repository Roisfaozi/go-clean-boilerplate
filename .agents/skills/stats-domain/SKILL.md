---
name: stats-domain
description: Use when changing dashboard stats, activity, insights, or metrics broadcaster behavior in the Casbin repo.
---

# Stats Domain: Dashboard and Metrics

**Announce at start:** "I'm using the stats-domain skill to preserve dashboard stats and broadcaster behavior."

## Read Order

1. `llm/cache/stats-system.md`
2. `internal/modules/stats/module.go`
3. `internal/config/app.go`
4. `internal/router/router.go`

## Workflow

### Phase 1 — Surface Check

- dashboard endpoint change
- data aggregation change
- broadcaster side-effect change

### Phase 2 — Patch

- keep stats usecase isolated
- keep broadcaster timing/side effect awareness explicit

### Phase 3 — Verify

- stats usecase tests
- router-authenticated route checks
- broadcaster behavior review if metrics payload changes
