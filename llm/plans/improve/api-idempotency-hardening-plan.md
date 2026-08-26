# API Idempotency Hardening Plan

## Goal

Kurangi retry-induced duplicate side effects pada API write path tanpa refactor besar lintas repo di fase awal.

## Inputs

- `llm/research/api-idempotency-audit.md`
- `internal/router/router.go`
- `internal/modules/auth/*`
- `internal/modules/organization/*`
- `internal/modules/api_key/*`
- `internal/modules/webhook/*`
- `internal/modules/permission/*`
- `internal/modules/access/*`
- `pkg/tus/*`
- `internal/worker/*`

## Non-Goals

- Tidak mendesain exactly-once semantics global.
- Tidak mengubah seluruh route `POST` menjadi `PUT`/`PATCH`.
- Tidak menambah distributed transaction baru.

## Phase 1 — Low-Risk Target-State Semantics

### Scope

Normalisasi endpoint yang seharusnya aman saat di-retry agar return success ketika target state sudah tercapai.

### Candidate endpoints

- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/verify-email`
- `DELETE` revoke/delete endpoints di permission, access, user, project, org, api-key, webhook
- `POST /api/v1/organizations/:id/restore`
- `DELETE /api/v1/organizations/:id/members/:userId`

### Work

- ubah usecase agar missing target relation/session/resource bisa dianggap no-op success bila aman
- bedakan genuine authorization/not-found boundary vs already-applied command
- tambahkan regression test untuk repeat delete/revoke/restore

### Verify

- repeat command kedua return success stabil
- tidak ada side effect tambahan pada retry kedua

## Phase 2 — Duplicate-as-Success for Relation Mutations

### Scope

Harden relation-style writes agar duplicate grant/link/assign tidak gagal secara noisy.

### Candidate endpoints

- `POST /api/v1/permissions/grant`
- `POST /api/v1/permissions/inheritance`
- `POST /api/v1/permissions/assign-access-right`
- `POST /api/v1/access-rights/link`
- matching revoke/unlink paths

### Work

- precheck existence atau normalize duplicate-store/enforcer error menjadi success
- tambahkan unique constraints/join uniqueness bila belum ada
- expose `unchanged`/`already_exists` semantics bila response wrapper mendukung

### Verify

- same relation add twice -> one durable relation, second response stable success
- same relation remove twice -> second response stable success

## Phase 3 — Command Idempotency for High-Risk Create Flows

### Scope

Tambahkan replay-safe protection pada endpoint yang paling berbahaya saat client/proxy retry.

### Highest-priority endpoints

- `POST /api/v1/api-keys`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/resend-verification`
- `POST /api/v1/organizations/:id/members/invite`
- `POST /api/v1/projects`
- `POST /api/v1/webhooks`
- `POST /api/v1/access-rights`
- `POST /api/v1/endpoints`

### Work

- add `Idempotency-Key` middleware/store
- persist request fingerprint + response snapshot + processing state
- reject same key with different fingerprint
- replay stored response for same key and same payload
- add TTL and cleanup strategy

### Verify

- same request + same key -> exact same response
- same key + different payload -> conflict
- concurrent duplicate requests -> single durable create result

## Phase 4 — Auth Token and SSO Retry Safety

### Scope

Harden auth flows yang sulit diberi idempotency key murni.

### Candidate flows

- login
- refresh
- SSO callback
- websocket ticket issue

### Work

- single-flight / short dedupe window for refresh
- recent successful login replay window per identity/device fingerprint
- one-time SSO state use + callback dedupe by provider identity or code hash
- recent unused ticket replay for websocket bootstrap

### Verify

- retry refresh tidak gagal spuriously setelah first success
- duplicate SSO callback tidak create duplicate user/org/identity
- repeated ticket request in short window behaves predictably

## Phase 5 — Invite, Email, and Token Lifecycle Hardening

### Scope

Pastikan token/email flows tidak spam atau menghasilkan token aktif berlebihan.

### Work

- one active password reset token per user
- one active verification token per user
- one active invite per `(org_id, email)`
- resend cooldown policy
- worker email dedupe by business key

### Verify

- repeated forgot-password/resend-verification/invite within cooldown do not multiply active token count
- email task retry does not send duplicate email beyond intended policy

## Phase 6 — Upload and Worker Side-Effect Idempotency

### Scope

Stateful async boundary yang tidak aman terhadap retry.

### Candidate flows

- TUS completion hook
- webhook delivery
- audit task sync
- email task handling

### Work

- persist processed upload IDs
- dedupe outbound webhook delivery by business event key where feasible
- ensure worker retries do not duplicate durable audit rows
- document remaining at-least-once semantics where exactly-once not feasible

### Verify

- replay same upload completion does not reapply avatar mutation twice
- duplicate worker payload does not create duplicate durable row/effect where promised

## Persistence and Schema Candidates

- unique user email
- unique username
- unique SSO identity `(provider, provider_id)`
- active invite uniqueness for `(organization_id, email)`
- unique endpoint `(method, path)`
- unique access-right link `(access_right_id, endpoint_id)`
- optional webhook business uniqueness if product semantics allow

## Testing Strategy

1. unit tests for no-op success semantics
2. package integration tests for duplicate create/assign/revoke behavior
3. concurrency tests for create endpoints with same idempotency key/business key
4. worker/upload retry tests for duplicate payload processing
5. auth/session retry tests for refresh and SSO callback

## Rollout Notes

- start with endpoints that create credentials or outbound fanout targets
- keep behavior changes additive and backwards-compatible where possible
- where response semantics change from error to success-no-op, note this in API changelog/release notes

## Recommended First Slice

1. make delete/revoke/restore/logout/verify-email no-op-safe
2. normalize duplicate permission/access relation mutations to success
3. add idempotency to `POST /api/v1/api-keys`
4. add one-active-invite policy to org invite flow
5. harden TUS completion hook against duplicate processing
