---
name: webhook-domain
description: Use when changing webhook CRUD, event subscription behavior, webhook logs, or async webhook trigger dispatch in this Casbin repo.
---

# Webhook Domain

## Overview

Webhook domain combines tenant-scoped configuration with worker-owned async dispatch.

Changes must preserve organization scope, event semantics, and async delivery assumptions unless product intent explicitly changes.

## When To Use

Use this skill when:

- changing webhook create/update/delete/find flows
- changing subscribed events or trigger conditions
- changing webhook logs query behavior
- changing worker dispatch path for webhook delivery

## Read Order

1. `AGENTS.md`
2. `llm/cache/webhook-system.md`
3. `llm/cache/worker-audit-webhook-system.md`
4. `internal/router/router.go`
5. `internal/modules/webhook/module.go`
6. `internal/modules/webhook/delivery/http/webhook_routes.go`
7. `internal/modules/webhook/usecase/webhook_usecase.go`
8. `internal/worker/tasks/webhook.go`

## Runtime Truth To Preserve

- webhook routes are registered under `tenantAuthorized` group via `webhookHttp.RegisterWebhookRoutes(...)`
- webhook operations are organization-scoped
- trigger dispatch is async through worker path, not synchronous inline delivery

## Workflow

### Step 1 — Classify Webhook Concern

State whether change affects:

- CRUD/config
- event subscription set
- trigger dispatch
- delivery log visibility

### Step 2 — Review Organization Scope

Confirm webhook operation still respects:

- tenant organization context
- route protection in tenantAuthorized path
- event filtering within organization ownership

### Step 3 — Patch Async Behavior Carefully

- preserve worker-based dispatch unless product intent explicitly changes
- preserve trigger side effect timing expectations
- keep controller thin and usecase-owned orchestration

### Step 4 — Verify

Start with:

- `internal/modules/webhook/test/*`

Add worker-path verification when dispatch semantics change.

## Common Mistakes

- turning async trigger into sync call accidentally
- weakening organization scoping on list, logs, or trigger paths
- changing event names or filters without tracing producers and consumers

## Stop Conditions

- stop if dispatch behavior changes but worker consumer path is not reviewed
- stop if webhook event semantics change and upstream trigger sources are not traced

## Completion Output

Report:

- webhook surface changed
- async and org-scope implications
- files changed
- verification run and exact result
