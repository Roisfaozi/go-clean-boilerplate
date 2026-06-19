# Database Migration Workflow

## Use when

- adding or changing SQL schema under `db/migrations`
- changing persistence contract that requires backfill or schema evolution
- changing seed/bootstrap behavior tied to schema changes

## Read first

- `llm/conventions/database.md`
- `llm/cache/backend-map.md`
- `llm/cache/domain-rules.md`
- owning repository/model/usecase files

## Live code to inspect

- `db/migrations`
- `db/seeds/main.go`
- affected GORM models under `internal/modules/*/model` or similar
- affected repository queries and tests
- Makefile migration targets

## Steps

1. confirm schema change is required and cannot be solved in code only.
2. add paired up/down migration files.
3. update affected model/repository/query assumptions.
4. update seed code only if bootstrap data truly depends on new schema.
5. add or adjust tests for persistence behavior where feasible.
6. verify migration command surface exists and document any local blockers.

## Verification commands

- migration command surface: `make migrate-up`, `make migrate-down`
- backend tests for affected repository/module
- integration tests when schema affects request lifecycle: `pnpm go:test-integration`

## Review checklist

- up/down files both exist
- column/index/default names match live query usage
- tenant, audit, webhook, and worker side effects still compatible
- no destructive migration done without explicit user intent

## Stop conditions / needs confirmation

- requested migration is destructive or irreversible without approval
- existing data backfill strategy is unclear
- schema change spans modules with conflicting ownership
