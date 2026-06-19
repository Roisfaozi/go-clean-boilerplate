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
3. locate lowest layer that owns behavior.
4. patch root cause, not only response formatting or UI symptom.
5. add or adjust regression test if adjacent test pattern exists.
6. run narrow check first, then broader check if boundary changed.
7. document unrelated failures separately.

## Verification commands

- middleware bug: targeted `go test` on `internal/middleware/...`
- module bug: targeted `go test` on owning package under `internal/modules/...`
- workspace surface touched: `pnpm typecheck`, `pnpm build`, or app-specific checks as relevant
- integration/E2E when request lifecycle or persistence semantics changed: `pnpm go:test-integration`, `pnpm go:test-e2e`

## Review checklist

- root cause proven before fix
- regression path covered by test or explicit code trace
- no unrelated workaround hidden in outer layer
- docs/task notes updated only if durable and verified

## Stop conditions / needs confirmation

- cannot reproduce and no code-path evidence narrows failure
- issue likely depends on external service or infra unavailable locally
- multiple plausible root causes remain after targeted tracing
