package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/usecase"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		expectedEntities := []entity.AccessRight{
			{ID: 1, Name: "view_dashboard"},
			{ID: 2, Name: "edit_settings"},
		}
		mockRepo.On("GetAllAccessRights", ctx).Return(expectedEntities, nil).Once()
		results, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Data, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - No Data", func(t *testing.T) {
		mockRepo.On("GetAllAccessRights", ctx).Return([]entity.AccessRight{}, nil).Once()
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
		req := model.LinkEndpointRequest{AccessRightID: 1, EndpointID: 2}
		mockRepo.On("LinkEndpointToAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(nil).Once()
		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
