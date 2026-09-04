package usecase_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type roleTestDeps struct {
	Repo         *mocks.MockRoleRepository
	TM           *mocking.MockWithTransactionManager
	PermissionUC *permMocks.MockIPermissionUseCase
}

func setupRoleTest() (*roleTestDeps, usecase.RoleUseCase) {
	deps := &roleTestDeps{
		Repo:         new(mocks.MockRoleRepository),
		TM:           new(mocking.MockWithTransactionManager),
		PermissionUC: new(permMocks.MockIPermissionUseCase),
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, deps.PermissionUC)
	return deps, uc
}

func TestRoleUseCase_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "NewRole", Description: "Desc"}

		deps.Repo.On("FindByNameInScope", ctx, "NewRole", (*string)(nil)).Return(nil, exception.ErrNotFound)
		deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

		res, err := uc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "NewRole", res.Name)
	})

	t.Run("Conflict", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "ExistingRole", Description: "Desc"}

		deps.Repo.On("FindByNameInScope", ctx, "ExistingRole", (*string)(nil)).Return(&entity.Role{}, nil)

		res, err := uc.Create(ctx, req)
		assert.ErrorIs(t, err, exception.ErrConflict)
		assert.Nil(t, res)
	})

	t.Run("DBErrorFind", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "Role", Description: "Desc"}

		deps.Repo.On("FindByNameInScope", ctx, "Role", (*string)(nil)).Return(nil, errors.New("db error"))

		res, err := uc.Create(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("DBErrorCreate", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "Role", Description: "Desc"}

		deps.Repo.On("FindByNameInScope", ctx, "Role", (*string)(nil)).Return(nil, exception.ErrNotFound)
		deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(errors.New("db error"))

		res, err := uc.Create(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}

func TestRoleUseCase_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewDesc"}
		id := "role-1"

		role := &entity.Role{ID: id, Name: "Role", Description: "OldDesc"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)
		deps.Repo.On("Update", ctx, mock.MatchedBy(func(r *entity.Role) bool {
			return r.Description == "NewDesc"
		})).Return(nil)

		res, err := uc.Update(ctx, id, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "NewDesc", res.Description)
	})

	t.Run("NotFound", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewDesc"}
		id := "role-1"

		deps.Repo.On("FindByID", ctx, id).Return(nil, exception.ErrNotFound)

		res, err := uc.Update(ctx, id, req)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, res)
	})

	t.Run("DBErrorFind", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewDesc"}
		id := "role-1"

		deps.Repo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

		res, err := uc.Update(ctx, id, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("DBErrorUpdate", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewDesc"}
		id := "role-1"

		role := &entity.Role{ID: id, Name: "Role"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)
		deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(errors.New("db error"))

		res, err := uc.Update(ctx, id, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}

func TestRoleUseCase_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()

		roles := []*entity.Role{
			{ID: "1", Name: "Role1"},
			{ID: "2", Name: "Role2"},
		}
		deps.Repo.On("FindAll", ctx).Return(roles, nil)

		res, err := uc.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()

		deps.Repo.On("FindAll", ctx).Return(nil, errors.New("db error"))

		res, err := uc.GetAll(ctx)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}

func TestRoleUseCase_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-1"

		role := &entity.Role{ID: id, Name: "NormalRole"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)
		deps.Repo.On("Delete", ctx, id).Return(nil)
		deps.PermissionUC.On("DeleteRole", ctx, "NormalRole").Return(nil)

		err := uc.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("ForbiddenSuperadmin", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-super"

		role := &entity.Role{ID: id, Name: "role:superadmin"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)

		err := uc.Delete(ctx, id)
		assert.ErrorIs(t, err, exception.ErrForbidden)
	})

	t.Run("NotFound", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-1"

		deps.Repo.On("FindByID", ctx, id).Return(nil, exception.ErrNotFound)

		err := uc.Delete(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("DBErrorFind", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-1"

		deps.Repo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

		err := uc.Delete(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("DBErrorDelete", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-1"

		role := &entity.Role{ID: id, Name: "NormalRole"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)
		deps.Repo.On("Delete", ctx, id).Return(errors.New("db error"))

		err := uc.Delete(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("DBErrorPermissionCleanup", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		id := "role-1"

		role := &entity.Role{ID: id, Name: "NormalRole"}
		deps.Repo.On("FindByID", ctx, id).Return(role, nil)
		deps.Repo.On("Delete", ctx, id).Return(nil)
		deps.PermissionUC.On("DeleteRole", ctx, "NormalRole").Return(errors.New("perm error"))

		err := uc.Delete(ctx, id)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestRoleUseCase_DeleteForOrganization(t *testing.T) {
	t.Run("Success - Calls DeleteInOrg", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		orgID := "org-1"
		roleID := "role-1"
		role := &entity.Role{ID: roleID, Name: "OrgRole", OrganizationID: &orgID}

		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(role, nil)
		deps.Repo.On("DeleteInOrg", ctx, orgID, roleID).Return(nil)
		deps.PermissionUC.On("DeleteRoleInOrg", ctx, "OrgRole", orgID).Return(nil)

		err := uc.DeleteForOrganization(ctx, orgID, roleID)
		assert.NoError(t, err)
	})

	t.Run("NotFoundInOrg", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		orgID := "org-1"
		roleID := "role-1"

		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(nil, exception.ErrNotFound)

		err := uc.DeleteForOrganization(ctx, orgID, roleID)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("DeleteInOrgRecordNotFound", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		orgID := "org-1"
		roleID := "role-1"
		role := &entity.Role{ID: roleID, Name: "OrgRole", OrganizationID: &orgID}

		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(role, nil)
		deps.Repo.On("DeleteInOrg", ctx, orgID, roleID).Return(exception.ErrNotFound)

		err := uc.DeleteForOrganization(ctx, orgID, roleID)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})
}

func TestRoleUseCase_GetAllRolesDynamic(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		filter := &querybuilder.DynamicFilter{}

		roles := []*entity.Role{
			{ID: "1", Name: "Role1"},
		}
		deps.Repo.On("FindAllDynamic", ctx, filter).Return(roles, nil)

		res, err := uc.GetAllRolesDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("DBError", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		filter := &querybuilder.DynamicFilter{}

		deps.Repo.On("FindAllDynamic", ctx, filter).Return(nil, errors.New("db error"))

		res, err := uc.GetAllRolesDynamic(ctx, filter)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}

func TestRoleUseCase_CreateForOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "OrgRole", Description: "Desc"}
		orgID := "org-1"

		deps.Repo.On("FindByNameInScope", ctx, "OrgRole", &orgID).Return(nil, exception.ErrNotFound)
		deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

		res, err := uc.CreateForOrganization(ctx, orgID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "OrgRole", res.Name)
	})

	t.Run("EmptyOrgID", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.CreateRoleRequest{Name: "OrgRole", Description: "Desc"}

		res, err := uc.CreateForOrganization(ctx, "", req)
		assert.ErrorIs(t, err, exception.ErrBadRequest)
		assert.Nil(t, res)

		deps.Repo.AssertNotCalled(t, "FindByNameInScope", mock.Anything, mock.Anything, mock.Anything)
		deps.Repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}

func TestRoleUseCase_GetOrganizationRoles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		orgID := "org-1"

		roles := []*entity.Role{
			{ID: "1", Name: "Role1", OrganizationID: &orgID},
			{ID: "2", Name: "Role2", OrganizationID: &orgID},
		}
		deps.Repo.On("FindOrganizationRoles", ctx, orgID).Return(roles, nil)

		res, err := uc.GetOrganizationRoles(ctx, orgID)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("EmptyOrgID", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()

		res, err := uc.GetOrganizationRoles(ctx, "")
		assert.ErrorIs(t, err, exception.ErrBadRequest)
		assert.Nil(t, res)

		deps.Repo.AssertNotCalled(t, "FindOrganizationRoles", mock.Anything, mock.Anything)
	})

	t.Run("DBError", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		orgID := "org-1"

		deps.Repo.On("FindOrganizationRoles", ctx, orgID).Return(nil, errors.New("db error"))

		res, err := uc.GetOrganizationRoles(ctx, orgID)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}

func TestRoleUseCase_UpdateForOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewOrgDesc"}
		orgID := "org-1"
		roleID := "role-1"

		role := &entity.Role{ID: roleID, Name: "OrgRole", Description: "OldDesc", OrganizationID: &orgID}
		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(role, nil)
		deps.Repo.On("Update", ctx, mock.MatchedBy(func(r *entity.Role) bool {
			return r.Description == "NewOrgDesc"
		})).Return(nil)

		res, err := uc.UpdateForOrganization(ctx, orgID, roleID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "NewOrgDesc", res.Description)
	})

	t.Run("InvalidInputs", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewOrgDesc"}

		res, err := uc.UpdateForOrganization(ctx, "", "role-1", req)
		assert.ErrorIs(t, err, exception.ErrBadRequest)
		assert.Nil(t, res)

		res, err = uc.UpdateForOrganization(ctx, "org-1", "", req)
		assert.ErrorIs(t, err, exception.ErrBadRequest)
		assert.Nil(t, res)

		deps.Repo.AssertNotCalled(t, "FindOrganizationRoleByID", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("NotFound", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewOrgDesc"}
		orgID := "org-1"
		roleID := "role-1"

		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(nil, exception.ErrNotFound)

		res, err := uc.UpdateForOrganization(ctx, orgID, roleID, req)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, res)
	})

	t.Run("DBErrorFind", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewOrgDesc"}
		orgID := "org-1"
		roleID := "role-1"

		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(nil, errors.New("db error"))

		res, err := uc.UpdateForOrganization(ctx, orgID, roleID, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})

	t.Run("DBErrorUpdate", func(t *testing.T) {
		deps, uc := setupRoleTest()
		ctx := context.Background()
		req := &model.UpdateRoleRequest{Description: "NewOrgDesc"}
		orgID := "org-1"
		roleID := "role-1"

		role := &entity.Role{ID: roleID, Name: "OrgRole", OrganizationID: &orgID}
		deps.Repo.On("FindOrganizationRoleByID", ctx, orgID, roleID).Return(role, nil)
		deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(errors.New("db error"))

		res, err := uc.UpdateForOrganization(ctx, orgID, roleID, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
	})
}
