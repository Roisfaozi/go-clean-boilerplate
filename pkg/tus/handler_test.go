package tus

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
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
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	t.Run("terminates upload when store supports termination", func(t *testing.T) {
		upload := &fakeTerminatableUpload{}
		store := &fakeTerminatableStore{upload: upload}

		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)

		assert.True(t, upload.terminated)
	})

	t.Run("no panic when store does not support termination", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", logger)
		})
	})

	t.Run("no panic when upload lookup fails", func(t *testing.T) {
		store := &fakeTerminatableStore{getErr: errors.New("not found")}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)
		})
	})

	t.Run("no panic when termination fails", func(t *testing.T) {
		upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", logger)
		})
		assert.False(t, upload.terminated)
	})

	t.Run("no panic when logger is nil", func(t *testing.T) {
		upload := &fakeTerminatableUpload{}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", nil)
		})

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeTerminatableStore{getErr: errors.New("not found")}, "upload-1", nil)
		})

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeTerminatableStore{upload: &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}}, "upload-1", nil)
		})
	})
}

// MockHook for background dispatcher testing
type testMockHook struct {
	mu     sync.Mutex
	called bool
	err    error
	done   chan struct{}
}

func (m *testMockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()

	if m.done != nil {
		select {
		case m.done <- struct{}{}:
		default:
		}
	}

	return m.err
}

func (m *testMockHook) isCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func TestNewHandler_Success(t *testing.T) {
	registry := NewRegistry()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}

	h, err := NewHandler(cfg, registry, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewHandler_S3_Success(t *testing.T) {
	registry := NewRegistry()
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "my-bucket",
		S3Endpoint:    "http://localhost:9000",
		BasePath:      "/files/",
	}

	h, err := NewHandler(cfg, registry, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewHandler_LocalMkdirError(t *testing.T) {
	registry := NewRegistry()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: "\x00invalidpath", // Invalid path
		BasePath:      "/files/",
	}

	h, err := NewHandler(cfg, registry, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, h)
}

func TestNewHandler_PreUploadCreateCallback(t *testing.T) {
	registry := NewRegistry()
	registry.Register("avatar", &testMockHook{})

	// Create request with unauthenticated context
	req, _ := http.NewRequest("POST", "/files/", nil)
	hookEvent := handler.HookEvent{
		Context: req.Context(),
		Upload: handler.FileInfo{
			MetaData: handler.MetaData{"type": "avatar"},
		},
	}

	// PreUploadCreateCallback uses BindAuthenticatedMetadata and ValidateUploadMetadata
	// We can test those functions directly since we can't access h.config
	_, _, err := BindAuthenticatedMetadata(hookEvent)
	assert.Error(t, err) // Unauthenticated

	// Create request with authenticated context but missing type
	ctx := authcontext.WithUserID(req.Context(), "user-123")
	hookEvent.Context = ctx
	hookEvent.Upload.MetaData = handler.MetaData{} // Missing type

	_, _, err = ValidateUploadMetadata(hookEvent.Upload.MetaData, registry)
	assert.Error(t, err) // Missing type

	// Create request with authenticated context and invalid type
	hookEvent.Upload.MetaData = handler.MetaData{"type": "invalid"}
	_, _, err = ValidateUploadMetadata(hookEvent.Upload.MetaData, registry)
	assert.Error(t, err) // Invalid type

	// Create request with authenticated context and valid type
	hookEvent.Upload.MetaData = handler.MetaData{"type": "avatar"}
	_, _, err = ValidateUploadMetadata(hookEvent.Upload.MetaData, registry)
	assert.NoError(t, err) // Success
}

func TestNewHandler_BackgroundDispatcher(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	t.Run("local dispatcher success", func(t *testing.T) {
		registry := NewRegistry()
		hook := &testMockHook{done: make(chan struct{}, 1)}
		registry.Register("avatar", hook)

		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: t.TempDir(),
			BasePath:      "/files",
		}

		h, err := NewHandler(cfg, registry, nil, logger)
		assert.NoError(t, err)

		h.CompleteUploads <- handler.HookEvent{
			Upload: handler.FileInfo{
				ID: "upload-1",
				MetaData: handler.MetaData{
					"type": "avatar",
				},
			},
		}

		select {
		case <-hook.done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for hook")
		}

		assert.True(t, hook.isCalled())
	})

	t.Run("s3 dispatcher success", func(t *testing.T) {
		registry := NewRegistry()
		hook2 := &testMockHook{done: make(chan struct{}, 1)}
		registry.Register("avatar2", hook2)

		cfg := Config{
			StorageDriver: "s3",
			S3Bucket:      "my-bucket",
			S3Endpoint:    "http://localhost:9000",
			BasePath:      "/files",
		}

		h, err := NewHandler(cfg, registry, nil, logger)
		assert.NoError(t, err)

		h.CompleteUploads <- handler.HookEvent{
			Upload: handler.FileInfo{
				ID: "upload-2",
				MetaData: handler.MetaData{
					"type": "avatar2",
				},
			},
		}

		select {
		case <-hook2.done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for hook")
		}

		assert.True(t, hook2.isCalled())
	})

	t.Run("dispatcher hook error", func(t *testing.T) {
		registry := NewRegistry()
		errHook := &testMockHook{err: errors.New("hook error"), done: make(chan struct{}, 1)}
		registry.Register("document", errHook)

		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: t.TempDir(),
			BasePath:      "/files",
		}

		h, err := NewHandler(cfg, registry, nil, logger)
		assert.NoError(t, err)

		h.CompleteUploads <- handler.HookEvent{
			Upload: handler.FileInfo{
				ID: "upload-3",
				MetaData: handler.MetaData{
					"type": "document",
				},
			},
		}

		select {
		case <-errHook.done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for hook")
		}

		assert.True(t, errHook.isCalled())
	})

	t.Run("dispatcher hook error with nil logger", func(t *testing.T) {
		registry := NewRegistry()
		errHook2 := &testMockHook{err: errors.New("hook error 2"), done: make(chan struct{}, 1)}
		registry.Register("document2", errHook2)

		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: t.TempDir(),
			BasePath:      "/files",
		}

		h, err := NewHandler(cfg, registry, nil, nil)
		assert.NoError(t, err)

		h.CompleteUploads <- handler.HookEvent{
			Upload: handler.FileInfo{
				ID: "upload-4",
				MetaData: handler.MetaData{
					"type": "document2",
				},
			},
		}

		select {
		case <-errHook2.done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for hook")
		}

		assert.True(t, errHook2.isCalled())
	})
}
