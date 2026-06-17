# Bugfix Workflow

## Use when

- fixing incorrect behavior in existing code
- closing regressions found by tests, runtime audit, or security review
- correcting docs/cache drift against live code

## Read first

- relevant cache files for affected layer
- target tests near failing behavior
- `llm/conventions/testing.md`
- `AGENTS.md` high-risk rules if bug touches auth, tenant, Casbin, API key, upload, realtime, or worker paths

## Live code to inspect

- failing test or route path
- runtime wiring in `internal/config/app.go` when dependency lifecycle is involved
- route/middleware in `internal/router/router.go` if behavior is HTTP-visible
- target module controller/usecase/repository
- frontend proxy/client code if bug crosses browser/backend boundary

## Steps

1. state observed bug and expected behavior.
2. reproduce with existing test, route path, or code trace.
3. locate the lowest layer that owns the behavior.
4. patch root cause, not only response formatting or UI symptom.
5. add/adjust regression test if adjacent test pattern exists.
6. run narrow check first, then broader check if boundary changed.
7. document unrelated failures separately.

## Verification matrix

- middleware bug: `internal/middleware` tests.
- tenant/security bug: tenant/permission integration and E2E scenarios.
- query/filter bug: `pkg/querybuilder` and affected repository dynamic tests.
- frontend proxy bug: affected frontend build/typecheck plus backend route check.
- worker bug: `internal/worker` package tests plus integration where async side effects matter.
