package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccessUseCase struct {
	repo repository.AccessRepository
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
func NewAccessUseCase(repo repository.AccessRepository, log *logrus.Logger) IAccessUseCase {
	return &AccessUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *AccessUseCase) CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error) {
	accessRightEntity := &entity.AccessRight{
		Name:        pkg.SanitizeString(req.Name),
		Description: pkg.SanitizeString(req.Description),
	}

	if err := uc.repo.CreateAccessRight(ctx, accessRightEntity); err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to create access right in repository")
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("Successfully created access right '%s'", accessRightEntity.Name)

	return model.ConvertAccessRightToResponse(accessRightEntity), nil
}

func (uc *AccessUseCase) GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error) {
	uc.log.WithContext(ctx).Info("Retrieving all access rights")
	accessRightEntities, err := uc.repo.GetAccessRights(ctx)
	if err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to get all access rights from repository")
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
		uc.log.WithContext(ctx).WithError(err).Error("Failed to create endpoint in repository")
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("Successfully created endpoint: %s %s", endpointEntity.Method, endpointEntity.Path)

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
		uc.log.WithContext(ctx).WithError(err).Error("Failed to link endpoint to access right in repository")
		return err
	}

	uc.log.WithContext(ctx).Infof("Successfully linked endpoint %s to access right %s", req.EndpointID, req.AccessRightID)
	return nil
}

func (uc *AccessUseCase) DeleteAccessRight(ctx context.Context, id string) error {
	uc.log.WithContext(ctx).Infof("Attempting to delete access right with ID: %s", id)
	_, err := uc.repo.GetAccessRightByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Access right with ID %s not found for deletion", id)
			return exception.ErrNotFound
		}
		uc.log.WithContext(ctx).WithError(err).Errorf("Failed to find access right with ID %s: %v", id, err)
		return exception.ErrInternalServer
	}

	if err := uc.repo.DeleteAccessRight(ctx, id); err != nil {
		uc.log.WithContext(ctx).WithError(err).Errorf("Failed to delete access right with ID %s: %v", id, err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("Successfully deleted access right with ID: %s", id)
	return nil
}

func (uc *AccessUseCase) DeleteEndpoint(ctx context.Context, id string) error {
	uc.log.WithContext(ctx).Infof("Attempting to delete endpoint with ID: %s", id)

	if err := uc.repo.DeleteEndpoint(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Endpoint with ID %s not found for deletion", id)
			return exception.ErrNotFound
		}
		uc.log.WithContext(ctx).WithError(err).Errorf("Failed to delete endpoint with ID %s: %v", id, err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("Successfully deleted endpoint with ID: %s", id)
	return nil
}

func (uc *AccessUseCase) GetEndpointsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.EndpointResponse, int64, error) {
	uc.log.WithContext(ctx).Info("Retrieving endpoints dynamically")
	endpointEntities, total, err := uc.repo.FindEndpointsDynamic(ctx, filter)
	if err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to get endpoints dynamically from repository")
		return nil, 0, exception.ErrInternalServer
	}

	var responses []*model.EndpointResponse
	for _, ep := range endpointEntities {
		responses = append(responses, &model.EndpointResponse{
			ID:        ep.ID,
			Path:      ep.Path,
			Method:    ep.Method,
			CreatedAt: ep.CreatedAt,
		})
	}
	return responses, total, nil
}

func (uc *AccessUseCase) GetAccessRightsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) (*model.AccessRightListResponse, int64, error) {
	uc.log.WithContext(ctx).Info("Retrieving access rights dynamically")
	accessRightEntities, total, err := uc.repo.FindAccessRightsDynamic(ctx, filter)
	if err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to get access rights dynamically from repository")
		return nil, 0, exception.ErrInternalServer
	}
	return model.ConvertAccessRightListToResponse(accessRightEntities), total, nil
}
