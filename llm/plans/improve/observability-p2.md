# Observability P2 Plan - Hardening and Maintainability

## Objective

Reduce operational noise, cardinality, privacy risk, and inconsistent logging
after the P0 and P1 behavior is proven.

## P2-1 - Normalize HTTP Log Schema

### Files

- `internal/middleware/request_logger.go`
- `internal/middleware/request_logger_test.go`
- logging documentation

### Planned Schema

```text
type=http_request
request_id
trace_id
span_id
method
route
status
latency_ms
response_bytes
client_ip
user_agent
auth_method (bounded)
service
environment
instance_id
```

### Rules

- `route` uses `c.FullPath()` and falls back to `unknown`.
- Do not index raw dynamic URL paths by default.
- Do not log query strings, request bodies, cookies, Authorization, API keys,
  CSRF tokens, or upload metadata without explicit redaction.
- Keep `latency_ms`; remove `latency_ns` if no consumer requires both.
- Use consistent field names across request and panic logs.

## P2-2 - Use One Panic Recovery Owner

### Files

- `internal/router/router.go`
- `internal/middleware/recovery.go`
- `internal/middleware/recovery_test.go`

### Planned Change

- Remove `gin.Recovery()` and retain the custom structured recovery middleware,
  provided tests prove response and stack logging behavior.
- Ensure recovery is early enough in the chain to catch downstream panics and
  late enough to use request ID/trace context.
- Add tests for panic before and after headers are written.
- Add `app_panics_total{route}` with a bounded route label if useful.

## P2-3 - Standardize Frontend and Proxy Correlation

### Files

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- frontend API clients
- frontend error boundaries

### Planned Change

- Forward backend `X-Request-ID` from both proxies to browser clients.
- Preserve incoming request IDs only under an approved trust model; otherwise
  let the backend generate one.
- Replace ad hoc proxy `console.log` calls with a small server-side structured
  logger or consistent JSON records.
- Include request ID in frontend-visible support/error metadata without
  exposing internal error details.
- Remove debug console statements from production bundles.
- Select an error tracking product only through a separate design decision;
  this plan does not assume Sentry or another vendor.

## P2-4 - Establish PII and Secret Redaction Policy

### Initial Audit Targets

- email recipient logging in worker email handler;
- session IDs in auth logs;
- user email in SSO success logs;
- user IDs, organization IDs, webhook URLs, storage paths;
- error values from external systems that may embed credentials.

### Rules

- Passwords, access/refresh tokens, API keys, cookies, CSRF tokens, reset and
  verification tokens, SMTP passwords, storage secrets, and webhook secrets
  must never be logged.
- Email should be masked or replaced with an internal user ID.
- Session IDs should be hashed or omitted.
- IDs may be logged only when operationally necessary and retention/access are
  controlled.
- External errors should be reviewed for URL credential leakage.

### Tests

Add table-driven redaction tests for shared redaction helpers. Do not add a
helper until at least two real call sites need identical behavior.

## P2-5 - Add Dependency and Realtime Metrics Carefully

Potential additions after measured need:

- database pool gauges from `sql.DB.Stats()`;
- Redis operation failures and latency at selected boundaries;
- SSE active connections;
- WebSocket send drops and subscription failures;
- webhook attempts, outcomes, and duration;
- email outcomes and duration;
- audit outbox pending/oldest age;
- TUS upload completion, failures, size, and duration;
- rate-limit rejections by bounded limiter type.

Cardinality rules:

- never label by user, organization, project, request, session, task ID, email,
  URL, raw error, or raw route path;
- labels must have a finite reviewed vocabulary;
- prefer logs/traces for high-cardinality event details;
- document metric type, unit, labels, owner, and dashboard consumer.

## P2-6 - Unify Application Lifecycle Logging

### Files

- `cmd/api/main.go`
- `internal/config/app.go`
- worker/scheduler lifecycle code

### Planned Change

- Use the configured structured logger after application construction.
- Emit service name, environment, version, and instance ID as stable fields.
- Keep pre-application config failures on standard logger because Logrus is not
  available yet.
- Avoid `Fatal` inside library constructors where returning an error is
  practical; process exit should be owned by `main`.
- Preserve shutdown ordering and report each failed cleanup without hiding
  later cleanup attempts.

## P2 Verification Matrix

```bash
go test ./internal/middleware ./internal/config ./internal/worker/... ./pkg/...
pnpm --filter casbin-web lint
pnpm --filter casbin-web typecheck
pnpm --filter casbin-web test
pnpm --filter casbin-client typecheck
pnpm --filter casbin-client test:e2e
```

Manual log review should confirm:

- one request log per request;
- one structured panic record;
- stable route names;
- no known secrets or direct email/session values;
- matching request ID across browser proxy and backend;
- production output remains valid line-delimited JSON.

## P2 Exit Gate

- HTTP logging schema is documented and tested.
- Panic recovery has one owner.
- Both frontend proxies preserve request correlation.
- PII/secret redaction rules cover known high-risk call sites.
- New metrics meet cardinality rules.
- Lifecycle logs use the shared structured logger where available.
