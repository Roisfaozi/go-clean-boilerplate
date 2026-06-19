---
name: webhook-domain
description: Use when changing webhook CRUD, subscription filters, trigger dispatch, or webhook logs in the Casbin repo.
---

# Webhook Domain: Trigger Dispatch and Logs

**Announce at start:** "I'm using the webhook-domain skill to preserve webhook CRUD, trigger, and async dispatch behavior."

## Read Order

1. `llm/cache/webhook-system.md`
2. `internal/modules/webhook/module.go`
3. `internal/modules/webhook/usecase/webhook_usecase.go`
4. `internal/worker/tasks/webhook.go`

## Workflow

### Phase 1 — Classify Change

- CRUD
- trigger event routing
- log retrieval
- worker payload behavior

### Phase 2 — Patch

- preserve async dispatch intent
- preserve org scope
- keep payload compatible with worker

### Phase 3 — Verify

- webhook usecase/controller tests
- worker webhook task tests if payload changes
