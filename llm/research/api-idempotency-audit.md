# API Idempotency Audit

## Scope

Audit ini memetakan endpoint dan flow fitur yang tidak punya konsep idempotency eksplisit, lalu memberi rekomendasi fix per endpoint.

Fokus:

- HTTP write endpoints
- auth/session/token flows
- tenant/member/project CRUD
- permission/Casbin mutation
- access-right and endpoint registry mutation
- webhook/API-key creation
- upload/TUS completion side effects
- worker/email/audit side effects yang ikut terpanggil dari request path

## Evidence Paths

- `internal/router/router.go`
- `internal/modules/auth/delivery/http/auth_routes.go`
- `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/usecase/auth_session.go`
- `internal/modules/auth/usecase/auth_recovery_sso.go`
- `internal/modules/user/delivery/http/user_routes.go`
- `internal/modules/user/usecase/user_usecase.go`
- `internal/modules/organization/delivery/http/organization_routes.go`
- `internal/modules/organization/usecase/organization_member_usecase.go`
- `internal/modules/organization/repository/organization_repository.go`
- `internal/modules/organization/repository/invitation_repository.go`
- `internal/modules/project/usecase/project_usecase.go`
- `internal/modules/api_key/delivery/http/api_key_routes.go`
- `internal/modules/api_key/usecase/api_key_usecase.go`
- `internal/modules/webhook/delivery/http/webhook_routes.go`
- `internal/modules/webhook/usecase/webhook_usecase.go`
- `internal/modules/permission/delivery/http/permission_routes.go`
- `internal/modules/permission/usecase/permission_usecase.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/modules/access/delivery/http/access_routes.go`
- `internal/modules/access/usecase/access_usecase.go`
- `pkg/tus/*`
- `internal/worker/*`

## Verified Facts

- Repo tidak punya implementasi repo-wide `Idempotency-Key` atau command dedupe layer.
- Banyak endpoint write masih berbentuk command-style `POST` yang membuat token, key, UUID, relasi, atau task async baru pada tiap call.
- Banyak flow request punya side effect tambahan: email task, audit log, webhook task, cache invalidation, session/token rotation.
- Beberapa endpoint `DELETE` dan revoke sudah dekat ke idempotent secara hasil akhir, tapi belum punya contract no-op formal ketika target state sudah tercapai.
- `AcceptInvitation` adalah salah satu flow yang sudah punya guard idempotent parsial; code-nya punya cabang `If already active, do nothing (idempotent).`
- Upload TUS adalah stateful protocol; idempotency ada di level resumable upload protocol, bukan di model REST create/update biasa.

## Classification Rules Used

- **Idempotent**: request yang diulang dengan payload sama menghasilkan final state yang sama dan idealnya response success yang stabil.
- **Quasi-idempotent**: final state sering sama, tapi response, audit, task, cache, atau error path bisa beda saat repeat.
- **Non-idempotent**: repeat request yang sama bisa membuat row baru, token baru, relasi baru, side effect baru, atau command failure yang berbeda secara material.

## Endpoint Matrix

