package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
)

type UserUseCase interface {
	Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*model.UserResponse, error)
	// GetAllUsers retrieves all users
	GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, error)
	Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error)
	Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
}
