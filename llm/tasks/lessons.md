# Lessons

## Phase 1

- Root repo is hybrid: Go backend core plus active `apps/web` and `apps/client` frontends.
- `package.json`, `go.mod`, `Makefile`, `.env.example`, Docker Compose, and CI are enough to ground toolchain analysis before deeper architecture work.
- Env ownership can be mapped concretely from `internal/config/config.go`, `apps/web`, and `apps/client` code usage.

## Phase 2

- `internal/config/app.go` is the highest-value file for runtime truth.
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
