# Query Builder Security

## Purpose

Durable map for dynamic filter/sort behavior in `pkg/querybuilder` and every repository endpoint that exposes user-driven filtering.

This package is small but security-sensitive. Treat changes here as authorization-adjacent hardening work, not harmless DX cleanup.

## Primary source of truth

1. `pkg/querybuilder/query_builder.go`
2. `pkg/querybuilder/raw_builder.go`
3. `pkg/querybuilder/query_builder_test.go`
4. `pkg/querybuilder/raw_builder_test.go`
5. repositories that call `BuildRawQuery`
6. `llm/cache/user-system.md`
7. `llm/cache/audit-system.md`

## Runtime ownership

### Package responsibilities

`pkg/querybuilder/query_builder.go` and `pkg/querybuilder/raw_builder.go` own:

- mapping requested fields to DB column names (`GetDBFieldName`)
- validating filter/sort field legitimacy against model metadata (`BuildRawQuery`)
- building parameterized `WHERE` clauses (`WhereSQL` + `Args`)
- building safe `ORDER BY` clauses (`OrderBy`) and clamped `Limit`/`Offset`
- denying obviously sensitive fields (`isSensitiveField`)

### Call-site responsibilities

Repositories still own:

- which model metadata is exposed
- whether endpoint should allow dynamic filtering at all
- pagination/count behavior
- surrounding tenant/auth restrictions and `deleted_at = 0` filtering

Do not push repository access policy into querybuilder alone.

## Proven field-resolution behavior

`GetDBFieldName` resolves requested field by trying, in effect:

1. direct struct field name
2. case-insensitive struct field match
3. `db` tag column name
4. `json` tag name
5. snake_case version of struct field name (`ToSnakeCase`)

Column name extraction prefers the `db` tag (`extractColumnNameFromDBTag`) and falls back to `ToSnakeCase(field.Name)`. If field is not found (or matches `isSensitiveField`), query generation fails closed with error.

That is important. Do not degrade invalid field handling into silent skip.

## Sensitive-field behavior

Current `isSensitiveField` denies these top-level names:

- `Password`
- `Token`
- `Secret`
- `Key`
- `Salt`

Also note:

- denial happens on requested input name and resolved struct field name
- current check is exact-name based, not broad substring matching

Implication:

- adding alias fields or new sensitive model fields may require explicit test and review
- if repo later exposes fields like `ApiKeyHash` or `ResetToken`, do not assume current deny list already covers every variant

## Supported filter operators

Current supported types in `BuildRawQuery` (`buildSingleCondition`):

- `equals`
- `contains`
- `in`
- `between`
- `gt`
- `gte`
- `lt`
- `lte`
- `ne`

Unsupported operator returns error.

## SQL safety model

Package safety depends on two separate properties:

### Value safety

- values are collected into `SQLQueryResult.Args` and bound via SQL placeholders (`?`) in `sqlx` queries
- querybuilder does not interpolate raw filter values directly into SQL

### Field safety

- column names are derived from model metadata (`db` tag, `json` tag, or `ToSnakeCase`), not accepted raw from caller
- sort direction defaults to `ASC`
- only case-insensitive `desc` selects `DESC` order

If either property regresses, injection surface opens.

## Sort behavior

`buildOrderClause` (called by `BuildRawQuery`):

- returns empty string when sort is absent
- validates each requested sort field through `GetDBFieldName`
- formats `<column> ASC|DESC` into `SQLQueryResult.OrderBy`
- fails on invalid field instead of silently sorting on raw input

## Known consumers and risk level

Dynamic filtering appears in high-value admin/listing flows such as:

- user listing/search paths
- audit log search
- access endpoint listing
- other repository-backed admin surfaces

These are dangerous because they combine privileged data access with flexible query input.

## Known sharp edges

- sensitive-field deny list is exact-name based and may miss future semantically sensitive aliases.
- `db` tag or `json` tag exposure can widen allowed query surface even if handler code is unchanged.
- changing model fields can silently change querybuilder allow-surface.
- sorting by newly exposed field may leak sensitive ordering behavior even if field not returned in API response.

## Change checklist

Before editing querybuilder or adding new dynamic query endpoint, prove these answers:

1. Which repository endpoints call this package?
2. Which model fields become queryable through struct names or tags?
3. Could any new field represent secret, token, hash, status, tenant, or soft-delete data that should stay restricted?
4. Should invalid fields fail closed with error?
5. Do tests cover denied fields, invalid fields, operator errors, and expected SQL shape?
6. Does endpoint also need tenant or role restrictions outside querybuilder?

## Verification paths

Narrow first:

- `pkg/querybuilder/query_builder_test.go`
- direct repository tests for affected list/search endpoint

Broader when endpoint contract changes:

- target module tests
- integration coverage for admin listing path if security boundary changed

## Hard rules

- Do not accept raw SQL field names from client input.
- Do not silently ignore invalid fields; fail closed.
- Do not weaken sensitive-field deny behavior without explicit security reason.
- Do not assume tag changes are cosmetic; they can widen query surface.
- Add tests when new operator, field-resolution path, or deny behavior changes.
