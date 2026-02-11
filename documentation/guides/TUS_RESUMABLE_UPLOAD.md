# Global Resumable Upload Guide (Tus + RustFS)

This document details the architecture and implementation of the **Global Resumable Upload Service** in the Go Clean Boilerplate project.

## 1. Overview & Architecture

We have replaced traditional single-shot uploads with the **Tus Protocol** (via `tusd`) to support resumable uploads, suitable for unstable networks and large files.

### Key Concepts

1.  **Global Service**: A single endpoint (`/api/v1/upload/files/`) handles ALL file types (Avatars, Documents, Galleries).
2.  **Storage Agnostic**: Uses **RustFS** (S3-Compatible) for physical storage.
3.  **Event-Driven (Registry Pattern)**: The core Tus handler is generic. Logic for specific features (e.g., "Update User Avatar in DB") is decoupled using **Hooks**.

### Data Flow

1.  **Client** starts upload with metadata: `{ type: "avatar", user_id: "123" }`.
2.  **Tus Handler** receives chunks and saves them to RustFS.
3.  **Completion**: When upload finishes, Tus Handler looks up the `type` ("avatar") in the **Registry**.
4.  **Hook Execution**: The registered `AvatarHook` is executed to update the database.

---

## 2. Infrastructure Setup

### 2.1 Docker Compose (RustFS)

Add the RustFS service to your `docker-compose.dev.yml`:

```yaml
services:
  rustfs:
    image: rustfs/rustfs:latest
    container_name: rustfs-server
    ports:
      - '9000:9000' # S3 API
      - '9001:9001' # Console UI
    environment:
      - RUSTFS_ACCESS_KEY=rustfsadmin
      - RUSTFS_SECRET_KEY=rustfsadmin
      - RUSTFS_CONSOLE_ENABLE=true
      - RUSTFS_VOLUMES=/data
    volumes:
      - ./db/rustfs-data:/data
```

### 2.2 Environment Variables (`.env`)

Configure the S3 driver to point to RustFS:

```ini
# Storage Configuration
STORAGE_DRIVER=s3
STORAGE_S3_ENDPOINT=http://localhost:9000
STORAGE_S3_BUCKET=uploads-bucket
STORAGE_S3_ACCESS_KEY=rustfsadmin
STORAGE_S3_SECRET_KEY=rustfsadmin
STORAGE_S3_REGION=us-east-1
STORAGE_S3_USE_SSL=false
STORAGE_S3_FORCE_PATH_STYLE=true

# Tus Configuration
TUS_BASE_PATH=/api/v1/upload/files/
```

---

## 3. Backend Implementation (Core)

### 3.1 The Registry (`pkg/tus/registry.go`)

This defines the contract for other modules.

```go
package tus

import "context"

type UploadEvent struct {
	UploadID string            // The Tus ID (e.g., abc-123)
	FileURL  string            // The Full S3 URL
	Metadata map[string]string // Metadata sent by client
}

// UploadHook must be implemented by Feature Modules (User, Project, etc)
type UploadHook interface {
	HandleUpload(ctx context.Context, event UploadEvent) error
}

type Registry struct {
	hooks map[string]UploadHook
}

func NewRegistry() *Registry {
	return &Registry{hooks: make(map[string]UploadHook)}
}

func (r *Registry) Register(uploadType string, hook UploadHook) {
	r.hooks[uploadType] = hook
}

func (r *Registry) Get(uploadType string) UploadHook {
	return r.hooks[uploadType]
}
```

### 3.2 The Handler (`pkg/tus/handler.go`)

Initializes `tusd` and dispatches events.

