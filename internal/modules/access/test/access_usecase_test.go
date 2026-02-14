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
	tests := []struct {
		name           string
		req            model.CreateAccessRightRequest
		setupMock      func(deps *accessTestDeps)
		verifyResponse func(t *testing.T, res *model.AccessRightResponse, err error)
	}{
		{
			name: "Success - Create Valid Access Right",
			req: model.CreateAccessRightRequest{
				Name:        "view_dashboard",
				Description: "Allows viewing the main dashboard",
			},
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("CreateAccessRight", mock.Anything, mock.MatchedBy(func(ar *entity.AccessRight) bool {
					return ar.Name == "view_dashboard"
				})).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "view_dashboard", res.Name)
			},
		},
		{
			name: "Error - Repository Create Fails",
			req:  model.CreateAccessRightRequest{Name: "error_right"},
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("CreateAccessRight", mock.Anything, mock.AnythingOfType("*entity.AccessRight")).
					Return(errors.New("db error")).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, "db error", err.Error())
			},
		},
		{
			name: "Success - Sanitize Inputs",
			req: model.CreateAccessRightRequest{
				Name:        "<b>Bold Name</b>",
				Description: "<script>alert('xss')</script>",
			},
			setupMock: func(deps *accessTestDeps) {
				expectedName := "Bold Name"
				expectedDesc := "alert(&#39;xss&#39;)"
				deps.Repo.On("CreateAccessRight", mock.Anything, mock.MatchedBy(func(ar *entity.AccessRight) bool {
					return ar.Name == expectedName && ar.Description == expectedDesc
				})).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "Bold Name", res.Name)
				assert.Equal(t, "alert(&#39;xss&#39;)", res.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			if tt.setupMock != nil {
				tt.setupMock(deps)
			}

			res, err := uc.CreateAccessRight(ctx, tt.req)
			tt.verifyResponse(t, res, err)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestGetAllAccessRights(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(deps *accessTestDeps)
		verifyResponse func(t *testing.T, res *model.AccessRightListResponse, err error)
	}{
		{
			name: "Success - Has Data",
			setupMock: func(deps *accessTestDeps) {
				expectedEntities := []*entity.AccessRight{
					{ID: "1", Name: "view_dashboard"},
					{ID: "2", Name: "edit_settings"},
				}
				deps.Repo.On("GetAccessRights", mock.Anything).Return(expectedEntities, nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightListResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Len(t, res.Data, 2)
			},
		},
		{
			name: "Success - No Data",
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("GetAccessRights", mock.Anything).Return([]*entity.AccessRight{}, nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightListResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Len(t, res.Data, 0)
			},
		},
		{
			name: "Error - Repository Fails",
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("GetAccessRights", mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			verifyResponse: func(t *testing.T, res *model.AccessRightListResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
				assert.Equal(t, "db error", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			if tt.setupMock != nil {
				tt.setupMock(deps)
			}

			res, err := uc.GetAllAccessRights(ctx)
			tt.verifyResponse(t, res, err)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestCreateEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		req            model.CreateEndpointRequest
		setupMock      func(deps *accessTestDeps)
		verifyResponse func(t *testing.T, res *model.EndpointResponse, err error)
	}{
		{
			name: "Success - Create Valid Endpoint",
			req: model.CreateEndpointRequest{
				Path:   "/api/v1/test",
				Method: "GET",
			},
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e *entity.Endpoint) bool {
					return e.Path == "/api/v1/test" && e.Method == "GET"
				})).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.EndpointResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "/api/v1/test", res.Path)
			},
		},
		{
			name: "Error - Repository Create Fails",
			req:  model.CreateEndpointRequest{Path: "/error", Method: "POST"},
			setupMock: func(deps *accessTestDeps) {
				deps.Repo.On("CreateEndpoint", mock.Anything, mock.AnythingOfType("*entity.Endpoint")).
					Return(errors.New("db error")).Once()
			},
			verifyResponse: func(t *testing.T, res *model.EndpointResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, "db error", err.Error())
			},
		},
		{
			name: "Success - Sanitize Inputs",
			req: model.CreateEndpointRequest{
				Path:   "/api/v1/test/<script>alert(1)</script>",
				Method: "GET",
			},
			setupMock: func(deps *accessTestDeps) {
				expectedPath := "/api/v1/test/alert(1)"
				deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e *entity.Endpoint) bool {
					return e.Path == expectedPath
				})).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res *model.EndpointResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "/api/v1/test/alert(1)", res.Path)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			if tt.setupMock != nil {
				tt.setupMock(deps)
			}

			res, err := uc.CreateEndpoint(ctx, tt.req)
			tt.verifyResponse(t, res, err)
			deps.Repo.AssertExpectations(t)
		})
	}
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
