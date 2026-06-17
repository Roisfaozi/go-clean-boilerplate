# Database Conventions

## Runtime access

- Use GORM as the primary ORM/data access layer.
- Repository packages own DB access for modules.
- Use module repositories instead of calling GORM directly from controllers.
- Tenant-aware reads/writes should respect organization context and middleware/usecase boundaries.
- Dynamic query construction should use `pkg/querybuilder` or GORM placeholders, not interpolated SQL.

## Schema management

- SQL migrations live under `db/migrations` with paired `.up.sql` / `.down.sql` files.
- Migration files follow the numeric prefix pattern already present, for example `000025_*.up.sql` and `000025_*.down.sql`.
- Seed entrypoint is `db/seeds/main.go`.
- Migration commands come from `Makefile`: `make migrate-create`, `make migrate-up`, `make migrate-down`, `make migrate-version`, `make seed-up`.
- Do not invent another migration tool; README and Makefile indicate `golang-migrate`.

## Data boundaries

- Users, organizations, roles, access rights, Casbin rules, projects, audit logs, API keys, webhooks, token records, and outbox data are DB-backed.
- Redis-backed state is still part of data correctness for sessions, API key cache, presence, tickets, rate limits, and workers.
- Changes that mutate both DB and Casbin policy state need transaction-aware handling.

Tenant data patterns seen in code:

- organization/member data often requires cached membership checks and invalidation after updates.
- soft-delete and restore/hard-delete style flows exist for admin-style organization behavior.
- invitation/token tables and audit/outbox tables are separate persisted concerns, not collapsed into one table.

## Query safety

- Use parameterized GORM queries.
- Query-builder field names should come from struct metadata, not raw user input.
- Sensitive fields must remain blocked from generic filtering/sorting.
- Do not expose deleted/soft-deleted records unless route/usecase explicitly supports restore/admin flow.

## Verification

- For schema work, confirm migration pair, affected entities/models, repositories, seeds, and tests.
- For tenant-sensitive data, run or add integration/E2E coverage around tenant isolation.
- For query-builder work, run `pkg/querybuilder` tests and adjacent repository dynamic-search tests.
- For migration execution, use `make migrate-up` / `make migrate-down` in a configured environment.

Checkpoints before merging DB work:

- route/controller response shape still matches models and Swagger after schema changes
- repository queries still respect tenant and soft-delete behavior
- worker side effects still persist or consume intended rows after schema adjustments
