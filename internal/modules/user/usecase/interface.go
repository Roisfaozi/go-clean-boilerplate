package usecase

import (
	"context"
	"io"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type UserUseCase interface {
	Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*model.UserResponse, error)
	GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, int64, error)
	GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, int64, error)
	Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error)
	Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error)
	UpdateStatus(ctx context.Context, userID, status string) error
	UpdateAvatar(ctx context.Context, actorUserID string, userID string, file io.Reader, filename string, contentType string) (*model.UserResponse, error)
	SetAvatarURL(ctx context.Context, userID string, url string) error
	HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error
	DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error
}
