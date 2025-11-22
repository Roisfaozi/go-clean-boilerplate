package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model/converter"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userUseCase struct {
	Log            *logrus.Logger
	Validate       *validator.Validate
	TM             tx.WithTransactionManager
	UserRepository repository.UserRepository
}

// NewUserUseCase creates a new instance of UserUseCase.
//
// Parameters:
// - logger: The logger instance.
// - validate: The validator instance.
// - tm: The transaction manager instance.
// - userRepository: The user repository instance.
//
// Returns:
// - A pointer to the newly created UserUseCase instance.
func NewUserUseCase(logger *logrus.Logger, validate *validator.Validate, tm tx.WithTransactionManager,
	userRepository repository.UserRepository) UserUseCase {
	return &userUseCase{
		Log:            logger,
		Validate:       validate,
		TM:             tm,
		UserRepository: userRepository,
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
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, exception.ErrBadRequest
	}

	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := c.UserRepository.FindByID(txCtx, request.ID)
		if err == nil {
			c.Log.Warnf("User already exists : %+v", request.ID)
			return exception.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrInternalServer
		}

		password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Log.Warnf("Failed to encrypt password : %+v", err)
			return exception.ErrInternalServer
		}

		newUser := &entity.User{
			ID:       request.ID,
			Password: string(password),
			Name:     request.Name,
			Token:    uuid.NewString(),
		}

		if err := c.UserRepository.Create(txCtx, newUser); err != nil {
			c.Log.Warnf("Failed to insert user : %+v", err)
			return exception.ErrInternalServer
		}

		response = converter.UserToResponse(newUser)
		return nil
	})

	return response, err
}

func (c *userUseCase) Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, exception.ErrBadRequest
	}

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
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, exception.ErrBadRequest
	}

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
