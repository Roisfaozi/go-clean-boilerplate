# Security Boundary Change-Type Playbooks

## Purpose

Reusable verification flow for security-sensitive backend, realtime, upload, worker, and frontend-contract changes.

Use this before claiming a phase complete. Prefer narrow proof first, then broader suites.

## Environment Rules

- Use user-installed Go when Snap Go appears in PATH:
  - `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ...`
  - `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test`
- Integration and E2E suites require Docker/Testcontainers.
- Restricted sandbox may block Docker, sockets, or cache writes; report exact blocker.
- Broad suites do not replace focused proof for changed code.
- Race detector proof only counts when command used `-race` and exited successfully.

## Auth / Session

### Scope

Login, refresh, logout, revoke-all, password reset, SSO callback, current-user, and routes requiring Redis-backed sessions.

### Narrow Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/middleware/... ./internal/modules/user/...`
- Add target package tests for changed handler/usecase/repository.

### Integration / E2E Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-integration`
- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-e2e`

### Expected Result

- expired/revoked sessions fail even with structurally valid JWT.
- authenticated routes reject API-key-only access when user session is required.
- cookies and bearer fallback behave consistently through frontend proxies.

## Tenant / Casbin / API Key

### Scope

Organization context, membership cache, Casbin domain enforcement, API-key auth, and scope checks.

### Narrow Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/middleware/... ./internal/modules/organization/... ./internal/modules/permission/... ./internal/modules/api_key/...`
- Add route or usecase tests for changed permission boundary.

### Integration / E2E Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/integration/... -tags=integration -p 1 -timeout=10m`
- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/e2e/... -tags=e2e -p 1 -timeout=15m`

### Expected Result

- tenant-scoped routes require correct organization context.
- Casbin domain is `organization_id` unless route is intentionally global.
- API-key wildcard and project/org scopes work only where documented.
- user-session-only routes reject API-key-only calls.

## Upload / Storage

### Scope

TUS upload route, upload metadata, storage providers, completion hook, avatar updates, cleanup after hook failure.

### Narrow Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./pkg/tus ./pkg/storage/...`
- Include target module package when upload completion mutates domain state.

### E2E Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/e2e/modules -tags=e2e -run TestTUS -p 1 -timeout=15m`

### Expected Result

- `Upload-Metadata` values are valid base64 where TUS requires it.
- avatar completion trusts server-bound `authenticated_user_id`, not client-only `user_id`.
- failed completion hook attempts cleanup for completed uploads.
- local and S3-backed termination interfaces remain available.

## Worker / Audit / Webhook

### Scope

Asynq tasks, audit outbox, webhook dispatch, webhook delivery logs, email or cleanup jobs.

### Narrow Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/worker/... ./internal/modules/audit/... ./internal/modules/webhook/...`

### Integration Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-integration`

### Expected Result

- DB transaction and task enqueue order are intentional.
- replay/idempotency does not duplicate audit logs.
- webhook delivery failure and delivery-log persistence failure stay observable.
- retryable failures return errors to worker path.

## Realtime / SSE / WebSocket

### Scope

WS tickets, origin checks, Redis presence, SSE clients, join/leave behavior, distributed broadcasts.

### Narrow Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./pkg/ws ./pkg/sse`
- For targeted regression:
  - `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./pkg/ws -run 'TestWebSocketManager_UnregisterKeepsPresenceForOtherConnections|TestWebSocketManager_UnregisterNilSendClient|TestRedisTicketManager|TestPresenceManager' -v`

### Race Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -race ./pkg/ws ./pkg/sse`

### E2E Proof

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/e2e/realtime -tags=e2e -p 1 -timeout=15m`

### Expected Result

- WS tickets are single-purpose and expire correctly.
- closing one connection does not emit leave while same user/org still has another active connection.
- SSE unregister removes client without blocking broadcasts.
- no race warnings in stateful managers.

## Frontend Contract

### Scope

Backend response shape, error envelope, auth cookies, proxy headers, shared API types, upload/realtime frontend consumers.

### Narrow Proof

- `pnpm --filter casbin-web typecheck`
- `pnpm --filter casbin-client typecheck`
- `pnpm --filter @casbin/api-types typecheck`

### Runtime Proof

- `pnpm --filter casbin-web build`
- `pnpm --filter casbin-client build`
- `pnpm --filter casbin-client test:e2e` when browser flow changed.

### Expected Result

- both proxies still preserve required cookies and response headers.
- shared types match backend response changes used by both apps.
- app-local types remain local only when one app owns flow or transport differs.
- `apps/client` lint runs Biome; use `typecheck` for TypeScript validation.

## Broad Suite Separation

Use broad commands after focused proof, not before:

1. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test`
2. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-integration`
3. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-e2e`
4. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-race`
5. `pnpm typecheck`
6. `pnpm build`

Report each command separately with pass/fail/blocker.
