---
name: project-domain
description: Use when changing tenant-scoped project CRUD or project-specific permissions in the Casbin repo.
---

# Project Domain: Tenant-Scoped Project CRUD

**Announce at start:** "I'm using the project-domain skill to preserve project tenant scope and API-key scope rules."

## Read Order

1. `llm/cache/project-system.md`
2. `llm/cache/api-key-system.md`
3. `llm/cache/tenant-organization-system.md`
4. `internal/modules/project/module.go`
5. `internal/router/router.go`

## Workflow

### Phase 1 — Scope Check

- confirm route is under tenantAuthorized group
- confirm API-key scope string(s) in router
- confirm org context is required

### Phase 2 — Patch

- keep project controller thin
- keep repository isolated
- keep tenant-specific permission behavior explicit

### Phase 3 — Verify

- route access tests
- scope tests
- tenant path tests if project access changes
