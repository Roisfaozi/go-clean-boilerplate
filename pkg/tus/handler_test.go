package tus

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
    "bytes"
    "sync"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tushandler "github.com/tus/tusd/v2/pkg/handler"
)

// mockStore is a basic mock for handler.DataStore
type mockStore struct {
	tushandler.DataStore
	UseInCalled bool
}

func (m *mockStore) UseIn(c *tushandler.StoreComposer) {
	m.UseInCalled = true
	c.UseCore(m)
}

// Ensure mockStore implements terminater if needed
type terminaterMockStore struct {
	mockStore
	TerminateCalled bool
	TerminateError  error
	GetUploadError  error
}

func (m *terminaterMockStore) AsTerminatableUpload(upload tushandler.Upload) tushandler.TerminatableUpload {
	return &terminatableUploadMock{store: m}
}
func (m *terminaterMockStore) GetUpload(ctx context.Context, id string) (tushandler.Upload, error) {
	if m.GetUploadError != nil {
		return nil, m.GetUploadError
	}
	return &uploadMock{}, nil
}

type terminatableUploadMock struct {
	store *terminaterMockStore
}

func (t *terminatableUploadMock) Terminate(ctx context.Context) error {
	t.store.TerminateCalled = true
	return t.store.TerminateError
}

type uploadMock struct{
	tushandler.Upload
}

type mockHook struct {
	Err error
    Called bool
    mu sync.Mutex
    done chan struct{}
}

func (m *mockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.Called = true
    if m.done != nil {
        select {
        case m.done <- struct{}{}:
        default:
        }
    }
	return m.Err
}

func TestNewHandler_Success(t *testing.T) {
	registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

	t.Run("Local Storage", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "tus-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		cfg := Config{
			StorageDriver: "local",
			LocalRootPath: tempDir,
			BasePath:      "/files/",
		}

		h, err := NewHandler(cfg, registry, nil, log)
		require.NoError(t, err)
		require.NotNil(t, h)
	})

    t.Run("Local Storage - Mkdir error", func(t *testing.T) {
        cfg := Config{
            StorageDriver: "local",
            LocalRootPath: "/dev/null/invalid",
            BasePath:      "/files/",
        }

        h, err := NewHandler(cfg, registry, nil, log)
        require.Error(t, err)
        require.Nil(t, h)
    })

	t.Run("S3 Storage", func(t *testing.T) {
		cfg := Config{
			StorageDriver: "s3",
			S3Bucket:      "my-bucket",
			BasePath:      "/files/",
		}

		h, err := NewHandler(cfg, registry, &s3.Client{}, log)
		require.NoError(t, err)
		require.NotNil(t, h)
	})
}

func TestCleanupFailedCompletedUpload(t *testing.T) {
    log := logrus.New()
    log.SetOutput(io.Discard)

    t.Run("Not Terminatable", func(t *testing.T) {
        store := &mockStore{}
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", log)
    })

    t.Run("GetUpload Error", func(t *testing.T) {
        store := &terminaterMockStore{
            GetUploadError: errors.New("get upload error"),
        }
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", log)
        assert.False(t, store.TerminateCalled)
    })

    t.Run("Terminate Error", func(t *testing.T) {
        store := &terminaterMockStore{
            TerminateError: errors.New("terminate error"),
        }
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", log)
        assert.True(t, store.TerminateCalled)
    })

    t.Run("Success", func(t *testing.T) {
        store := &terminaterMockStore{}
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", log)
        assert.True(t, store.TerminateCalled)
    })

    t.Run("Nil Logger", func(t *testing.T) {
        store := &terminaterMockStore{}
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", nil)
        assert.True(t, store.TerminateCalled)
    })

    t.Run("Nil Logger Not Terminatable", func(t *testing.T) {
        store := &mockStore{}
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", nil)
    })

    t.Run("Nil Logger GetUpload Error", func(t *testing.T) {
        store := &terminaterMockStore{
            GetUploadError: errors.New("get upload error"),
        }
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", nil)
        assert.False(t, store.TerminateCalled)
    })

    t.Run("Nil Logger Terminate Error", func(t *testing.T) {
        store := &terminaterMockStore{
            TerminateError: errors.New("terminate error"),
        }
        cleanupFailedCompletedUpload(context.Background(), store, "test-id", nil)
        assert.True(t, store.TerminateCalled)
    })
}

func TestBackgroundDispatcher(t *testing.T) {
    registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

    hook := &mockHook{
        done: make(chan struct{}, 1),
    }
    registry.Register("test-type", hook)

    tempDir, err := os.MkdirTemp("", "tus-test")
    require.NoError(t, err)
    defer func() { _ = os.RemoveAll(tempDir) }()

    cfg := Config{
        StorageDriver: "local",
        LocalRootPath: tempDir,
        BasePath:      "/files/",
    }

    h, err := NewHandler(cfg, registry, nil, log)
    require.NoError(t, err)
    require.NotNil(t, h)

    // Simulate upload event
    h.CompleteUploads <- tushandler.HookEvent{
        Upload: tushandler.FileInfo{
            ID: "test-id",
            MetaData: map[string]string{
                "type": "test-type",
            },
        },
    }

    select {
    case <-hook.done:
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for hook")
    }

    hook.mu.Lock()
    assert.True(t, hook.Called)
    hook.mu.Unlock()
}