### Auth

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/auth/register` | Non-idempotent | create user + default role + default org + auto-login | duplicate account/org/session under retry or race | require `Idempotency-Key`; replay saved result; enforce stable unique handling on `email` and `username` | P0 |
| `POST /api/v1/auth/login` | Non-idempotent | issue access/refresh tokens and session each call | token churn, duplicate sessions | short dedupe window keyed by successful identity + device fingerprint; return same token pair for same in-flight/recent retry | P1 |
| `POST /api/v1/auth/refresh` | Non-idempotent | refresh rotates tokens; second retry can hit used token | first call succeeds, retry fails | single-flight refresh; store mapping from old refresh token to new pair for short TTL | P0 |
| `POST /api/v1/auth/logout` | Quasi-idempotent | revoke session then clear cookie; second call may miss session | unstable response on retry | treat already-revoked/missing session as success | P1 |
| `POST /api/v1/auth/forgot-password` | Non-idempotent | generates reset token, queues email, writes audit | duplicate emails, token churn | one active token per user; resend cooldown; dedupe email task by business key | P0 |
| `POST /api/v1/auth/reset-password` | Non-idempotent | updates password, revokes sessions, deletes token | retry sees invalid token or mismatched result | mark token as used and return stable duplicate-completion result when applicable | P1 |
| `POST /api/v1/auth/verify-email` | Non-idempotent | sets verified flag, deletes token | second call returns invalid/already verified | if already verified, return success target-state response | P1 |
| `POST /api/v1/auth/resend-verification` | Non-idempotent | new token + email each call | duplicate emails/tokens | one active verification token, cooldown, dedupe task | P0 |
| `POST /api/v1/auth/ticket` | Non-idempotent | creates one-time WS ticket each call | ticket churn under reconnect/retry | return same recent unused ticket for short TTL per session | P2 |
| `GET /api/v1/auth/sso/:provider/callback` | Non-idempotent | exchanges code, may auto-provision user/org, link SSO identity, create session | duplicate provisioning/session race | one-time state, dedupe callback by provider/code or provider identity, unique-safe upsert path, replay success | P0 |

### User

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/users/register` | Non-idempotent | creates user via admin-ish path | duplicate user on retry/race | `Idempotency-Key`; unique-safe insert and replay | P1 |
| `PUT /api/v1/users/me` | Mostly idempotent | overwrites fields | repeated audit/task on identical write | skip no-op updates; only emit side effect on actual diff | P3 |
| `PATCH /api/v1/users/me/avatar` | Quasi-idempotent | storage upload and avatar mutation are stateful | orphan files, repeated writes | dedupe by upload ID/object key; make completion hook last-write-wins and cleanup old file after commit | P1 |
| `PATCH /api/v1/users/:id/status` | Mostly idempotent | sets target status | repeated audit/callbacks | if status already target, return success no-op and skip side effects | P2 |
| `DELETE /api/v1/users/:id` | Quasi-idempotent | delete repeat can become not-found | unstable retry result | return success when already deleted | P2 |

### Organization and Membership

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/organizations` | Non-idempotent | creates org | duplicate org or slug conflict | `Idempotency-Key`; unique slug policy and replay result | P1 |
| `PUT /api/v1/organizations/:id` | Mostly idempotent | overwrites org fields | extra cache invalidation and audit | no-op detection before save | P3 |
| `DELETE /api/v1/organizations/:id` | Quasi-idempotent | soft delete repeat can diverge response | unstable delete retry | already-deleted should return success | P2 |
| `POST /api/v1/organizations/:id/restore` | Quasi-idempotent | restore command on stateful resource | retry may hit different branch | already-active should return success | P2 |
| `DELETE /api/v1/organizations/:id/hard` | Quasi-idempotent | permanent delete | second retry sees not found | already-gone should return success or stable tombstone result | P2 |
| `POST /api/v1/organizations/:id/members/invite` | Non-idempotent | creates invitation token and likely email side effect | duplicate invitation/email | one active invite per `(org_id, email)`; return existing invite if still valid; dedupe email task | P0 |
| `POST /api/v1/organizations/invitations/accept` | Quasi-idempotent | partial idempotent guard exists, but token deletion/policy/cache still stateful | retry after success may fail by token absence | if membership already active for invite target, return success even when invitation already consumed | P1 |
| `PATCH /api/v1/organizations/:id/members/:userId` | Mostly idempotent | sets target role | repeated Casbin churn | if role already target, return no-op success | P2 |
| `DELETE /api/v1/organizations/:id/members/:userId` | Quasi-idempotent | removes member and Casbin relation | second retry not found | already-absent member should return success | P2 |

### Project

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/projects` | Non-idempotent | plain create | duplicate project | `Idempotency-Key` or domain uniqueness on `(org_id, name)` if valid | P1 |
| `PUT /api/v1/projects/:id` | Mostly idempotent | overwrite update | repeated side effects | no-op detection | P3 |
| `DELETE /api/v1/projects/:id` | Quasi-idempotent | second delete differs | unstable retry result | already-deleted returns success | P2 |

