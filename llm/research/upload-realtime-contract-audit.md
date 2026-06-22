# Upload Realtime Contract Audit

## Scope

Phase 6 audit for TUS upload, storage, SSE, WS, and presence.

## Evidence Paths

- `pkg/tus/*`
- `pkg/storage/*`
- `pkg/ws/*`
- `pkg/sse/*`
- `internal/router/router.go`
- `internal/modules/user/*`

## Failure Modes

- upload metadata can act as trust boundary if client-supplied fields override server intent
- completion hook may mutate wrong user/org if token or metadata mismatches
- WS presence can drift for multi-connection users if unregister logic is too simple
- Redis pub/sub can duplicate broadcast or leak stale presence if cleanup incomplete

## What to Verify

- upload route is session-only and rejects API-key-only access
- avatar completion only mutates owner target
- upload completion failures clean storage or leave explicit retry trail
- presence update handles multiple connections per user

## Confirmed Implementation Notes

- Current hook registry in app wiring registers `avatar` as the active upload type.
- TUS create callback binds server-side `user_id` and `authenticated_user_id` from request context and overwrites any client-provided `user_id` metadata.
- Avatar upload completion now requires `authenticated_user_id`; legacy/client-controlled `user_id` alone is rejected and cannot mutate another user's avatar.
- TUS local and S3 stores are registered through tusd `UseIn`, preserving termination support; when a completion hook fails, completed upload storage is terminated when the store supports it and emits explicit warnings otherwise.
- WebSocket manager already preserves presence for remaining same-user/same-org connections before emitting leave/offline.
- Redis ticket and SSE HTTP tests require localhost sockets in sandbox; the user-provided host run verified them successfully.

## Current Verification

- Passed: `go test ./pkg/tus ./internal/modules/user/test -run 'TestBindAuthenticatedMetadata|TestValidateUploadMetadata|TestAvatarHook' -v`
- Passed: `go test ./pkg/tus ./internal/modules/user/test -run 'TestBindAuthenticatedMetadata|TestValidateUploadMetadata|TestAvatarHook|TestCleanupFailedCompletedUpload' -v`
- Passed: `go test ./pkg/storage/...`
- Passed from host/socket-capable shell: `go test ./pkg/ws -run 'TestWebSocketManager_UnregisterKeepsPresenceForOtherConnections|TestWebSocketManager_UnregisterNilSendClient|TestRedisTicketManager|TestPresenceManager' -v && go test ./pkg/sse -v`
