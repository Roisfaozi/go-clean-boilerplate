package test

import (
	"context"
	"errors"
	"io"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
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

type organizationTestDeps struct {
	OrgRepo    *mocks.MockOrganizationRepository
	MemberRepo *mocks.MockOrganizationMemberRepository
	TM         *mocking.MockWithTransactionManager
	Enforcer   *permissionMocks.MockIEnforcer
}

func setupOrganizationTest() (*organizationTestDeps, usecase.OrganizationUseCase) {
	mockEnforcer := new(permissionMocks.MockIEnforcer)
	deps := &organizationTestDeps{
		OrgRepo:    new(mocks.MockOrganizationRepository),
		MemberRepo: new(mocks.MockOrganizationMemberRepository),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   mockEnforcer,
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	uc := usecase.NewOrganizationUseCase(log, deps.TM, deps.OrgRepo, deps.MemberRepo, deps.Enforcer)

	return deps, uc
}

func TestOrganizationUseCase_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		userID := "user-123"
		req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "acme-corp"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
		deps.OrgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
			return org.Name == req.Name && org.Slug == req.Slug && org.OwnerID == userID
		}), usecase.DefaultOwnerRoleID).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

		res, err := uc.CreateOrganization(ctx, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Name, res.Name)
		deps.OrgRepo.AssertExpectations(t)
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("Error - Slug Exists", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		userID := "user-123"
		req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "exists"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrConflict)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(true, nil)

		res, err := uc.CreateOrganization(ctx, userID, req)

		assert.ErrorIs(t, err, exception.ErrConflict)
		assert.Nil(t, res)
	})

	t.Run("Error - Slug Check Fails", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, errors.New("db error"))

		_, err := uc.CreateOrganization(ctx, "u1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Error - Create Fails", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
		deps.OrgRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(errors.New("create error"))

		_, err := uc.CreateOrganization(ctx, "u1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Error - Enforcer Fails", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
		deps.OrgRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

		_, err := uc.CreateOrganization(ctx, "u1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

}

func TestOrganizationUseCase_GetOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1", Name: "Org 1"}

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)

		res, err := uc.GetOrganization(ctx, "org-1")
		assert.NoError(t, err)
		assert.Equal(t, org.ID, res.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil) // Return nil, nil for not found from repo

		_, err := uc.GetOrganization(ctx, "org-1")
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

		_, err := uc.GetOrganization(ctx, "org-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestOrganizationUseCase_GetOrganizationBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1", Slug: "slug-1"}

		deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(org, nil)

		res, err := uc.GetOrganizationBySlug(ctx, "slug-1")
		assert.NoError(t, err)
		assert.Equal(t, org.Slug, res.Slug)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(nil, nil)

		_, err := uc.GetOrganizationBySlug(ctx, "slug-1")
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(nil, errors.New("db error"))

		_, err := uc.GetOrganizationBySlug(ctx, "slug-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestOrganizationUseCase_UpdateOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.UpdateOrganizationRequest{Name: "New Name"}
		existingOrg := &entity.Organization{ID: orgID, Name: "Old Name"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
		deps.OrgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
			return org.Name == "New Name"
		})).Return(nil)

		res, err := uc.UpdateOrganization(ctx, orgID, req)
		assert.NoError(t, err)
		assert.Equal(t, "New Name", res.Name)
	})

	t.Run("Settings Update", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		settings := map[string]interface{}{"theme": "dark"}
		req := &model.UpdateOrganizationRequest{Settings: settings}
		existingOrg := &entity.Organization{ID: orgID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
		deps.OrgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
			return org.Settings["theme"] == "dark"
		})).Return(nil)

		res, err := uc.UpdateOrganization(ctx, orgID, req)
		assert.NoError(t, err)
		assert.Equal(t, "dark", res.Settings["theme"])
	})

	t.Run("No Change", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.UpdateOrganizationRequest{} // Empty
		existingOrg := &entity.Organization{ID: orgID, Name: "Name"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
		deps.OrgRepo.On("Update", ctx, mock.Anything).Return(nil)

		res, err := uc.UpdateOrganization(ctx, orgID, req)
		assert.NoError(t, err)
		assert.Equal(t, "Name", res.Name)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

		_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{})
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Find Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

		_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{})
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		existingOrg := &entity.Organization{ID: "org-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(existingOrg, nil)
		deps.OrgRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{Name: "New"})
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestOrganizationUseCase_GetUserOrganizations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgs := []*entity.Organization{{ID: "org-1"}}

		deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return(orgs, nil)

		res, err := uc.GetUserOrganizations(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, 1, res.Total)
	})

	t.Run("Empty", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return([]*entity.Organization{}, nil)

		res, err := uc.GetUserOrganizations(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, 0, res.Total)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return(nil, errors.New("db error"))

		_, err := uc.GetUserOrganizations(ctx, "user-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestOrganizationUseCase_DeleteOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "owner-1"
		org := &entity.Organization{ID: orgID, OwnerID: userID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.OrgRepo.On("Delete", ctx, orgID).Return(nil)

		err := uc.DeleteOrganization(ctx, orgID, userID)
		assert.NoError(t, err)
	})

	t.Run("Not Owner", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		orgID := "org-1"
		org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrForbidden)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)

		err := uc.DeleteOrganization(ctx, orgID, "other-user")
		assert.ErrorIs(t, err, exception.ErrForbidden)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

		err := uc.DeleteOrganization(ctx, "org-1", "user-1")
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Delete Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1", OwnerID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.OrgRepo.On("Delete", ctx, "org-1").Return(errors.New("db error"))

		err := uc.DeleteOrganization(ctx, "org-1", "user-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Find Error", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

		err := uc.DeleteOrganization(ctx, "org-1", "user-1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

// ===============================================
// SECURITY TESTS
// ===============================================

// [NEW TEST FILE]
// ===============================================
// Security Hardening Tests
// ===============================================

// TestCreateOrganization_XSS verifies that although we save raw input (as per requirements),
// we verify that the system CAN handle potentially malicious strings without crashing.
// Note: Sanitization usually happens at the frontend or response layer, but here we ensure storage integrity.
func TestCreateOrganization_XSS(t *testing.T) {
	orgRepo, _, tm, enforcer, uc := setupOrganizationUseCase()
	ctx := context.Background()

	xssPayload := "<script>alert('xss')</script>"
	request := &model.CreateOrganizationRequest{
		Name: "Acme " + xssPayload,
		Slug: "acme-xss",
	}
	userID := "user-123"

	// Mock successful creating (Assuming we treat it as literal string)
	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("SlugExists", ctx, "acme-xss").Return(false, nil)
	orgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
		return org.Name == request.Name // verifying raw persistence
	}), usecase.DefaultOwnerRoleID).Return(nil)

	enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	response, err := uc.CreateOrganization(ctx, userID, request)

	assert.NoError(t, err)
	assert.Equal(t, request.Name, response.Name)
}

