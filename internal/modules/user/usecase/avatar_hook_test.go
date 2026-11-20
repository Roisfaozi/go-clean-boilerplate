package usecase

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAvatarUserUseCase struct {
	UserUseCase
	setAvatarFn func(ctx context.Context, userID, url string) error
	calls       atomic.Int32
}

func (s *stubAvatarUserUseCase) SetAvatarURL(ctx context.Context, userID, url string) error {
	s.calls.Add(1)
	if s.setAvatarFn != nil {
		return s.setAvatarFn(ctx, userID, url)
	}
	return nil
}

type stubUploadTracker struct {
	mu      sync.Mutex
	seen    map[string]bool
	unmarks []string
	markErr error
}

func (s *stubUploadTracker) MarkProcessedIfAbsent(_ context.Context, uploadID string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.markErr != nil {
		return false, s.markErr
	}
	if s.seen == nil {
		s.seen = make(map[string]bool)
	}
	if s.seen[uploadID] {
		return false, nil
	}
	s.seen[uploadID] = true
	return true, nil
}

func (s *stubUploadTracker) UnmarkProcessed(_ context.Context, uploadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.seen, uploadID)
	s.unmarks = append(s.unmarks, uploadID)
	return nil
}

func TestAvatarHook_HandleUpload_DuplicateUploadIDSkipped(t *testing.T) {
	uc := &stubAvatarUserUseCase{}
	hook := &AvatarHook{UserUseCase: uc}

	event := tus.UploadEvent{
		UploadID: "upload-123",
		FileURL:  "https://cdn.example.com/avatars/upload-123.png",
		Metadata: map[string]string{
			"authenticated_user_id": "user-1",
		},
	}

	require.NoError(t, hook.HandleUpload(context.Background(), event))
	require.NoError(t, hook.HandleUpload(context.Background(), event))

	assert.Equal(t, int32(1), uc.calls.Load(), "replaying the same uploadID must not invoke SetAvatarURL twice")
}

