package usecase_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	txMock "github.com/Roisfaozi/go-clean-boilerplate/pkg/tx/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupOrganizationUseCase() (*mocks.MockOrganizationRepository, *mocks.MockOrganizationMemberRepository, *txMock.MockTransactionManager, *permissionMocks.IEnforcer, usecase.OrganizationUseCase) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel) // Suppress log output in tests

	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockOrganizationMemberRepository)
	tm := new(txMock.MockTransactionManager)
	enforcer := new(permissionMocks.IEnforcer)

	// Default behavior for enforcer with context to return itself
	enforcer.On("WithContext", mock.Anything).Return(enforcer)

	uc := usecase.NewOrganizationUseCase(log, tm, orgRepo, memberRepo, enforcer)
	return orgRepo, memberRepo, tm, enforcer, uc
}

// ===============================================
// CreateOrganization Tests
// ===============================================

func TestCreateOrganization_Success(t *testing.T) {
	orgRepo, _, tm, enforcer, uc := setupOrganizationUseCase()
	ctx := context.Background()

	request := &model.CreateOrganizationRequest{
		Name: "Acme Corp",
		Slug: "acme-corp",
	}
	userID := "user-123"

	// Setup mocks
	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("SlugExists", ctx, "acme-corp").Return(false, nil)
	orgRepo.On("Create", ctx, mock.AnythingOfType("*entity.Organization"), usecase.DefaultOwnerRoleID).Return(nil)
	enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	// Execute
	response, err := uc.CreateOrganization(ctx, userID, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Acme Corp", response.Name)
	assert.Equal(t, "acme-corp", response.Slug)
	assert.Equal(t, userID, response.OwnerID)
	assert.NotEmpty(t, response.ID)

	orgRepo.AssertExpectations(t)
	tm.AssertExpectations(t)
}

func TestCreateOrganization_SlugExists(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	request := &model.CreateOrganizationRequest{
		Name: "Acme Corp",
		Slug: "existing-slug",
	}
	userID := "user-123"

	// Setup mocks
	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(exception.ErrConflict)

	orgRepo.On("SlugExists", ctx, "existing-slug").Return(true, nil)

	// Execute
	response, err := uc.CreateOrganization(ctx, userID, request)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrConflict, err)
	assert.Nil(t, response)

	orgRepo.AssertExpectations(t)
	tm.AssertExpectations(t)
}

// ===============================================
// GetOrganization Tests
// ===============================================

func TestGetOrganization_Success(t *testing.T) {
	orgRepo, _, _, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	org := &entity.Organization{
		ID:      "org-123",
		Name:    "Acme Corp",
		Slug:    "acme-corp",
		OwnerID: "user-123",
		Status:  entity.OrgStatusActive,
	}

	orgRepo.On("FindByID", ctx, "org-123").Return(org, nil)

	// Execute
	response, err := uc.GetOrganization(ctx, "org-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "org-123", response.ID)
	assert.Equal(t, "Acme Corp", response.Name)

	orgRepo.AssertExpectations(t)
}

func TestGetOrganization_NotFound(t *testing.T) {
	orgRepo, _, _, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgRepo.On("FindByID", ctx, "non-existent").Return(nil, nil)

	// Execute
	response, err := uc.GetOrganization(ctx, "non-existent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
	assert.Nil(t, response)

	orgRepo.AssertExpectations(t)
}

// ===============================================
// GetOrganizationBySlug Tests
// ===============================================

func TestGetOrganizationBySlug_Success(t *testing.T) {
	orgRepo, _, _, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	org := &entity.Organization{
		ID:      "org-123",
		Name:    "Acme Corp",
		Slug:    "acme-corp",
		OwnerID: "user-123",
		Status:  entity.OrgStatusActive,
	}

	orgRepo.On("FindBySlug", ctx, "acme-corp").Return(org, nil)

	// Execute
	response, err := uc.GetOrganizationBySlug(ctx, "acme-corp")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "acme-corp", response.Slug)

	orgRepo.AssertExpectations(t)
}

// ===============================================
// UpdateOrganization Tests
// ===============================================

func TestUpdateOrganization_Success(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	existingOrg := &entity.Organization{
		ID:      "org-123",
		Name:    "Old Name",
		Slug:    "acme-corp",
		OwnerID: "user-123",
		Status:  entity.OrgStatusActive,
	}

	request := &model.UpdateOrganizationRequest{
		Name: "New Name",
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, "org-123").Return(existingOrg, nil)
	orgRepo.On("Update", ctx, mock.AnythingOfType("*entity.Organization")).Return(nil)

	// Execute
	response, err := uc.UpdateOrganization(ctx, "org-123", request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "New Name", response.Name)

	orgRepo.AssertExpectations(t)
	tm.AssertExpectations(t)
}

func TestUpdateOrganization_NotFound(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	request := &model.UpdateOrganizationRequest{
		Name: "New Name",
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(exception.ErrNotFound)

	orgRepo.On("FindByID", ctx, "non-existent").Return(nil, nil)

	// Execute
	response, err := uc.UpdateOrganization(ctx, "non-existent", request)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
	assert.Nil(t, response)

	tm.AssertExpectations(t)
}

// ===============================================
// GetUserOrganizations Tests
// ===============================================

func TestGetUserOrganizations_Success(t *testing.T) {
	orgRepo, _, _, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgs := []*entity.Organization{
		{ID: "org-1", Name: "Org 1", Slug: "org-1", OwnerID: "user-123"},
		{ID: "org-2", Name: "Org 2", Slug: "org-2", OwnerID: "user-456"},
	}

	orgRepo.On("FindUserOrganizations", ctx, "user-123").Return(orgs, nil)

	// Execute
	response, err := uc.GetUserOrganizations(ctx, "user-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Organizations, 2)

	orgRepo.AssertExpectations(t)
}

func TestGetUserOrganizations_Empty(t *testing.T) {
	orgRepo, _, _, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgRepo.On("FindUserOrganizations", ctx, "user-no-orgs").Return([]*entity.Organization{}, nil)

	// Execute
	response, err := uc.GetUserOrganizations(ctx, "user-no-orgs")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, response.Total)
	assert.Empty(t, response.Organizations)

	orgRepo.AssertExpectations(t)
}

// ===============================================
// DeleteOrganization Tests
// ===============================================

func TestDeleteOrganization_Success(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	org := &entity.Organization{
		ID:      "org-123",
		Name:    "Acme Corp",
		Slug:    "acme-corp",
		OwnerID: "user-123",
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, "org-123").Return(org, nil)
	orgRepo.On("Delete", ctx, "org-123").Return(nil)

	// Execute
	err := uc.DeleteOrganization(ctx, "org-123", "user-123")

	// Assert
	assert.NoError(t, err)

	orgRepo.AssertExpectations(t)
	tm.AssertExpectations(t)
}

func TestDeleteOrganization_NotOwner(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	org := &entity.Organization{
		ID:      "org-123",
		Name:    "Acme Corp",
		OwnerID: "actual-owner",
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(exception.ErrForbidden)

	orgRepo.On("FindByID", ctx, "org-123").Return(org, nil)

	// Execute - trying to delete as non-owner
	err := uc.DeleteOrganization(ctx, "org-123", "not-the-owner")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrForbidden, err)

	tm.AssertExpectations(t)
}

func TestDeleteOrganization_NotFound(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(exception.ErrNotFound)

	orgRepo.On("FindByID", ctx, "non-existent").Return(nil, nil)

	// Execute
	err := uc.DeleteOrganization(ctx, "non-existent", "user-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)

	tm.AssertExpectations(t)
}
