package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
)

var ErrAuthenticatedUploadUserRequired = errors.New("authenticated upload user metadata is required")

type AvatarHook struct {
	UserUseCase UserUseCase
}

func (h *AvatarHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
	userID := event.Metadata["authenticated_user_id"]
	if userID == "" {
		return ErrAuthenticatedUploadUserRequired
	}

	return h.UserUseCase.SetAvatarURL(ctx, userID, event.FileURL)
}
