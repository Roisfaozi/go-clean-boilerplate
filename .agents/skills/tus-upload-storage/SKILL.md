---
name: tus-upload-storage
description: Use when changing TUS upload handling, upload metadata, storage provider behavior, hook dispatch, avatar updates, or upload completion flows.
---

# TUS Upload Storage: Upload Lifecycle Boundary

**Announce at start:** "I'm using the tus-upload-storage skill because upload metadata and hooks are high-risk boundaries."

## Read Order

1. `llm/cache/domain-rules.md`
2. `llm/cache/backend-map.md`
3. `internal/config/app.go`
4. `pkg/tus`
5. `pkg/storage`
6. upload route in `internal/router/router.go`
7. upload/storage tests

## Workflow

### Phase 1 — Trace Upload Route

- TUS route is separate from normal JSON CRUD.
- Confirm auth/status middleware and upload handler wiring.
- Confirm storage provider: local/S3-compatible config.

### Phase 2 — Metadata Boundary

- Identify upload type metadata.
- Validate completion hook routing.
- Confirm avatar/user profile update path if relevant.

### Phase 3 — Patch Rules

- Preserve context propagation into storage operations.
- Do not trust upload metadata without existing validation pattern.
- Keep hook side effects intentional and idempotent where possible.

### Phase 4 — Verify

- storage provider tests
- TUS upload tests
- integration checks when route lifecycle or hooks changed

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
