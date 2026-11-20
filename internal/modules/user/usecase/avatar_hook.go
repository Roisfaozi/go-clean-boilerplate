package usecase

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
)

const (
	defaultProcessedUploadTTL = 24 * time.Hour
	defaultMaxTrackedUploads  = 10000
)

var ErrAuthenticatedUploadUserRequired = errors.New("authenticated upload user metadata is required")

// ProcessedUploadTracker optionally backs upload completion deduplication across instances (e.g. Redis).
type ProcessedUploadTracker interface {
	MarkProcessedIfAbsent(ctx context.Context, uploadID string, ttl time.Duration) (bool, error)
	UnmarkProcessed(ctx context.Context, uploadID string) error
}

type avatarUploadFlight struct {
	done chan struct{}
	err  error
}

type AvatarHook struct {
	UserUseCase       UserUseCase
	ProcessedTTL      time.Duration
	MaxTrackedUploads int
	Tracker           ProcessedUploadTracker

	mu        sync.Mutex
	processed map[string]time.Time
	inFlight  map[string]*avatarUploadFlight
	nowFunc   func() time.Time
}

func (h *AvatarHook) now() time.Time {
	if h.nowFunc != nil {
		return h.nowFunc()
	}
	return time.Now()
}

func (h *AvatarHook) processedTTL() time.Duration {
	if h.ProcessedTTL > 0 {
		return h.ProcessedTTL
	}
	return defaultProcessedUploadTTL
}

func (h *AvatarHook) maxTrackedUploads() int {
	if h.MaxTrackedUploads > 0 {
		return h.MaxTrackedUploads
	}
	return defaultMaxTrackedUploads
}

func (h *AvatarHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
	userID := event.Metadata["authenticated_user_id"]
	if userID == "" {
		return ErrAuthenticatedUploadUserRequired
	}

	if event.UploadID == "" {
		return h.UserUseCase.SetAvatarURL(ctx, userID, event.FileURL)
	}

	flight, alreadyProcessed, existingFlight := h.beginUpload(event.UploadID)
	if alreadyProcessed {
		return nil
	}
	if existingFlight != nil {
		select {
		case <-existingFlight.done:
			return existingFlight.err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	ttl := h.processedTTL()
	claimedExternally := false
	if h.Tracker != nil {
		claimed, err := h.Tracker.MarkProcessedIfAbsent(ctx, event.UploadID, ttl)
		if err != nil {
			h.finishUpload(event.UploadID, flight, false, err)
			return err
		}
		if !claimed {
			h.finishUpload(event.UploadID, flight, true, nil)
			return nil
		}
		claimedExternally = true
	}

	err := h.UserUseCase.SetAvatarURL(ctx, userID, event.FileURL)
	if err != nil {
		if claimedExternally {
			_ = h.Tracker.UnmarkProcessed(ctx, event.UploadID)
		}
		h.finishUpload(event.UploadID, flight, false, err)
		return err
	}

	h.finishUpload(event.UploadID, flight, true, nil)
	return nil
}

func (h *AvatarHook) beginUpload(uploadID string) (flight *avatarUploadFlight, alreadyProcessed bool, existing *avatarUploadFlight) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.now()
	h.cleanupExpiredLocked(now)

	if exp, ok := h.processed[uploadID]; ok && exp.After(now) {
		return nil, true, nil
	}
	if inFlight, ok := h.inFlight[uploadID]; ok {
		return nil, false, inFlight
	}

	if h.inFlight == nil {
		h.inFlight = make(map[string]*avatarUploadFlight)
	}
	flight = &avatarUploadFlight{done: make(chan struct{})}
	h.inFlight[uploadID] = flight
	return flight, false, nil
}

func (h *AvatarHook) finishUpload(uploadID string, flight *avatarUploadFlight, markProcessed bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if markProcessed {
		now := h.now()
		if h.processed == nil {
			h.processed = make(map[string]time.Time)
		}
		h.cleanupExpiredLocked(now)
		h.evictIfFullLocked(uploadID)
		h.processed[uploadID] = now.Add(h.processedTTL())
	}

	flight.err = err
	delete(h.inFlight, uploadID)
	close(flight.done)
}

func (h *AvatarHook) cleanupExpiredLocked(now time.Time) {
	for id, exp := range h.processed {
		if !exp.After(now) {
			delete(h.processed, id)
		}
	}
}

func (h *AvatarHook) evictIfFullLocked(incomingID string) {
	limit := h.maxTrackedUploads()
	if _, exists := h.processed[incomingID]; exists || len(h.processed) < limit {
		return
	}

	var oldestID string
	var oldestExp time.Time
	first := true
	for id, exp := range h.processed {
		if first || exp.Before(oldestExp) {
			oldestID = id
			oldestExp = exp
			first = false
		}
	}
	if !first {
		delete(h.processed, oldestID)
	}
}
