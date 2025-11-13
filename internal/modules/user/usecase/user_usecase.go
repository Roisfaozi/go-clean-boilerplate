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
	TM             tx.TransactionManager
	UserRepository repository.UserRepository
	//UserProducer   *messaging.UserProducer
}

func NewUserUseCase(logger *logrus.Logger, validate *validator.Validate, tm tx.TransactionManager,
	userRepository repository.UserRepository) UserUseCase {
	return &userUseCase{
		Log:            logger,
		Validate:       validate,
		TM:             tm,
		UserRepository: userRepository,
	}
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	var user *entity.User
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = uc.UserRepository.FindByID(txCtx, id)
		return err
	})
	return user, err
}

func (c *userUseCase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, exception.ErrBadRequest
	}

	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := c.UserRepository.FindByID(txCtx, request.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrInternalServer
		}

		if user != nil && user.ID != "" {
			c.Log.Warnf("User already exists : %+v", request.ID)
			return exception.ErrConflict
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
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrNotFound
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
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrNotFound
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

func (c *userUseCase) Logout(ctx context.Context, request *model.LogoutUserRequest) (*model.UserResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, exception.ErrBadRequest
	}

	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := c.UserRepository.FindByID(txCtx, request.ID)
		if err != nil {
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrNotFound
		}

		user.Token = ""

		if err := c.UserRepository.Update(txCtx, user); err != nil {
			c.Log.Warnf("Failed update user : %+v", err)
			return exception.ErrInternalServer
		}

		response = converter.UserToResponse(user)
		return nil
	})

	return response, err
}
