package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userUseCase struct {
	Log            *logrus.Logger
	TM             tx.WithTransactionManager
	UserRepository repository.UserRepository
	Enforcer       usecase.IEnforcer
}

// NewUserUseCase creates a new instance of UserUseCase.
//
// Parameters:
// - logger: The logger instance.
// - tm: The transaction manager instance.
// - userRepository: The user repository instance.
// - enforcer: The Casbin enforcer instance.
//
// Returns:
// - A pointer to the newly created UserUseCase instance.
func NewUserUseCase(logger *logrus.Logger, tm tx.WithTransactionManager,
	userRepository repository.UserRepository, enforcer usecase.IEnforcer) UserUseCase {
	return &userUseCase{
		Log:            logger,
		TM:             tm,
		UserRepository: userRepository,
		Enforcer:       enforcer,
	}
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id string) (*model.UserResponse, error) {
	var user *entity.User
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = uc.UserRepository.FindByID(txCtx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				uc.Log.Warnf("User with id %s not found", id)
				return exception.ErrNotFound
			}
			uc.Log.Errorf("Failed to find user by id %s: %v", id, err)
			return exception.ErrInternalServer
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return converter.UserToResponse(user), nil
}

func (c *userUseCase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := c.UserRepository.FindByUsername(txCtx, request.Username)
		if err == nil {
			c.Log.Warnf("User already exists : %+v", request.Username)
			return exception.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Errorf("Failed find user by username : %+v", err)
			return exception.ErrInternalServer
		}

		_, err = c.UserRepository.FindByEmail(txCtx, request.Email)
		if err == nil {
			c.Log.Warnf("User already exists : %+v", request.Email)
			return exception.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Errorf("Failed find user by email : %+v", err)
			return exception.ErrInternalServer
		}

		password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Log.Errorf("Failed to encrypt password : %+v", err)
			return exception.ErrInternalServer
		}

		newID, err := uuid.NewV7()
		if err != nil {
			c.Log.Errorf("Failed to generate UUID: %v", err)
			return exception.ErrInternalServer
		}

		token, err := uuid.NewV7()
		if err != nil {
			c.Log.Errorf("Failed to generate UUID: %v", err)
			return exception.ErrInternalServer
		}

		newUser := &entity.User{
			ID:       newID.String(),
			Username: request.Username,
			Email:    request.Email,
			Password: string(password),
			Name:     request.Name,
			Token:    token.String(),
		}

		if err := c.UserRepository.Create(txCtx, newUser); err != nil {
			c.Log.Errorf("Failed to insert user : %+v", err)
			if strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry") {
				return exception.ErrConflict
			}
			return exception.ErrInternalServer
		}

		// Assign default role "role:user"
		if _, err := c.Enforcer.AddGroupingPolicy(newUser.ID, "role:user"); err != nil {
			c.Log.Errorf("Failed to assign default role to user : %+v", err)
			return exception.ErrInternalServer
		}

		response = converter.UserToResponse(newUser)
		return nil
	})

	return response, err
}

func (c *userUseCase) Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := c.UserRepository.FindByID(txCtx, request.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Log.Warnf("User with id %s not found", request.ID)
				return exception.ErrNotFound
			}
			c.Log.Errorf("Failed to find user by id %s: %v", request.ID, err)
			return exception.ErrInternalServer
		}
		response = converter.UserToResponse(user)
		return nil
	})

	return response, err
}

func (c *userUseCase) Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error) {
	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := c.UserRepository.FindByID(txCtx, request.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Log.Warnf("User with id %s not found for update", request.ID)
				return exception.ErrNotFound
			}
			c.Log.Errorf("Failed to find user by id %s for update: %v", request.ID, err)
			return exception.ErrInternalServer
		}

		if len(request.Password) > 0 {
			password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
			if err != nil {
				c.Log.Warnf("Failed to encrypt password : %+v", err)
				return exception.ErrInternalServer
			}
			user.Password = string(password)
		}

		if len(request.Name) > 0 {
			user.Name = request.Name
		}

		if err := c.UserRepository.Update(txCtx, user); err != nil {
			c.Log.Warnf("Failed update user : %+v", err)
			return exception.ErrInternalServer
		}

		response = converter.UserToResponse(user)
		return nil
	})

	return response, err
}

func (u *userUseCase) GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, error) {
	var users []*entity.User
	err := u.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		users, err = u.UserRepository.FindAll(txCtx, request)
		if err != nil {
			u.Log.Errorf("Failed to find all users: %v", err)
			return exception.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var responses []*model.UserResponse
	for _, user := range users {
		responses = append(responses, converter.UserToResponse(user))
	}

	return responses, nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, id string) error {
	return u.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := u.UserRepository.FindByID(txCtx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				u.Log.Warnf("User with id %s not found for deletion", id)
				return exception.ErrNotFound
			}
			u.Log.Errorf("Failed to find user by id %s for deletion: %v", id, err)
			return exception.ErrInternalServer
		}

		if err := u.UserRepository.Delete(txCtx, id); err != nil {
			u.Log.Errorf("Failed to delete user with id %s: %v", id, err)
			return exception.ErrInternalServer
		}
		return nil
	})
}

func (u *userUseCase) GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, error) {
	var users []*entity.User
	err := u.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		users, err = u.UserRepository.FindAllDynamic(txCtx, filter)
		if err != nil {
			u.Log.Errorf("Failed to find users dynamically: %v", err)
			return exception.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var responses []*model.UserResponse
	for _, user := range users {
		responses = append(responses, converter.UserToResponse(user))
	}

	return responses, nil
}
