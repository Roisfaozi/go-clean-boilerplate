package tus

import (
	"context"
	"errors"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tus/tusd/v2/pkg/handler"
)

type fakeTerminatableStore struct {
	upload *fakeTerminatableUpload
	getErr error
}

func (s *fakeTerminatableStore) NewUpload(ctx context.Context, info handler.FileInfo) (handler.Upload, error) {
	return s.upload, nil
}

func (s *fakeTerminatableStore) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.upload, nil
}

func (s *fakeTerminatableStore) AsTerminatableUpload(upload handler.Upload) handler.TerminatableUpload {
	return upload.(handler.TerminatableUpload)
}

func (s *fakeTerminatableStore) UseIn(composer *handler.StoreComposer) {
	composer.UseCore(s)
	composer.UseTerminater(s)
}

type fakeCoreStore struct{}

func (s *fakeCoreStore) NewUpload(ctx context.Context, info handler.FileInfo) (handler.Upload, error) {
	return &fakeTerminatableUpload{}, nil
}

func (s *fakeCoreStore) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	return &fakeTerminatableUpload{}, nil
}

type fakeTerminatableUpload struct {
	terminateErr error
	terminated   bool
}

func (u *fakeTerminatableUpload) WriteChunk(ctx context.Context, offset int64, src io.Reader) (int64, error) {
	return 0, nil
}

func (u *fakeTerminatableUpload) GetInfo(ctx context.Context) (handler.FileInfo, error) {
	return handler.FileInfo{ID: "upload-1"}, nil
}

func (u *fakeTerminatableUpload) GetReader(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

func (u *fakeTerminatableUpload) FinishUpload(ctx context.Context) error {
	return nil
}

func (u *fakeTerminatableUpload) Terminate(ctx context.Context) error {
	if u.terminateErr != nil {
		return u.terminateErr
	}
	u.terminated = true
	return nil
}

func TestCleanupFailedCompletedUpload(t *testing.T) {
	t.Run("terminates upload when store supports termination", func(t *testing.T) {
		upload := &fakeTerminatableUpload{}
		store := &fakeTerminatableStore{upload: upload}

		logger := logrus.New()
		logger.SetOutput(io.Discard)

		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)

		assert.True(t, upload.terminated)
	})

	t.Run("no panic when store does not support termination", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", logger)
		})
	})

	t.Run("no panic when upload lookup fails", func(t *testing.T) {
		store := &fakeTerminatableStore{getErr: errors.New("not found")}

		assert.NotPanics(t, func() {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)
		})
	})

	t.Run("no panic when termination fails", func(t *testing.T) {
		upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)
		})
		assert.False(t, upload.terminated)
	})

	t.Run("nil logger", func(t *testing.T) {
		upload := &fakeTerminatableUpload{}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})
		assert.True(t, upload.terminated)
	})

	t.Run("nil logger errors", func(t *testing.T) {
		store := &fakeTerminatableStore{getErr: errors.New("not found")}
		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})

		upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
		store = &fakeTerminatableStore{upload: upload}
		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", nil)
		})
	})
}

type mockHook struct {
	handleUploadFunc func(ctx context.Context, event UploadEvent) error
}

func (m *mockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	if m.handleUploadFunc != nil {
		return m.handleUploadFunc(ctx, event)
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	t.Run("s3 store", func(t *testing.T) {
		cfg := Config{
			StorageDriver: "s3",
			S3Bucket:      "test-bucket",
			S3Endpoint:    "http://s3.local",
			BasePath:      "/files/",
		}
		reg := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		// This will create a basic S3 client with fake credentials if needed,
		// but since s3store.New doesn't immediately ping S3, passing nil might work,
		// let's try nil for now, or just a generic empty one.
		_, err := NewHandler(cfg, reg, &s3.Client{}, log)
		assert.NoError(t, err)
	})

	t.Run("local store success", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: tmpDir,

			BasePath: "/files/",
		}
		reg := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		_, err := NewHandler(cfg, reg, nil, log)
		assert.NoError(t, err)
	})

	t.Run("local store failure", func(t *testing.T) {
		// use a null byte to fail MkdirAll
		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: "\x00invalidpath",
			BasePath:      "/files/",
		}
		reg := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		_, err := NewHandler(cfg, reg, nil, log)
		assert.Error(t, err)
	})

	// We will test background dispatcher indirectly by triggering a mock event
}
func TestBackgroundDispatcher(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: tmpDir,

		BasePath: "/files/",
	}
	reg := NewRegistry()

	handledChan := make(chan bool, 1)
	mock := &mockHook{
		handleUploadFunc: func(ctx context.Context, event UploadEvent) error {
			assert.Equal(t, "upload-id-123", event.UploadID)
			assert.Equal(t, "/files//upload-id-123", event.FileURL)
			handledChan <- true
			return nil
		},
	}
	reg.Register("user_avatar", mock)

	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)

	// Simulate complete upload event
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "upload-id-123",
			MetaData: handler.MetaData{
				"type": "user_avatar",
			},
		},
	}

	select {
	case <-handledChan:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for hook")
	}
}

func TestBackgroundDispatcherS3(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "bucket",
		S3Endpoint:    "http://s3",
	}
	reg := NewRegistry()

	handledChan := make(chan bool, 1)
	mock := &mockHook{
		handleUploadFunc: func(ctx context.Context, event UploadEvent) error {
			assert.Equal(t, "upload-id-456", event.UploadID)
			assert.Equal(t, "http://s3/bucket/upload-id-456", event.FileURL)
			handledChan <- true
			return nil
		},
	}
	reg.Register("doc", mock)

	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, &s3.Client{}, log)
	assert.NoError(t, err)

	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "upload-id-456",
			MetaData: handler.MetaData{
				"type": "doc",
			},
		},
	}

	select {
	case <-handledChan:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for hook")
	}
}

func TestBackgroundDispatcherError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: tmpDir,
	}
	reg := NewRegistry()

	handledChan := make(chan bool, 1)
	mock := &mockHook{
		handleUploadFunc: func(ctx context.Context, event UploadEvent) error {
			handledChan <- true
			return errors.New("hook error")
		},
	}
	reg.Register("error_type", mock)

	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)

	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "upload-id-err",
			MetaData: handler.MetaData{
				"type": "error_type",
			},
		},
	}

	select {
	case <-handledChan:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for hook")
	}
	// wait slightly for cleanup to run (it runs asynchronously anyway, but since it's the same goroutine...)
	// Actually the dispatcher goroutine runs it sequentially.

}
func TestPreUploadCreateCallback(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: tmpDir,
	}
	reg := NewRegistry()
	reg.Register("test_type", &mockHook{})

	log := logrus.New()
	log.SetOutput(io.Discard)

	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Length", "100")
		// "type test_type" -> dHlwZSB0ZXN0X3R5cGU=
		// actually standard tus metadata parsing parses comma separated key space base64value.
		req.Header.Set("Upload-Metadata", "type dGVzdF90eXBl")

		ctx := authcontext.WithUserID(req.Context(), "user-123")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Length", "100")

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid metadata type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Length", "100")
		// invalid type
		req.Header.Set("Upload-Metadata", "type aW52YWxpZA==") // base64 for 'invalid'

		ctx := authcontext.WithUserID(req.Context(), "user-123")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNewHandlerError(t *testing.T) {
	// This requires handler.NewHandler to fail, which is hard unless StoreComposer is nil or invalid.
	// Since we're using "local" default and current directory (LocalRootPath = ""),
	// MkdirAll might succeed but then filestore path is empty.

	// Actually we only need MkdirAll to fail, which we already covered.

}
