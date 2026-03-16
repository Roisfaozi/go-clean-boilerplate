package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type accessTestDeps struct {
	Repo *mocks.MockAccessRepository
}

func setupAccessTest() (*accessTestDeps, usecase.IAccessUseCase) {
	deps := &accessTestDeps{
		Repo: new(mocks.MockAccessRepository),
	}
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(deps.Repo, log)
	return deps, uc
}

func TestCreateAccessRight(t *testing.T) {
	t.Run("Success - Create Valid Access Right", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).Return(nil).Once()
		req := model.CreateAccessRightRequest{
			Name:        "view_dashboard",
			Description: "Allows viewing the main dashboard",
		}
		createdAccessRight, err := uc.CreateAccessRight(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdAccessRight)
		assert.Equal(t, req.Name, createdAccessRight.Name)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Create Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.CreateAccessRightRequest{Name: "error_right"}
		repoErr := errors.New("db error")
		deps.Repo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).Return(repoErr).Once()

		_, err := uc.CreateAccessRight(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestGetAllAccessRights(t *testing.T) {
	t.Run("Success - Has Data", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		expectedEntities := []*entity.AccessRight{
			{ID: "1", Name: "view_dashboard"},
			{ID: "2", Name: "edit_settings"},
		}
		deps.Repo.On("GetAccessRights", ctx).Return(expectedEntities, nil).Once()
		results, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Data, 2)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Success - No Data", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRights", ctx).Return([]*entity.AccessRight{}, nil).Once()
		results, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Data, 0)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		repoErr := errors.New("db error")
		deps.Repo.On("GetAccessRights", ctx).Return(nil, repoErr).Once()

		results, err := uc.GetAllAccessRights(ctx)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Success - Sanitize Inputs", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.CreateAccessRightRequest{
			Name:        "<b>Bold Name</b>",
			Description: "<script>alert('xss')</script>",
		}

		// Expect HTML escaped strings (pkg.SanitizeString escapes HTML tags)
		expectedName := "&lt;b&gt;Bold Name&lt;/b&gt;"
		expectedDesc := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"

		deps.Repo.On("CreateAccessRight", ctx, mock.MatchedBy(func(ar *entity.AccessRight) bool {
			return ar.Name == expectedName && ar.Description == expectedDesc
		})).Return(nil).Once()

		createdAccessRight, err := uc.CreateAccessRight(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdAccessRight)
		assert.Equal(t, expectedName, createdAccessRight.Name)
		assert.Equal(t, expectedDesc, createdAccessRight.Description)
		deps.Repo.AssertExpectations(t)
	})

}

