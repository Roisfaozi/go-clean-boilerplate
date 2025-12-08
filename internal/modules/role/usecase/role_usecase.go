package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model/converter"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type roleUseCase struct {
	Log            *logrus.Logger
	TM             tx.WithTransactionManager
	RoleRepository repository.RoleRepository
}

func NewRoleUseCase(log *logrus.Logger, tm tx.WithTransactionManager, roleRepository repository.RoleRepository) RoleUseCase {
	return &roleUseCase{
		Log:            log,
		TM:             tm,
		RoleRepository: roleRepository,
	}
}

func (uc *roleUseCase) Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error) {
	var response *model.RoleResponse
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := uc.RoleRepository.FindByName(txCtx, request.Name)
		if err == nil {
			uc.Log.Warnf("Role with name %s already exists", request.Name)
			return exception.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			uc.Log.Errorf("Failed to find role by name: %v", err)
			return exception.ErrInternalServer
		}

		newID, err := uuid.NewV7()
		if err != nil {
			uc.Log.Errorf("Failed to generate UUID: %v", err)
			return exception.ErrInternalServer
		}

		newRole := &entity.Role{
			ID:          newID.String(),
			Name:        request.Name,
			Description: request.Description,
		}

		if err := uc.RoleRepository.Create(txCtx, newRole); err != nil {
			uc.Log.Errorf("Failed to create role: %v", err)
			return exception.ErrInternalServer
		}

		response = converter.RoleToResponse(newRole)
		return nil
	})

	return response, err
}

func (uc *roleUseCase) GetAll(ctx context.Context) ([]model.RoleResponse, error) {
	var roles []*entity.Role
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		roles, err = uc.RoleRepository.FindAll(txCtx)
		if err != nil {
			uc.Log.Errorf("Failed to get all roles: %v", err)
			return exception.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return converter.RolesToResponse(roles), nil
}

func (uc *roleUseCase) Delete(ctx context.Context, id string) error {
	return uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		role, err := uc.RoleRepository.FindByID(txCtx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				uc.Log.Warnf("Role with id %s not found for deletion", id)
				return exception.ErrNotFound
			}
			uc.Log.Errorf("Failed to find role by id: %v", err)
			return exception.ErrInternalServer
		}

		// Prevent deleting superadmin role
		if role.Name == "role:superadmin" {
			uc.Log.Warn("Attempt to delete superadmin role blocked")
			return exception.ErrForbidden
		}

		if err := uc.RoleRepository.Delete(txCtx, id); err != nil {
			uc.Log.Errorf("Failed to delete role: %v", err)
			return exception.ErrInternalServer
		}

		return nil
	})
}