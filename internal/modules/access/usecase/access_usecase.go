package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/repository"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// AccessUseCase implements the access use case.
type AccessUseCase struct {
	repo     repository.IAccessRepository
	log      *logrus.Logger
	validate *validator.Validate
}

// NewAccessUseCase creates a new AccessUseCase.
func NewAccessUseCase(repo repository.IAccessRepository, log *logrus.Logger, validate *validator.Validate) IAccessUseCase {
	return &AccessUseCase{
		repo:     repo,
		log:      log,
		validate: validate,
	}
}

// CreateAccessRight handles the business logic for creating a new access right.
func (uc *AccessUseCase) CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error) {
	if err := uc.validate.Struct(req); err != nil {
		uc.log.WithError(err).Error("Validation failed for creating access right")
		return nil, err
	}

	accessRightEntity := &entity.AccessRight{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := uc.repo.CreateAccessRight(ctx, accessRightEntity); err != nil {
		uc.log.WithError(err).Error("Failed to create access right in repository")
		return nil, err
	}

	uc.log.Infof("Successfully created access right '%s'", accessRightEntity.Name)

	return model.ConvertAccessRightToResponse(accessRightEntity), nil
}

// GetAllAccessRights retrieves all access rights.
func (uc *AccessUseCase) GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error) {
	uc.log.Info("Retrieving all access rights")
	accessRightEntities, err := uc.repo.GetAllAccessRights(ctx)
	if err != nil {
		uc.log.WithError(err).Error("Failed to get all access rights from repository")
		return nil, err
	}

	return model.ConvertAccessRightListToResponse(accessRightEntities), nil
}

// CreateEndpoint handles the business logic for creating a new endpoint.
func (uc *AccessUseCase) CreateEndpoint(ctx context.Context, req model.CreateEndpointRequest) (*model.EndpointResponse, error) {
	if err := uc.validate.Struct(req); err != nil {
		uc.log.WithError(err).Error("Validation failed for creating endpoint")
		return nil, err
	}

	endpointEntity := &entity.Endpoint{
		Path:   req.Path,
		Method: req.Method,
	}

	if err := uc.repo.CreateEndpoint(ctx, endpointEntity); err != nil {
		uc.log.WithError(err).Error("Failed to create endpoint in repository")
		return nil, err
	}

	uc.log.Infof("Successfully created endpoint: %s %s", endpointEntity.Method, endpointEntity.Path)

	return &model.EndpointResponse{
		ID:        endpointEntity.ID,
		Path:      endpointEntity.Path,
		Method:    endpointEntity.Method,
		CreatedAt: endpointEntity.CreatedAt,
	}, nil
}

// LinkEndpointToAccessRight handles the business logic for linking an endpoint to an access right.
func (uc *AccessUseCase) LinkEndpointToAccessRight(ctx context.Context, req model.LinkEndpointRequest) error {
	if err := uc.validate.Struct(req); err != nil {
		uc.log.WithError(err).Error("Validation failed for linking endpoint")
		return err
	}

	err := uc.repo.LinkEndpointToAccessRight(ctx, req.AccessRightID, req.EndpointID)
	if err != nil {
		uc.log.WithError(err).Error("Failed to link endpoint to access right in repository")
		return err
	}

	uc.log.Infof("Successfully linked endpoint %d to access right %d", req.EndpointID, req.AccessRightID)
	return nil
}
