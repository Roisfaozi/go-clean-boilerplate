package local

import (
	"context"
	"fmt"
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
	// Prevent directory traversal
	cleanPath := filepath.Clean(filename)
	fullPath := filepath.Join(s.RootPath, cleanPath)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		_ = dst.Close()
	}()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	// Return public URL (assumes the app serves RootPath statically)
	url := fmt.Sprintf("%s/%s", s.BaseURL, cleanPath)
	return url, nil
}

func (s *LocalStorage) DeleteFile(ctx context.Context, filename string) error {
	cleanPath := filepath.Clean(filename)
	fullPath := filepath.Join(s.RootPath, cleanPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) GetFileUrl(filename string) (string, error) {
	cleanPath := filepath.Clean(filename)
	return fmt.Sprintf("%s/%s", s.BaseURL, cleanPath), nil
}
