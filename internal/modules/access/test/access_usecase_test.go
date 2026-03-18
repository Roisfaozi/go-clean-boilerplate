package test

import (
	"context"
	"errors"
	"io"
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

type accessTestDeps struct {
	Repo *mocks.MockAccessRepository
}

func setupAccessTest() (*accessTestDeps, usecase.IAccessUseCase) {
	deps := &accessTestDeps{
		Repo: new(mocks.MockAccessRepository),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewAccessUseCase(deps.Repo, log)
	return deps, uc
}

func TestAccessUseCase_CreateAccessRight(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		req := model.CreateAccessRightRequest{
			Name:        "<script>Admin</script>", // Sanitize test
			Description: "Full access",
		}

		deps.Repo.EXPECT().CreateAccessRight(ctx, mock.AnythingOfType("*entity.AccessRight")).Return(nil).Once()

		res, err := uc.CreateAccessRight(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "&lt;script&gt;Admin&lt;/script&gt;", res.Name) // Should be HTML escaped
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		req := model.CreateAccessRightRequest{
			Name: "Admin",
		}

		deps.Repo.EXPECT().CreateAccessRight(ctx, mock.AnythingOfType("*entity.AccessRight")).Return(errors.New("db error")).Once()

		res, err := uc.CreateAccessRight(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetAllAccessRights(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		mockEntities := []*entity.AccessRight{
			{ID: "1", Name: "Admin"},
		}
		deps.Repo.EXPECT().GetAccessRights(ctx).Return(mockEntities, nil).Once()

		res, err := uc.GetAllAccessRights(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Data, 1)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().GetAccessRights(ctx).Return(nil, errors.New("db error")).Once()

		res, err := uc.GetAllAccessRights(ctx)
		assert.Error(t, err)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_CreateEndpoint(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		req := model.CreateEndpointRequest{
			Path:   "<b>/api/users</b>", // Sanitize test
			Method: "GET",
		}

		deps.Repo.EXPECT().CreateEndpoint(ctx, mock.AnythingOfType("*entity.Endpoint")).Return(nil).Once()

		res, err := uc.CreateEndpoint(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "&lt;b&gt;/api/users&lt;/b&gt;", res.Path) // Should be HTML escaped
		assert.Equal(t, "GET", res.Method)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		req := model.CreateEndpointRequest{
			Path:   "/api/users",
			Method: "GET",
		}

		deps.Repo.EXPECT().CreateEndpoint(ctx, mock.AnythingOfType("*entity.Endpoint")).Return(errors.New("db error")).Once()

		res, err := uc.CreateEndpoint(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_LinkUnlinkEndpoint(t *testing.T) {
	ctx := context.TODO()
	req := model.LinkEndpointRequest{
		AccessRightID: "ar-1",
		EndpointID:    "ep-1",
	}

	t.Run("Link Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().LinkEndpointToAccessRight(ctx, "ar-1", "ep-1").Return(nil).Once()

		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Link Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().LinkEndpointToAccessRight(ctx, "ar-1", "ep-1").Return(errors.New("db error")).Once()

		err := uc.LinkEndpointToAccessRight(ctx, req)
		assert.Error(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Unlink Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().UnlinkEndpointFromAccessRight(ctx, "ar-1", "ep-1").Return(nil).Once()

		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Unlink Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().UnlinkEndpointFromAccessRight(ctx, "ar-1", "ep-1").Return(errors.New("db error")).Once()

		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.Error(t, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_DeleteAccessRight(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		mockAR := &entity.AccessRight{ID: "ar-1"}
		deps.Repo.EXPECT().GetAccessRightByID(ctx, "ar-1").Return(mockAR, nil).Once()
		deps.Repo.EXPECT().DeleteAccessRight(ctx, "ar-1").Return(nil).Once()

		err := uc.DeleteAccessRight(ctx, "ar-1")
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().GetAccessRightByID(ctx, "ar-1").Return(nil, gorm.ErrRecordNotFound).Once()

		err := uc.DeleteAccessRight(ctx, "ar-1")
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Find Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().GetAccessRightByID(ctx, "ar-1").Return(nil, errors.New("db error")).Once()

		err := uc.DeleteAccessRight(ctx, "ar-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Delete Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		mockAR := &entity.AccessRight{ID: "ar-1"}
		deps.Repo.EXPECT().GetAccessRightByID(ctx, "ar-1").Return(mockAR, nil).Once()
		deps.Repo.EXPECT().DeleteAccessRight(ctx, "ar-1").Return(errors.New("db error")).Once()

		err := uc.DeleteAccessRight(ctx, "ar-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_DeleteEndpoint(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().DeleteEndpoint(ctx, "ep-1").Return(nil).Once()

		err := uc.DeleteEndpoint(ctx, "ep-1")
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().DeleteEndpoint(ctx, "ep-1").Return(gorm.ErrRecordNotFound).Once()

		err := uc.DeleteEndpoint(ctx, "ep-1")
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Delete Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		deps.Repo.EXPECT().DeleteEndpoint(ctx, "ep-1").Return(errors.New("db error")).Once()

		err := uc.DeleteEndpoint(ctx, "ep-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetDynamic(t *testing.T) {
	ctx := context.TODO()

	t.Run("GetEndpointsDynamic Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		filter := &querybuilder.DynamicFilter{Page: 1, PageSize: 10}
		mockEntities := []*entity.Endpoint{{ID: "1", Path: "/api", Method: "GET"}}

		deps.Repo.EXPECT().FindEndpointsDynamic(ctx, filter).Return(mockEntities, int64(1), nil).Once()

		res, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "/api", res[0].Path)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("GetEndpointsDynamic Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		filter := &querybuilder.DynamicFilter{}

		deps.Repo.EXPECT().FindEndpointsDynamic(ctx, filter).Return(nil, int64(0), errors.New("db error")).Once()

		res, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		assert.Equal(t, int64(0), total)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("GetAccessRightsDynamic Success", func(t *testing.T) {
		deps, uc := setupAccessTest()
		filter := &querybuilder.DynamicFilter{Page: 1, PageSize: 10}
		mockEntities := []*entity.AccessRight{{ID: "1", Name: "Admin"}}

		deps.Repo.EXPECT().FindAccessRightsDynamic(ctx, filter).Return(mockEntities, int64(1), nil).Once()

		res, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Data, 1)
		assert.Equal(t, int64(1), total)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("GetAccessRightsDynamic Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		filter := &querybuilder.DynamicFilter{}

		deps.Repo.EXPECT().FindAccessRightsDynamic(ctx, filter).Return(nil, int64(0), errors.New("db error")).Once()

		res, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		assert.Equal(t, int64(0), total)
		deps.Repo.AssertExpectations(t)
	})
}