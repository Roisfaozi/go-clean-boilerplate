package local

import (
	"context"
	"fmt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	RootPath string
	BaseURL  string
}

func NewLocalStorage(rootPath, baseURL string) (*LocalStorage, error) {
	if rootPath == "" {
		rootPath = "./uploads"
	}
	// Ensure directory exists
	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &LocalStorage{
		RootPath: rootPath,
		BaseURL:  baseURL,
	}, nil
}

func (s *LocalStorage) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// Prevent directory traversal by using filepath.Base
	baseName := filepath.Base(filename)
	fullPath := filepath.Join(s.RootPath, baseName)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		telemetry.StorageUploadsTotal.WithLabelValues("local", "failed").Inc()
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		_ = dst.Close()
	}()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		telemetry.StorageUploadsTotal.WithLabelValues("local", "failed").Inc()
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	// Return public URL (assumes the app serves RootPath statically)
	telemetry.StorageUploadsTotal.WithLabelValues("local", "success").Inc()
	url := fmt.Sprintf("%s/%s", s.BaseURL, baseName)
	return url, nil
}

func (s *LocalStorage) DeleteFile(ctx context.Context, filename string) error {
	baseName := filepath.Base(filename)
	fullPath := filepath.Join(s.RootPath, baseName)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) GetFileUrl(ctx context.Context, filename string) (string, error) {
	baseName := filepath.Base(filename)
	return fmt.Sprintf("%s/%s", s.BaseURL, baseName), nil
}
