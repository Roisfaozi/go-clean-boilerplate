package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
)

// userUseCase defines the interface for user use case operations
type UserUseCase interface {
	Register(ctx context.Context, request model.RegisterUserRequest) (*model.UserResponse, error)
	Login(ctx context.Context, request model.LoginUserRequest) (*model.UserResponse, error)
	GetByID(ctx context.Context, id string) (*model.UserResponse, error)
	Update(ctx context.Context, id string, request model.UpdateUserRequest) (*model.UserResponse, error)
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context, limit, offset int) ([]*model.UserResponse, error)
	VerifyToken(ctx context.Context, token string) (*model.UserResponse, error)
}
