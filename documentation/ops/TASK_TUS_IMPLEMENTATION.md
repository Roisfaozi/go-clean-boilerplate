# TASK: Implement Global Resumable Upload (Tus + RustFS)

## 1. Objective

Replace single-shot uploads with a robust, resumable upload service using the **Tus Protocol** backed by **RustFS** (S3-compatible). The implementation must use a **Registry Pattern** to allow modular extensions (Avatar, Documents, etc.) without modifying the core upload handler.

## 2. Pre-requisites

- [x] Go 1.25+ installed.
- [x] Docker & Docker Compose running.
- [x] RustFS container configuration ready.

---

## 3. Implementation Plan

### Phase 1: Infrastructure & Dependencies (Estimated: 1 Hour)

**Goal:** Prepare the environment and install necessary libraries.

- [x] **1.1 Install Go Libraries:**
  - `github.com/tus/tusd/v2`
  - `github.com/aws/aws-sdk-go-v2` (Core, Config, S3, Credentials)
- [x] **1.2 Configure Docker Compose:**
  - Add `rustfs` service to `docker-compose.dev.yml`.
  - Map port `9000` (API) and `9001` (Console).
- [x] **1.3 Update Environment Variables:**
  - Add `STORAGE_S3_*` configs pointing to RustFS in `.env` and `.env.example`.
  - Add `TUS_BASE_PATH`.

### Phase 2: Core TUS Implementation (Estimated: 2 Hours)

**Goal:** Build the generic TUS handler and Event Registry.

- [x] **2.1 Create Registry Package (`pkg/tus/registry.go`):**
  - Define `UploadEvent` struct.
  - Define `UploadHook` interface.
  - Implement thread-safe `Registry` struct.
- [x] **2.2 Create Handler Package (`pkg/tus/handler.go`):**
  - Implement `NewHandler` function.
  - Configure AWS SDK v2 connection to RustFS.
  - Setup `tusd` Composer with S3Store.
  - Implement background goroutine to listen for `CompleteUploads` and dispatch events to Registry.

### Phase 3: Middleware & Routing (Estimated: 1 Hour)

**Goal:** Expose the TUS handler via Gin router securely.

- [x] **3.1 Update CORS Middleware (`internal/middleware/cors_middleware.go`):**
  - Add `Tus-Resumable`, `Upload-Length`, `Upload-Metadata` to `AllowHeaders`.
  - Add `Location`, `Upload-Offset` to `ExposeHeaders`.
  - Allow `PATCH`, `HEAD` methods.
- [x] **3.2 Wiring in Router (`internal/router/router.go`):**
  - Initialize TUS Registry & Handler in `SetupRouter`.
  - Create group `/api/v1/upload`.
  - Apply `AuthMiddleware`.
  - Mount handler with `http.StripPrefix`.

### Phase 4: Feature Integration (Avatar Example) (Estimated: 1 Hour)

**Goal:** Connect the User module to the new TUS system.

- [x] **4.1 Create Avatar Hook (`internal/modules/user/usecase/avatar_hook.go`):**
  - Implement `HandleUpload` method.
  - Call `UserUseCase.UpdateAvatarUrl` with the new S3 URL.
- [x] **4.2 Register Hook (`internal/config/app.go`):**
  - Register the avatar hook with key `"avatar"` into the TUS Registry.

### Phase 5: Verification & Cleanup (Estimated: 1 Hour)

**Goal:** Verify end-to-end functionality.

- [x] **5.1 Manual Testing:**
  - Run `make docker-dev` and `make run`.
  - Create bucket `uploads-bucket` in RustFS Console.
  - Use Postman/Curl to hit `HEAD /api/v1/upload/files/` (Check headers).
  - Use a TUS client (e.g., Uppy demo) to upload a file.
  - Verify file appears in RustFS.
  - Verify database is updated (if testing Avatar flow).
- [x] **5.2 Documentation:**
  - Update API documentation if necessary.

---

## 4. Rollback Plan

If critical issues arise:

1.  Revert changes to `internal/router/router.go` (Disable TUS route).
2.  Revert `internal/middleware/cors_middleware.go` (Restore original CORS).
3.  The original `PATCH /users/me/avatar` endpoint remains untouched and functional throughout this process.
