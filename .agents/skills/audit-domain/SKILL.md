---
name: audit-domain
description: Use when changing audit logs, audit outbox, or request side effects that emit audit data in the Casbin repo.
---

# Audit Domain: Audit Logs and Outbox

**Announce at start:** "I'm using the audit-domain skill to preserve audit visibility and outbox behavior."

## Read Order

1. `llm/cache/audit-system.md`
2. `internal/modules/audit/module.go`
3. `internal/modules/audit/usecase/audit_usecase.go`
4. `internal/modules/audit/repository/audit_repository.go`

## Workflow

### Phase 1 — Identify Direction

- log listing/querying
- outbox sync
- write-following side effect

### Phase 2 — Patch

- preserve persistence behavior
- preserve any request/worker coupling

### Phase 3 — Verify

- audit controller/usecase/repository tests
