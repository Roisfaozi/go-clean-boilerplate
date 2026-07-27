package tus

import (
	"context"
	"errors"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"io"
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

func TestNewHandler(t *testing.T) {
	t.Run("s3 storage driver", func(t *testing.T) {
		cfg := Config{
			StorageDriver: "s3",
			S3Bucket:      "my-bucket",
			BasePath:      "/files/",
		}
		registry := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		handler, err := NewHandler(cfg, registry, nil, log)

		assert.NoError(t, err)
		assert.NotNil(t, handler)
	})
}

func TestNewHandler_LocalStore(t *testing.T) {
	t.Run("local storage driver success", func(t *testing.T) {
		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: "/tmp/tus-test",
			BasePath:      "/files/",
		}
		registry := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		handler, err := NewHandler(cfg, registry, nil, log)

		assert.NoError(t, err)
		assert.NotNil(t, handler)
	})

	t.Run("local storage driver mkdir failure", func(t *testing.T) {
		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: "\x00invalidpath",
			BasePath:      "/files/",
		}
		registry := NewRegistry()
		log := logrus.New()
		log.SetOutput(io.Discard)

		handler, err := NewHandler(cfg, registry, nil, log)

		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "failed to create tus directory")
	})
}

type fakeHook struct {
	handled chan bool
	err     error
}

func (h *fakeHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	h.handled <- true
	return h.err
}

func TestNewHandler_Dispatcher(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: "/tmp/tus-test-dispatcher-local",
		BasePath:      "/files/",
	}
	registry := NewRegistry()
	hook := &fakeHook{handled: make(chan bool, 1)}
	registry.Register("avatar", hook)

	log := logrus.New()
	log.SetOutput(io.Discard)

	tusHandler, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)

	// Send an event
	event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
		Upload: handler.FileInfo{
			ID: "upload123",
			MetaData: map[string]string{
				"type": "avatar",
			},
		},
	}

	tusHandler.CompleteUploads <- event

	// Wait for goroutine
	<-hook.handled
}

func TestNewHandler_Dispatcher_S3(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "my-bucket",
		S3Endpoint:    "http://s3.local",
	}
	registry := NewRegistry()
	hook := &fakeHook{handled: make(chan bool, 1)}
	registry.Register("avatar", hook)

	log := logrus.New()
	log.SetOutput(io.Discard)

	tusHandler, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)

	// Send an event
	event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
		Upload: handler.FileInfo{
			ID: "upload123",
			MetaData: map[string]string{
				"type": "avatar",
			},
		},
	}

	tusHandler.CompleteUploads <- event
	<-hook.handled
}

func TestNewHandler_Dispatcher_HookError(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: "/tmp/tus-test-dispatcher-error",
		BasePath:      "/files/",
	}
	registry := NewRegistry()
	hook := &fakeHook{err: errors.New("hook error"), handled: make(chan bool, 1)}
	registry.Register("avatar", hook)

	log := logrus.New()
	log.SetOutput(io.Discard)

	tusHandler, err := NewHandler(cfg, registry, nil, log)
	assert.NoError(t, err)

	event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
		Upload: handler.FileInfo{
			ID: "upload123",
			MetaData: map[string]string{
				"type": "avatar",
			},
		},
	}

	tusHandler.CompleteUploads <- event
	<-hook.handled
}

func TestNewHandler_Dispatcher_HookError_NilLogger(t *testing.T) {
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: "/tmp/tus-test-dispatcher-error-nil-log",
		BasePath:      "/files/",
	}
	registry := NewRegistry()
	hook := &fakeHook{err: errors.New("hook error"), handled: make(chan bool, 1)}
	registry.Register("avatar", hook)

	tusHandler, err := NewHandler(cfg, registry, nil, nil)
	assert.NoError(t, err)

	event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
		Upload: handler.FileInfo{
			ID: "upload123",
			MetaData: map[string]string{
				"type": "avatar",
			},
		},
	}

	tusHandler.CompleteUploads <- event
	<-hook.handled
}

func TestCleanupFailedCompletedUpload_Logs(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	t.Run("terminates upload when store supports termination", func(t *testing.T) {
		upload := &fakeTerminatableUpload{}
		store := &fakeTerminatableStore{upload: upload}

		cleanupFailedCompletedUpload(context.Background(), store, "upload-1", log)

		assert.True(t, upload.terminated)
	})

	t.Run("no panic when store does not support termination", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), &fakeCoreStore{}, "upload-1", log)
		})
	})

	t.Run("no panic when upload lookup fails", func(t *testing.T) {
		store := &fakeTerminatableStore{getErr: errors.New("not found")}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", log)
		})
	})

	t.Run("no panic when termination fails", func(t *testing.T) {
		upload := &fakeTerminatableUpload{terminateErr: errors.New("delete failed")}
		store := &fakeTerminatableStore{upload: upload}

		assert.NotPanics(t, func() {
			cleanupFailedCompletedUpload(context.Background(), store, "upload-1", log)
		})
		assert.False(t, upload.terminated)
	})
}

func TestDefaultPreUploadCreateCallback(t *testing.T) {
	registry := NewRegistry()
	hook := &fakeHook{handled: make(chan bool, 1)}
	registry.Register("avatar", hook)

	t.Run("success", func(t *testing.T) {
		event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
			HTTPRequest: handler.HTTPRequest{
				Header: map[string][]string{
					"X-Organization-Id": {"org123"},
					"X-User-Id":         {"user123"},
				},
			},
			Upload: handler.FileInfo{
				MetaData: map[string]string{
					"type": "avatar",
				},
			},
		}

		_, _, err := defaultPreUploadCreateCallback(event, registry)
		assert.NoError(t, err)
	})

	t.Run("auth metadata bind failure", func(t *testing.T) {
		event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
			HTTPRequest: handler.HTTPRequest{
				Header: map[string][]string{},
			},
			Upload: handler.FileInfo{
				MetaData: map[string]string{},
			},
		}

		_, _, err := defaultPreUploadCreateCallback(event, registry)
		assert.Error(t, err)
	})

	t.Run("validate upload metadata failure", func(t *testing.T) {
		event := handler.HookEvent{Context: authcontext.WithUserID(context.Background(), "user123"),
			HTTPRequest: handler.HTTPRequest{
				Header: map[string][]string{
					"X-Organization-Id": {"org123"},
					"X-User-Id":         {"user123"},
				},
			},
			Upload: handler.FileInfo{
				MetaData: map[string]string{
					"type": "invalid-type",
				},
			},
		}

		_, _, err := defaultPreUploadCreateCallback(event, registry)
		assert.Error(t, err)
	})
}