func TestBackgroundDispatcher_S3(t *testing.T) {
    registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

    hook := &mockHook{
        done: make(chan struct{}, 1),
    }
    registry.Register("test-type", hook)

    cfg := Config{
        StorageDriver: "s3",
        S3Bucket:      "my-bucket",
        S3Endpoint:    "http://localhost:9000",
        BasePath:      "/files/",
    }

    h, err := NewHandler(cfg, registry, &s3.Client{}, log)
    require.NoError(t, err)
    require.NotNil(t, h)

    // Simulate upload event
    h.CompleteUploads <- tushandler.HookEvent{
        Upload: tushandler.FileInfo{
            ID: "test-id",
            MetaData: map[string]string{
                "type": "test-type",
            },
        },
    }

    select {
    case <-hook.done:
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for hook")
    }

    hook.mu.Lock()
    assert.True(t, hook.Called)
    hook.mu.Unlock()
}

func TestBackgroundDispatcher_Error_NilLogger(t *testing.T) {
    registry := NewRegistry()

    hook := &mockHook{
        Err: errors.New("hook error"),
        done: make(chan struct{}, 1),
    }
    registry.Register("test-type", hook)

    cfg := Config{
        StorageDriver: "s3",
        S3Bucket:      "my-bucket",
        S3Endpoint:    "http://localhost:9000",
        BasePath:      "/files/",
    }

    h, err := NewHandler(cfg, registry, &s3.Client{}, nil)
    require.NoError(t, err)
    require.NotNil(t, h)

    // Simulate upload event
    h.CompleteUploads <- tushandler.HookEvent{
        Upload: tushandler.FileInfo{
            ID: "test-id",
            MetaData: map[string]string{
                "type": "test-type",
            },
        },
    }

    select {
    case <-hook.done:
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for hook")
    }

    hook.mu.Lock()
    assert.True(t, hook.Called)
    hook.mu.Unlock()
}

func TestBackgroundDispatcher_Error(t *testing.T) {
    registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

    hook := &mockHook{
        Err: errors.New("hook error"),
        done: make(chan struct{}, 1),
    }
    registry.Register("test-type", hook)

    tempDir, err := os.MkdirTemp("", "tus-test")
    require.NoError(t, err)
    defer func() { _ = os.RemoveAll(tempDir) }()

    cfg := Config{
        StorageDriver: "local",
        LocalRootPath: tempDir,
        BasePath:      "/files/",
    }

    h, err := NewHandler(cfg, registry, nil, log)
    require.NoError(t, err)
    require.NotNil(t, h)

    // Simulate upload event
    h.CompleteUploads <- tushandler.HookEvent{
        Upload: tushandler.FileInfo{
            ID: "test-id",
            MetaData: map[string]string{
                "type": "test-type",
            },
        },
    }

    select {
    case <-hook.done:
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for hook")
    }

    hook.mu.Lock()
    assert.True(t, hook.Called)
    hook.mu.Unlock()
}

func TestPreUploadCreateCallback_HTTP(t *testing.T) {
    registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

    hook := &mockHook{}
    registry.Register("test-type", hook)

	tempDir, err := os.MkdirTemp("", "tus-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: tempDir,
		BasePath:      "/files/",
	}

	h, err := NewHandler(cfg, registry, nil, log)
	require.NoError(t, err)
	require.NotNil(t, h)

    mux := http.NewServeMux()
    mux.Handle("/files/", http.StripPrefix("/files/", h))

    // Inject auth context manually using a wrapper handler
    authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := authcontext.WithUserID(r.Context(), "user-id-123")
        r = r.WithContext(ctx)
        mux.ServeHTTP(w, r)
    })

    req := httptest.NewRequest(http.MethodPost, "/files/", bytes.NewReader(nil))
    req.Header.Set("Tus-Resumable", "1.0.0")
    req.Header.Set("Upload-Length", "100")
    req.Header.Set("Upload-Metadata", "type dGVzdC10eXBl")

    w := httptest.NewRecorder()
    authHandler.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    req2 := httptest.NewRequest(http.MethodPost, "/files/", bytes.NewReader(nil))
    req2.Header.Set("Tus-Resumable", "1.0.0")
    req2.Header.Set("Upload-Length", "100")
    req2.Header.Set("Upload-Metadata", "type aW52YWxpZC10eXBl")

    w2 := httptest.NewRecorder()
    authHandler.ServeHTTP(w2, req2)

    assert.Equal(t, http.StatusBadRequest, w2.Code)

    req3 := httptest.NewRequest(http.MethodPost, "/files/", bytes.NewReader(nil))
    req3.Header.Set("Tus-Resumable", "1.0.0")
    req3.Header.Set("Upload-Length", "100")

    w3 := httptest.NewRecorder()
    mux.ServeHTTP(w3, req3)

    assert.Equal(t, http.StatusUnauthorized, w3.Code)
}
