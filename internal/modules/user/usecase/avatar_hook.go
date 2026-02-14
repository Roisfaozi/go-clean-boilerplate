package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
)

type AvatarHook struct {
	UserUseCase UserUseCase
}

func (h *AvatarHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
	userID := event.Metadata["user_id"]
	if userID == "" {
		return nil
	}

	return h.UserUseCase.SetAvatarURL(ctx, userID, event.FileURL)
}
