# Project System

## Purpose

Durable map for tenant-scoped project CRUD, explicit project API-key scopes, and project access behavior under tenant-authorized routing.

## Runtime truth

- Module root: `internal/modules/project/`
- Wiring entry: `internal/modules/project/module.go`
- Controller: `internal/modules/project/delivery/http/project_controller.go`
- Usecase: `internal/modules/project/usecase/project_usecase.go`
- Repository: `internal/modules/project/repository/project_repository.go`
- Entity and model paths: `internal/modules/project/entity/*`, `internal/modules/project/model/*`

## Route ownership

`internal/router/router.go` registers `/api/v1/projects` directly under `tenantAuthorized` group:

- `POST /projects` requires `project:manage`
- `GET /projects` requires `project:view` or `project:manage`
- `GET /projects/:id` requires `project:view` or `project:manage`
- `PUT /projects/:id` requires `project:manage`
- `DELETE /projects/:id` requires `project:manage`

The `tenantAuthorized` group already applies API-key auth, JWT/session, auto scope, user status, required organization, and Casbin middleware.

## Behavior surfaces

- tenant-scoped project CRUD
- explicit project API-key scope overrides on routes
- project visibility within organization context

## Coupling to other systems

- project access depends on tenant organization context
- API-key scope semantics interact with project route behavior
- Casbin enforcement and required organization middleware both affect project access

## Hard rules

- Project routes are tenant-scoped; preserve required organization context.
- Preserve explicit project API-key scope overrides in router.
- Do not move project access checks into frontend-only logic.
- Repository and usecase changes should maintain tenant isolation semantics.

## Verification and evidence paths

- `internal/modules/project/test/project_usecase_test.go`
- `internal/modules/project/test/mocks/mock_project_repository.go`
- `internal/modules/project/usecase/project_usecase.go`
- `internal/router/router.go`
