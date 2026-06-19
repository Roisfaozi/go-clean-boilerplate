# TUS Upload System

## Purpose

Durable map for resumable uploads in this repo:

- upload route ownership
- auth boundary for upload requests
- TUS handler wiring
- metadata binding and trusted fields
- hook dispatch by upload type
- storage backend selection
- avatar completion side effect

Use this file before changing `pkg/tus`, upload routes, storage drivers, avatar upload flow, or metadata-handling code.

## Primary source of truth

1. `internal/router/router.go`
2. `internal/config/app.go`
3. `pkg/tus/handler.go`
4. `pkg/tus/auth.go`
5. `pkg/tus/registry.go`
6. `pkg/storage/*`
7. `internal/modules/user/usecase/avatar_hook.go`
8. `internal/modules/user/test/avatar_hook_test.go`

## Runtime ownership

### Route ownership

- upload route is separate group: `/api/v1/upload`
- upload route is not under normal `authenticated`, `tenantAuthorized`, or `authorized` CRUD groups
- route is `Any("/files/*any", ...)` and wrapped with `http.StripPrefix("/api/v1/upload/files/", tusHandler)`

### Middleware ownership

Upload group currently uses:

1. `authMiddleware.ValidateToken()`
2. `UserStatusMiddleware`

Implication:

- upload access requires authenticated user token
- user status still enforced
- tenant middleware and Casbin are not in route group by default, so metadata and hook logic must preserve trust boundaries carefully

### App wiring ownership

`internal/config/app.go` proves:

- TUS registry is created at boot
- upload type `avatar` is registered to `userUseCase.AvatarHook`
- storage provider and TUS handler are initialized from env-backed storage config
- AWS S3 client for TUS can be initialized separately from generic storage provider

## Handler behavior

`pkg/tus/handler.go` owns three critical behaviors:

### Storage selection

- `storage_driver == s3` uses `tusd` S3 store
- otherwise defaults to local file store
- local mode creates root path with `os.MkdirAll`

### Pre-create metadata binding

- `PreUploadCreateCallback` calls `BindAuthenticatedMetadata(hook)`
- this is trust boundary: authenticated server-side identity can override or augment client metadata before upload record is created

### Completion dispatch

- `NotifyCompleteUploads` is enabled
- background goroutine reads `tusHandler.CompleteUploads`
- `meta["type"]` selects hook from registry
- hook receives `UploadEvent{UploadID, FileURL, Metadata}`

## File URL behavior

Completion URL is built differently by storage mode:

- S3: `S3Endpoint/S3Bucket/<uploadID>`
- local: `BasePath/<uploadID>`

If changing URL construction, audit frontend/client expectations and avatar persistence semantics.

## Avatar upload contract

Current concrete hook path:

- upload metadata type `avatar`
- registry lookup finds `userUseCase.AvatarHook`
- hook calls `SetAvatarURL`
- tests prove `authenticated_user_id` should take precedence over client-provided `user_id`

That precedence is security-relevant. Do not relax it.

## Trust and security boundaries

- TUS route is not ordinary JSON controller flow.
- upload metadata is partly client-controlled and partly server-bound.
- completion hooks can mutate durable user data.
- authenticated identity must remain authoritative for user-owned uploads.
- storage driver or endpoint config can change resulting file URL contract.

## Cross-system coupling

Upload changes can affect:

- `user` profile/avatar behavior
- `pkg/storage` provider semantics
- auth middleware and session rules
- frontend avatar/profile flows
- object storage infra and local filesystem permissions

Read with:

- `llm/cache/user-system.md`
- `llm/cache/authentication-system.md`
- `llm/cache/tenant-organization-system.md` if uploads ever become org-scoped

## Known sharp edges

- route uses `Any`; method-specific assumptions can be wrong.
- local and S3 stores produce different file URL forms.
- hook dispatch depends on metadata `type`; typo or renamed type silently changes behavior.
- background completion dispatcher means side effects happen after upload completion event, not in same request-response style as CRUD handler.

## Change checklist

Before editing upload code, prove these answers:

1. Is change on route protection, metadata binding, storage backend, or hook dispatch?
2. Does metadata contain trusted server-derived identity fields?
3. Does changed upload type map to registered hook in app wiring?
4. Does file URL shape remain compatible with consumer code?
5. Does avatar update still prefer authenticated identity over client metadata?
6. Does change require frontend profile or file-preview audit?

## Verification paths

Narrow first:

- `pkg/tus/*_test.go`
- `internal/modules/user/test/avatar_hook_test.go`
- `pkg/storage/*_test.go`
- `internal/router/router.go`
- `internal/config/app.go`

Broader when auth or avatar flow changed:

- target user module tests
- integration flow covering authenticated upload + profile/avatar readback

## Hard rules

- Do not treat TUS path like normal JSON CRUD route.
- Do not trust client `user_id` over authenticated metadata.
- Do not drop auth or user-status middleware from upload route.
- Do not change upload `type` semantics without updating registry wiring.
- Do not break storage-context propagation or file URL consistency without consumer audit.
