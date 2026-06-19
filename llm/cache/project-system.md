# Project System

## Purpose

Durable map for tenant-scoped project CRUD:

- required organization context
- explicit project API-key scopes
- Casbin-protected route behavior
- repository/usecase tenant isolation expectations

Use before changing `internal/modules/project`, `/api/v1/projects` routes, project scopes, or frontend project API contracts.

## Primary source of truth

1. `internal/router/router.go`
2. `internal/modules/project/delivery/http/project_controller.go`
3. `internal/modules/project/usecase/project_usecase.go`
4. `internal/modules/project/repository/project_repository.go`
5. `internal/modules/project/module.go`
6. `llm/cache/api-key-system.md`
7. `llm/cache/tenant-organization-system.md`
8. `llm/cache/casbin-permission-system.md`

## Runtime ownership

- `NewProjectModule` wires DB and validator.
- project repository owns DB persistence.
- project usecase owns project business behavior.
- router owns explicit scope requirements and tenant route placement.

## Route ownership

`/api/v1/projects` is registered directly inside `tenantAuthorized` route group.

Route scope matrix:

| Method/path            | Handler | API-key scope                      |
| ---------------------- | ------- | ---------------------------------- |
| `POST /projects`       | create  | `project:manage`                   |
| `GET /projects`        | list    | `project:view` or `project:manage` |
| `GET /projects/:id`    | detail  | `project:view` or `project:manage` |
| `PUT /projects/:id`    | update  | `project:manage`                   |
| `DELETE /projects/:id` | delete  | `project:manage`                   |

Tenant-authorized route group also applies token validation, user status, required organization, and Casbin middleware.

## Tenant semantics

Project is not global CRUD.

Project access must preserve:

- organization context from tenant middleware
- API-key organization identity if API key used
- Casbin domain enforcement
- project visibility within current organization

## Cross-system coupling

Project changes can affect:

- API-key scope contracts
- frontend project pages/forms
- Casbin route permissions
- tenant membership access
- audit/webhook only if side effects are added later

## Known sharp edges

- explicit scope overrides are easy to lose during route refactor.
- auto-scope alone may not match project intent.
- repository tests may pass while tenant route protection is wrong.
- frontend route hiding does not replace tenant/Casbin enforcement.

## Change checklist

Before editing project code, prove:

1. Does route stay under `tenantAuthorized`?
2. Does method still require right `project:*` scope?
3. Does usecase/repository preserve organization isolation?
4. Does frontend consume changed request/response shape?
5. Does Casbin policy/access registry need update?

## Verification paths

- `internal/modules/project/test/project_usecase_test.go`
- `internal/modules/project/test/mocks/mock_project_repository.go`
- `internal/modules/project/usecase/project_usecase.go`
- `internal/router/router.go`
- integration route test if tenant or scope behavior changed

## Hard rules

- Preserve required organization context.
- Preserve explicit project API-key scopes.
- Do not move project authorization into frontend.
- Verify route group and middleware order when changing paths.
