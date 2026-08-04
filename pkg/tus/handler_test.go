package tus

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
)

type mockUploadHook struct {
	mock.Mock
	mu sync.Mutex
}

func (m *mockUploadHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestNewHandler_LocalStore(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: filepath.Join(tempDir, "uploads"),
		BasePath:      "/uploads/",
	}

	registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

	h, err := NewHandler(cfg, registry, nil, log)
	require.NoError(t, err)
	assert.NotNil(t, h)

	// Test Mkdir failure by passing an invalid path
	cfgInvalid := Config{
		StorageDriver: "local",
		LocalRootPath: "\x00invalidpath",
	}
	_, err = NewHandler(cfgInvalid, registry, nil, log)
	assert.Error(t, err)
}

func TestNewHandler_S3Store(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "test-bucket",
		S3Endpoint:    "http://localhost:9000",
		BasePath:      "/uploads/",
	}

	registry := NewRegistry()
	log := logrus.New()
	log.SetOutput(io.Discard)

	s3Client := s3.New(s3.Options{})
	h, err := NewHandler(cfg, registry, s3Client, log)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

func TestPreUploadCreateCallback(t *testing.T) {
	registry := NewRegistry()
	registry.Register("user_avatar", &mockUploadHook{})

	cb := PreUploadCreateCallback(registry)

	// Valid case
	req, _ := http.NewRequest("POST", "/", nil)

	meta := map[string]string{
		"type": "user_avatar",
		"filename": "test.jpg",
	}

	ctx := context.Background()
	ctx = authcontext.WithUserID(ctx, "user1")

	hook := handler.HookEvent{
		Context: ctx,
		HTTPRequest: handler.HTTPRequest{
			RemoteAddr: "127.0.0.1",
		},
		Upload: handler.FileInfo{
			MetaData: meta,
		},
	}

	hook.HTTPRequest.Header = req.Header

	resp, changes, err := cb(hook)
	assert.NoError(t, err)
	assert.Equal(t, "user1", changes.MetaData["user_id"])
	assert.Equal(t, 0, resp.StatusCode)

	// Error from Auth (missing org id / unauthenticated)
	req2, _ := http.NewRequest("POST", "/", nil)
	hook.HTTPRequest.Header = req2.Header
	hook.Context = context.Background() // Missing user ID
	_, _, err = cb(hook)
	assert.Error(t, err)

	// Error from Meta (missing type)
	hook.Context = ctx
	hook.Upload.MetaData = map[string]string{"filename": "test"}
	_, _, err = cb(hook)
	assert.Error(t, err)
}

func TestBackgroundDispatcher(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		StorageDriver: "local",
		LocalRootPath: filepath.Join(tempDir, "uploads"),
		BasePath:      "/uploads/",
	}

	registry := NewRegistry()
	hookMock := &mockUploadHook{}
	registry.Register("user_avatar", hookMock)

	log := logrus.New()
	log.SetOutput(io.Discard)

	h, err := NewHandler(cfg, registry, nil, log)
	require.NoError(t, err)

	handledChan := make(chan struct{})

	// Add expectation
	hookMock.On("HandleUpload", mock.Anything, mock.AnythingOfType("tus.UploadEvent")).Run(func(args mock.Arguments) {
		close(handledChan)
	}).Return(nil)

	// Send an event
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-id",
			MetaData: map[string]string{
				"type": "user_avatar",
			},
		},
	}

	select {
	case <-handledChan:
		// Passed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for background dispatcher")
	}

	hookMock.AssertExpectations(t)
}

func TestBackgroundDispatcher_Error(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "test-bucket",
		S3Endpoint:    "http://localhost:9000",
		BasePath:      "/uploads/",
	}

	registry := NewRegistry()
	hookMock := &mockUploadHook{}
	registry.Register("doc", hookMock)

	log := logrus.New()
	log.SetOutput(io.Discard)

	s3Client := s3.New(s3.Options{})
	h, err := NewHandler(cfg, registry, s3Client, log)
	require.NoError(t, err)

	handledChan := make(chan struct{})

	// Add expectation to return error
	hookMock.On("HandleUpload", mock.Anything, mock.AnythingOfType("tus.UploadEvent")).Run(func(args mock.Arguments) {
		close(handledChan)
	}).Return(assert.AnError)

	// Send an event
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-id",
			MetaData: map[string]string{
				"type": "doc",
			},
		},
	}

	select {
	case <-handledChan:
		// Passed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for background dispatcher error case")
	}

	// Let cleanup run
	time.Sleep(100 * time.Millisecond)

	hookMock.AssertExpectations(t)
}

