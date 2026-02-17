package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrganizationUseCase_Extended(t *testing.T) {
	t.Run("CreateOrganization - Enforcer Error (Simulated)", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		userID := "user-123"
		req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "acme-corp"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
		deps.OrgRepo.On("Create", ctx, mock.Anything, usecase.DefaultOwnerRoleID).Return(nil)
		// Enforcer error
		deps.Enforcer.On("AddGroupingPolicy", userID, usecase.DefaultOwnerRoleID, mock.AnythingOfType("string")).Return(false, errors.New("casbin error"))

		res, err := uc.CreateOrganization(ctx, userID, req)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("CreateOrganization - Slug Check Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		userID := "user-123"
		req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "acme-corp"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, errors.New("db check error"))

		res, err := uc.CreateOrganization(ctx, userID, req)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("UpdateOrganization - Repo Update Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.UpdateOrganizationRequest{Name: "New Name"}
		existingOrg := &entity.Organization{ID: orgID, Name: "Old Name"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
		deps.OrgRepo.On("Update", ctx, mock.Anything).Return(errors.New("db update error"))

		res, err := uc.UpdateOrganization(ctx, orgID, req)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("DeleteOrganization - Repo Delete Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "owner-1"
		org := &entity.Organization{ID: orgID, OwnerID: userID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.OrgRepo.On("Delete", ctx, orgID).Return(errors.New("db delete error"))

		err := uc.DeleteOrganization(ctx, orgID, userID)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("GetOrganization - FindByID Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, errors.New("db error"))

		res, err := uc.GetOrganization(ctx, orgID)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("GetOrganizationBySlug - FindBySlug Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		slug := "slug-1"

		deps.OrgRepo.On("FindBySlug", ctx, slug).Return(nil, errors.New("db error"))

		res, err := uc.GetOrganizationBySlug(ctx, slug)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("GetUserOrganizations - FindUserOrganizations Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		userID := "user-1"

		deps.OrgRepo.On("FindUserOrganizations", ctx, userID).Return(nil, errors.New("db error"))

		res, err := uc.GetUserOrganizations(ctx, userID)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}
