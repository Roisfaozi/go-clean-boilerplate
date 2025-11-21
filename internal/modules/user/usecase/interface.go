package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
)

// UserUseCase defines the interface for user use case operations
type UserUseCase interface {
	// Create creates a new user
	Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error)
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	// GetAllUsers retrieves all users
	GetAllUsers(ctx context.Context) ([]*model.UserResponse, error)
	// Current gets the current user's information
	Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error)
	// Update updates a user's information
	Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error)
	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
	// Logout handles user logout
	Logout(ctx context.Context, request *model.LogoutUserRequest) (*model.UserResponse, error)
}
