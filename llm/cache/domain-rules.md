# Domain Rules

## Purpose

Cross-domain rulebook for behavior that repeatedly matters across modules, route groups, and integrations.

This file is for durable rules, not one-off implementation notes.

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

## Auth and session rules

- Registration creates user state and provisions default workspace or organization in a transaction.
- Login and protected session lifecycle use JWT plus Redis-backed session state.
- Auth middleware accepts bearer or cookie token paths and validates session state before protected handlers run.
- WebSocket handoff uses short-lived Redis ticket flow, not direct raw access token at the WS route.
- Logout and session revocation need both token handling and Redis cleanup semantics.

## Organization and tenant rules

- Organization is tenant backbone.
- Tenant-protected routes require organization context before Casbin authorization.
- Organization module owns lifecycle, membership, invitations, and cached organization reader behavior.
- Project routes are tenant-scoped and live under tenant-authorized route groups.
- Membership cache invalidation matters after org/member changes.
- Owner-only and admin-only constraints must not be bypassed in invitation or member update flows.

## Casbin and permission rules

- Casbin is DB-backed and initialized through `internal/config`.
- Production guard requires Casbin enabled and policies loaded.
- Permission module owns policy operations, role assignment, access-right expansion, and batch checks.
- Route authorization combines subject, domain, object path, and method through middleware and enforcer behavior.
- Transaction-bound policy changes should use transactional enforcer patterns.

## API-key rules

- API-key middleware participates in authenticated and tenant-authorized route groups.
- API keys have scopes and must pass auto or explicit scope checks on protected routes.
- API keys are organization-scoped and should be checked with scope-aware middleware, not identity only.

## Audit, webhook, and worker rules

- Audit can run as synchronous module behavior or asynchronous worker side effect.
- Audit outbox and worker handlers support side-effect processing outside primary request path.
- Webhook dispatch is asynchronous through worker/task distributor.
- Worker jobs include email delivery, audit syncing, cleanup, and webhook dispatch.

## Upload and storage rules

- TUS upload handling is separated from normal CRUD endpoints.
- Upload hooks route completion behavior by upload metadata or type.
- Storage provider abstraction supports local or S3-compatible backends.
- Avatar upload is concrete hook path from upload completion into user profile update logic.

## Query builder and filtering rules

- `pkg/querybuilder` validates fields against struct fields or tags.
- Sensitive fields such as password, token, secret, key, and salt are denied for filtering or sorting.
- GORM placeholders are used for values; field names come from safe metadata.
- Sorting and filtering rules are part of security model, not convenience helpers only.

## Realtime rules

- SSE route requires auth token.
- WebSocket route uses ticket validation.
- Presence and distributed WebSocket behavior use Redis-backed managers.
- Stats broadcaster can publish realtime payloads that share assumptions with websocket consumers.

## Known pitfalls and security learnings

- WebSocket origin validation is security-sensitive and must not be left wide open.
- Dynamic query helpers can leak sensitive field information if field access is not restricted.
- Auth and session changes should be evaluated against Redis-backed session lifecycle, not JWT parsing alone.
- Route matcher and API-key/Casbin layering matter for tenant authorization correctness.

## How to use this file

- use it when a change spans multiple modules or boundaries
- use it to sanity-check auth, tenant, permission, worker, upload, or querybuilder decisions
- do not use it as replacement for reading target module live code
