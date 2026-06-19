---
name: user-domain
description: Use when changing user profile, status, avatar, list/search, registration, or user-related side effects in the Casbin repo.
---

# User Domain: User Lifecycle and Profile Surface

**Announce at start:** "I'm using the user-domain skill to preserve user, avatar, search, audit, and storage boundaries."

## Read Order

1. `llm/cache/user-system.md`
2. `llm/cache/querybuilder-security.md`
3. `llm/cache/tus-upload-system.md`
4. `llm/cache/worker-audit-webhook-system.md`
5. `internal/modules/user/module.go`
6. `internal/modules/user/delivery/http/user_controller.go`

## Workflow

### Phase 1 — Trace Feature Surface

Identify whether change touches:

- register/login-adjacent user data
- profile/self update
- avatar upload hook
- status/admin management
- dynamic search/list filtering

### Phase 2 — Boundary Checks

- validate with querybuilder security if list/search fields change
- validate storage/TUS if avatar or upload flow changes
- validate audit/webhook if user write triggers side effects

### Phase 3 — Patch

- keep controller thin
- keep user business rules in usecase
- keep persistence in repository
- keep storage and hook logic explicit

### Phase 4 — Verify

- controller/usecase/repository tests
- avatar hook tests if upload changes
- list/search tests if filter surface changes
