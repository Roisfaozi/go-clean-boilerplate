# Stats System

## Purpose

Durable map for dashboard statistics, activity, insights, and real-time metrics dependencies.

## Runtime truth

- Module root: `internal/modules/stats/module.go`.
- Controller: `internal/modules/stats/delivery/http/stats_controller.go`.
- Usecase: `internal/modules/stats/usecase/stats_usecase.go`.
- Model: `internal/modules/stats/model/stats_model.go`.
- `NewStatsModule` wires DB and logger into stats usecase.

## Route ownership

`internal/router/router.go` registers stats routes under `authenticated` group:

- `GET /api/v1/stats/summary`
- `GET /api/v1/stats/activity`
- `GET /api/v1/stats/insights`

The authenticated group applies API-key auth, JWT/session, API-key auto scope, user-session requirement, user status, and optional auth rate limiter.

## Runtime side effect

`internal/config/app.go` uses stats usecase inside a real-time metrics broadcaster goroutine that periodically broadcasts `metrics_update` payloads over WebSocket channel `system:metrics` and prunes stale presence users.

## Hard rules

- Stats changes can affect both HTTP dashboard endpoints and WebSocket metrics broadcasting.
- Do not add expensive unbounded queries without considering periodic broadcaster load.
- Preserve authenticated route boundary.

## Tests and evidence paths

- `internal/modules/stats/test/stats_usecase_test.go`
- `internal/modules/stats/usecase/stats_usecase.go`
- `internal/config/app.go`
- `internal/router/router.go`
