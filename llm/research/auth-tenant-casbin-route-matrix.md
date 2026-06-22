# Auth Tenant Casbin Deep Audit

## Scope

Audit ini fokus ke boundary `auth -> session -> API key -> tenant resolution -> Casbin`.

## Evidence Paths

- `internal/router/router.go`
- `internal/middleware/auth_middleware.go`
- `internal/middleware/api_key_middleware.go`
- `internal/middleware/tenant_middleware.go`
- `internal/middleware/casbin_middleware.go`
- `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/repository/token_repository.go`
- `internal/modules/organization/usecase/reader.go`
- `internal/modules/organization/usecase/organization_usecase.go`
- `internal/modules/permission/usecase/permission_usecase.go`

## Route Truth Matrix

### Public

- `public` group expose login, refresh, forgot/reset password, verify email, register, and SSO callbacks.
- Main risk: brute force, token misuse, SSO callback abuse, account enumeration.
- Existing control: dedicated critical limiter on login path plus public limiter.

### Authenticated

Order in `authenticated` group:

1. `apiKeyMiddleware.Authenticate()`
2. `authMiddleware.ValidateToken()`
3. `apiKeyMiddleware.RequireScopeAuto()`
4. `apiKeyMiddleware.RequireUserSession()`
5. `UserStatusMiddleware`

Meaning:

- route can accept JWT or API key in first layers
- but `RequireUserSession()` blocks API-key-only access for this group
- effective truth: user session is mandatory here even if API key auth machinery runs first

Risk:

- engineers may assume API key works everywhere under authenticated routes because `Authenticate()` is first
- adding routes here without understanding `RequireUserSession()` can create wrong product assumptions or flaky tests

### TenantAuthorized

Order in `tenantAuthorized` group:

1. `apiKeyMiddleware.Authenticate()`
2. `authMiddleware.ValidateToken()`
3. `apiKeyMiddleware.RequireScopeAuto()`
4. `UserStatusMiddleware`
5. `tenantMiddleware.RequireOrganization()`
6. `casbinMiddleware`

Meaning:

- JWT or API key identity established first
- tenant/org context resolved before Casbin
- Casbin domain depends on `organization_id` set by tenant middleware

Risk:

- wrong org resolution means wrong authorization domain
- route-level explicit scopes plus Casbin policy plus tenant membership must all align
- this is highest-risk route family for cross-tenant bugs

### Authorized Admin

Order in `authorized` group:

1. `apiKeyMiddleware.Authenticate()`
2. `authMiddleware.ValidateToken()`
3. `apiKeyMiddleware.RequireScopes("admin:manage")`
4. `UserStatusMiddleware`
5. `tenantMiddleware.OptionalOrganization()`
6. `casbinMiddleware`

Meaning:

- admin scope gate runs before optional org resolution
- Casbin may operate in `global` or resolved org domain depending on optional tenant context

Risk:

- mixed global/org semantics can be confusing
- deleted-org and superadmin behavior need extra review because optional org context changes domain used by Casbin

## Verified Flow Details

### Auth Middleware Truth

- `ValidateToken()` skips JWT validation entirely when request already marked as API-key auth.
- For non-API-key auth, token source can be bearer header or `access_token` cookie.
- JWT validation alone is not enough; middleware calls `AuthUseCase.Verify()` to load Redis-backed session.
- Middleware writes `user_id`, `session_id`, `user_role`, `username` into Gin context and request context.

Implication:

- session revocation truth lives in Redis session store, not in JWT signature only
- if Redis/session store fails, authenticated request can fail closed with internal error

### API Key Boundary Truth

- API-key middleware sits before auth middleware in protected groups.
- `ValidateToken()` explicitly no-ops when API key auth is active.
- scope enforcement can be auto-derived from method or route-specific via `RequireScopes`.
- some groups still require user session, others do not.

Implication:

- API key is not a simple alternate bearer token; it is separate identity path with route-family-specific semantics
- risk is not only unauthorized access, but also unintended denial from mismatched route placement

### Tenant Resolution Truth

- organization can be requested from header, existing context, route id, or slug.
- tenant middleware supports both `RequireOrganization()` and `OptionalOrganization()`.
- membership validity uses Redis-cached `CachedOrgReader` with 5-minute TTL and negative caching.
- member role is cached separately.

