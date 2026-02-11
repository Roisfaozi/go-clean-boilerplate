package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAvatarHook_HandleUpload(t *testing.T) {
	t.Run("Positive: Success with valid user_id", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/user123.png",
			Metadata: map[string]string{
				"user_id": "user123",
				"type":    "avatar",
			},
		}

		mockUC.On("SetAvatarURL", ctx, "user123", event.FileURL).Return(nil).Once()

		err := hook.HandleUpload(ctx, event)

		assert.NoError(t, err)
		mockUC.AssertExpectations(t)
	})

	t.Run("Negative: Missing user_id in metadata", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/user123.png",
			Metadata: map[string]string{
				"type": "avatar", // Missing user_id
			},
		}

		// UseCase should NOT be called
		err := hook.HandleUpload(ctx, event)

		assert.NoError(t, err) // Early exit returns nil
		mockUC.AssertNotCalled(t, "SetAvatarURL", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Negative: UseCase returns error", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/user123.png",
			Metadata: map[string]string{
				"user_id": "user123",
			},
		}

		expectedErr := errors.New("database connection failed")
		mockUC.On("SetAvatarURL", ctx, "user123", event.FileURL).Return(expectedErr).Once()

		err := hook.HandleUpload(ctx, event)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockUC.AssertExpectations(t)
	})

	t.Run("Edge: Empty user_id value", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/user123.png",
			Metadata: map[string]string{
				"user_id": "", // Empty string
			},
		}

		// UseCase should NOT be called
		err := hook.HandleUpload(ctx, event)

		assert.NoError(t, err)
		mockUC.AssertNotCalled(t, "SetAvatarURL", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Edge: Extra metadata fields", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/user123.png",
			Metadata: map[string]string{
				"user_id":  "user123",
				"extra":    "value",
				"filename": "test.png",
			},
		}

		mockUC.On("SetAvatarURL", ctx, "user123", event.FileURL).Return(nil).Once()

		err := hook.HandleUpload(ctx, event)

		assert.NoError(t, err)
		mockUC.AssertExpectations(t)
	})

	t.Run("Security: UserID with injection payload", func(t *testing.T) {
		mockUC := new(mocks.MockUserUseCase)
		hook := &usecase.AvatarHook{
			UserUseCase: mockUC,
		}
		ctx := context.Background()
		injectionPayload := "'; DROP TABLE users; --"
		event := tus.UploadEvent{
			FileURL: "https://s3.example.com/avatars/malicious.png",
			Metadata: map[string]string{
				"user_id": injectionPayload,
			},
		}

		// Hook should pass it to UC which handles security
		mockUC.On("SetAvatarURL", ctx, injectionPayload, event.FileURL).Return(nil).Once()

		err := hook.HandleUpload(ctx, event)

		assert.NoError(t, err)
		mockUC.AssertExpectations(t)
	})
}
