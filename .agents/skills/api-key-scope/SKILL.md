---
name: api-key-scope
description: Use when changing protected endpoints, API-key authentication, API-key scopes, organization-scoped API keys, or route access behavior involving API keys.
---

# API Key Scope: Protected Route Capability

**Announce at start:** "I'm using the api-key-scope skill to preserve protected endpoint scope behavior."

## Read Order

1. `llm/cache/domain-rules.md`
2. `llm/workflows/api-endpoint.md`
3. `internal/router/router.go`
4. `internal/middleware/api_key_middleware.go`
5. `internal/modules/api_key`
6. affected route/controller tests

## Workflow

### Phase 1 — Route Group

Identify if route is `authenticated`, `tenantAuthorized`, or `authorized`.

### Phase 2 — Scope Decision

Decide:

- no API-key access
- auto scope derived from route
- explicit scope required
- organization-scoped key behavior

### Phase 3 — Patch

- Preserve API-key identity and scope separation.
- Preserve org-scoped behavior and Redis use.
- Keep API-key checks in middleware/router boundary.

### Phase 4 — Verification

- Test allowed key.
- Test missing key/token.
- Test wrong scope.
- Test wrong organization when route is tenant-scoped.

## Red Flags

- accepting identity without scope
- adding route to protected group without scope decision
- checking scope ad hoc in controller while middleware supports it

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
