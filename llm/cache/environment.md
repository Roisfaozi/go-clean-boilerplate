# Environment and Execution Context

## Environment files confirmed

- `.env.example` exists at repo root.
- Repo guidance instructs copying `.env.example` to `.env` before local run.

## Confirmed backend environment categories

From `README.md` and `.env.example`, this repo uses environment-driven toggles and secrets for:

- server port, timeouts, app env, app name, trusted proxies, and frontend base URL
- database credentials and pooling
- Redis credentials and pooling
- JWT secrets and token durations
- Casbin enablement, model, default role/domain, and watcher behavior
- CORS origins and trusted proxy config
- rate limit enablement, burst, RPS, and storage mode
- SMTP sender/host/port/credentials for worker email delivery
- websocket write/pong/ping/max message settings and distributed mode
- OpenTelemetry enablement, service name, and collector URL
- storage driver plus local/S3 configuration
- TUS base path
- pprof enablement and port
- circuit breaker enablement and thresholds
- seeding/admin bootstrap password

These categories come from `internal/config/config.go`, `internal/config/app.go`, and `.env.example`.

Defaults matter here: `internal/config/config.go` sets many defaults via Viper, so env files can override runtime behavior without changing code.

## Local run expectations confirmed

Backend local run path:

- copy `.env.example` to `.env`
- fill required secrets and connection values
- use `make run` or `pnpm go:run`

Recommended local infra path:

- `docker compose -f docker-compose.dev.yml up --build`
- or `make docker-dev`

## Build and verification commands confirmed

Workspace commands from `package.json`:

- `pnpm build`
- `pnpm lint`
- `pnpm test`
- `pnpm typecheck`

Backend commands from `package.json` and `Makefile`:

- `pnpm go:run`
- `pnpm go:test`
- `pnpm go:test-integration`
- `pnpm go:test-e2e`
- `pnpm go:docs`
- `make run`
- `make build`
- `make test`
- `make test-integration`
- `make test-e2e`

The Makefile is confirmed active and should be treated as a current source for backend workflow commands.

Frontend package commands confirmed

`apps/web`:

- `next dev --turbo`
- `next build`
- `next start`
- `eslint src`
- `tsc --noEmit`

`apps/client`:

- `react-router dev`
- `react-router build`
- `react-router-serve ./build/server/index.js`
- `react-router typegen && tsc`
- `playwright test`

Active command caveat:

- `apps/client` lint runs Biome; `pnpm typecheck` remains the TypeScript validation gate.
- `make docker-dev` is an active repo command; it is the preferred wrapper for dev infra if the user wants one command.

## Local infra files confirmed

- `docker-compose.yml`
- `docker-compose.dev.yml`
- `docker-compose.prod.yml`

## CI verification surface confirmed

Primary CI workflow `.github/workflows/ci.yml` runs:

- Go lint via `golangci-lint`
- unit tests via `make test-coverage`
- integration tests via `make test-integration`
- E2E tests via `make test-e2e`
- benchmarks via `make bench`
- build check via `go build -v -o /dev/null ./cmd/api/main.go`

## Env ownership by module

### Backend config root

`internal/config/config.go` is the main backend env source of truth. It loads app configuration through Viper/automatic env binding and maps env-backed settings for:

- `server.*`
- `mysql.*`
- `redis.*`
- `jwt.*`
- `casbin.*`
- `casbin.watcher.*`
- `cors.allowed_origins`
- `metrics.*`
- `security.*`
- `storage.*`
- `tus.*`
- `pprof.*`
- `websocket.*`
- `sso.google.*`, `sso.microsoft.*`, `sso.github.*`
- `rate_limit.*`
- `smtp.*`
- `circuit_breaker.*`

Module-level ownership note:

- `internal/config/app.go` consumes the config root to wire runtime-specific values into workers, storage, websockets, SSO, and auth modules.
- `internal/modules/auth` depends on JWT, SSO, Redis session storage, and ticket manager values from config.
- `internal/modules/organization` depends on `frontend_base_url` for invite/member flows.
- `internal/worker` depends on SMTP and rate-limit related config carried through app wiring.
- `apps/web` and `apps/client` use `process.env` values local to each app and should not be conflated with backend config.

### `apps/web`

`apps/web` uses `process.env` / `NEXT_PUBLIC_*` for client-visible and server-side app settings:

- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_APP_URL`
- `NEXT_PUBLIC_WS_URL`
- `NEXT_PUBLIC_COOKIE_SECURE`
- `NEXT_PUBLIC_ACCESS_TOKEN_MAX_AGE`
- `NEXT_PUBLIC_REFRESH_TOKEN_MAX_AGE`

### `apps/client`

`apps/client` uses `process.env` / `import.meta.env` for browser/runtime settings:

- `CI`
- `VITE_NEWSLETTER_PROVIDER`
- `VITE_NEWSLETTER_ENDPOINT`
- `NODE_ENV`

### Scope note

This is module ownership only for Phase 1. It does not yet claim which env vars are required in every deployment path, only where they are consumed in code.