func TestCreateEndpoint(t *testing.T) {
	t.Run("Success - Create Valid Endpoint", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).Return(nil).Once()
		req := model.CreateEndpointRequest{Path: "/api/v1/test", Method: "GET"}
		createdEndpoint, err := uc.CreateEndpoint(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdEndpoint)
		assert.Equal(t, req.Path, createdEndpoint.Path)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Create Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.CreateEndpointRequest{Path: "/error", Method: "POST"}
		repoErr := errors.New("db error")
		deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).Return(repoErr).Once()

		_, err := uc.CreateEndpoint(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestLinkEndpointToAccessRight(t *testing.T) {
	t.Run("Success - Link Valid IDs", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		deps.Repo.On("LinkEndpointToAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(nil).Once()
		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		repoErr := errors.New("db error")
		deps.Repo.On("LinkEndpointToAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(repoErr).Once()

		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestDeleteAccessRight(t *testing.T) {
	id := "1"

	t.Run("Success - Delete Access Right", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
		deps.Repo.On("DeleteAccessRight", ctx, id).Return(nil).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRightByID", ctx, id).Return(nil, gorm.ErrRecordNotFound).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - GetAccessRightByID Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		repoErr := errors.New("db error")
		deps.Repo.On("GetAccessRightByID", ctx, id).Return(nil, repoErr).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - DeleteAccessRight Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
		repoErr := errors.New("db error")
		deps.Repo.On("DeleteAccessRight", ctx, id).Return(repoErr).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestDeleteEndpoint(t *testing.T) {
	id := "1"

	t.Run("Success - Delete Endpoint", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("DeleteEndpoint", ctx, id).Return(nil).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Not Found (GORM delete behavior)", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("DeleteEndpoint", ctx, id).Return(gorm.ErrRecordNotFound).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - DeleteEndpoint Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		repoErr := errors.New("db error")
		deps.Repo.On("DeleteEndpoint", ctx, id).Return(repoErr).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetEndpointsDynamic(t *testing.T) {
	t.Run("Success - Get Endpoints Dynamically", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Method": {Type: "equals", From: "GET"},
			},
		}
		expectedEndpoints := []*entity.Endpoint{
			{ID: "1", Path: "/api/test", Method: "GET"},
		}
		deps.Repo.On("FindEndpointsDynamic", ctx, filter).Return(expectedEndpoints, int64(1), nil).Once()

		results, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "GET", results[0].Method)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		deps.Repo.On("FindEndpointsDynamic", ctx, filter).Return(nil, int64(0), repoError).Once()

		results, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, int64(0), total)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetAccessRightsDynamic(t *testing.T) {
	t.Run("Success - Get Access Rights Dynamically", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Manage"},
			},
		}
		expectedAccessRights := []*entity.AccessRight{
			{ID: "1", Name: "Manage Users"},
		}
		deps.Repo.On("FindAccessRightsDynamic", ctx, filter).Return(expectedAccessRights, int64(1), nil).Once()

		results, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results.Data, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Manage Users", results.Data[0].Name)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		deps.Repo.On("FindAccessRightsDynamic", ctx, filter).Return(nil, int64(0), repoError).Once()

		results, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, int64(0), total)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestCreateAccessRight_Sanitization(t *testing.T) {
	deps, uc := setupAccessTest()
	ctx := context.Background()

	// Capture the entity passed to CreateAccessRight to verify sanitization
	var capturedEntity *entity.AccessRight
	deps.Repo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).
		Run(func(args mock.Arguments) {
			capturedEntity = args.Get(1).(*entity.AccessRight)
		}).
		Return(nil).Once()

	req := model.CreateAccessRightRequest{
		Name:        "<b>Bold</b> Right",
		Description: "<script>alert('xss')</script> Description",
	}

	createdAccessRight, err := uc.CreateAccessRight(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccessRight)

	// Verify that the entity passed to repo was HTML escaped
	assert.Equal(t, "&lt;b&gt;Bold&lt;/b&gt; Right", capturedEntity.Name)
	assert.Equal(t, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; Description", capturedEntity.Description)

	deps.Repo.AssertExpectations(t)
}

func TestCreateEndpoint_Sanitization(t *testing.T) {
	deps, uc := setupAccessTest()
	ctx := context.Background()

	var capturedEntity *entity.Endpoint
	deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).
		Run(func(args mock.Arguments) {
			capturedEntity = args.Get(1).(*entity.Endpoint)
		}).
		Return(nil).Once()

	req := model.CreateEndpointRequest{
		Path:   "/api/v1/test/<script>alert(1)</script>",
		Method: "GET",
	}

	createdEndpoint, err := uc.CreateEndpoint(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdEndpoint)

	// Verify that the entity passed to repo was HTML escaped
	assert.Equal(t, "/api/v1/test/&lt;script&gt;alert(1)&lt;/script&gt;", capturedEntity.Path)

	deps.Repo.AssertExpectations(t)
}

func TestCreateEndpoint_DuplicateDetection(t *testing.T) {
	deps, uc := setupAccessTest()

	req := model.CreateEndpointRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	expectedErr := exception.ErrConflict
	deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
		return true
	})).Return(expectedErr).Once()

	resp, err := uc.CreateEndpoint(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	deps.Repo.On("CreateEndpoint", mock.Anything, mock.Anything).Return(nil).Once()

}

func TestLinkEndpointToAccessRight_Duplicate(t *testing.T) {
	deps, uc := setupAccessTest()

	req := model.LinkEndpointRequest{
		AccessRightID: "access-right-uuid",
		EndpointID:    "endpoint-uuid",
	}

	// Case: Duplicate link
	deps.Repo.On("LinkEndpointToAccessRight", mock.Anything, req.AccessRightID, req.EndpointID).
		Return(errors.New("duplicate entry")) // Simulate DB error

	err := uc.LinkEndpointToAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestUnlinkEndpointFromAccessRight(t *testing.T) {
	t.Run("Success - Unlink Valid IDs", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		deps.Repo.On("UnlinkEndpointFromAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(nil).Once()
		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		repoErr := errors.New("db error")
		deps.Repo.On("UnlinkEndpointFromAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(repoErr).Once()

		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})
}
