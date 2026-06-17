# Go Service Workflow

## Use when

- creating or extending backend service/module behavior in Go
- changing usecase/repository/controller contracts
- adding cross-module dependencies

## Read first

- `llm/cache/backend-map.md`
- `llm/cache/module-map.md`
- `llm/conventions/golang.md`
- `llm/conventions/database.md`
- `llm/conventions/testing.md`

## Live code to inspect

- target `internal/modules/<name>/module.go`
- adjacent `entity`, `model`, `repository`, `usecase`, `delivery/http` files
- `internal/config/app.go` if constructor wiring changes
- tests under target module, `internal`, `pkg`, `tests/integration`, and `tests/e2e`

## Steps

1. identify owning module and existing analog.
2. update entity/model if data shape changes.
3. update repository interface/implementation if persistence changes.
4. update usecase business logic and transactions.
5. update controller and route only if HTTP surface changes.
6. wire constructor dependencies explicitly.
7. update mocks when interfaces change.
8. update unit/integration tests.

## Guardrails

- do not add package-level global clients.
- do not bypass transaction manager for multi-step writes.
- do not bypass permission/tenant middleware by checking only in frontend.
- do not create repositories inside controllers.

## Verification

- package unit tests for changed module.
- `make mocks` if interfaces changed.
- integration tests if DB/Redis/Casbin/worker/storage behavior changed.
- E2E tests if route lifecycle changed.
