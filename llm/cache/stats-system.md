# Stats System

## Purpose

Durable map for stats-domain behavior:

- dashboard summary
- activity stats
- system insights
- authenticated stats API routes
- realtime metrics broadcaster coupling

Use before changing `internal/modules/stats`, dashboard metrics payloads, or realtime metrics broadcasting.

## Primary source of truth

1. `internal/modules/stats/delivery/http/stats_controller.go`
2. `internal/modules/stats/usecase/stats_usecase.go`
3. `internal/modules/stats/model/stats_model.go`
4. `internal/modules/stats/module.go`
5. `internal/config/app.go`
6. `internal/router/router.go`
7. `llm/cache/realtime-system.md`

## Runtime ownership

- `NewStatsModule` wires DB and logger into stats usecase.
- stats controller exposes HTTP endpoints.
- stats usecase owns aggregation logic.
- `internal/config/app.go` consumes stats usecase from outside module for realtime broadcaster.

## Route ownership

Stats routes are under authenticated route group:

- `GET /api/v1/stats/summary`
- `GET /api/v1/stats/activity`
- `GET /api/v1/stats/insights`

Authenticated group includes API-key authenticate, JWT/session validation, auto scope, user-session requirement, user status, and optional rate limiter.

## Realtime coupling

`internal/config/app.go` periodically calls stats usecase inside metrics goroutine.

Current broadcaster semantics:

- interval is 2 seconds
- reads system insights and dashboard summary
- publishes JSON with `type: metrics_update`
- broadcasts to websocket channel `system:metrics`
- also prunes stale presence users in same loop

This means expensive stats queries can affect realtime loop health.

## Behavior surfaces

- HTTP dashboard stats
- system insight calculation
- activity feed calculation
- realtime metrics payload
- websocket consumer expectations

## Known sharp edges

- stats module looks isolated, but broadcaster runs outside module.
- a slow DB aggregation can affect periodic websocket metrics.
- payload field changes can break dashboard consumers even if HTTP endpoints still pass.
- authenticated session rules still apply; stats are not public metrics by default.

## Change checklist

Before editing stats code, prove:

1. Is change for HTTP endpoint, realtime broadcaster, or both?
2. Does aggregation run every request or every 2-second broadcaster tick?
3. Does payload field shape change?
4. Do frontend consumers expect current keys?
5. Does query need pagination, limits, indexes, or caching?

## Verification paths

- `internal/modules/stats/test/stats_usecase_test.go`
- `internal/modules/stats/usecase/stats_usecase.go`
- `internal/modules/stats/delivery/http/stats_controller.go`
- `internal/config/app.go`
- `pkg/ws/*` when realtime payload changed

## Hard rules

- Do not add unbounded expensive queries to broadcaster path.
- Do not change `system:metrics` payload semantics without frontend audit.
- Preserve authenticated route boundary.
- Keep stats aggregation in usecase, not controller.
