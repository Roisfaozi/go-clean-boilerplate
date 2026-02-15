# Multi-Provider Storage Guide

This project implements a Strategy Pattern for file storage, allowing you to switch between local storage and S3-compatible providers (AWS S3, MinIO, Cloudflare R2) via simple configuration.

## Interface Definition

The storage abstraction is defined in `pkg/storage/provider.go`:

```go
type Provider interface {
    UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
    DeleteFile(ctx context.Context, filename string) error
    GetFileUrl(filename string) (string, error)
}
```

## Available Providers

### 1. Local Storage (`local`)
Stores files on the server's local disk.
- **Use Case**: Development, single-instance deployments.
- **Config**:
  - `STORAGE_DRIVER=local`
  - `STORAGE_LOCAL_ROOT_PATH=./uploads`
  - `STORAGE_LOCAL_BASE_URL=http://localhost:8080/uploads`

### 2. S3-Compatible Storage (`s3`)
Supports any S3-compatible API.
- **Use Case**: Production, horizontal scaling, cloud-native deployments.
- **Providers Tested**: AWS S3, MinIO, Cloudflare R2.
- **Config**:
  - `STORAGE_DRIVER=s3`
  - `STORAGE_S3_ENDPOINT`: Your provider endpoint (e.g., `https://<accountid>.r2.cloudflarestorage.com`)
  - `STORAGE_S3_BUCKET`: Target bucket name.
  - `STORAGE_S3_FORCE_PATH_STYLE`: Set `true` for MinIO.

## How to Use in UseCases

1.  **Inject the Provider**: Ensure your UseCase struct includes `storage.Provider`.
2.  **Call Upload**:

```go
func (u *myUseCase) UpdateAvatar(ctx context.Context, file io.Reader, filename string) error {
    // 1. Upload to storage
    url, err := u.Storage.UploadFile(ctx, file, "avatars/"+filename, "image/png")
    if err != nil {
        return err
    }

    // 2. Save URL to database
    return u.Repo.UpdateAvatarUrl(ctx, url)
}
```

## Security Considerations

-   **Path Sanitization**: The Local provider uses `filepath.Clean` to prevent directory traversal attacks.
-   **Presigned URLs**: For the S3 provider, `GetFileUrl` generates a time-limited presigned URL (default: 1 hour) instead of exposing public ACLs.
-   **Size Limits**: Always validate file size in the Controller layer before passing it to the UseCase.
