package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type organizationTestDeps struct {
	OrgRepo    *mocks.MockOrganizationRepository
	MemberRepo *mocks.MockOrganizationMemberRepository
	TM         *mocking.MockWithTransactionManager
	Enforcer   *permissionMocks.IEnforcer
}

func setupOrganizationTest() (*organizationTestDeps, usecase.OrganizationUseCase) {
	deps := &organizationTestDeps{
		OrgRepo:    new(mocks.MockOrganizationRepository),
		MemberRepo: new(mocks.MockOrganizationMemberRepository),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   new(permissionMocks.IEnforcer),
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
		deps.Enforcer.On("AddGroupingPolicy", userID, usecase.DefaultOwnerRoleID, mock.AnythingOfType("string")).Return(true, nil)

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
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("casbin error"))

		_, err := uc.CreateOrganization(ctx, "u1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("XSS Payload", func(t *testing.T) {
		deps, uc := setupOrganizationTest()
		ctx := context.Background()
		xssName := "<script>alert(1)</script>Org"
		req := &model.CreateOrganizationRequest{Name: xssName, Slug: "xss-org"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
		deps.OrgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
			return org.Name == xssName // Verify raw storage
		}), mock.Anything).Return(nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		res, err := uc.CreateOrganization(ctx, "u1", req)
		assert.NoError(t, err)
		assert.Equal(t, xssName, res.Name)
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
