# TUS Upload System

## Purpose

Durable map for TUS upload route, storage providers, metadata hooks, and upload completion behavior.

## Runtime truth

- Upload route is separate from normal CRUD groups in `internal/router/router.go`.
- `/api/v1/upload/files/*any` uses `authMiddleware.ValidateToken()` plus `UserStatusMiddleware`, then wraps `tusHandler`.
- `pkg/tus` owns auth helper, handler setup, registry, and Swagger docs.
- `pkg/storage` provides storage abstraction used by app/module wiring.
- `internal/config/app.go` wires storage provider and TUS handler.

## Storage support

- `pkg/tus/handler.go` supports local file store and S3-compatible store through tusd.
- Storage config is environment-driven through app config and `.env.example` categories.

## Hard rules

- Do not treat TUS route as normal JSON CRUD.
- Preserve auth/status middleware on upload route.
- Upload metadata and completion hook routing are trust-boundary sensitive.
- Preserve context propagation into storage/request-scoped operations.

## Tests and evidence paths

- `pkg/tus/*_test.go`
- `pkg/storage`
- `internal/router/router.go`
- `internal/config/app.go`
