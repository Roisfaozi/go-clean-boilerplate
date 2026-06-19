---
name: test-driven-development
description: Use when implementing a feature or bugfix with a meaningful test seam in the Casbin backend/frontend.
---

# Test Driven Development: Casbin Test-First Flow

**Announce at start:** "I'm using TDD where this repo has a meaningful test seam."

## Repo Test Targets

- Go usecase/repository/middleware: package-local tests
- route/auth/tenant/Casbin/API-key: integration/E2E when needed
- querybuilder: `pkg/querybuilder` tests
- upload/storage: `pkg/tus`, `pkg/storage`, integration tests
- frontend: app typecheck/build and existing UI/E2E patterns

## Red-Green-Refactor

### RED

- Write minimal failing test for desired behavior.
- Confirm failure is expected, not compile/setup error.

### GREEN

- Implement minimal code to pass.
- Do not add extra behavior.

### REFACTOR

- Clean names/duplication after green.
- Keep tests green.

## Exceptions

If no practical test seam exists, document:

- why no test was added
- exact code trace/manual verification used instead