func TestBackgroundDispatcher_Error_NoLog(t *testing.T) {
	cfg := Config{
		StorageDriver: "s3",
		S3Bucket:      "test-bucket",
		S3Endpoint:    "http://localhost:9000",
		BasePath:      "/uploads/",
	}

	registry := NewRegistry()
	hookMock := &mockUploadHook{}
	registry.Register("doc", hookMock)

	s3Client := s3.New(s3.Options{})
	h, err := NewHandler(cfg, registry, s3Client, nil)
	require.NoError(t, err)

	handledChan := make(chan struct{})

	// Add expectation to return error
	hookMock.On("HandleUpload", mock.Anything, mock.AnythingOfType("tus.UploadEvent")).Run(func(args mock.Arguments) {
		close(handledChan)
	}).Return(assert.AnError)

	// Send an event
	h.CompleteUploads <- handler.HookEvent{
		Upload: handler.FileInfo{
			ID: "test-id",
			MetaData: map[string]string{
				"type": "doc",
			},
		},
	}

	select {
	case <-handledChan:
		// Passed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for background dispatcher error case")
	}

	// Let cleanup run
	time.Sleep(100 * time.Millisecond)

	hookMock.AssertExpectations(t)
}

type mockTerminaterDataStore struct {
	handler.DataStore
	mock.Mock
}

func (m *mockTerminaterDataStore) AsTerminatableUpload(upload handler.Upload) handler.TerminatableUpload {
	args := m.Called(upload)
	return args.Get(0).(handler.TerminatableUpload)
}

func (m *mockTerminaterDataStore) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(handler.Upload), args.Error(1)
}

type mockTerminatableUpload struct {
	mock.Mock
}

func (m *mockTerminatableUpload) Terminate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestCleanupFailedCompletedUpload(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Case 1: not terminatable
	storeNotTerminatable := struct {
		handler.DataStore
	}{}
	cleanupFailedCompletedUpload(context.Background(), storeNotTerminatable, "upload1", log)
	cleanupFailedCompletedUpload(context.Background(), storeNotTerminatable, "upload1", nil)

	// Case 2: GetUpload error
	storeTerminatable := &mockTerminaterDataStore{}
	storeTerminatable.On("GetUpload", mock.Anything, "upload2").Return(handler.Upload(nil), assert.AnError)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable, "upload2", log)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable, "upload2", nil)

	// Case 3: Terminate error
	storeTerminatable2 := &mockTerminaterDataStore{}
	mockUpload := handler.Upload(nil)
	storeTerminatable2.On("GetUpload", mock.Anything, "upload3").Return(mockUpload, nil)
	mockTerminatable := &mockTerminatableUpload{}
	mockTerminatable.On("Terminate", mock.Anything).Return(assert.AnError)
	storeTerminatable2.On("AsTerminatableUpload", mockUpload).Return(mockTerminatable)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable2, "upload3", log)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable2, "upload3", nil)

	// Case 4: Success
	storeTerminatable3 := &mockTerminaterDataStore{}
	mockUpload2 := handler.Upload(nil)
	storeTerminatable3.On("GetUpload", mock.Anything, "upload4").Return(mockUpload2, nil)
	mockTerminatable2 := &mockTerminatableUpload{}
	mockTerminatable2.On("Terminate", mock.Anything).Return(nil)
	storeTerminatable3.On("AsTerminatableUpload", mockUpload2).Return(mockTerminatable2)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable3, "upload4", log)
	cleanupFailedCompletedUpload(context.Background(), storeTerminatable3, "upload4", nil)
}