### API Key

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/api-keys` | Non-idempotent | generates new secret and DB row each call | duplicate active keys, client loses first secret | require `Idempotency-Key`; persist first generated secret in replay store for TTL; return exact same response on retry | P0 |
| `DELETE /api/v1/api-keys/:id` | Quasi-idempotent | revoke/delete repeat can become not-found | unstable retry result | already-revoked should return success | P2 |

### Webhook

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/webhooks` | Non-idempotent | create new webhook config | duplicate delivery target and multiplied outgoing events | unique business key on `(organization_id, url, event_set_hash)` or `Idempotency-Key`; replay existing row on retry | P0 |
| `PUT /api/v1/webhooks/:id` | Mostly idempotent | overwrite config | repeated audit/version change | no-op detection before update | P3 |
| `DELETE /api/v1/webhooks/:id` | Quasi-idempotent | repeated delete not stable | retry instability | already-deleted returns success | P2 |

### Permission and Casbin

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/permissions/assign-role` | Quasi-idempotent | removes old role then adds target role | repeated policy churn | if target role already assigned in domain, return success unchanged and skip mutation | P1 |
| `DELETE /api/v1/permissions/revoke-role` | Quasi-idempotent | second revoke sees absent relation | retry error drift | missing relation should return success | P1 |
| `POST /api/v1/permissions/grant` | Quasi-idempotent | AddPolicy duplicate may error/conflict | repeated retries unstable | treat already-existing policy as success; precheck or normalize duplicate error | P1 |
| `DELETE /api/v1/permissions/revoke` | Quasi-idempotent | remove policy when absent can differ | retry instability | missing policy should return success | P1 |
| `POST /api/v1/permissions/inheritance` | Quasi-idempotent | add grouping policy duplicate-sensitive | repeated retries unstable | normalize existing relation to success | P1 |
| `DELETE /api/v1/permissions/inheritance` | Quasi-idempotent | second delete may fail | retry instability | missing inheritance should return success | P1 |
| `POST /api/v1/permissions/assign-access-right` | Quasi-idempotent | relation assignment duplicate-sensitive | duplicate relation/conflict | unique constraint + duplicate-as-success | P1 |
| `DELETE /api/v1/permissions/revoke-access-right` | Quasi-idempotent | repeated revoke may fail | retry instability | missing relation should return success | P1 |

### Access Registry

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `POST /api/v1/access-rights` | Non-idempotent | create new registry row | duplicate access right or conflict | unique logical key and `Idempotency-Key` for create command | P1 |
| `DELETE /api/v1/access-rights/:id` | Quasi-idempotent | repeated delete differs | unstable result | already-deleted returns success | P2 |
| `POST /api/v1/access-rights/link` | Quasi-idempotent | join relation duplicate-sensitive | duplicate relation/conflict | unique `(access_right_id, endpoint_id)` + duplicate-as-success | P1 |
| `POST /api/v1/access-rights/unlink` | Quasi-idempotent | repeated unlink differs | retry instability | absent relation returns success | P1 |
| `POST /api/v1/endpoints` | Non-idempotent | create endpoint registry row | duplicate endpoint rows/conflict | unique `(method, path)` and `Idempotency-Key` if needed | P1 |
| `DELETE /api/v1/endpoints/:id` | Quasi-idempotent | repeated delete differs | unstable result | already-deleted returns success | P2 |

### Upload and TUS

| Route | Status | Reason | Impact | Recommended Fix | Priority |
|---|---|---|---|---|---|
| `ANY /api/v1/upload/files/*any` | Stateful protocol | resumable upload and completion hook mutate domain state | orphaned file, repeated hook mutation, race on retry/resume | persist processed upload IDs; make completion hook idempotent by upload ID and target object | P0 |

## Cross-Cutting Recommendations

### 1. Add an idempotency middleware for critical command APIs

Apply first to:

- `POST /auth/register`
- `POST /auth/refresh`
- `POST /auth/forgot-password`
- `POST /auth/resend-verification`
- `POST /organizations/:id/members/invite`
- `POST /projects`
- `POST /api-keys`
- `POST /webhooks`
- `POST /access-rights`
- `POST /endpoints`

Store fields:

- idempotency key
- request fingerprint
- authenticated actor
- route/method
- response status/body
- expiration time
- processing state (`in_progress`, `completed`, `failed_replayable`)

Behavior:

- same key + same fingerprint -> replay stored response
- same key + different fingerprint -> reject with conflict
- in-progress duplicate -> return 409 or wait/replay based on timeout policy

### 2. Normalize target-state commands to no-op success

Apply to:

- logout
- delete/revoke endpoints
- restore endpoints
- verify-email when already verified
- remove member when already absent
- revoke role/policy/relation when already absent

This is low-risk and high-value.

### 3. Enforce uniqueness at persistence boundaries

Candidates:

- users by `email`, `username`
- SSO identities by `(provider, provider_id)`
- active invitation by `(organization_id, email)` or by explicit active-token selector
- projects by domain-specific business key if valid
- access-right links by `(access_right_id, endpoint_id)`
- endpoints by `(method, path)`
- webhook uniqueness by business key if product semantics allow it

Use unique violation handling to convert duplicate create into stable application response where safe.

### 4. Make worker and email side effects dedupe-aware

Use business keys such as:

- `forgot_password:user_id:bucket`
- `verification_email:user_id:bucket`
- `invite_member:org_id:email`
- `avatar_hook:upload_id`
- `webhook_delivery:webhook_id:event_id`

Handler should check whether business event already processed.

### 5. Add no-op detection before update writes

Useful for:

- profile update
- org update
- project update
- webhook update
- member role update
- user status update

Skip save, audit, cache invalidation, and webhook/task side effects when no field changes.

## Suggested Implementation Order

### Phase 1: low-risk behavior hardening

- no-op success for delete/revoke/restore/logout/verify-email
- duplicate-as-success for permission/access relation grants where relation already exists
- no-op detection for common update endpoints

### Phase 2: persistence safety

- unique constraints and duplicate handling for relation/link tables and endpoint registry
- one active invite per org/email
- one active verification/reset token policy per user

### Phase 3: command replay safety

- idempotency middleware/store for high-risk create/auth endpoints
- API-key creation replay support with secret retention for replay TTL

### Phase 4: async and upload safety

- worker dedupe keys
- TUS completion idempotent processing by upload ID
- audit/email/webhook duplicate suppression where business-safe

## Highest-Risk Gaps Right Now

1. `POST /api/v1/api-keys`
2. `POST /api/v1/auth/register`
3. `POST /api/v1/auth/refresh`
4. `GET /api/v1/auth/sso/:provider/callback`
5. `POST /api/v1/organizations/:id/members/invite`
6. `POST /api/v1/webhooks`
7. `ANY /api/v1/upload/files/*any` completion path

## Verification Ideas

- repeat same request with same payload and same idempotency key -> same response body and status
- repeat same revoke/delete/restore without idempotency key -> stable success
- trigger duplicate worker task payload -> single durable side effect
- replay upload completion callback with same upload ID -> no second domain mutation
- race test create endpoints with same business key -> one row, stable replay/conflict semantics

## Inference

- Repo saat ini lebih kuat di authorization layering daripada retry safety.
- Endpoint auth, invite, API key, webhook, dan upload punya blast radius tertinggi karena mereka membuat credential, token, outbound delivery target, atau state async baru.
- Banyak perbaikan bisa dilakukan tanpa redesign besar: normalize no-op success, add duplicate-as-success on relation mutations, and add idempotency only to highest-risk commands first.