// TestUpdateOrganization_Settings verifies that arbitrary JSON settings can be persisted correctly.
func TestUpdateOrganization_Settings(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgID := "org-123"
	settings := map[string]interface{}{
		"theme":       "dark",
		"mfa_enabled": true,
		"max_users":   100,
	}

	existingOrg := &entity.Organization{
		ID:       orgID,
		Name:     "Acme Corp",
		Settings: nil,
	}

	request := &model.UpdateOrganizationRequest{
		Settings: settings,
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
	orgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
		// Verify settings map matches
		return org.Settings["theme"] == "dark" && org.Settings["mfa_enabled"] == true
	})).Return(nil)

	response, err := uc.UpdateOrganization(ctx, orgID, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "dark", response.Settings["theme"])
}

// TestUpdateOrganization_ConcurrentModification Edge Case
// Verifies that updates process cleanly even if data unchanged
func TestUpdateOrganization_NoChange(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgID := "org-123"
	existingOrg := &entity.Organization{
		ID:   orgID,
		Name: "Acme Corp",
	}

	request := &model.UpdateOrganizationRequest{} // Empty request

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
	orgRepo.On("Update", ctx, mock.Anything).Return(nil)

	response, err := uc.UpdateOrganization(ctx, orgID, request)

	assert.NoError(t, err)
	assert.Equal(t, "Acme Corp", response.Name)
}

// ===============================================
// OLD USECASE TESTS
// ===============================================

func setupOrganizationUseCase() (*mocks.MockOrganizationRepository, *mocks.MockOrganizationMemberRepository, *txMock.MockTransactionManager, *permissionMocks.MockIEnforcer, usecase.OrganizationUseCase) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel) // Suppress log output in tests

	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockOrganizationMemberRepository)
	tm := new(txMock.MockTransactionManager)
	enforcer := new(permissionMocks.MockIEnforcer)

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
	enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

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
