package local_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalStorage(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Create Success", func(t *testing.T) {
		ls, err := local.NewLocalStorage(tempDir, "http://localhost")
		assert.NoError(t, err)
		assert.NotNil(t, ls)
	})

	t.Run("Create Default Path", func(t *testing.T) {
		// Caution: this creates ./uploads in the current directory
		// We should clean it up or skip if we want to avoid side effects in source tree
		// But for coverage we might need to test it.
		// Alternatively, we skip this if we can't easily mock os.MkdirAll failure here.
		// Let's pass empty string which defaults to ./uploads
		ls, err := local.NewLocalStorage("", "http://localhost")
		assert.NoError(t, err)
		assert.Equal(t, "./uploads", ls.RootPath)
		_ = os.RemoveAll("./uploads")
	})
}

func TestUploadFile(t *testing.T) {
	tempDir := t.TempDir()
	baseURL := "http://localhost/files"
	ls, err := local.NewLocalStorage(tempDir, baseURL)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		content := "test content"
		reader := strings.NewReader(content)
		filename := "test.txt"

		url, err := ls.UploadFile(context.Background(), reader, filename, "text/plain")
		assert.NoError(t, err)
		assert.Equal(t, baseURL+"/"+filename, url)

		// Verify file exists
		data, err := os.ReadFile(filepath.Join(tempDir, filename))
		assert.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("Path Traversal", func(t *testing.T) {
		content := "hacker content"
		reader := strings.NewReader(content)
		// Should be cleaned to just "passwd" inside root path
		filename := "../../../etc/passwd"

		url, err := ls.UploadFile(context.Background(), reader, filename, "text/plain")
		assert.NoError(t, err)
		assert.Equal(t, baseURL+"/passwd", url) // filepath.Clean removes ..

		// Verify it was written to tempDir/passwd
		_, err = os.Stat(filepath.Join(tempDir, "passwd"))
		assert.NoError(t, err)
	})
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	baseURL := "http://localhost/files"
	ls, err := local.NewLocalStorage(tempDir, baseURL)
	require.NoError(t, err)

	filename := "delete_me.txt"
	fullPath := filepath.Join(tempDir, filename)
	err = os.WriteFile(fullPath, []byte("content"), 0644)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := ls.DeleteFile(context.Background(), filename)
		assert.NoError(t, err)

		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Not Exist", func(t *testing.T) {
		err := ls.DeleteFile(context.Background(), "nonexistent.txt")
		assert.NoError(t, err) // Should return nil
	})
}

func TestGetFileUrl(t *testing.T) {
	ls := &local.LocalStorage{BaseURL: "http://localhost"}

	url, err := ls.GetFileUrl("test.png")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/test.png", url)

	url, err = ls.GetFileUrl("../test.png")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/test.png", url)
}

func TestUploadFile_CreateError(t *testing.T) {
	// Use a read-only directory to trigger error
	tempDir := t.TempDir()
	err := os.Chmod(tempDir, 0500) // Read/Execute only, no Write
	require.NoError(t, err)
	defer func() { _ = os.Chmod(tempDir, 0700) }() // Cleanup

	ls := &local.LocalStorage{RootPath: tempDir, BaseURL: "http://localhost"}

	reader := strings.NewReader("content")
	_, err = ls.UploadFile(context.Background(), reader, "fail.txt", "text/plain")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create destination file")
}
