package tus

import (
	"context"
	"errors"
	"io"
	"testing"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"time"

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

		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)

		assert.True(t, upload.terminated)
	})

	t.Run("no panic when store does not support termination", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", nil)
		})
	})

	t.Run("no panic when upload lookup fails", func(t *testing.T) {
		store := &fakeTerminatableStore{getErr: errors.New("not found")}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})
	})

	t.Run("no panic when termination fails", func(t *testing.T) {
		upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", nil)
		})
		assert.False(t, upload.terminated)
	})
}

// Add test to cover log output in cleanupFailedCompletedUpload
func TestCleanupFailedCompletedUpload_WithLog(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// terminater fails to get upload
	storeGetErr := &fakeTerminatableStore{getErr: errors.New("not-found")}
	cleanupFailedCompletedUpload(context.Background(), storeGetErr, "not-found", log)

	// terminater fails to terminate
	upload := &fakeTerminatableUpload{terminateErr: errors.New("term failed")}
	storeTermErr := &fakeTerminatableStore{upload: upload}
	cleanupFailedCompletedUpload(context.Background(), storeTermErr, "test-id", log)

	// store not terminater
	storeNotTerminater := &fakeCoreStore{}
	cleanupFailedCompletedUpload(context.Background(), storeNotTerminater, "test-id", log)

	// success
	uploadSuccess := &fakeTerminatableUpload{}
	storeSuccess := &fakeTerminatableStore{upload: uploadSuccess}
	cleanupFailedCompletedUpload(context.Background(), storeSuccess, "test-id", log)
}

type mockHook struct {
	handledChan chan struct{}
	err         error
	received    UploadEvent
}

func (m *mockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	m.received = event
	if m.handledChan != nil {
		m.handledChan <- struct{}{}
	}
	return m.err
}

func TestNewHandler_BackgroundDispatcher_HookError_WithLog(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	registry := NewRegistry()

	handledChan := make(chan struct{}, 1)
	mockHookInstance := &mockHook{
		handledChan: handledChan,
		err:         errors.New("hook failed"),
	}
	registry.Register("test-type", mockHookInstance)

	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)

	// Trigger the dispatcher
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-upload-id",
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}

	select {
	case <-handledChan:
		// wait a little bit to make sure cleanup code ran
		time.Sleep(50 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for background dispatcher")
	}
}

func TestNewHandler_LocalStore(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir() + "/\x00invalidpath", // Invalid path
	}
	registry := NewRegistry()
	h, err := NewHandler(cfg, registry, nil, log)
	assert.Error(t, err)
	assert.Nil(t, h)
}

func TestNewHandler_S3Store(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "test-bucket",
		S3Endpoint:    "http://localhost:9000",
	}
	registry := NewRegistry()
	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewHandler_Success(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	registry := NewRegistry()
	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewHandler_BackgroundDispatcher(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	registry := NewRegistry()

	handledChan := make(chan struct{}, 1)
	mockHookInstance := &mockHook{handledChan: handledChan}
	registry.Register("test-type", mockHookInstance)

	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)

	// Trigger the dispatcher
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-upload-id",
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}

	select {
	case <-handledChan:
		assert.Equal(t, "test-upload-id", mockHookInstance.received.UploadID)
		assert.Equal(t, "/files//test-upload-id", mockHookInstance.received.FileURL)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for background dispatcher")
	}
}

func TestNewHandler_BackgroundDispatcher_S3(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "test-bucket",
		S3Endpoint:    "http://localhost:9000",
	}
	registry := NewRegistry()

	handledChan := make(chan struct{}, 1)
	mockHookInstance := &mockHook{handledChan: handledChan}
	registry.Register("test-type", mockHookInstance)

	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)

	// Trigger the dispatcher
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-upload-id",
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}

	select {
	case <-handledChan:
		assert.Equal(t, "test-upload-id", mockHookInstance.received.UploadID)
		assert.Equal(t, "http://localhost:9000/test-bucket/test-upload-id", mockHookInstance.received.FileURL)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for background dispatcher")
	}
}

func TestNewHandler_BackgroundDispatcher_HookError(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	registry := NewRegistry()

	handledChan := make(chan struct{}, 1)
	mockHookInstance := &mockHook{
		handledChan: handledChan,
		err:         errors.New("hook failed"),
	}
	registry.Register("test-type", mockHookInstance)

	h, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)

	// Trigger the dispatcher
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-upload-id",
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}

	select {
	case <-handledChan:
		// wait a little bit to make sure cleanup code ran
		time.Sleep(50 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for background dispatcher")
	}
}

func TestNewHandler_BackgroundDispatcher_HookError_NilLog(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	registry := NewRegistry()

	handledChan := make(chan struct{}, 1)
	mockHookInstance := &mockHook{
		handledChan: handledChan,
		err:         errors.New("hook failed"),
	}
	registry.Register("test-type", mockHookInstance)

	h, err := NewHandler(cfg, registry, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, h)

	// Trigger the dispatcher
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-upload-id",
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}

	select {
	case <-handledChan:
		// wait a little bit to make sure cleanup code ran
		time.Sleep(50 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for background dispatcher")
	}
}

func TestPreUploadCreateCallback_Extracted(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test-type", &mockHook{})
	cb := GetPreUploadCreateCallback(registry)

	// 1. Bind error (No user ID in context)
	ctx := context.Background()
	hookFailBind := handler.HookEvent{
		Context: ctx,
		Upload: handler.FileInfo{
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}
	_, _, err := cb(hookFailBind)
	assert.Error(t, err)
}

func TestPreUploadCreateCallback_Extracted_ValidateError(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test-type", &mockHook{})
	cb := GetPreUploadCreateCallback(registry)

	ctx := authcontext.WithUserID(context.Background(), "user-123")
	hookFailValidate := handler.HookEvent{
		Context: ctx,
		Upload: handler.FileInfo{
			MetaData: handler.MetaData{
				// missing "type"
			},
		},
	}
	_, _, err := cb(hookFailValidate)
	assert.Error(t, err)
}

func TestPreUploadCreateCallback_Extracted_Success(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test-type", &mockHook{})
	cb := GetPreUploadCreateCallback(registry)

	ctx := authcontext.WithUserID(context.Background(), "user-123")
	hookSuccess := handler.HookEvent{
		Context: ctx,
		Upload: handler.FileInfo{
			MetaData: handler.MetaData{
				"type": "test-type",
			},
		},
	}
	_, changes, err := cb(hookSuccess)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", changes.MetaData["user_id"])
}
