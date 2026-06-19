# Project System

## Purpose

Durable map for tenant-scoped project CRUD behavior.

## Runtime truth

- Module root: `internal/modules/project/module.go`.
- Controller: `internal/modules/project/delivery/http/project_controller.go`.
- Usecase: `internal/modules/project/usecase/project_usecase.go`.
- Repository: `internal/modules/project/repository/project_repository.go`.
- Entity/model: `internal/modules/project/entity` and `internal/modules/project/model`.

## Route ownership

`internal/router/router.go` registers `/api/v1/projects` under `tenantAuthorized` group:

- `POST /projects` requires `project:manage` API-key scope.
- `GET /projects` requires `project:view` or `project:manage`.
- `GET /projects/:id` requires `project:view` or `project:manage`.
- `PUT /projects/:id` requires `project:manage`.
- `DELETE /projects/:id` requires `project:manage`.

The `tenantAuthorized` group applies API-key auth, JWT/session, API-key auto scope, user status, required organization, and Casbin middleware.

## Hard rules

- Project routes are tenant-scoped; preserve required organization context.
- Do not move project access checks into frontend-only logic.
- Preserve explicit project API-key scope overrides in router.
- Repository/usecase changes should maintain tenant isolation semantics.

## Tests and evidence paths

- `internal/modules/project/test/project_usecase_test.go`
- `internal/modules/project/test/mocks/mock_project_repository.go`
- `internal/router/router.go`
