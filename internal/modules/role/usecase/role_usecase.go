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
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type roleUseCase struct {
	Log            *logrus.Logger
	Validate       *validator.Validate
	TM             tx.WithTransactionManager
	RoleRepository repository.RoleRepository
}

func NewRoleUseCase(log *logrus.Logger, validate *validator.Validate, tm tx.WithTransactionManager, roleRepository repository.RoleRepository) RoleUseCase {
	return &roleUseCase{
		Log:            log,
		Validate:       validate,
		TM:             tm,
		RoleRepository: roleRepository,
	}
}

func (uc *roleUseCase) Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error) {
	if err := uc.Validate.Struct(request); err != nil {
		uc.Log.Warnf("Invalid request body for create role: %+v", err)
		return nil, exception.ErrBadRequest
	}

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

		newRole := &entity.Role{
			ID:          uuid.New().String(),
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