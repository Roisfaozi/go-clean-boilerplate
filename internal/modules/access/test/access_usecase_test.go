package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// nullWriter is used to discard log output during tests.
type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestCreateAccessRight(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})

	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Create Valid Access Right", func(t *testing.T) {
		mockRepo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).Return(nil).Once()
		req := model.CreateAccessRightRequest{
			Name:        "view_dashboard",
			Description: "Allows viewing the main dashboard",
		}
		createdAccessRight, err := uc.CreateAccessRight(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdAccessRight)
		assert.Equal(t, req.Name, createdAccessRight.Name)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllAccessRights(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Has Data", func(t *testing.T) {
		expectedEntities := []*entity.AccessRight{
			{ID: "1", Name: "view_dashboard"},
			{ID: "2", Name: "edit_settings"},
		}
		mockRepo.On("GetAccessRights", ctx).Return(expectedEntities, nil).Once()
		results, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Data, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - No Data", func(t *testing.T) {
		mockRepo.On("GetAccessRights", ctx).Return([]*entity.AccessRight{}, nil).Once()
		results, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Data, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestCreateEndpoint(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Create Valid Endpoint", func(t *testing.T) {
		mockRepo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).Return(nil).Once()
		req := model.CreateEndpointRequest{Path: "/api/v1/test", Method: "GET"}
		createdEndpoint, err := uc.CreateEndpoint(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdEndpoint)
		assert.Equal(t, req.Path, createdEndpoint.Path)
		mockRepo.AssertExpectations(t)
	})
}

func TestLinkEndpointToAccessRight(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Link Valid IDs", func(t *testing.T) {
		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		mockRepo.On("LinkEndpointToAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(nil).Once()
		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteAccessRight(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()
	id := "1"

	t.Run("Success - Delete Access Right", func(t *testing.T) {
		mockRepo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
		mockRepo.On("DeleteAccessRight", ctx, id).Return(nil).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		mockRepo.On("GetAccessRightByID", ctx, id).Return(nil, gorm.ErrRecordNotFound).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteEndpoint(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()
	id := "1" // Changed to string

	t.Run("Success - Delete Endpoint", func(t *testing.T) {
		mockRepo.On("DeleteEndpoint", ctx, id).Return(nil).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Not Found (GORM delete behavior)", func(t *testing.T) {
		mockRepo.On("DeleteEndpoint", ctx, id).Return(gorm.ErrRecordNotFound).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetEndpointsDynamic(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Get Endpoints Dynamically", func(t *testing.T) {
		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Method": {Type: "equals", From: "GET"},
			},
		}
		expectedEndpoints := []*entity.Endpoint{
			{ID: "1", Path: "/api/test", Method: "GET"},
		}
		mockRepo.On("FindEndpointsDynamic", ctx, filter).Return(expectedEndpoints, nil).Once()

		results, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "GET", results[0].Method)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		mockRepo.On("FindEndpointsDynamic", ctx, filter).Return(nil, repoError).Once()

		results, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		mockRepo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetAccessRightsDynamic(t *testing.T) {
	mockRepo := new(mocks.MockAccessRepository)
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(mockRepo, log)
	ctx := context.Background()

	t.Run("Success - Get Access Rights Dynamically", func(t *testing.T) {
		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Manage"},
			},
		}
		expectedAccessRights := []*entity.AccessRight{
			{ID: "1", Name: "Manage Users"},
		}
		mockRepo.On("FindAccessRightsDynamic", ctx, filter).Return(expectedAccessRights, nil).Once()

		results, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results.Data, 1)
		assert.Equal(t, "Manage Users", results.Data[0].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		mockRepo.On("FindAccessRightsDynamic", ctx, filter).Return(nil, repoError).Once()

		results, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		mockRepo.AssertExpectations(t)
	})
}