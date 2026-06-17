# Database Migration Workflow

## Use when

- changing schema, indexes, seed assumptions, or repository behavior tied to schema
- adding/removing columns, tables, constraints, FK relationships, or unique indexes
- changing Casbin/API-key/webhook/audit/project persistence behavior

## Read first

- `llm/cache/backend-map.md`
- `llm/conventions/database.md`
- `AGENTS.md` migration and database guidance
- target repository/model/usecase files

## Live code to inspect

- `db/migrations`
- `db/seeds/main.go`
- affected repositories and entities/models
- integration/E2E tests touching affected tables
- `Makefile` migration and seed targets

## Steps

1. inspect existing migration numbering and naming pattern.
2. create paired `.up.sql` and `.down.sql` files.
3. update entities/models/repositories/usecases that depend on schema change.
4. update seeds if initial data assumptions change.
5. check API response/request model impact.
6. run `make migrate-up` / `make migrate-down` in the proper environment when applicable.
7. verify with narrow package tests, then integration tests for affected modules.

## Guardrails

- never add only an up migration.
- keep down migration safe and honest; if destructive rollback is impossible, document why in migration comments and final report.
- do not change DB schema without updating runtime code that reads/writes the schema.
- verify tenant isolation when organization-scoped tables change.

## Verification

- migration command in a configured local DB environment.
- repository/unit tests for affected module.
- integration tests for schema and isolation behavior.
- E2E tests if route behavior changes.
