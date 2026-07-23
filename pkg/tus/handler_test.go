package tus

import (
	"context"
	"errors"
	"io"
	"testing"
	"sync"
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

type fakeHook struct {
	err error
	wg  *sync.WaitGroup
}

func (h *fakeHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	if h.wg != nil {
		h.wg.Done()
	}
	return h.err
}

func TestNewHandler_DispatcherSuccess(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	reg := NewRegistry()
	var wg sync.WaitGroup
	wg.Add(1)
	reg.Register("test_type", &fakeHook{err: nil, wg: &wg})

	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)
	_ = h

	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "upload-123",
			MetaData: map[string]string{
				"type": "test_type",
			},
		},
	}
	wg.Wait()
}

func TestNewHandler_DispatcherError(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	reg := NewRegistry()
	var wg sync.WaitGroup
	wg.Add(1)
	reg.Register("test_type", &fakeHook{err: errors.New("hook error"), wg: &wg})

	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)
	_ = h

	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "upload-123",
			MetaData: map[string]string{
				"type": "test_type",
			},
		},
	}
	wg.Wait()
	// give dispatcher time to call cleanup
	time.Sleep(50 * time.Millisecond)
}

func TestNewHandler_LocalDriver(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: t.TempDir(),
		BasePath:      "/files/",
	}
	reg := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewHandler_MkdirError(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		// use null byte to force an error in MkdirAll
		LocalRootPath: "\x00invalidpath",
		BasePath:      "/files/",
	}
	reg := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)
	h, err := NewHandler(cfg, reg, nil, log)
	assert.Error(t, err)
	assert.Nil(t, h)
}

func TestNewHandler_S3Driver(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "my-bucket",
		S3Endpoint:    "http://localhost:9000",
		BasePath:      "/files/",
	}
	reg := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)
	// pass nil for s3 client since s3store.New doesn't immediately use it to connect in v2
	h, err := NewHandler(cfg, reg, nil, log)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestCleanupFailedCompletedUpload_LogNoTerminater(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	assert.NotPanics(t, func() {
		cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", log)
	})
}

func TestCleanupFailedCompletedUpload_LogTerminateError(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
	store := &fakeTerminatableStore{upload: upload}

	assert.NotPanics(t, func() {
		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", log)
	})
}

func TestCleanupFailedCompletedUpload_LogTerminateSuccess(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	upload := &fakeTerminatableUpload{}
	store := &fakeTerminatableStore{upload: upload}

	assert.NotPanics(t, func() {
		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", log)
	})
}