```go
package tus

// ... imports (tusd, aws-sdk-v2) ...

func NewHandler(cfg Config, registry *Registry, s3Client *s3.Client, log *logrus.Logger) (*handler.Handler, error) {
    // 1. Configure AWS SDK v2 for RustFS
    // ... setup s3Client ...
    store := s3store.New(cfg.S3Bucket, s3Client)

    // 2. Create Composer
    composer := handler.NewStoreComposer()
    composer.UseCore(store)

    // 3. Create Handler with Notifications Enabled
    tusHandler, err := handler.NewHandler(handler.Config{
        BasePath:              cfg.BasePath,
        StoreComposer:         composer,
        NotifyCompleteUploads: true, // Critical for hooks
    })

    // 4. Background Dispatcher
    go func() {
        for {
            event := <-tusHandler.CompleteUploads
            meta := event.Upload.MetaData
            uploadType := meta["type"]

            if hook := registry.Get(uploadType); hook != nil {
                fileURL := fmt.Sprintf("%s/%s/%s", cfg.S3Endpoint, cfg.S3Bucket, event.Upload.ID)

                // Dispatch to specific module
                err := hook.HandleUpload(context.Background(), UploadEvent{
                    UploadID: event.Upload.ID,
                    FileURL:  fileURL,
                    Metadata: meta,
                })
                if err != nil {
                    log.Errorf("Hook error for %s: %v", uploadType, err)
                }
            }
        }
    }()

    return tusHandler, nil
}
```

---

## 4. Middleware Configuration

### 4.1 CORS (`internal/middleware/cors_middleware.go`)

Tus requires specific headers. Update your CORS config:

```go
config.AllowHeaders = []string{
    "Authorization", "Content-Type", "X-Org-ID",
    // Tus Headers:
    "Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset",
    "Upload-Protocol", "Upload-Draft-Interop-Version",
}
config.ExposeHeaders = []string{
    "Upload-Offset", "Location", "Upload-Length", "Tus-Version",
    "Tus-Resumable", "Tus-Max-Size", "Tus-Extension", "Upload-Metadata",
}
config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
```

### 4.2 Routing (`internal/router/router.go`)

```go
// 1. Init Registry & Register Hooks
registry := tus.NewRegistry()
registry.Register("avatar", userModule.AvatarHook)

// 2. Init Handler
tusHandler, _ := tus.NewHandler(cfg, registry, s3Client, logger)

// 3. Register Route
uploadGroup := router.Group("/api/v1/upload")
uploadGroup.Use(authMiddleware.ValidateToken())
{
    // StripPrefix is required for tusd to handle relative paths correctly
    uploadGroup.Any("/files/*any", gin.WrapH(http.StripPrefix("/api/v1/upload/files/", tusHandler)))
}
```

---

## 5. Feature Implementation (How to Add New Uploads)

### Example: Adding Avatar Upload

You do **NOT** modify `pkg/tus`. You only work in `internal/modules/user`.

1.  **Create Hook:** `internal/modules/user/usecase/avatar_hook.go`

    ```go
    type AvatarHook struct {
        UserUseCase UserUseCase
    }

    func (h *AvatarHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
        userID := event.Metadata["user_id"]
        return h.UserUseCase.UpdateAvatarUrl(ctx, userID, event.FileURL)
    }
    ```

2.  **Register:** In `internal/config/app.go`, register it with the key `"avatar"`.

### Example: Adding Document Upload (Future)

1.  **Create Hook:** `internal/modules/project/usecase/doc_hook.go`
    ```go
    func (h *DocHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
        projectID := event.Metadata["project_id"]
        return h.Repo.SaveDocument(ctx, projectID, event.FileURL)
    }
    ```
2.  **Register:** Register with key `"project_doc"`.

---

## 6. Frontend Usage (Client)

Use `tus-js-client` or `uppy`. The critical part is the **Metadata**.

```javascript
import * as tus from 'tus-js-client'

const upload = new tus.Upload(file, {
  endpoint: 'http://localhost:8080/api/v1/upload/files/',
  headers: {
    Authorization: 'Bearer <YOUR_JWT>',
  },
  // METADATA DRIVES THE LOGIC
  metadata: {
    type: 'avatar', // Must match the key registered in backend
    user_id: 'user-123', // Data needed by the Hook
    filename: file.name,
  },
  onSuccess: () => console.log('Upload finished:', upload.url),
})

upload.start()
```

## 7. Troubleshooting

- **CORS Error:** Check if `Tus-Resumable` and `Upload-Offset` are in `ExposeHeaders`.
- **404 on PATCH:** Ensure `http.StripPrefix` matches your `BasePath`.
- **RustFS Error:** Ensure `ForcePathStyle=true` is set in AWS Config (RustFS requires path-style access).
