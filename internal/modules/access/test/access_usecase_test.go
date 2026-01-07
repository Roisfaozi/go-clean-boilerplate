package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	accessMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type accessTestDeps struct {
	Repo *accessMocks.MockAccessRepository
}

func setupAccessTest() (*accessTestDeps, usecase.IAccessUseCase) {
	deps := &accessTestDeps{
		Repo: new(accessMocks.MockAccessRepository),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewAccessUseCase(deps.Repo, log)
	return deps, uc
}

func TestAccessUseCase_CreateAccessRight_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	req := model.CreateAccessRightRequest{Name: "view_dashboard", Description: "View Dashboard"}

	deps.Repo.On("CreateAccessRight", mock.Anything, mock.MatchedBy(func(ar *entity.AccessRight) bool {
		return ar.Name == "view_dashboard" && ar.Description == "View Dashboard"
	})).Return(nil)

	res, err := uc.CreateAccessRight(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "view_dashboard", res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestAccessUseCase_CreateAccessRight_RepoError(t *testing.T) {
	deps, uc := setupAccessTest()
	req := model.CreateAccessRightRequest{Name: "err", Description: "err"}

	deps.Repo.On("CreateAccessRight", mock.Anything, mock.Anything).Return(errors.New("db error"))

	res, err := uc.CreateAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestAccessUseCase_GetAllAccessRights_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	rights := []*entity.AccessRight{{Name: "r1"}, {Name: "r2"}}

	deps.Repo.On("GetAccessRights", mock.Anything).Return(rights, nil)

	res, err := uc.GetAllAccessRights(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res.Data, 2)
}

func TestAccessUseCase_CreateEndpoint_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	req := model.CreateEndpointRequest{Path: "/api/v1/users", Method: "GET"}

	deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e *entity.Endpoint) bool {
		return e.Path == "/api/v1/users" && e.Method == "GET"
	})).Return(nil)

	res, err := uc.CreateEndpoint(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/users", res.Path)
}

func TestAccessUseCase_LinkEndpointToAccessRight_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	req := model.LinkEndpointRequest{AccessRightID: "ar1", EndpointID: "ep1"}

	deps.Repo.On("LinkEndpointToAccessRight", mock.Anything, "ar1", "ep1").Return(nil)

	err := uc.LinkEndpointToAccessRight(context.Background(), req)

	assert.NoError(t, err)
}

func TestAccessUseCase_DeleteAccessRight_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	id := "ar1"

	deps.Repo.On("GetAccessRightByID", mock.Anything, id).Return(&entity.AccessRight{ID: id}, nil)
	deps.Repo.On("DeleteAccessRight", mock.Anything, id).Return(nil)

	err := uc.DeleteAccessRight(context.Background(), id)

	assert.NoError(t, err)
}

func TestAccessUseCase_DeleteAccessRight_NotFound(t *testing.T) {
	deps, uc := setupAccessTest()
	id := "unknown"

	deps.Repo.On("GetAccessRightByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)

	err := uc.DeleteAccessRight(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestAccessUseCase_DeleteEndpoint_Success(t *testing.T) {
	deps, uc := setupAccessTest()
	id := "ep1"

	deps.Repo.On("DeleteEndpoint", mock.Anything, id).Return(nil)

	err := uc.DeleteEndpoint(context.Background(), id)

	assert.NoError(t, err)
}

func TestAccessUseCase_DeleteEndpoint_NotFound(t *testing.T) {
	deps, uc := setupAccessTest()
	id := "unknown"

	deps.Repo.On("DeleteEndpoint", mock.Anything, id).Return(gorm.ErrRecordNotFound)

	err := uc.DeleteEndpoint(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrNotFound)
}
