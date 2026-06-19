# Stats System

## Purpose

Durable map for dashboard statistics, activity, insights, and realtime metrics dependencies.

## Runtime truth

- Module root: `internal/modules/stats/`
- Wiring entry: `internal/modules/stats/module.go`
- Controller: `internal/modules/stats/delivery/http/stats_controller.go`
- Usecase: `internal/modules/stats/usecase/stats_usecase.go`
- Model: `internal/modules/stats/model/stats_model.go`
- `NewStatsModule` wires DB and logger into stats usecase.

## Route ownership

`internal/router/router.go` registers stats routes under `authenticated` group:

- `GET /api/v1/stats/summary`
- `GET /api/v1/stats/activity`
- `GET /api/v1/stats/insights`

Authenticated layering here includes API-key auth, JWT/session, auto scope, user-session requirement, user status, and optional rate limiter.

## Realtime coupling

- `internal/config/app.go` uses stats usecase inside realtime metrics broadcaster goroutine.
- That broadcaster publishes `metrics_update` payloads over websocket channel `system:metrics` and prunes stale presence users.

## Behavior surfaces

- HTTP stats endpoints for summary, activity, and insights
- stats aggregation logic used by dashboard behavior
- realtime metrics broadcasting path

## Hard rules

- Stats changes can affect both HTTP endpoints and realtime metrics broadcasting.
- Do not add expensive unbounded queries without considering periodic broadcaster load.
- Preserve authenticated route boundary and session semantics.
- If payload semantics change, check realtime consumer expectations too.

## Verification and evidence paths

- `internal/modules/stats/test/stats_usecase_test.go`
- `internal/modules/stats/usecase/stats_usecase.go`
- `internal/config/app.go`
- `internal/router/router.go`
