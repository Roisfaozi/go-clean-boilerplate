package storage

import (
	"context"
	"io"
)

// Provider defines the interface for file storage operations.
type Provider interface {
	// UploadFile uploads a file with the given content and returns its URL/path
	UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)

	// DeleteFile removes a file from storage
	DeleteFile(ctx context.Context, filename string) error

	// GetFileUrl returns the accessible URL for the file
	GetFileUrl(filename string) (string, error)
}
