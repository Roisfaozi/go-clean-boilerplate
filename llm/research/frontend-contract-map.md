# Frontend Contract Map

## Scope

Phase 7 contract map for backend routes and two frontend apps.

## Evidence Paths

- `internal/router/router.go`
- `apps/web/src/proxy.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`
- `packages/hooks/*`
- `packages/ui/*`

## Contract Surface

- backend route shape
- shared types
- web proxy behavior
- client proxy behavior
- auth cookie/header semantics
- org scope and tenant-aware endpoints

## Consumer Map

Use this map whenever backend route shape, payload fields, response envelope, cookie/header semantics, upload behavior, realtime event payloads, or tenant/auth requirements change.

| Backend surface                                                                              | Shared type/source                                                                       | `apps/web` consumer/proxy                                                        | `apps/client` consumer/proxy                                                     | Required check                                                                                             |
| -------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `/api/v1/auth/*`                                                                             | `packages/api-types/src/index.ts` auth request/response types when imported by consumers | `apps/web/src/app/api/v1/[...path]/route.ts`; auth UI/hooks under `apps/web/src` | `apps/client/app/routes/api-proxy.ts`; auth routes/hooks under `apps/client/app` | verify cookie forwarding, `Set-Cookie`, bearer fallback, logout/session expiry states                      |
| `/api/v1/users/*`                                                                            | `User`, user schemas, profile/avatar fields in `packages/api-types/src/index.ts`         | web profile/user callers plus proxy                                              | client user/profile callers plus proxy                                           | verify `avatar_url`/legacy aliases, status fields, query params, validation errors                         |
| `/api/v1/organizations/*`                                                                    | `Organization`, `OrgMember` in `packages/api-types/src/index.ts`                         | organization switch/list consumers plus proxy                                    | organization switch/list consumers plus proxy                                    | verify tenant context, membership status, org cookie/header behavior                                       |
| `/api/v1/roles/*`, `/api/v1/permissions/*`, `/api/v1/access-rights/*`, `/api/v1/endpoints/*` | `Role`, `Permission`, `AccessRight`, `Endpoint` in `packages/api-types/src/index.ts`     | admin/permission surfaces plus proxy                                             | admin/permission surfaces plus proxy                                             | verify Casbin-facing IDs/actions/resources stay aligned                                                    |
| `/api/v1/projects/*`                                                                         | `Project` in `packages/api-types/src/index.ts`                                           | project surfaces plus proxy                                                      | project surfaces plus proxy                                                      | verify organization scope and API-key project scope coupling                                               |
| `/api/v1/api-keys/*`                                                                         | API-key types if added to `packages/api-types`; otherwise local app types                | API-key UI/hooks plus proxy                                                      | API-key UI/hooks plus proxy                                                      | verify scope names, masked secret fields, organization/project scoping                                     |
| `/api/v1/audit-logs/*`                                                                       | audit types if added to `packages/api-types`; otherwise local app types                  | audit/admin pages plus proxy                                                     | audit/admin pages plus proxy                                                     | verify filter/sort allowlist, timestamp shape, organization visibility                                     |
| `/api/v1/webhooks/*`                                                                         | webhook types if added to `packages/api-types`; otherwise local app types                | webhook pages/hooks plus proxy                                                   | webhook pages/hooks plus proxy                                                   | verify event subscription names, delivery log fields, secret masking                                       |
| `/api/v1/upload/files/*`                                                                     | upload metadata contract in backend TUS package and app upload helpers                   | upload/avatar callers plus proxy only if routed through Next app                 | `apps/client/app/lib/upload` and proxy when same-origin route is used            | verify `Upload-Metadata` base64 fields, server-bound `authenticated_user_id`, completion-hook error states |
| `/api/v1/events`                                                                             | SSE event payload docs/types when added                                                  | web realtime hooks if present                                                    | client realtime hooks if present                                                 | verify event names, reconnect behavior, auth/tenant visibility                                             |
| `/api/v1/ws` and websocket ticket route                                                      | WS ticket/realtime types when added                                                      | web websocket consumers if present                                               | client websocket consumers if present                                            | verify ticket issuance, origin, presence join/leave semantics                                              |
| `/api/v1/stats/*`                                                                            | `DashboardStats`, `ActivityMetric`, `SystemInsight` in `packages/api-types/src/index.ts` | dashboard/stat hooks plus proxy                                                  | dashboard/stat hooks plus proxy                                                  | verify numeric field names and realtime stat refresh behavior                                              |

## Backend API Change Checklist

For every backend route contract change, complete this before commit:

1. Identify route registration in `internal/router/router.go` or target module route file.
2. Identify request/response structs, validation tags, and error envelope used by the handler/usecase.
3. Check whether `packages/api-types/src/index.ts` already models the changed shape.
4. Search both apps for route path, type name, field name, and hook/client helper.
5. Verify `apps/web/src/app/api/v1/[...path]/route.ts` still preserves required cookies, auth headers, and response headers.
6. Verify `apps/client/app/routes/api-proxy.ts` still preserves required cookies, `Set-Cookie`, stream/body behavior, and response headers.
7. Run owning app typecheck; run build when route/component wiring changes; run E2E when auth, tenant, upload, or realtime flow changes.

## Validation Gate

- `apps/web`: `pnpm --filter casbin-web typecheck`; add `pnpm --filter casbin-web build` for runtime route/component changes.
- `apps/client`: `pnpm --filter casbin-client typecheck`; `pnpm --filter casbin-client lint` runs Biome instead of placeholder output.
- `packages/api-types`: `pnpm --filter @casbin/api-types typecheck` when shared contract types change.
- Browser/E2E: `pnpm --filter casbin-client test:e2e` or targeted backend E2E when auth, tenant, upload, or realtime behavior changes.

## Next Actions

- Keep this map updated in same commit as backend response-shape changes.
- Prefer shared `packages/api-types` updates when both active apps consume same shape.
- Keep app-local types only when a feature exists in one app or transport differs materially.
