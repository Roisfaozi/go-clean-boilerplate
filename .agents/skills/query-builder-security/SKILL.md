---
name: query-builder-security
description: Use when changing filtering, sorting, dynamic query fields, repository list endpoints, or `pkg/querybuilder` behavior.
---

# Query Builder Security: Dynamic Field Safety

**Announce at start:** "I'm using the query-builder-security skill because filtering/sorting is part of this repo's security model."

## Read Order

1. `llm/cache/domain-rules.md`
2. `llm/conventions/database.md`
3. `pkg/querybuilder`
4. affected repository/list endpoint
5. existing querybuilder tests

## Iron Rules

- Field names must come from whitelisted struct metadata, not raw user input.
- Sensitive fields stay denied: password, token, secret, key, salt, and equivalent secrets.
- Values use GORM placeholders.
- Sorting/filtering convenience must not weaken security.

## Workflow

### Phase 1 — Identify Input Surface

- query params
- filter object
- sort field/order
- repository list method

### Phase 2 — Validate Field Semantics

- Confirm allowlist/denylist behavior.
- Confirm struct tags/field metadata are intended.
- Confirm sensitive fields remain unqueryable.

### Phase 3 — Patch and Test

- Add tests for allowed field.
- Add tests for denied sensitive field.
- Add tests for invalid/unknown field.
- Add affected repository/list endpoint tests when possible.

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
