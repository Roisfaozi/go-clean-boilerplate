//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	roleUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationRoleEndpoints_ScopeIsolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	pUC := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
	rUC := roleUseCase.NewRoleUseCase(env.Logger, tm, rRepo, pUC)
	orgUC := orgUsecase.NewOrganizationUseCase(env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo)

	ownerA := setup.CreateTestUser(t, env.DB, "epOwnerA", "epownera@example.com", "Password123!")
	ownerB := setup.CreateTestUser(t, env.DB, "epOwnerB", "epownerb@example.com", "Password123!")

	orgA, err := orgUC.CreateOrganization(context.Background(), ownerA.ID, &orgModel.CreateOrganizationRequest{Name: "EP Org A", Slug: "ep-org-a"})
	require.NoError(t, err)
	orgB, err := orgUC.CreateOrganization(context.Background(), ownerB.ID, &orgModel.CreateOrganizationRequest{Name: "EP Org B", Slug: "ep-org-b"})
	require.NoError(t, err)

	ctx := context.Background()

	// CREATE: role tercipta dengan organization_id yang benar
	created, err := rUC.CreateForOrganization(ctx, orgA.ID, &roleModel.CreateRoleRequest{
		Name:        "ep_editor",
		Description: "Editor",
	})
	require.NoError(t, err)
	require.NotNil(t, created.OrganizationID)
	assert.Equal(t, orgA.ID, *created.OrganizationID)

	// CREATE: nama sama boleh dipakai di organisasi berbeda (unique per org)
	createdB, err := rUC.CreateForOrganization(ctx, orgB.ID, &roleModel.CreateRoleRequest{
		Name:        "ep_editor",
		Description: "Editor B",
	})
	require.NoError(t, err, "nama role sama harus boleh di organisasi berbeda")

	// CREATE: reserved name ditolak
	_, err = rUC.CreateForOrganization(ctx, orgA.ID, &roleModel.CreateRoleRequest{Name: "role:admin"})
	require.ErrorIs(t, err, exception.ErrBadRequest)

	// CREATE: orgID kosong ditolak
	_, err = rUC.CreateForOrganization(ctx, "", &roleModel.CreateRoleRequest{Name: "no_org"})
	require.ErrorIs(t, err, exception.ErrBadRequest)

	// LIST: hanya role milik org sendiri
	listA, err := rUC.GetOrganizationRoles(ctx, orgA.ID)
	require.NoError(t, err)
	for _, r := range listA {
		require.NotNil(t, r.OrganizationID)
		assert.Equal(t, orgA.ID, *r.OrganizationID, "list tidak boleh bocor lintas tenant")
	}

	// UPDATE cross-tenant HARUS ditolak
	_, err = rUC.UpdateForOrganization(ctx, orgA.ID, createdB.ID, &roleModel.UpdateRoleRequest{
		Description: "hijacked",
	})
	require.ErrorIs(t, err, exception.ErrNotFound, "org A tidak boleh mengubah role org B")

	// pastikan data org B tidak berubah
	stillB, err := rRepo.FindOrganizationRoleByID(ctx, orgB.ID, createdB.ID)
	require.NoError(t, err)
	assert.Equal(t, "Editor B", stillB.Description)

	// UPDATE own-tenant berhasil
	updated, err := rUC.UpdateForOrganization(ctx, orgA.ID, created.ID, &roleModel.UpdateRoleRequest{
		Description: "Editor Updated",
	})
	require.NoError(t, err)
	assert.Equal(t, "Editor Updated", updated.Description)

	// DELETE cross-tenant HARUS ditolak
	err = rUC.DeleteForOrganization(ctx, orgA.ID, createdB.ID)
	require.ErrorIs(t, err, exception.ErrNotFound, "org A tidak boleh menghapus role org B")

	// DELETE own-tenant berhasil
	require.NoError(t, rUC.DeleteForOrganization(ctx, orgA.ID, created.ID))

	// role org B tetap ada setelah delete milik org A
	_, err = rRepo.FindOrganizationRoleByID(ctx, orgB.ID, createdB.ID)
	require.NoError(t, err, "role org B harus tetap utuh")
}
