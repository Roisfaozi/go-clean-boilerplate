package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/repository"
	"github.com/sirupsen/logrus"
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
