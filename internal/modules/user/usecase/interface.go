package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
)

// UserUseCase defines the interface for user use case operations
type UserUseCase interface {
	Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*model.UserResponse, error)
	GetAllUsers(ctx context.Context) ([]*model.UserResponse, error)
	Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error)
	Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	Logout(ctx context.Context, request *model.LogoutUserRequest) (*model.UserResponse, error)
}
