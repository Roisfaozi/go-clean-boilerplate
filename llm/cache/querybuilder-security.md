# Query Builder Security

## Purpose

Durable map for dynamic filtering and sorting security in `pkg/querybuilder` and repository list endpoints that depend on it.

## Runtime truth

- `pkg/querybuilder/query_builder.go` builds dynamic GORM filter and sort clauses.
- Field names are resolved through struct fields, JSON tags, GORM tags, or snake_case conversion.
- Values use GORM placeholders rather than unsafe interpolation.
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
- Sort direction defaults to `asc`.
- Only `desc` is accepted as descending.

## Security semantics

- Dynamic query behavior is security-sensitive because repositories can expose model-backed fields through user input.
- The package is designed around model metadata allow/deny behavior, not raw user-defined field strings.
- Invalid or sensitive fields should fail closed rather than degrade into unsafe querying.

## Coupling to other systems

- user list and search paths are especially sensitive because they combine admin-facing query flexibility with user data
- repository list endpoints across modules can depend on querybuilder behavior for pagination, filter, and sort semantics

## Hard rules

- Field names must come from safe model metadata, not raw SQL from user input.
- Sensitive field filters and sorts must stay denied: password, token, secret, key, salt, and equivalents.
- Query-builder changes are security changes, not convenience-only refactors.
- Add denied-field and invalid-field tests when changing this package.

## Verification and evidence paths

- `pkg/querybuilder/query_builder.go`
- `pkg/querybuilder/query_builder_test.go`
- `pkg/querybuilder/README.md`
- repository list endpoints that rely on dynamic filtering or sorting