Implication:

- membership and role correctness can drift briefly if invalidation misses an edge path
- negative caching reduces DB load but increases stale-deny risk after recent membership creation or restore

### Casbin Truth

- Casbin middleware reads `user_id` from context and uses request path + method + org domain.
- domain defaults to `global` when `organization_id` absent.
- in non-release mode, nil enforcer leads to pass-through.

Implication:

- route path exactness matters for authorization
- optional tenant context changes domain and thus policy meaning
- dev/test can hide missing-enforcer problems unless explicitly tested

## Deep Weaknesses

### 1. Security Depends on Middleware Order

This stack is correct only when current order is preserved.

Failure modes:

- move `RequireOrganization()` after Casbin -> policy evaluated in wrong domain
- move `RequireUserSession()` out of authenticated group -> API keys may reach user-session-only routes
- add route into wrong group -> auth semantics silently change without usecase code change

This is biggest maintainability weakness.

### 2. Auth Has Broad Blast Radius

`AuthUseCase` handles:

- login lockouts
- password checks
- session issuance and revocation
- registration default role assignment
- default workspace creation
- audit enqueue
- password reset
- email verification
- SSO login/callback session creation
- websocket ticket issuance

Weakness:

- auth is not narrowly auth; it is orchestration for user/org/worker/policy/session concerns
- bugfix in one auth path can regress org provisioning, audit, or session lifecycle

### 3. Session Truth Split Across JWT and Redis

Strength:

- revocation supported
- stale JWT alone cannot keep access

Weakness:

- availability and consistency now depend on Redis session store
- refresh flow revokes old session and creates new one; failures in between can strand user or create inconsistent lifecycle
- tests must cover not only token validity but session index and revoke-all behavior

### 4. Tenant Caching Has Staleness Tradeoff

Membership cache:

- positive cache for membership
- negative cache for non-membership
- separate role cache
- org-wide invalidation uses Redis SCAN by pattern

Weakness:

- SCAN invalidation can be expensive and operationally noisy at scale
- negative cache can deny newly invited/newly restored access until invalidation or TTL expiry
- missing invalidation path means stale authZ for up to 5 minutes

### 5. Global vs Org Domain Semantics Are Easy To Misread

Casbin domain fallback to `global` is convenient, but risky when route family uses optional organization context.

Weakness:

- reviewers must reason about whether missing `organization_id` is intended or accidental
- admin/global routes and org routes share middleware style but not identical semantics

### 6. Policy Writes Need Transactional Discipline

Permission usecase mutates Casbin policies directly.

Weakness:

- if business write and policy write are not coupled transactionally, auth state can drift from DB state
- route/policy path string exactness means refactors can break authorization without schema or type errors

## Codebase Gaps

- no generated route-to-middleware-to-domain matrix from `internal/router/router.go`
- no single artifact that documents which protected routes allow API keys, require user sessions, or require org domain
- no explicit stale-cache playbook for membership troubleshooting in docs
- no static guard preventing route registration in wrong protection group

## High-Risk Scenarios

1. new protected endpoint added to wrong router group
2. org member updated but cache invalidation missed
3. refresh token flow partially fails during old-session revoke/new-session issue
4. API key wildcard scope plus tenant route mismatch
5. admin route unexpectedly evaluated in `global` domain
6. SSO-created session diverges from password-login session assumptions

## Recommendations

### P0

- create durable route matrix documenting each protected endpoint and required middleware/domain
- add tests that assert representative endpoints belong to expected protection group semantics
- explicitly test cache invalidation paths after invite, accept, remove, restore, soft delete, hard delete

### P1

- reduce auth orchestration sprawl by separating provisioning, session lifecycle, and SSO/session issuance concerns
- add targeted tests for refresh partial-failure semantics and revoke-all consistency
- add explicit observability for tenant resolution source: header vs slug vs route id

### P2

- consider startup guard stronger than release-mode-only check for nil Casbin enforcer in environments meant to simulate production auth
- add developer-facing docs explaining when to use `authenticated` vs `tenantAuthorized` vs `authorized`

## Needs Confirmation

- whether all route additions follow a code review checklist for protection group placement
- whether every membership mutation path invalidates both membership and role cache consistently
- whether permission writes that coincide with DB writes always use transactional enforcer path
