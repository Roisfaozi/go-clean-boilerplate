# Comprehensive Testing Strategy Guide

This document defines the official testing standards for this project, covering Unit, Integration, and E2E testing layers.

---

## 🏗️ 1. Unit Testing (Isolated Logic)

**Goal:** Verify business logic in isolation with zero external dependencies.

- **Libraries**: `testify`, `mockery`.
- **Pattern**: **Dependency Struct Pattern** for UseCases.

### Standard Setup

```go
type userTestDeps struct {
    Repo     *mocks.MockUserRepository
    TM       *mocking.MockWithTransactionManager
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
    deps := &userTestDeps{
        Repo: new(mocks.MockUserRepository),
        TM:   new(mocking.MockWithTransactionManager),
    }
    uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, ...)
    return deps, uc
}
```

---

## 🔗 2. Integration Testing (Real Infrastructure)

**Goal:** Verify interaction with real MySQL and Redis instances using **Singleton Containers**.

- **Optimization**: We use the **Singleton Container Pattern** to start Docker only once per test suite.
- **Cleanup**: Use `TRUNCATE` between tests instead of restarting containers.

### Execution

```bash
make test-integration
```

---

## 🌍 3. End-to-End (E2E) Testing (Full Flow)

**Goal:** Verify the complete HTTP request-response cycle from the client's perspective.

- **Setup**: Uses `httptest.Server` connected to the singleton integration containers.
- **Client**: Uses a custom `TestClient` wrapper for easy JSON assertions.
- **Worker Management**: The `TestServer` explicitly manages `Scheduler` and `TaskProcessor` lifetimes. This ensures that asynchronous side-effects (like Audit Log syncing or Email delivery) actually execute during the test window.

### Execution

```bash
make test-e2e
```

---

## 🛡️ 4. Security Testing

Every module must include:

- **SQL Injection Tests**: Ensuring inputs are parameterized.
- **RBAC Tests**: Verifying Casbin policies correctly block unauthorized roles.
- **Validation Tests**: Checking for invalid formats and required fields.

---

## ⚡ 5. Generating Mocks

When you modify an interface, you MUST regenerate mocks:

```bash
make mocks
```

---

## 🧪 6. Frontend Testing (apps/web)

`apps/web` uses **Vitest** with Testing Library in a jsdom environment.

```bash
pnpm --filter casbin-web test
```

| Path | Role |
|---|---|
| `apps/web/vitest.config.ts` | jsdom env, `~` alias, `src/**/*.test.{ts,tsx}` |
| `apps/web/src/test/setup.ts` | cleanup + jsdom shims (`matchMedia`, `ResizeObserver`, `scrollIntoView`) |

`vitest.config.ts` is excluded from the app `tsconfig.json` because
`@vitejs/plugin-react` ships an exports map that `moduleResolution: "node"`
cannot resolve.

### Conventions

- Tests sit next to the component they cover.
- Mock the API module, not `fetch`, so payload mapping stays asserted.
- Use `vi.hoisted` for mock state referenced inside a `vi.mock` factory.
- When mocking a module you also import values from, spread `importOriginal()`
  so schemas and helpers survive.
- Add `beforeEach(() => vi.clearAllMocks())`; call history leaks between tests
  otherwise.
- Jest-DOM matchers are not installed. Assert with plain values
  (`expect(el.value).toBe(...)`, `expect(queryBy...).toBeNull()`).

### Known gap

Vitest covers contexts, forms, and dialog logic. There is **no browser E2E** for
`apps/web`; Playwright exists only in `apps/client`. Full-page flows still need
manual verification.

### Full gate

```bash
pnpm --filter casbin-web test
pnpm --filter casbin-web typecheck
pnpm exec biome lint apps/web
pnpm --filter casbin-web build
```
