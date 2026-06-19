# Query Builder Security

## Purpose

Durable map for dynamic filter/sort behavior in `pkg/querybuilder` and every repository endpoint that exposes user-driven filtering.

This package is small but security-sensitive. Treat changes here as authorization-adjacent hardening work, not harmless DX cleanup.

## Primary source of truth

1. `pkg/querybuilder/query_builder.go`
2. `pkg/querybuilder/query_builder_test.go`
3. `pkg/querybuilder/README.md`
4. repositories that call `GenerateDynamicQuery` and `GenerateDynamicSort`
5. `llm/cache/user-system.md`
6. `llm/cache/audit-system.md`

## Runtime ownership

### Package responsibilities

`pkg/querybuilder/query_builder.go` owns:

- mapping requested fields to DB column names
- validating filter/sort field legitimacy against model metadata
- building supported filter clauses
- building safe order clauses
- denying obviously sensitive fields

### Call-site responsibilities

Repositories still own:

- which model metadata is exposed
- whether endpoint should allow dynamic filtering at all
- pagination/count behavior
- surrounding tenant/auth restrictions

Do not push repository access policy into querybuilder alone.

## Proven field-resolution behavior

`GetDBFieldName` resolves requested field by trying, in effect:

1. direct struct field name
2. case-insensitive struct field match
3. JSON tag name
4. GORM column tag
5. snake_case version of struct field name

If field is not found, query generation fails closed with error.

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

Current supported types in `GenerateDynamicQuery`:

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

- values are passed through GORM placeholders like `?`
- querybuilder does not interpolate raw filter values directly into SQL

### Field safety

- column names are derived from model metadata, not accepted raw from caller
- sort direction defaults to `asc`
- only exact `desc` selects descending order

If either property regresses, injection surface opens.

## Sort behavior

`GenerateDynamicSort`:

- no-ops when sort absent
- validates each requested sort field through same metadata path
- appends `ORDER BY <column> asc|desc`
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
- JSON tag or GORM tag exposure can widen allowed query surface even if handler code unchanged.
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
