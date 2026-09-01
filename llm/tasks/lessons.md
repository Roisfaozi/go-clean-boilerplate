# Lessons

## Phase 1

- Root repo is hybrid: Go backend core plus active `apps/web` and `apps/client` frontends.
- `package.json`, `go.mod`, `Makefile`, `.env.example`, Docker Compose, and CI are enough to ground toolchain analysis before deeper architecture work.
- Env ownership can be mapped concretely from `internal/config/config.go`, `apps/web`, and `apps/client` code usage.

## Phase 2

- `internal/config/app.go` remains highest-value file for runtime truth, but helper extraction means `internal/config/app_helpers.go` must also be checked for module wiring, broadcaster behavior, SSO providers, and TUS setup.
- `internal/router/router.go` is the clearest single file for route strata, middleware layering, and upload/realtime exposure.
- `internal/modules/*/module.go` files are the best compact view of real dependency boundaries.

## Phase 3

- Organization is the tenant backbone; many auth/permission behaviors depend on org/member context.
- Auth correctness depends on JWT plus Redis-backed session behavior, not JWT parsing alone.
- API key and Casbin layering must be reviewed together for protected routes.
- Sentinel guidance matters for WebSocket origin validation and reflection-based query safety.

## Phase 4

- Conventions in this repo are driven more by live module patterns and Makefile/CI than by style docs alone.
- Frontend apps are active; `apps/client` lint now runs Biome; `typecheck` remains the separate TypeScript gate.
- Integration/E2E validation expectations are strong because auth, tenant, worker, upload, and realtime flows are infrastructure-heavy.

## Phase 5+

- Proxy behavior in `apps/web` and `apps/client` is part of the real API contract surface and should be audited with backend route changes.
- Existing `documentation/llm/*` docs are helpful, but live code remains authoritative when there is drift.
- `documentation/api/AI_STREAMING_CONTRACT.md` currently reads as supporting/planned contract documentation, not confirmed live backend routing.

## Mockery & Testing Migration Lessons

- **Mockery v2.53.6 Configuration**: Default output writes 1 file per interface (`mock_{{.InterfaceName}}.go`). Specifying `filename:` in `.mockery.yml` per-package causes interfaces to overwrite each other. CLI `--filename` combined with a config file applies globally and overwrites all generated mocks into a single file.
- **Variadic Method Signature in Mocks**: Mockery v2.53.6 generates variadic parameters (`params ...interface{}` or `domain ...string`) as individual arguments, requiring test expectations to pass positional arguments (`"arg1", "arg2"`) rather than slices (`[]interface{}{"arg1", "arg2"}`).
- **False-Positive Unit Tests**:
  - `organization_member_usecase_test.go`: Tests without `usecase.WithActorUserID(ctx, actorID)` failed early inside `authorizeMemberManagement` with `ErrForbidden`, bypassing intended test logic and masking missing mock setups.
  - `audit_usecase_test.go`: `CreateAuditLogRequest{UserID: "u1"}` failed mandatory field validation (`Action`/`Entity`), returning validation errors before reaching repository layer.
  - `auth_usecase_test.go`: `HandleSSOCallback` uses `StoreToken`, not `StoreSession`. Registering expectations on uncalled methods with `NewMockX(t)` triggers test cleanup failures.


## API Key Scope Enforcement (verified against internal/router/router.go + internal/middleware/api_key_middleware.go)

- Only two route groups accept API-key auth: `tenantAuthorized` and `authorized`. The `authenticated` group is JWT-only (`RequireUserSession` rejects API keys), so scopes like `user:*`, `stats:*`, `permission:*` unlock nothing for API keys.
- `authorized` group requires `admin:manage` for API keys (`RequireScopes("admin:manage")` at group level) — permission, access-right, audit, user-admin, and org-admin endpoints are gated by `admin:manage`, not by resource-scoped scopes.
- `tenantAuthorized` combines `RequireScopeAuto()` (derives `resource:view|create|update|delete` from pathParts[2], e.g. `/api/v1/webhooks` → `webhook:view`) with explicit `RequireScopes(...)`; a key must satisfy BOTH. `manage` covers every action on the same resource (`hasRequiredScope`).
- Effective UI-relevant scope set for API keys: `org:view|manage`, `project:view|manage`, `role:view|manage`, `member:manage`, `presence:view`, `webhook:manage`, `admin:manage`. Member/presence endpoints additionally require `org:view` (auto-scope from `/organizations/...`).
- Only `user.created` is emitted as a webhook event in the backend (`internal/modules/user/usecase/user_usecase.go:149-160`); UI event presets beyond it do not fire until triggers are registered.
- Web frontend tests now run via Vitest in `apps/web` (`pnpm --filter casbin-web test`); `vitest.config.ts` is excluded from the app tsconfig because `@vitejs/plugin-react` uses an exports map that `moduleResolution: "node"` cannot resolve.

## Observability Configuration Lessons

- Viper `AutomaticEnv()` does not make arbitrary environment-only keys visible
  to `Unmarshal`; keys must be registered through defaults, config, flags, or
  explicit `BindEnv`, or assigned through explicit `v.Get*` calls. This is why
  `envOnlyKeys` exists in `internal/config/config.go`.
- `env:` and `envDefault:` tags on `AppConfig` fields are inactive in this
  repository because configuration is loaded by Viper and no environment-tag
  parser dependency is present. Effective environment names come from Viper
  key paths plus the `.` to `_` replacer.
- OTEL config uses documented `OTEL_*` names, so its Viper namespace must be
  `otel.*`; using the struct field name `telemetry.*` would map to the wrong
  environment variables.
- Metrics, tracing, and dashboard stats are three separate runtime surfaces:
  Prometheus metrics may be real while authenticated dashboard stats still use
  placeholder values. Audits must inspect producers and both frontend
  consumers, not infer correctness from the existence of `/metrics`.
- Context-aware Logrus hooks only enrich entries created with
  `WithContext(ctx)`, and typed context keys from different packages do not
  match even when their string values look similar.

## Time & Soft-Delete Normalization

- `deleted_at` across migrations and GORM models uses `BIGINT NOT NULL DEFAULT 0` sentinel (with `gorm.io/plugin/soft_delete` tag `softDelete:milli`).
- Instant timestamps are epoch milliseconds (`BIGINT`). Conversion functions and helpers live in `pkg/epochms` (Go) and `packages/utils/src/time.ts` (TS).
- Timezones must be explicit parameters (`timeZone` option in JS, `In(tz)` in Go). Server or browser implicit fallback is blocked by `./scripts/guard-time-conventions.sh` in pre-commit hooks.
