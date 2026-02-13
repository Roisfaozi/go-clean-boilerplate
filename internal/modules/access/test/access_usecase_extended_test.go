package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

func TestDeleteAccessRight_Extended(t *testing.T) {
	id := "ext-1"

	t.Run("Error - GetAccessRightByID Generic Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		repoErr := errors.New("db connection failed")
		deps.Repo.On("GetAccessRightByID", ctx, id).Return(nil, repoErr).Once()

		err := uc.DeleteAccessRight(ctx, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - DeleteAccessRight Generic Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		// First it gets the entity
		deps.Repo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
		// Then it tries to delete but fails
		repoErr := errors.New("delete failed")
		deps.Repo.On("DeleteAccessRight", ctx, id).Return(repoErr).Once()

		err := uc.DeleteAccessRight(ctx, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestDeleteEndpoint_Extended(t *testing.T) {
	id := "ext-1"

	t.Run("Error - DeleteEndpoint Generic Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		repoErr := errors.New("delete failed")
		deps.Repo.On("DeleteEndpoint", ctx, id).Return(repoErr).Once()

		err := uc.DeleteEndpoint(ctx, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - DeleteEndpoint Not Found", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("DeleteEndpoint", ctx, id).Return(gorm.ErrRecordNotFound).Once()

		err := uc.DeleteEndpoint(ctx, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})
}
