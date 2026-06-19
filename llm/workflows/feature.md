# Feature Workflow

## Use when

- adding a new backend feature
- adding a new frontend feature in `apps/web` or `apps/client`
- extending an existing module with new user-visible behavior

## Read first

- `llm/cache/project-overview.md`
- `llm/cache/architecture.md`
- `llm/cache/domain-rules.md`
- `llm/cache/module-map.md`
- relevant `llm/conventions/*.md`
- more specific workflow if task becomes API-only, DB-only, or cross-stack

## Live code to inspect

- `internal/config/app.go`
- `internal/router/router.go`
- target module under `internal/modules/*`
- target frontend app under `apps/web` or `apps/client`
- shared packages under `packages/*` if reused by feature

## Steps

1. locate owning module and user flow.
2. identify backend, frontend, worker, or storage boundaries touched.
3. define smallest coherent slice that preserves auth, tenant, and contract rules.
4. implement backend-first when feature changes persisted behavior or API contract.
5. implement frontend consumption only after contract is clear.
6. add or adjust tests at narrowest layer that covers change.
7. update task state in `llm/tasks/todo.md` for multi-step work.

## Verification commands

- backend unit slice: `pnpm go:test` or narrower package test command
- frontend app checks: `pnpm --filter casbin-web typecheck`, `pnpm --filter casbin-web lint`, `pnpm --filter casbin-client typecheck`
- cross-stack changes: combine backend verification with affected frontend app checks
- integration/E2E when route, auth, tenant, worker, upload, or cookie behavior changed: `pnpm go:test-integration`, `pnpm go:test-e2e`

## Review checklist

- correct layer owns behavior
- auth / tenant / API-key / Casbin boundaries preserved
- request and response contract still match consumers
- shared package reuse preferred over duplication
- unrelated files unchanged

## Stop conditions / needs confirmation

- route ownership unclear between modules
- contract source of truth conflicts between backend and frontend
- env/runtime dependency needed but not verifiable locally
- migration or destructive data change implied by feature without explicit user request
