---
name: database-transactions
description: Use when changing GORM transactions, all-or-nothing writes, schema migrations, seed data, Casbin policy writes tied to DB state, or tenant-sensitive repository behavior.
---

# Database Transactions: GORM and Casbin Consistency

**Announce at start:** "I'm using the database-transactions skill to preserve DB, tenant, and Casbin consistency."

## When To Use

- multi-table writes
- create/update/delete with audit/webhook side effects
- membership/role/policy changes
- migration or seed changes
- querybuilder/list filter behavior

## Read Order

1. `llm/conventions/database.md`
2. `llm/cache/domain-rules.md`
3. `llm/workflows/database-migration.md`
4. target model/repository/usecase
5. `db/migrations` and `db/seeds/main.go` when schema/seed changes

## Workflow

### Phase 1 — Transaction Boundary

Identify all writes that must commit/rollback together:

- DB rows
- Casbin policies
- audit outbox
- webhook task enqueue
- Redis/cache invalidation

### Phase 2 — Implementation Rules

- Use repo transaction helper/pattern already present.
- Do not split dependent writes into independent commits.
- Use transactional enforcer patterns when policy and DB writes must align.
- Keep tenant constraints inside transaction path when relevant.

### Phase 3 — Migration Rules

- Add paired up/down SQL files only when schema change required.
- No destructive migration without explicit approval.
- Update models/repositories/tests consistently.

### Phase 4 — Verification

- Unit tests for transaction rollback when pattern exists.
- Integration tests for DB/Redis/Casbin interactions.
- Migration up/down commands when environment supports it.

## Review Checklist

- [ ] rollback path understood
- [ ] policy state cannot diverge from DB state
- [ ] tenant/org cache invalidation preserved
- [ ] no sensitive querybuilder field exposure

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
