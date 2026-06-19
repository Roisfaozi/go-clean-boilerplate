---
name: access-domain
description: Use when changing access-right registry, endpoint definitions, or permission expansion logic in the Casbin repo.
---

# Access Domain: Access-Right Registry

**Announce at start:** "I'm using the access-domain skill to preserve access-right registry and permission semantics."

## Read Order

1. `llm/cache/access-right-system.md`
2. `llm/cache/casbin-permission-system.md`
3. `internal/modules/access/module.go`
4. `internal/modules/permission/usecase/access_right_assignment.go`

## Workflow

### Phase 1 — Registry Impact

- identify resource/action semantics
- identify who consumes access-right data

### Phase 2 — Patch

- keep access controller/usecase/repository boundaries intact
- keep permission expansion consistent

### Phase 3 — Verify

- access repository/usecase tests
- permission assignment tests when expansion changes
