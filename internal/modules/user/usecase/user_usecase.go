package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type userUseCaseImpl struct {
	DB       tx.WithTransactionManager
	Log      *logrus.Logger
	Repo     repository.UserRepository
	Enforcer permissionUseCase.IEnforcer
	AuditUC  auditUseCase.AuditUseCase
	AuthUC   authUseCase.AuthUseCase
	Storage  storage.Provider
}

func NewUserUseCase(
	db tx.WithTransactionManager,
	log *logrus.Logger,
	repo repository.UserRepository,
	enforcer permissionUseCase.IEnforcer,
	auditUC auditUseCase.AuditUseCase,
	authUC authUseCase.AuthUseCase,
	storage storage.Provider,
) UserUseCase {
	return &userUseCaseImpl{
		DB:       db,
		Log:      log,
		Repo:     repo,
		Enforcer: enforcer,
		AuditUC:  auditUC,
		AuthUC:   authUC,
		Storage:  storage,
	}
}

func (u *userUseCaseImpl) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	request.Name = pkg.SanitizeString(request.Name)
	request.Username = pkg.SanitizeString(request.Username)

	existingUser, err := u.Repo.FindByUsername(ctx, request.Username)
	if err == nil && existingUser != nil {
		u.Log.Warnf("Username already exists: %s", request.Username)
		return nil, exception.ErrConflict
	}

	existingEmail, err := u.Repo.FindByEmail(ctx, request.Email)
	if err == nil && existingEmail != nil {
		u.Log.Warnf("Email already exists: %s", request.Email)
		return nil, exception.ErrConflict
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		u.Log.Errorf("Failed to hash password: %v", err)
		return nil, exception.ErrInternalServer
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:       userID.String(),
		Username: request.Username,
		Email:    request.Email,
		Password: string(hashedPassword),
		Name:     request.Name,
		Status:   entity.UserStatusActive,
	}

	err = u.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.Repo.Create(txCtx, user); err != nil {
			u.Log.Errorf("Failed to create user: %v", err)
			return exception.ErrInternalServer
		}

		roleAdded := false
		if u.Enforcer != nil {
			_, err := u.Enforcer.AddGroupingPolicy(user.ID, "role:user", "global")
			if err != nil {
				u.Log.Errorf("Failed to assign default role: %v", err)
				return exception.ErrInternalServer
			}
			roleAdded = true
		}

		if u.AuditUC != nil {

			err := u.AuditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
				UserID:    user.ID,
				Action:    "CREATE",
				Entity:    "User",
				EntityID:  user.ID,
				NewValues: map[string]interface{}{"username": user.Username, "email": user.Email, "Name": user.Name},
			})
			if err != nil {
				u.Log.Errorf("Failed to create audit log (rollback triggered): %v", err)

				if roleAdded && u.Enforcer != nil {
					if _, errComp := u.Enforcer.RemoveFilteredGroupingPolicy(0, user.ID, "", "global"); errComp != nil {
						u.Log.Errorf("Failed to rollback Casbin policy: %v", errComp)
					}
				}
				return exception.ErrInternalServer
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	telemetry.UserRegistrationsTotal.Inc()

	return converter.UserToResponse(user), nil
}

func (u *userUseCaseImpl) GetUserByID(ctx context.Context, id string) (*model.UserResponse, error) {
	if pkg.ContainsSQLInjection(id) {
		u.Log.Warnf("Potential SQL Injection detected in ID: %s", id)
		return nil, exception.ErrBadRequest
	}

	user, err := u.Repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}

	return converter.UserToResponse(user), nil
}

func (u *userUseCaseImpl) GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, int64, error) {
	users, total, err := u.Repo.FindAll(ctx, request)
	if err != nil {
		u.Log.Errorf("Failed to get all users: %v", err)
		return nil, 0, exception.ErrInternalServer
	}

	var userResponses []*model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, converter.UserToResponse(user))
	}

	return userResponses, total, nil
}

func (u *userUseCaseImpl) GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, int64, error) {
	users, total, err := u.Repo.FindAllDynamic(ctx, filter)
	if err != nil {
		u.Log.Errorf("Failed to find users dynamically: %v", err)
		return nil, 0, exception.ErrInternalServer
	}

	var userResponses []*model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, converter.UserToResponse(user))
	}

	return userResponses, total, nil
}

func (u *userUseCaseImpl) Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	user, err := u.Repo.FindByID(ctx, request.ID)
	if err != nil {
		return nil, exception.ErrNotFound
	}

	return converter.UserToResponse(user), nil
}

func (u *userUseCaseImpl) Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error) {
	request.Name = pkg.SanitizeString(request.Name)

	user, err := u.Repo.FindByID(ctx, request.ID)
	if err != nil {
		return nil, exception.ErrNotFound
	}

	if request.Username != "" {
		if request.Username != user.Username {
			if existing, _ := u.Repo.FindByUsername(ctx, request.Username); existing != nil {
				return nil, exception.ErrConflict
			}
		}
		user.Username = request.Username
	}

	if request.Name != "" {
		user.Name = request.Name
	}

	if request.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			u.Log.Errorf("Failed to hash password: %v", err)
			return nil, exception.ErrInternalServer
		}
		user.Password = string(hashedPassword)
	}

	err = u.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.Repo.Update(txCtx, user); err != nil {
			u.Log.Errorf("Failed to update user: %v", err)
			return exception.ErrInternalServer
		}

		if u.AuditUC != nil {
			newVals := make(map[string]interface{})
			if request.Name != "" {
				newVals["name"] = request.Name
			}
			if request.Username != "" {
				newVals["username"] = request.Username
			}

			if err := u.AuditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
				UserID:    user.ID,
				Action:    "UPDATE",
				Entity:    "User",
				EntityID:  user.ID,
				NewValues: newVals,
			}); err != nil {
				u.Log.Errorf("Failed to log activity for update: %v", err)
				return exception.ErrInternalServer
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return converter.UserToResponse(user), nil
}

