# Auth Boundary Split Audit

## Scope

Phase 2 audit untuk memecah blast radius `internal/modules/auth/usecase/auth_usecase.go` tanpa langsung mengubah behavior.

## Evidence Paths

- `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/test/auth_usecase_test.go`
- `internal/modules/auth/repository/token_repository.go`
- `internal/middleware/auth_middleware.go`

## Current Auth Responsibilities

- credential login and lockout
- JWT pair generation
- Redis-backed session store/verify/revoke/revoke-all
- refresh-token rotation
- registration user creation
- default role assignment
- default organization provisioning
- password reset
- email verification
- audit task enqueue
- SSO login/callback session creation
- WebSocket ticket issuance

## Existing Test Signals

- login success/failure, timing and lockout paths
- token refresh success/failure and cleanup failure behavior
- revoke token and revoke-all behavior
- reset password, verify email, replay edge cases
- registration success and provisioning path
- middleware session verification path

## Split Seams

### Session Service

Owns:

- token pair generation
- session persist
- verify session
- revoke one session
- revoke all sessions
- refresh rotation

Candidate owner files:

- new `internal/modules/auth/usecase/session_service.go`
- existing `internal/modules/auth/repository/token_repository.go`

### Credential Service

Owns:

- login lookup
- password comparison
- login attempts and lockout
- default role lookup for login response

Candidate owner files:

- new `internal/modules/auth/usecase/credential_service.go`

### Registration Provisioning Service

Owns:

- user creation
- default role assignment
- default organization creation
- registration audit enqueue

Candidate owner files:

- new `internal/modules/auth/usecase/registration_service.go`

### Recovery Service

Owns:

- forgot password token
- reset password
- email verification token
- verify email

Candidate owner files:

- new `internal/modules/auth/usecase/recovery_service.go`

### SSO Service

Owns:

- provider state/callback flow
- SSO user lookup/create/link
- SSO session issuance

Candidate owner files:

- new `internal/modules/auth/usecase/sso_service.go`

### Ticket Service

Owns:

- WebSocket ticket issuance and session context mapping

Candidate owner files:

- new `internal/modules/auth/usecase/ticket_service.go`

## Refactor Rules

- no behavior changes in first extraction pass
- move code behind same exported `AuthUseCase` interface
- keep public controller contract unchanged
- move tests only after narrow package tests pass
- do not split transaction scope accidentally

## Phase 2 Next Implementation Slices

1. Extract session helpers first because session tests are strongest.
2. Extract credential login logic second.
3. Extract registration provisioning third.
4. Extract recovery and SSO after behavior remains green.

## Verification

- `go test ./internal/modules/auth/... ./internal/middleware/...`
- integration auth tests if session repository behavior changes
- E2E auth tests if controller response shape changes
