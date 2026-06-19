---
name: tus-upload-storage
description: Use when changing TUS upload handling, upload metadata, storage provider behavior, hook dispatch, avatar updates, or upload completion flows in this Casbin repo.
---

# TUS Upload Storage

## Overview

Upload path here is separate trust boundary, not standard JSON CRUD.

It spans route middleware, TUS handler, storage provider abstraction, metadata, and completion hooks that can update domain state like avatars.

## Read Order

1. `AGENTS.md`
2. `llm/cache/tus-upload-system.md`
3. `llm/cache/user-system.md` when avatar flow is involved
4. `internal/router/router.go`
5. `internal/config/app.go`
6. `pkg/tus/*`
7. `pkg/storage/*`
8. affected hook or usecase paths

## Workflow

### Step 1 — Classify Upload Concern

State whether change affects:

- upload route behavior
- metadata parsing
- storage provider behavior
- completion hook dispatch
- avatar or domain update after upload

### Step 2 — Preserve Trust Boundary

Confirm:

- upload route keeps auth and user-status middleware
- metadata is treated as trust-sensitive input
- storage behavior still uses configured provider abstraction

### Step 3 — Patch With Context Integrity

- preserve context propagation into storage and request-scoped operations
- keep TUS path separate from normal JSON controller assumptions
- if avatar flow changes, trace user-side hook and storage completion logic together

### Step 4 — Verify

- `pkg/tus/*_test.go` where relevant
- storage package tests if provider logic changed
- avatar or downstream hook tests when upload completion updates domain state

## Common Mistakes

- treating upload route as ordinary API endpoint
- weakening metadata validation or trust boundary
- changing provider behavior without checking completion hook effects

## Stop Conditions

- stop if upload auth/status boundary would change implicitly
- stop if storage provider and completion hook implications are not both traced

## Completion Output

Report:

- upload surface changed
- trust-boundary implications
- files changed
- verification run and exact result
