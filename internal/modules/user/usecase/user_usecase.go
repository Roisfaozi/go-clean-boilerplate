package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
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

// STANDARD email regex
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

type userUseCase struct {
	Log            *logrus.Logger
	TM             tx.WithTransactionManager
	UserRepository repository.UserRepository
	Enforcer       usecase.IEnforcer
	auditUC        auditUseCase.AuditUseCase
}

func NewUserUseCase(logger *logrus.Logger, tm tx.WithTransactionManager,
	userRepository repository.UserRepository, enforcer usecase.IEnforcer,
	auditUC auditUseCase.AuditUseCase,
) UserUseCase {
	return &userUseCase{
		Log:            logger,
		TM:             tm,
		UserRepository: userRepository,
		Enforcer:       enforcer,
		auditUC:        auditUC,
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
	// Tightened Validation
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if email == "" || !emailRegex.MatchString(email) || strings.Contains(email, "..") {
		return nil, fmt.Errorf("invalid email format")
	}
	if len(request.Password) < 8 {
		return nil, fmt.Errorf("password too weak")
	}
	if len(request.Password) > 72 {
		return nil, fmt.Errorf("password too long")
	}

	var response *model.UserResponse
	err := c.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := c.UserRepository.FindByUsername(txCtx, request.Username)
		if err == nil {
			return exception.ErrConflict
		}
		_, err = c.UserRepository.FindByEmail(txCtx, request.Email)
		if err == nil {
			return exception.ErrConflict
		}

		password, _ := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		newID, _ := uuid.NewV7()
		token, _ := uuid.NewV7()

		newUser := &entity.User{
			ID:       newID.String(),
			Username: request.Username,
			Email:    request.Email,
			Password: string(password),
			Name:     request.Name,
			Token:    token.String(),
		}

		if err := c.UserRepository.Create(txCtx, newUser); err != nil {
			return exception.ErrInternalServer
		}

		_, _ = c.Enforcer.AddGroupingPolicy(newUser.ID, "role:user")

		if c.auditUC != nil {
			_ = c.auditUC.LogActivity(context.Background(), auditModel.CreateAuditLogRequest{
				UserID:   newUser.ID,
				Action:   "REGISTER",
				Entity:   "User",
				EntityID: newUser.ID,
				NewValues: map[string]interface{}{
					"username": newUser.Username,
					"email":    newUser.Email,
				},
				IPAddress: request.IPAddress,
				UserAgent: request.UserAgent,
			})
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
			return exception.ErrNotFound
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
			return exception.ErrNotFound
		}

		oldUserMap := map[string]interface{}{"name": user.Name, "email": user.Email}

		if len(request.Password) > 0 {
			if len(request.Password) < 8 { return fmt.Errorf("password too weak") }
			if len(request.Password) > 72 { return fmt.Errorf("password too long") }
			password, _ := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
			user.Password = string(password)
		}

		if len(request.Name) > 0 { user.Name = request.Name }

		if err := c.UserRepository.Update(txCtx, user); err != nil {
			return exception.ErrInternalServer
		}

		if c.auditUC != nil {
			_ = c.auditUC.LogActivity(context.Background(), auditModel.CreateAuditLogRequest{
				UserID:    user.ID,
				Action:    "UPDATE",
				Entity:    "User",
				EntityID:  user.ID,
				OldValues: oldUserMap,
				NewValues: map[string]interface{}{"name": user.Name, "email": user.Email},
				IPAddress: request.IPAddress,
				UserAgent: request.UserAgent,
			})
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
		return err
	})
	if err != nil { return nil, err }
	var responses []*model.UserResponse
	for _, user := range users { responses = append(responses, converter.UserToResponse(user)) }
	return responses, nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error {
	return u.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		userToDelete, err := u.UserRepository.FindByID(txCtx, request.ID)
		if err != nil { return exception.ErrNotFound }

		if err := u.UserRepository.Delete(txCtx, request.ID); err != nil {
			return exception.ErrInternalServer
		}

		if u.auditUC != nil {
			_ = u.auditUC.LogActivity(context.Background(), auditModel.CreateAuditLogRequest{
				UserID:   actorUserID,
				Action:   "DELETE",
				Entity:   "User",
				EntityID: request.ID,
				OldValues: map[string]interface{}{"username": userToDelete.Username},
				IPAddress: request.IPAddress,
				UserAgent: request.UserAgent,
			})
		}
		return nil
	})
}

func (u *userUseCase) GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, error) {
	var users []*entity.User
	err := u.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		users, err = u.UserRepository.FindAllDynamic(txCtx, filter)
		return err
	})
	if err != nil { return nil, err }
	var responses []*model.UserResponse
	for _, user := range users { responses = append(responses, converter.UserToResponse(user)) }
	return responses, nil
}
