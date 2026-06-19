# TUS Upload System

## Purpose

Durable map for TUS upload route, storage providers, metadata hooks, and upload completion behavior.

## Runtime truth

- Upload route is separate from normal CRUD groups in `internal/router/router.go`.
- `/api/v1/upload/files/*any` uses `authMiddleware.ValidateToken()` plus `UserStatusMiddleware`, then wraps `tusHandler`.
- `pkg/tus` owns auth helper, handler setup, registry, and upload-related docs.
- `pkg/storage` provides storage abstraction used by app and module wiring.
- `internal/config/app.go` wires storage provider and TUS handler.

## Route and trust boundary

- Upload route is not standard JSON CRUD route.
- Auth and user-status middleware still protect upload path.
- Upload metadata and completion hook behavior are trust-sensitive.

## Behavior surfaces

- TUS upload route handling
- metadata parsing and validation
- local or S3-compatible storage provider behavior
- completion hooks that may trigger downstream updates such as avatar behavior
- request-scoped context propagation into storage and hook logic

## Coupling to other systems

- user avatar flows can depend on upload completion
- storage provider configuration is environment-driven
- auth and status middleware still gate upload access

## Hard rules

- Do not treat TUS route as normal JSON CRUD.
- Preserve auth and user-status middleware on upload route.
- Preserve context propagation into storage and request-scoped operations.
- Treat metadata and completion-hook routing as trust-boundary sensitive.

## Verification and evidence paths

- `pkg/tus/*_test.go`
- `pkg/storage/*`
- `internal/router/router.go`
- `internal/config/app.go`
- `internal/modules/user/test/avatar_hook_test.go`
