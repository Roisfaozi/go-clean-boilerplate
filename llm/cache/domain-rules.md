# Domain Rules

## Purpose

Cross-domain rulebook for behavior that repeatedly matters across modules, route groups, and integrations.

Use this file when change spans multiple boundaries or when local code patch could violate repo-wide invariants.

This file does not replace live code. It tells you which invariants must survive.

## Primary source of truth

1. `internal/router/router.go`
2. `internal/config/app.go`
3. relevant middleware under `internal/middleware/*`
4. relevant module usecases
5. domain cache files under `llm/cache/*`

## Core entities

- user
- organization
- organization member
- role
- access right / endpoint access registry
- permission policy
- project
- API key
- audit log
- audit outbox
- webhook and webhook log
- password reset token
- email verification token
- SSO identity record

## Route-layer invariants

### Public routes

- used for login, refresh, forgot/reset password, verify email, register, SSO entry/callback, and selected public user/org flows
- public does not mean low-risk; auth bootstrap and password recovery live here

### Authenticated routes

Current invariant order:

1. optional API-key authenticate
2. token validate
3. API-key auto scope
4. user-session ban on API-key-only routes where required
5. user-status middleware

Meaning:

- JWT/session truth still matters
- API-key auth can participate but not replace every authenticated-human flow

### Tenant-authorized routes

Current invariant order:

1. optional API-key authenticate
2. token validate
3. API-key auto scope
4. user status
5. require organization
6. Casbin middleware

Meaning:

- tenant resolution must happen before Casbin
- org context is not optional in this route tier

### Authorized/admin routes

Current invariant order:

1. optional API-key authenticate
2. token validate
3. explicit admin API-key scope
4. user status
5. optional organization
6. Casbin middleware

Meaning:

- admin route protection is layered, not single-switch

## Auth and session rules

- JWT validity alone is not enough for protected flows; Redis-backed session validation matters.
- auth middleware can accept bearer or cookie token paths depending on surface.
- logout and revocation must clean session-side state, not only client token.
- WebSocket auth uses short-lived one-time ticket flow, not raw access token at WS upgrade boundary.
- registration is not pure auth concern; it also crosses user, role, audit, and webhook semantics.

## Organization and tenant rules

- organization is tenant backbone for tenant-authorized routes.
- organization context must be resolved before tenant-sensitive usecase or Casbin enforcement.
- organization module owns membership, invitations, cached org-reader behavior, restore, and admin lifecycle.
- membership cache invalidation matters after invite, accept, member update/removal, restore, and hard-delete-adjacent flows.
- owner/admin distinctions must not be bypassed in invite/member management.

## Casbin and permission rules

- Casbin is DB-backed and initialized through `internal/config`.
- production startup must not run fail-open with nil or zero-policy enforcer.
- permission module owns policy operations, assignment, expansion, and batch checks.
- authorization depends on subject + domain + object path + method, not user role string alone.
- policy writes coupled to DB writes should use transactional enforcer path.
- route protection truth belongs in router + middleware, not frontend hiding.

## API-key rules

- API keys are organization-scoped identities with scopes.
- API-key auth injects identity and org context into request path.
- API-key scope can be auto-derived or explicit per route.
- wildcard scope behavior exists and is security-sensitive.
- some authenticated routes explicitly forbid API-key-only use through `RequireUserSession()`.

## User rules

- registration includes default role assignment and audit side effects.
- `/users/me` self-service rules are separate from admin user management.
- admin list/search over users is sensitive because query flexibility meets personal data.
- avatar behavior spans direct API path and TUS completion-hook path.

## Audit, webhook, and worker rules

- audit can happen synchronously in request flow or asynchronously through outbox/worker.
- webhook dispatch is asynchronous through task distributor.
- request success does not always mean side-effect completion already happened.
- worker-owned flows include email delivery, audit sync, cleanup, and webhook dispatch.

## Upload and storage rules

- TUS upload handling is separate from normal CRUD controller flow.
- upload metadata is a trust boundary; authenticated metadata may override client-provided values.
- upload completion dispatch chooses hook by metadata `type`.
- avatar upload is concrete example of upload completion mutating user record.
- storage abstraction supports local and S3-compatible backends; URL shape can differ by driver.

## Query builder and filtering rules

- `pkg/querybuilder` allows dynamic filtering only through model metadata resolution.
- invalid fields must fail closed.
- sensitive fields such as password, token, secret, key, and salt must stay denied.
- query flexibility is part of security model, not helper sugar.

## Realtime rules

- SSE route requires auth token path.
- WebSocket route depends on one-time ticket validation.
- presence is organization-aware.
- stats broadcaster publishes payloads that websocket consumers may depend on.
- origin validation is security-sensitive.

## Frontend contract rules

- `apps/web` and `apps/client` are both active surfaces.
- backend contract changes may require auditing both proxies and both app consumers.
- proxy cookie/header forwarding is runtime behavior, not mere transport plumbing.
- frontend hiding or route gating does not replace backend authorization.

## When to stop and re-check

Stop and inspect more live code if any of these happen:

- route seems protected by multiple middlewares and you are not sure order matters
- module write also changes policy, cache, worker, or upload side effect
- change adds new filter/sort field on user/audit/admin list endpoint
- change touches org context, API key, or Casbin in same path
- response shape is consumed by both frontend apps

## Common failure patterns

- checking JWT and forgetting Redis session truth
- patching route handler without checking route-group middleware order
- widening query/search surface by tag rename or field exposure
- changing upload metadata or hook type without registry audit
- assuming API-key scope replaces Casbin or tenant checks
- changing organization membership without cache invalidation review

## Hard rules

- Live code stays source of truth.
- Tenant context is first-class boundary, not optional detail.
- Security-sensitive helpers like querybuilder, API-key scope, upload metadata, and origin checks need explicit review.
- Do not move authorization truth into frontend-only behavior.
- If invariant unclear, mark `needs confirmation` rather than inventing rule.
