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

## Next Actions

- list each changed backend route consumer in web/client
- verify both proxies after contract changes
- keep app typecheck/build mandatory for response-shape changes
