# Project Overview

This repository is a hybrid monorepo with a Go API backend as the operational core and multiple JavaScript/TypeScript workspace packages for web clients and shared frontend libraries.

## Repository shape

- Root workspace package name: `casbin-monorepo` from `package.json`.
- Root package manager: `pnpm@10.8.0` from `package.json`.
- Root build orchestrator: `turbo` from root scripts in `package.json`.
- Primary backend module: Go module `github.com/Roisfaozi/go-clean-boilerplate` from `go.mod`.
- Backend entrypoint: `cmd/api/main.go`.
- Secondary Go entrypoint: `cmd/gen/main.go` for module scaffolding.
- Frontend app packages present:
  - `apps/web` with `next dev --turbo` / `next build`
  - `apps/client` with `react-router dev` / `react-router build`

## Languages in active toolchain

- Go for API backend and backend test/build pipeline.
- TypeScript/JavaScript for workspace frontend apps and shared packages.

## Confirmed backend stack

- HTTP framework: Gin.
- ORM / database access: GORM.
- Database driver in module dependencies: MySQL.
- Authorization library: Casbin.
- Cache / session / worker backend: Redis.
- Background jobs: Asynq.
- Upload protocol: TUS (`tusd`).
- Observability: OpenTelemetry.
- API documentation: Swagger (`swag`).

These are confirmed from `go.mod` and root docs, not inferred from naming.

## Confirmed frontend stack present in workspace

- `apps/web`: Next.js 16 + React 19.
- `apps/client`: React Router 7 + Vite + React 19.
- `apps/web` and `apps/client` are active application surfaces and still need improvement toward production readiness.

Architecture and backend relationship details for these apps live in `llm/cache/frontend-map.md` and `llm/cache/api-contracts.md`.

## Workspace package structure

Shared workspace packages currently present:

- `packages/api-types`
- `packages/hooks`
- `packages/ui`
- `packages/utils`

These packages are consumed by frontend apps and are part of the active workspace, not standalone side directories.

## Canonical development commands confirmed

Root workspace:

- `pnpm build`
- `pnpm lint`
- `pnpm test`
- `pnpm typecheck`
- `pnpm go:run`
- `pnpm go:test`
- `pnpm go:test-integration`
- `pnpm go:test-e2e`

Backend via Makefile:

- `make run`
- `make build`
- `make test`
- `make test-unit`
- `make test-integration`
- `make test-e2e`
- `make test-all`
- `make docs`

The Makefile is confirmed active as an operational command surface for this repo.

Frontend app packages:

- `pnpm --filter casbin-web dev`
- `pnpm --filter casbin-client dev`

## Local infrastructure confirmed

- Docker Compose files exist: `docker-compose.yml`, `docker-compose.dev.yml`, `docker-compose.prod.yml`.
- `.env.example` exists and is referenced by repo guidance.

## Repo-level operational files

- `README.md` gives high-level repository guidance and command references.
- `AGENTS.md` is the agent entrypoint and points to concretized `llm/` context.
- `documentation/` contains supporting architecture, API, guide, and product docs.

## CI / delivery presence confirmed

- CI workflow exists at `.github/workflows/ci.yml`.
- Staging deploy workflow exists at `.github/workflows/cd-staging.yml`.
- Production deploy workflow exists at `.github/workflows/cd-production.yml`.

## Scope boundary

This file is the high-level map. Detailed runtime, module, auth, tenant, and request-flow behavior belongs in `llm/cache/architecture.md`, `llm/cache/backend-map.md`, `llm/cache/module-map.md`, and `llm/cache/domain-rules.md`.
