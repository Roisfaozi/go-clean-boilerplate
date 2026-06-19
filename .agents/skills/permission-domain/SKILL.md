---
name: permission-domain
description: Use when changing permission policy CRUD, role/user assignment, access-right expansion, or transactional Casbin behavior in the Casbin repo.
---

# Permission Domain: Policy and Assignment

**Announce at start:** "I'm using the permission-domain skill to preserve policy and assignment semantics."

## Read Order

1. `llm/cache/permission-system.md`
2. `llm/cache/access-right-system.md`
3. `llm/cache/casbin-permission-system.md`
4. `internal/modules/permission/module.go`
5. `internal/modules/permission/usecase/permission_usecase.go`

## Workflow

### Phase 1 — Classify Change

- policy CRUD
- role/user assignment
- batch check
- access-right expansion
- transactional enforcer behavior

### Phase 2 — Patch

- preserve Casbin transaction semantics
- preserve access-right and inheritance behavior
- keep controller thin

### Phase 3 — Verify

- permission usecase tests
- security tests for Casbin errors/failure paths
