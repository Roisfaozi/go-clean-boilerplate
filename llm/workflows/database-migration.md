# Database Migration Workflow

## Purpose

Primary workflow for schema and migration changes in this repo.

Use it to keep SQL changes reversible, runtime-aligned, and honest about data-risk.

## Use When

- adding table, column, index, constraint, or FK under `db/migrations`
- changing repository/entity behavior that requires schema delta
- modifying seed/bootstrap assumptions that depend on table shape
- fixing schema drift between live code and migration history

## Do Not Use When

- change is repository/query-only and no schema delta is required
- change is seed-only without schema or constraint impact
- request implies destructive production data rewrite without explicit approval

## Required Read Order

1. `AGENTS.md`
2. `llm/conventions/database.md`
3. relevant domain cache files
4. current migration files under `db/migrations`
5. target entity/model/repository/usecase code
6. `db/seeds` if seed flow depends on changed schema

## Live Code to Inspect

- `db/migrations/*`
- target repositories and entities/models
- `internal/config/app.go` if startup or module wiring depends on changed data
- `db/seeds/*` for bootstrap assumptions
- `Makefile` migration commands

## Repo-Specific Facts

- migration files are paired `*.up.sql` and `*.down.sql`
- root command surfaces exist: `make migrate-up`, `make migrate-down`, `make migrate-up-1`, `make migrate-down-1`
- many sensitive tables are auth, org, Casbin, audit, API-key, and webhook related; these need extra caution

## Workflow Steps

### Step 1 — Define Exact Schema Delta

State explicitly:

- table(s) affected
- columns/indexes/constraints/FKs changed
- default values and nullability change
- whether old rows need backfill or compatibility bridge

If you cannot state that clearly, you are not ready to edit migration yet.

### Step 2 — Map Runtime Owners

For every changed table/field, map:

- entity/model owner
- repository owner
- usecase owner
- seed/bootstrap owner if any
- route or worker path that reads/writes it

### Step 3 — Preserve Migration Discipline

- create or update matching up/down pair
- keep names sequence-safe and descriptive
- do not ship up migration without believable down path unless user explicitly accepts irreversible change

### Step 4 — Patch Runtime Code In Same Slice

Update only code that truly depends on new schema:

- entities/models
- repositories
- usecases
- validation or request models if externally visible
- seed files if bootstrap shape changed

### Step 5 — Check Risk Class

Treat migration as high-risk when touching:

- auth/session tables
- organization membership / tenant tables
- Casbin rule table
- audit/webhook/API-key tables
- soft-delete indexes or visibility constraints

For these, verify more than syntax. Verify behavioral intent.

## Common Mistakes

- editing repo code without adding migration
- adding migration without updating repo/entity assumptions
- making down migration fake or unsafe
- forgetting seed/bootstrap dependency
- hiding destructive data rewrite inside “schema cleanup”

## Verification

### Minimum

- inspect up/down pair exists and is coherent
- verify runtime code matches schema fields

### Add When Needed

- package tests for affected repositories/usecases
- integration tests when DB behavior, org scoping, or policy coupling changed
- docs generation only if public API surface changed because of schema-visible field

### Command Surface

- `make migrate-up`
- `make migrate-down`
- `make migrate-up-1`
- `make migrate-down-1`
- `pnpm go:test`
- `make test-integration`

## Stop Conditions

Stop and mark `needs confirmation` if:

- destructive backfill or data loss is implied
- old and new runtime compatibility window unclear
- auth/tenant/Casbin tables change without proven owner understanding
- migration order depends on unstated production data assumptions

## Completion Output

Report:

- migration files added/changed
- runtime files patched to match schema
- reversal safety summary
- verification run and blockers
