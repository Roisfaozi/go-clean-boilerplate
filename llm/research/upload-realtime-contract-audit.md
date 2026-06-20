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

## Needs Confirmation

- full hook registry and which types are trusted
- how avatar upload is keyed to authenticated identity in current flow
