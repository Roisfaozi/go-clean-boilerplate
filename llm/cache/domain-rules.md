# Domain Rules

## Core entities

- user
- organization
- organization member
- role
- access right / endpoint access registry
- project
- API key
- audit log
- audit outbox
- webhook and webhook log
- password reset token
- email verification token
- user SSO identity records

## Auth and session

- Registration creates user state and provisions a default workspace/organization in a transaction.
- Login/session lifecycle uses JWT plus Redis-backed session state.
- Auth middleware accepts bearer/cookie token paths and validates session state before protected handlers run.
- WebSocket handoff uses short-lived Redis ticket flow, not direct raw access token at the WS route.
- Password hashing uses project helper functions and bcrypt-backed behavior.
- Logout and session revocation need both token handling and Redis cleanup semantics.
- Password reset and email verification are tokenized flows with dedicated token records.

## Organization / tenant

- Organization is the tenant backbone.
- Tenant-protected routes require organization context before Casbin authorization.
- Organization module owns org lifecycle, membership, invitations, and cached organization reader behavior.
- Project routes are tenant-scoped and live under tenant-authorized route groups.
- Organization delete has soft/hard variants; hard delete and restore are admin-style flows.
- Invitation acceptance can activate users and must respect expiry checks.
- Member updates cannot silently change org owner behavior or bypass owner-only constraints.
- Membership cache invalidation matters after org/member changes.

## Casbin / permission

- Casbin is DB-backed and initialized through `internal/config`.
- Production guard requires Casbin enabled and policies loaded.
- Permission module owns policy operations, role assignment, access-right expansion, and batch checks.
- Route authorization combines subject, domain, object path, and method through middleware/enforcer behavior.
- Transaction-bound policy changes should use transactional enforcer patterns.
- Role inheritance, access-right assignment, and batch permission check are explicit backend flows.
- Authorized/admin management routes combine API key scope with Casbin policy allow.

## API key

- API key middleware participates in authenticated and tenant-authorized route groups.
- API keys have scopes and must pass auto or explicit scope checks on protected routes.
- API key usecase depends on API key repository, organization repository, user repository, and Redis.
- API keys are organization-scoped and should be checked with scope-aware middleware, not only identity.

## Audit, webhook, worker

- Audit can run as synchronous module behavior or asynchronous worker side effect.
- Audit outbox and worker handlers support side-effect processing outside primary request path.
- Webhook dispatch is asynchronous through worker/task distributor.
- Audit export has both synchronous and async variants, so response-time and background work matter separately.
- Worker jobs include email delivery, audit syncing, cleanup, and webhook dispatch.

## Upload / storage

- TUS upload handling is separated from normal CRUD endpoints.
- Upload hooks route completion behavior by upload metadata/type.
- Storage provider abstraction supports local or S3-compatible backends.
- Avatar upload is a concrete hook path from upload completion into user profile update logic.

## Query builder

- `pkg/querybuilder` validates fields against struct fields/tags.
- Sensitive fields such as password/token/secret/key/salt are denied for filtering/sorting.
- GORM placeholders are used for values; field names come from whitelisted struct metadata.
- Sorting and filtering rules are part of the security model, not only convenience helpers.

## Known pitfalls and security learnings

- WebSocket origin validation is a known security concern; sentinel guidance says origin checks must not be left wide open.
- Dynamic query helpers can leak sensitive field information if reflection-based field access is not restricted; sentinel guidance explicitly blocks sensitive field names.
- Auth/session changes should be evaluated against Redis-backed session lifecycle, not JWT parsing alone.
- Route matcher and API-key/Casbin layering matter for tenant authorization correctness.

## Realtime

- SSE route requires auth token.
- WebSocket route uses ticket validation.
- Presence and distributed WebSocket behavior use Redis-backed managers.
