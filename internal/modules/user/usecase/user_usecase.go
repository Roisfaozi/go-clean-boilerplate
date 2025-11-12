package usecase

import (
	"context"
	"errors"

	auth "github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model/converter"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils"
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
	JWTService usecase.AuthUseCase
}

func NewUserUseCase(logger *logrus.Logger, validate *validator.Validate, tm tx.TransactionManager,
	userRepository repository.UserRepository, userProducer *messaging.UserProducer, jwtService usecase.AuthUseCase) *userUseCase {
	return &userUseCase{
		Log:            logger,
		Validate:       validate,
		TM:             tm,
		UserRepository: userRepository,
		//UserProducer:   userProducer,
		JWTService: jwtService,
	}
}

func (uc *userUseCase) Register(ctx context.Context, request model.RegisterUserRequest) (*model.UserResponse, error) {
	if err := uc.Validate.Struct(request); err != nil {
		return nil, err
	}

	var userResponse *model.UserResponse
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		hashedPassword, err := utils.HashPassword(request.Password)
		if err != nil {
			return err
		}

		user := &entity.User{
			ID:       request.ID,
			Name:     request.Name,
			Password: hashedPassword,
		}

		if err := uc.UserRepository.Create(txCtx, user); err != nil {
			return err
		}

		userResponse = converter.UserToResponse(user)
		return nil
	})

	return userResponse, err
}

func (uc *userUseCase) Login(ctx context.Context, request auth.LoginRequest) (*auth.LoginResponse, string, error) {
	if err := uc.Validate.Struct(request); err != nil {
		return nil, "", err
	}

	var user *entity.User
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = uc.UserRepository.FindByUsername(txCtx, request.Username)
		if err != nil {
			return errors.New("invalid credentials")
		}

		if !utils.CheckPasswordHash(request.Password, user.Password) {
			return errors.New("invalid credentials")
		}

		sessionID := uuid.NewString()
		user.Token = sessionID
		if err := uc.UserRepository.Update(txCtx, user); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	accessToken, err := uc.JWTService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := uc.JWTService.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", err
	}

	loginResponse := &auth.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	return loginResponse, refreshToken, nil
}

func (uc *userUseCase) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, string, error) {
	claims, err := uc.JWTService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, "", err
	}

	user, err := uc.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, "", err
	}

	newAccessToken, err := uc.JWTService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	newRefreshToken, err := uc.JWTService.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", err
	}

	tokenResponse := &auth.TokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
	}

	return tokenResponse, newRefreshToken, nil
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

func (c *userUseCase) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	var auth *model.Auth
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := c.UserRepository.FindByID(txCtx, userID)
		if err != nil {
			c.Log.Warnf("Failed find user by id : %+v", err)
			return exception.ErrNotFound
		}

		if user.Token == "" || user.Token != sessionID {
			c.Log.Warnf("User token mismatch or empty for user : %s", userID)
			return exception.ErrUnauthorized // Token has been invalidated (logged out) or does not match
		}

		auth = &model.Auth{ID: user.ID}
		return nil
	})

	return auth, err
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
