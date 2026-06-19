---
name: role-domain
description: Use when changing roles, role validation, role-policy cleanup, or permission orchestration in the Casbin repo.
---

# Role Domain: Role and Permission Cleanup

**Announce at start:** "I'm using the role-domain skill to preserve role and permission orchestration."

## Read Order

1. `llm/cache/role-system.md`
2. `llm/cache/casbin-permission-system.md`
3. `llm/cache/access-right-system.md`
4. `internal/modules/role/module.go`
5. `internal/modules/permission/usecase/permission_usecase.go`

## Workflow

### Phase 1 — Identify Role Effect

- role CRUD only
- validation/converter changes
- policy/permission cleanup changes

### Phase 2 — Cross-Module Check

- confirm permission usecase dependency is preserved
- confirm access-right semantics not broken
- confirm transactional cleanup if policy state changes

### Phase 3 — Patch

- keep model conversion tests aligned
- keep controller thin
- keep role policy cleanup in permission flow

### Phase 4 — Verify

- role repository/usecase/controller tests
- permission security tests if policy cleanup changes
