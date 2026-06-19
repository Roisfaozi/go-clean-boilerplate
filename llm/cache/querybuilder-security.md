# Query Builder Security

## Purpose

Durable map for dynamic filtering/sorting security in `pkg/querybuilder` and repository list endpoints.

## Runtime truth

- `pkg/querybuilder/query_builder.go` builds dynamic GORM filter/sort clauses.
- Field names are resolved through struct fields, JSON tags, GORM tags, or snake_case conversion.
- Values use GORM placeholders.
- Sensitive top-level field names are denied by `isSensitiveField`.

## Supported filter types

- `equals`
- `contains`
- `in`
- `between`
- `gt`
- `gte`
- `lt`
- `lte`
- `ne`

## Sort behavior

- Sort fields use `GetDBFieldName`.
- Sort direction defaults to `asc`; only `desc` is accepted as descending.

## Hard rules

- Field names must come from whitelisted model metadata, not raw SQL from user input.
- Sensitive field filters/sorts must stay denied: Password, Token, Secret, Key, Salt.
- Query-builder changes are security changes, not convenience-only refactors.
- Add denied-field and invalid-field tests when changing this package.

## Tests and evidence paths

- `pkg/querybuilder/query_builder.go`
- `pkg/querybuilder/query_builder_test.go`
- `pkg/querybuilder/README.md`
