# Developer Flow

Practical reading and execution order for future developers.

## 1. Start Here

Read in this order before making changes:

1. `documentation/README.md`
2. `documentation/architecture/ARCHITECTURE.md`
3. `documentation/architecture/MULTI_TENANCY.md`
4. `documentation/architecture/ARCHITECTURE_VISUAL_AND_SEQUENCE.md`
5. `documentation/guides/GETTING_STARTED.md`
6. `documentation/guides/TESTING.md`

## 2. If You Change Backend API

Read:

- `documentation/guides/API_ACCESS_WORKFLOW.md`
- `documentation/guides/API_USAGE.md`
- `documentation/guides/ACCESS_RIGHTS_REFERENCE.md`

Then verify against live code:

- `internal/router/router.go`
- `internal/config/app.go`
- target module under `internal/modules/*`

## 3. If You Change Tenant Or Auth Behavior

Read:

- `documentation/architecture/MULTI_TENANCY.md`
- `documentation/guides/API_ACCESS_WORKFLOW.md`

Then verify against live code:

- `internal/middleware/auth_middleware.go`
- `internal/middleware/tenant_middleware.go`
- `internal/middleware/casbin_middleware.go`

## 4. If You Change Realtime

Read:

- `documentation/guides/REALTIME.md`
- `documentation/guides/PRESENCE_API.md`

Then verify against live code:

- `pkg/ws`
- `pkg/sse`
- `internal/router/router.go`

## 5. If You Change Upload Or Storage

Read:

- `documentation/guides/RESUMABLE_UPLOAD.md`
- `documentation/guides/CLIENT_UPLOAD_GUIDE.md`
- `documentation/guides/STORAGE.md`

Then verify against live code:

- `pkg/tus`
- `pkg/storage`
- target module using upload hooks

## 6. If You Change Frontend

Read:

- `documentation/guides/FRONTEND_STRUCTURE.md`

Then verify owner first:

- `apps/web`
- `apps/client`
- `packages/*`

## 7. Verification Order

1. narrowest package check
2. affected app `typecheck`
3. affected app `lint`
4. integration tests if DB, Redis, tenant, worker, upload, or Casbin touched
5. E2E if user flow changed

## 8. Planning Docs Rule

- `documentation/ops/*` = planning or migration notes
- `documentation/productplan/*` = backlog, vision, UX, wireframes
- do not treat them as runtime truth without checking live code
