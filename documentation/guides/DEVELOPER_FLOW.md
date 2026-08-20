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

## 2. Local Branch and Worktree Flow

This repo supports a simple worktree-based development flow without replacing the existing Go/Compose workflow.

Recommended branch model:

- `main` = production-ready
- `staging` = release candidate / final integration
- `dev` = daily integration branch

Default rule:

- `make wt-new feat/name` uses current checked out branch as base
- if you are on `dev`, new worktree is based on `dev`
- if you need explicit base, use `make wt-new feat/name staging`

Recommended day-to-day flow:

1. create feature branch from `dev`
2. create git worktree for that branch and move into it
3. run `make dev-up`
4. develop and test inside that worktree
5. merge back into `dev`
6. promote `dev -> staging -> main`

Useful targets:

- `make wt-new feat/name`
- `make wt-list`
- `make wt-path feat/name`
- `make wt-rm feat/name`
- `make wt-prune`
- `make env-init`
- `make env-sync`
- `make dev-up`
- `make dev-down`
- `make dev-reset`
- `make dev-status`
- `make migrate-up-local`
- `make migrate-down-local`
- `make test-local`
- `make doctor`

See also:

- `documentation/guides/WORKTREE_FLOW.md`

Auto behavior currently implemented:

- `make wt-new` creates worktree and bootstraps `.env.local`
- `make wt-enter` ensures `env-init` and `env-sync` for target worktree
- `make dev-up` auto-initializes `.env.local` when missing and runs `env-sync`
- `make dev-status` auto-initializes `.env.local` when missing

Worktree expectations:

- default worktree root is `.worktrees/` inside repo for restricted environments
- sibling worktree root can still be used by overriding `WORKTREE_ROOT`
- each worktree gets its own `.env.local`
- compose project name must be unique per worktree
- local ports should not collide between parallel branches
- feature streams may split across frontend and backend/caller logic worktrees

## 3. Setup Flow

### 3.1 Create Environment File

```bash
cp .env.example .env.local
```

### 3.2 Use Worktree Flow for Parallel Development

```bash
make wt-new feat/my-feature
cd .worktrees/feat-my-feature
make dev-up
```

If you already have a local override file and need to sync new template keys:

```bash
make env-sync
```

### 3.3 Start Infrastructure

Use the worktree-local compose stack:

```bash
make dev-up
```

Legacy alias still works:

```bash
make docker-dev
```

### 3.4 Run Migrations

Use the worktree-local database connection:

```bash
make migrate-up-local
```

### 3.5 Seed Initial Data

```bash
make seed-up
```

### 3.6 Run the Application

For direct Go execution:

```bash
make run
```

For worktree dev stack status:

```bash
make dev-status
```

## 4. If You Change Backend API

Read:

- `documentation/guides/API_ACCESS_WORKFLOW.md`
- `documentation/guides/API_USAGE.md`
- `documentation/guides/ACCESS_RIGHTS_REFERENCE.md`

Then verify against live code:

- `internal/router/router.go`
- `internal/config/app.go`
- target module under `internal/modules/*`

## 5. If You Change Tenant Or Auth Behavior

Read:

- `documentation/architecture/MULTI_TENANCY.md`
- `documentation/guides/API_ACCESS_WORKFLOW.md`

Then verify against live code:

- `internal/middleware/auth_middleware.go`
- `internal/middleware/tenant_middleware.go`
- `internal/middleware/casbin_middleware.go`

## 6. If You Change Realtime

Read:

- `documentation/guides/REALTIME.md`
- `documentation/guides/PRESENCE_API.md`

Then verify against live code:

- `pkg/ws`
- `pkg/sse`
- `internal/router/router.go`

## 7. If You Change Upload Or Storage

Read:

- `documentation/guides/RESUMABLE_UPLOAD.md`
- `documentation/guides/CLIENT_UPLOAD_GUIDE.md`
- `documentation/guides/STORAGE.md`

Then verify against live code:

- `pkg/tus`
- `pkg/storage`
- target module using upload hooks

## 8. If You Change Frontend

Read:

- `documentation/guides/FRONTEND_STRUCTURE.md`

Then verify owner first:

- `apps/web`
- `apps/client`
- `packages/*`

## 9. Verification Order

1. narrowest package check
2. affected app `typecheck`
3. affected app `lint`
4. integration tests if DB, Redis, tenant, worker, upload, or Casbin touched
5. E2E if user flow changed

## 10. Planning Docs Rule

- `documentation/ops/*` = planning or migration notes
- `documentation/productplan/*` = backlog, vision, UX, wireframes
- do not treat them as runtime truth without checking live code
