# Database Conventions

## Purpose

Guide for DB access, migrations, query safety, tenant-sensitive persistence, and transaction behavior in this repo.

## Primary runtime model

- GORM is the primary ORM and runtime DB access layer.
- Module repositories own DB access for module code.
- Controllers should not call GORM directly for business behavior.
- Tenant-aware reads and writes should respect organization context established by middleware or usecase boundaries.
- Dynamic query construction should use `pkg/querybuilder` or GORM placeholders, never interpolated SQL.

## Ownership rules

- controller
  - bind, validate, respond
- usecase
  - business rules, orchestration, transaction ownership when needed
- repository
  - query shape, GORM calls, persistence detail, tenant-aware scopes

If logic is deciding _what_ to write or _whether_ to write, it belongs above repository.
If logic is deciding _how_ to query or persist, it usually belongs in repository.

## Schema management

- SQL migrations live under `db/migrations` with paired `.up.sql` and `.down.sql` files.
- Migration files follow numeric prefix pattern already used in repo, for example `000025_*.up.sql` and `000025_*.down.sql`.
- Seed entrypoint is `db/seeds/main.go`.
- Migration commands come from `Makefile`, including:
  - `make migrate-create`
  - `make migrate-up`
  - `make migrate-down`
  - `make migrate-version`
  - `make seed-up`
- Do not invent another migration tool; repo guidance and Makefile indicate `golang-migrate`.

## Tenant and organization persistence rules

- organization and membership data often require cached membership checks and invalidation after updates
- tenant-sensitive repositories should preserve organization context and scope behavior
- soft-delete and restore or hard-delete style flows exist for admin-style organization behavior
- invitation or token tables and audit or outbox tables are separate persisted concerns, not one collapsed table

## Transaction and consistency rules

- changes that mutate both DB and Casbin policy state need transaction-aware handling
- request-owned writes and async side effects need explicit ordering
- repositories may consume transaction-scoped DB via context when `pkg/tx` is involved
- avoid doing irreversible external work before DB state is durable when semantics require atomicity

## Query safety rules

- always use parameterized GORM queries
- query-builder field names should come from struct metadata, not raw user input
- sensitive fields must remain blocked from generic filtering or sorting
- do not expose soft-deleted records unless route/usecase explicitly supports restore or admin flow
- if query flexibility expands, treat it as security-sensitive change

## Common change patterns

### Adding a column

- add migration pair
- update affected entity/model/repository logic
- verify response shape and docs if externally visible

### Adding tenant-sensitive query

- confirm organization context source
- preserve repository-level tenant scoping
- add integration/E2E coverage if route behavior depends on tenant isolation

### Changing list/filter behavior

- inspect `pkg/querybuilder`
- validate allowlist/denylist implications
- update repository dynamic-search tests

## Verification expectations

- for schema work, confirm migration pair, affected entities/models, repositories, seeds, and tests
- for tenant-sensitive data, run or add integration/E2E coverage around tenant isolation when practical
- for query-builder work, run `pkg/querybuilder` tests and adjacent repository dynamic-search tests
- for migration execution, use `make migrate-up` and `make migrate-down` in configured environment when appropriate

## Pre-merge checklist

- route or controller response shape still matches models and Swagger after schema changes
- repository queries still respect tenant and soft-delete behavior
- worker side effects still persist or consume intended rows after schema adjustments
- no hidden destructive data operation slipped in under ordinary refactor label