func (u *userUseCaseImpl) UpdateStatus(ctx context.Context, userID, status string) error {
	if status != entity.UserStatusActive && status != entity.UserStatusSuspended && status != entity.UserStatusBanned {
		return exception.ErrValidationError
	}

	_, err := u.Repo.FindByID(ctx, userID)
	if err != nil {
		return exception.ErrNotFound
	}

	return u.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.Repo.UpdateStatus(txCtx, userID, status); err != nil {
			u.Log.Errorf("Failed to update user status: %v", err)
			return exception.ErrInternalServer
		}

		if status == entity.UserStatusBanned || status == entity.UserStatusSuspended {
			if u.AuthUC != nil {
				if err := u.AuthUC.RevokeAllSessions(txCtx, userID); err != nil {
					u.Log.Errorf("Failed to revoke sessions for user %s: %v", userID, err)
					return exception.ErrInternalServer
				}
			}
		}

		if u.AuditUC != nil {
			if err := u.AuditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
				UserID:    userID,
				Action:    "UPDATE_STATUS",
				Entity:    "User",
				EntityID:  userID,
				NewValues: map[string]interface{}{"status": status},
			}); err != nil {
				u.Log.Errorf("Failed to log activity for status update: %v", err)
				return exception.ErrInternalServer
			}
		}
		return nil
	})
}

func (u *userUseCaseImpl) UpdateAvatar(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (*model.UserResponse, error) {
	// 1. Get User
	user, err := u.Repo.FindByID(ctx, userID)
	if err != nil {
		return nil, exception.ErrNotFound
	}

	// Security: Validate File Type using Magic Bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		u.Log.Warnf("Failed to read file for type detection: %v", err)
		return nil, exception.ErrBadRequest
	}
	buf = buf[:n]

	detectedType := http.DetectContentType(buf)

	// Whitelist allowed types and enforce extension
	allowedTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}

	ext, allowed := allowedTypes[detectedType]
	if !allowed {
		u.Log.Warnf("Blocked upload of type: %s", detectedType)
		return nil, exception.ErrValidationError
	}

	// Reconstruct the reader
	file = io.MultiReader(bytes.NewReader(buf), file)

	// 2. Generate unique filename to avoid collisions (force safe extension)
	newFilename := fmt.Sprintf("avatars/%s%s", userID, ext)

	// 3. Upload to Storage (use detected type)
	url, err := u.Storage.UploadFile(ctx, file, newFilename, detectedType)
	if err != nil {
		u.Log.Errorf("Failed to upload avatar: %v", err)
		return nil, exception.ErrInternalServer
	}

	// 4. Update Database
	user.AvatarURL = url
	if err := u.Repo.Update(ctx, user); err != nil {
		u.Log.Errorf("Failed to update user avatar URL: %v", err)
		if err := u.Storage.DeleteFile(ctx, newFilename); err != nil {
			u.Log.Errorf("Failed to cleanup avatar file %s after db error: %v", newFilename, err)
		}
		return nil, exception.ErrInternalServer
	}

	// 5. Audit Log
	if u.AuditUC != nil {
		_ = u.AuditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "UPDATE_AVATAR",
			Entity:   "User",
			EntityID: userID,
			NewValues: map[string]string{
				"avatar_url": url,
			},
		})
	}

	return converter.UserToResponse(user), nil
}

func (u *userUseCaseImpl) DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error {
	if pkg.ContainsSQLInjection(request.ID) {
		u.Log.Warnf("Potential SQL Injection in Delete User ID: %s", request.ID)
		return exception.ErrBadRequest
	}

	user, err := u.Repo.FindByID(ctx, request.ID)
	if err != nil {
		if err.Error() == "user not found" {
			return exception.ErrNotFound
		}
		return err
	}

	return u.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.Repo.Delete(txCtx, user.ID); err != nil {
			u.Log.Errorf("Failed to delete user: %v", err)
			return err
		}

		var oldRoles []string
		if u.Enforcer != nil {

			var err error
			oldRoles, err = u.Enforcer.GetRolesForUser(user.ID, "global")
			if err != nil {
				u.Log.Errorf("Failed to fetch roles for backup in delete: %v", err)
				return exception.ErrInternalServer
			}

			_, err = u.Enforcer.RemoveFilteredGroupingPolicy(0, user.ID, "", "global")
			if err != nil {
				u.Log.Errorf("Failed to remove user policies: %v", err)
				return exception.ErrInternalServer
			}
		}

		if u.AuditUC != nil {
			err := u.AuditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
				UserID:   actorUserID,
				Action:   "DELETE",
				Entity:   "User",
				EntityID: user.ID,
				OldValues: map[string]interface{}{
					"username": user.Username,
					"email":    user.Email,
				},
			})
			if err != nil {
				u.Log.Errorf("Failed to log audit for delete (rollback triggered): %v", err)

				if u.Enforcer != nil && len(oldRoles) > 0 {
					for _, role := range oldRoles {
						if _, errComp := u.Enforcer.AddGroupingPolicy(user.ID, role, "global"); errComp != nil {
							u.Log.Errorf("Failed to restore role %s during rollback: %v", role, errComp)
						}
					}
				}

				return exception.ErrInternalServer
			}
		}
		return nil
	})
}

func (u *userUseCaseImpl) HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error {
	if err := u.Repo.HardDeleteSoftDeletedUsers(ctx, retentionDays); err != nil {
		u.Log.Errorf("Failed to hard delete users: %v", err)
		return exception.ErrInternalServer
	}
	return nil
}