func TestAvatarHook_HandleUpload_DifferentUploadIDsAllowed(t *testing.T) {
	uc := &stubAvatarUserUseCase{}
	hook := &AvatarHook{UserUseCase: uc}

	first := tus.UploadEvent{
		UploadID: "upload-1",
		FileURL:  "https://cdn.example.com/avatars/upload-1.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}
	second := tus.UploadEvent{
		UploadID: "upload-2",
		FileURL:  "https://cdn.example.com/avatars/upload-2.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	require.NoError(t, hook.HandleUpload(context.Background(), first))
	require.NoError(t, hook.HandleUpload(context.Background(), second))

	assert.Equal(t, int32(2), uc.calls.Load())
}

func TestAvatarHook_HandleUpload_FailureAllowsRetry(t *testing.T) {
	failOnce := true
	uc := &stubAvatarUserUseCase{
		setAvatarFn: func(_ context.Context, _, _ string) error {
			if failOnce {
				failOnce = false
				return errors.New("temporary db failure")
			}
			return nil
		},
	}
	hook := &AvatarHook{UserUseCase: uc}

	event := tus.UploadEvent{
		UploadID: "upload-retry",
		FileURL:  "https://cdn.example.com/avatars/upload-retry.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	err := hook.HandleUpload(context.Background(), event)
	require.Error(t, err)
	assert.Equal(t, int32(1), uc.calls.Load())

	// Retry after failure must execute SetAvatarURL again and succeed
	require.NoError(t, hook.HandleUpload(context.Background(), event))
	assert.Equal(t, int32(2), uc.calls.Load())

	// Replay after success must be deduplicated
	require.NoError(t, hook.HandleUpload(context.Background(), event))
	assert.Equal(t, int32(2), uc.calls.Load())
}

func TestAvatarHook_HandleUpload_ConcurrentDuplicateUploadID(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	uc := &stubAvatarUserUseCase{
		setAvatarFn: func(_ context.Context, _, _ string) error {
			close(started)
			<-release
			return nil
		},
	}
	hook := &AvatarHook{UserUseCase: uc}

	event := tus.UploadEvent{
		UploadID: "upload-concurrent",
		FileURL:  "https://cdn.example.com/avatars/upload-concurrent.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	const goroutines = 5
	errs := make(chan error, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	go func() {
		defer wg.Done()
		errs <- hook.HandleUpload(context.Background(), event)
	}()

	<-started

	for i := 1; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			errs <- hook.HandleUpload(context.Background(), event)
		}()
	}

	time.Sleep(20 * time.Millisecond)
	close(release)

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), uc.calls.Load(), "concurrent duplicate upload callbacks must invoke SetAvatarURL once")
}

func TestAvatarHook_HandleUpload_TTLExpiryAndBoundedCapacity(t *testing.T) {
	currentTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	uc := &stubAvatarUserUseCase{}
	hook := &AvatarHook{
		UserUseCase:       uc,
		ProcessedTTL:      10 * time.Minute,
		MaxTrackedUploads: 2,
		nowFunc:           func() time.Time { return currentTime },
	}

	ev := func(id string) tus.UploadEvent {
		return tus.UploadEvent{
			UploadID: id,
			FileURL:  "https://cdn.example.com/avatars/" + id + ".png",
			Metadata: map[string]string{"authenticated_user_id": "user-1"},
		}
	}

	require.NoError(t, hook.HandleUpload(context.Background(), ev("u1")))
	currentTime = currentTime.Add(1 * time.Minute)
	require.NoError(t, hook.HandleUpload(context.Background(), ev("u2")))
	currentTime = currentTime.Add(1 * time.Minute)
	require.NoError(t, hook.HandleUpload(context.Background(), ev("u3")))

	// Capacity is bounded to 2, so oldest entry ("u1") was evicted while "u2" and "u3" remain.
	hook.mu.Lock()
	assert.Len(t, hook.processed, 2)
	_, hasU1 := hook.processed["u1"]
	_, hasU2 := hook.processed["u2"]
	_, hasU3 := hook.processed["u3"]
	hook.mu.Unlock()
	assert.False(t, hasU1)
	assert.True(t, hasU2)
	assert.True(t, hasU3)

	// Replaying u2 before TTL does not re-run SetAvatarURL.
	require.NoError(t, hook.HandleUpload(context.Background(), ev("u2")))
	assert.Equal(t, int32(3), uc.calls.Load())

	// Advancing past TTL expires u2 and u3.
	currentTime = currentTime.Add(15 * time.Minute)
	require.NoError(t, hook.HandleUpload(context.Background(), ev("u2")))
	assert.Equal(t, int32(4), uc.calls.Load())
}

func TestAvatarHook_HandleUpload_ExternalTracker(t *testing.T) {
	tracker := &stubUploadTracker{}
	uc1 := &stubAvatarUserUseCase{}
	uc2 := &stubAvatarUserUseCase{}

	hookA := &AvatarHook{UserUseCase: uc1, Tracker: tracker}
	hookB := &AvatarHook{UserUseCase: uc2, Tracker: tracker}

	event := tus.UploadEvent{
		UploadID: "upload-distributed",
		FileURL:  "https://cdn.example.com/avatars/upload-distributed.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	require.NoError(t, hookA.HandleUpload(context.Background(), event))
	require.NoError(t, hookB.HandleUpload(context.Background(), event))

	assert.Equal(t, int32(1), uc1.calls.Load())
	assert.Equal(t, int32(0), uc2.calls.Load(), "second instance sharing tracker must skip duplicate uploadID")
}

func TestAvatarHook_HandleUpload_ExternalTrackerUnmarksOnError(t *testing.T) {
	tracker := &stubUploadTracker{}
	uc := &stubAvatarUserUseCase{
		setAvatarFn: func(_ context.Context, _, _ string) error {
			return errors.New("write failed")
		},
	}
	hook := &AvatarHook{UserUseCase: uc, Tracker: tracker}

	event := tus.UploadEvent{
		UploadID: "upload-fail",
		FileURL:  "https://cdn.example.com/avatars/upload-fail.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	err := hook.HandleUpload(context.Background(), event)
	require.Error(t, err)
	assert.Equal(t, []string{"upload-fail"}, tracker.unmarks)
}

func TestAvatarHook_HandleUpload_EmptyUploadIDAndMissingUser(t *testing.T) {
	uc := &stubAvatarUserUseCase{}
	hook := &AvatarHook{UserUseCase: uc}

	missingUser := tus.UploadEvent{
		UploadID: "upload-no-user",
		FileURL:  "https://cdn.example.com/avatars/x.png",
		Metadata: map[string]string{},
	}
	assert.ErrorIs(t, hook.HandleUpload(context.Background(), missingUser), ErrAuthenticatedUploadUserRequired)
	assert.Equal(t, int32(0), uc.calls.Load())

	emptyUploadID := tus.UploadEvent{
		FileURL:  "https://cdn.example.com/avatars/y.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}
	require.NoError(t, hook.HandleUpload(context.Background(), emptyUploadID))
	require.NoError(t, hook.HandleUpload(context.Background(), emptyUploadID))
	assert.Equal(t, int32(2), uc.calls.Load(), "empty uploadID should not deduplicate unrelated events")
}
