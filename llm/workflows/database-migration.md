# Database Migration Workflow

## Purpose

Workflow ini untuk perubahan schema dan migration yang aman, reversible, dan aligned dengan repo runtime behavior.

## Use when

- adding or changing SQL schema under `db/migrations`
- changing persistence model that needs migration
- introducing column/index/constraint changes that affect runtime code

## Read first

1. `AGENTS.md`
2. `llm/conventions/database.md`
3. relevant domain cache files
4. current migration files under `db/migrations`
5. target repository/usecase code using affected tables

## Live code to inspect

- `db/migrations`
- target repositories and entities/models
- `internal/config/app.go` if startup or seed path depends on schema
- `db/seeds` if seed flow relies on changed table shape

## Workflow phases

### Phase 1 — Define schema delta

State exactly:

- table(s) affected
- columns/indexes/constraints changed
- whether data backfill or compatibility issue exists

### Phase 2 — Preserve migration discipline

- every migration must have matching up/down SQL file behavior
- name migration clearly
- keep runtime code and migration order aligned

### Phase 3 — Patch runtime owner code

Update only the model, repository, usecase, and seed logic that truly depend on new schema.

### Phase 4 — Verify safety

Check:

- migration applies cleanly
- rollback path still makes sense
- affected repository or integration path still works

## Review checklist

- up/down pair exists
- runtime code matches schema
- no hidden destructive data change slipped in
- seed or bootstrapping expectations checked when relevant

## Stop conditions / needs confirmation

- destructive data rewrite implied but not requested
- compatibility/backfill path is unclear
- migration affects production-sensitive auth, tenant, or permission tables without clear owner review
