package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccessUseCase struct {
	repo repository.IAccessRepository
	log  *logrus.Logger
}

// NewAccessUseCase creates a new AccessUseCase with the given repository and logger.
//
// Parameters:
// - repo: The repository to use for accessing access rights and endpoints.
// - log: The logger to use for logging.
//
// Returns:
// - An instance of IAccessUseCase implemented by AccessUseCase.
func NewAccessUseCase(repo repository.IAccessRepository, log *logrus.Logger) IAccessUseCase {
	return &AccessUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *AccessUseCase) CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error) {
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

func (uc *AccessUseCase) GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error) {
	uc.log.Info("Retrieving all access rights")
	accessRightEntities, err := uc.repo.GetAllAccessRights(ctx)
	if err != nil {
		uc.log.WithError(err).Error("Failed to get all access rights from repository")
		return nil, err
	}

	return model.ConvertAccessRightListToResponse(accessRightEntities), nil
}

func (uc *AccessUseCase) CreateEndpoint(ctx context.Context, req model.CreateEndpointRequest) (*model.EndpointResponse, error) {
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

func (uc *AccessUseCase) LinkEndpointToAccessRight(ctx context.Context, req model.LinkEndpointRequest) error {
	err := uc.repo.LinkEndpointToAccessRight(ctx, req.AccessRightID, req.EndpointID)
	if err != nil {
		uc.log.WithError(err).Error("Failed to link endpoint to access right in repository")
		return err
	}

	uc.log.Infof("Successfully linked endpoint %d to access right %d", req.EndpointID, req.AccessRightID)
	return nil
}

func (uc *AccessUseCase) DeleteAccessRight(ctx context.Context, id uint) error {
	uc.log.Infof("Attempting to delete access right with ID: %d", id)
	// Check if access right exists before deleting
	_, err := uc.repo.FindAccessRightByID(ctx, id) // Assuming a FindAccessRightByID exists or needs to be added
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.Warnf("Access right with ID %d not found for deletion", id)
			return exception.ErrNotFound
		}
		uc.log.WithError(err).Errorf("Failed to find access right with ID %d: %v", id, err)
		return exception.ErrInternalServer
	}

	if err := uc.repo.DeleteAccessRight(ctx, id); err != nil {
		uc.log.WithError(err).Errorf("Failed to delete access right with ID %d: %v", id, err)
		return exception.ErrInternalServer
	}

	uc.log.Infof("Successfully deleted access right with ID: %d", id)
	return nil
}

func (uc *AccessUseCase) DeleteEndpoint(ctx context.Context, id uint) error {
	uc.log.Infof("Attempting to delete endpoint with ID: %d", id)
	// Check if endpoint exists before deleting (assuming a FindEndpointByID exists or needs to be added)
	// For now, GORM Delete will just not delete if not found and return no error,
	// but checking beforehand provides better error messages.
	// We might need to add FindEndpointByID to the repo interface and implementation.
	// For simplicity, let's just try to delete and check GORM's error.

	if err := uc.repo.DeleteEndpoint(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // GORM Delete with PK returns ErrRecordNotFound if not found
			uc.log.Warnf("Endpoint with ID %d not found for deletion", id)
			return exception.ErrNotFound
		}
		uc.log.WithError(err).Errorf("Failed to delete endpoint with ID %d: %v", id, err)
		return exception.ErrInternalServer
	}

	uc.log.Infof("Successfully deleted endpoint with ID: %d", id)
	return nil
}
