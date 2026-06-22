# Phase Compliance Audit

## Phase 1 — Structure and toolchain

Status: complete.

Requirement coverage:

- package manager: `package.json`
- primary languages: `go.mod`, TypeScript package configs
- backend framework: `go.mod`
- frontend frameworks: `apps/web/package.json`, `apps/client/package.json`
- backend entrypoint: `cmd/api/main.go`
- frontend entrypoints: `apps/web/src/app/[locale]/layout.tsx`, `apps/client/app/root.tsx`, `apps/client/app/routes.ts`
- command surfaces: `package.json`, `Makefile`, app package scripts, CI
- env example: `.env.example`
- local infra: `docker-compose*.yml`
- CI: `.github/workflows/*.yml`
- testing/mocks/helpers: `tests`, package tests, `Makefile` mock target

Artifacts:

- `llm/cache/project-overview.md`
- `llm/cache/environment.md`
- `AGENTS.md`

## Phase 2 — Architecture and boundary

Status: complete.

Requirement coverage:

- backend folder and composition root: `internal/config/app.go`
- route setup and middleware stack: `internal/router/router.go`, `internal/middleware/*`
- module boundaries: `internal/modules/*/module.go`
- controller/usecase/repository flow: `internal/modules/*/{delivery,usecase,repository}`
- auth, tenant, Casbin, API key flows: router/middleware/usecase files
- worker/job/cron: `internal/worker/*`
- storage/upload: `pkg/storage`, `pkg/tus`
- realtime: `pkg/ws`, `pkg/sse`
- database/migrations: `db/migrations`, `db/seeds/main.go`
- frontend boundary: `apps/web`, `apps/client`

Artifacts:

- `llm/cache/architecture.md`
- `llm/cache/backend-map.md`
- `llm/cache/frontend-map.md`
- `llm/cache/api-contracts.md`
- `llm/cache/module-map.md`

## Phase 3 — Domain rules

Status: complete.

Requirement coverage:

- core domain entities listed explicitly
- auth/session/JWT/Redis rules from auth usecase and middleware
- tenant/org/member/invitation rules from organization usecases and tenant middleware
- permission/Casbin rules from permission module and router strata
- API key rules from API key module and middleware strata
- upload/TUS rules from TUS package and app wiring
- worker/audit/webhook rules from worker and module usecases
- error/validation notes from validator/controller patterns
- pitfalls from `.jules/sentinel.md` and querybuilder code

Artifact:

- `llm/cache/domain-rules.md`

## Phase 4 — Conventions

Status: complete.

Requirement coverage:

- Go patterns: architecture, DI, context, transactions, controller/usecase boundaries, security-sensitive rules
- database patterns: migration tool, schema pairing, tenant-aware persistence, query safety
- testing patterns: layer mapping, command mapping, infrastructure assumptions, validation strategy
- TypeScript/frontend patterns: app ownership, path aliases, proxy/API boundaries, verification caveats
- error/validation/lint/mock/test conventions are covered from live files and Makefile/package scripts

Artifacts:

- `llm/conventions/golang.md`
- `llm/conventions/database.md`
- `llm/conventions/testing.md`
- `llm/conventions/typescript.md`

## Phase 5 — Workflow concretization

Status: complete.

Requirement coverage:

- each workflow names when to use it
- each workflow names cache files to read
- each workflow names live code to inspect
- each workflow uses existing repo commands and paths
- each workflow has verification guidance
- workflow set covers feature, bugfix, API endpoint, Go service, cross-stack, and database migration

Artifacts:

- `llm/workflows/feature.md`
- `llm/workflows/bugfix.md`
- `llm/workflows/api-endpoint.md`
- `llm/workflows/go-service.md`
- `llm/workflows/cross-stack-change.md`
- `llm/workflows/database-migration.md`

## Phase 6 — End-to-end consistency

Status: complete.

Checks performed:

- required files exist and are non-empty
- no unresolved `needs confirmation` markers remain
- core referenced paths exist
- commands come from `package.json`, `Makefile`, app packages, or CI
- docs remain grounded in live code/config/tests, with docs used as support only
- AGENTS entrypoint routes agents into the concretized starter-pack

Quality pass additions:

- cache files were deepened with route trees, module dependency hotspots, env ownership, contract confidence, and runtime caveats
- convention files were expanded from generic guidance into repo-specific operational guidance
- workflow files were expanded into actionable playbooks with guardrails and verification matrices

Known non-blocking note:

- future work can keep sharpening docs as the repo evolves, but current starter-pack coverage is complete and operational for this checkout.

## Parity follow-up — 2026-06-18

Status: in progress.

Findings:

- core cache/convention/workflow set exists and remains aligned with repo runtime truth
- starter-pack knowledge lifecycle folders were missing and need restoration
- workflow files needed explicit verification/review/stop-condition sections for strict parity
- skill parity is blocked by current inability to write under `.agents/skills` in this environment

Parity closure — 2026-06-18

Status: complete.

Closure checks:

- missing lifecycle folders restored with README guidance
- workflow files now include explicit verification, review, and stop-condition sections
- starter-pack core skills now exist under `.agents/skills/` with repo-aware triggers
- retained existing repo-specific skills because they are additive, not conflicting
- no unresolved parity blocker remains in current workspace
