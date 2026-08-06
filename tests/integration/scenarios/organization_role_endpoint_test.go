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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type orgRoleEndpointFixture struct {
	roleUC roleUseCase.RoleUseCase
	rRepo  roleRepo.RoleRepository
	orgAID string
	orgBID string
}

func newOrgRoleEndpointFixture(t *testing.T, env *setup.TestEnvironment, slugPrefix string) *orgRoleEndpointFixture {
	t.Helper()

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

	ownerA := setup.CreateTestUser(t, env.DB, slugPrefix+"OwnerA", slugPrefix+"ownera@example.com", "Password123!")
	ownerB := setup.CreateTestUser(t, env.DB, slugPrefix+"OwnerB", slugPrefix+"ownerb@example.com", "Password123!")

	orgA, err := orgUC.CreateOrganization(context.Background(), ownerA.ID, &orgModel.CreateOrganizationRequest{
		Name: slugPrefix + " Org A", Slug: slugPrefix + "-org-a",
	})
	require.NoError(t, err)
	orgB, err := orgUC.CreateOrganization(context.Background(), ownerB.ID, &orgModel.CreateOrganizationRequest{
		Name: slugPrefix + " Org B", Slug: slugPrefix + "-org-b",
	})
	require.NoError(t, err)

	return &orgRoleEndpointFixture{roleUC: rUC, rRepo: rRepo, orgAID: orgA.ID, orgBID: orgB.ID}
}

func TestOrganizationRoleEndpoints_Create(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newOrgRoleEndpointFixture(t, env, "epcreate")

	tests := []struct {
		name        string
		orgID       string
		request     *roleModel.CreateRoleRequest
		expectedErr error
		assertRole  func(t *testing.T, resp *roleModel.RoleResponse)
	}{
		{
			name:  "creates role scoped to the organization",
			orgID: f.orgAID,
			request: &roleModel.CreateRoleRequest{
				Name: "ep_editor", Description: "Editor",
			},
			assertRole: func(t *testing.T, resp *roleModel.RoleResponse) {
				require.NotNil(t, resp.OrganizationID)
				assert.Equal(t, f.orgAID, *resp.OrganizationID)
			},
		},
		{
			name:  "allows same role name in a different organization",
			orgID: f.orgBID,
			request: &roleModel.CreateRoleRequest{
				Name: "ep_editor", Description: "Editor B",
			},
			assertRole: func(t *testing.T, resp *roleModel.RoleResponse) {
				require.NotNil(t, resp.OrganizationID)
				assert.Equal(t, f.orgBID, *resp.OrganizationID)
			},
		},
		{
			name:        "rejects duplicate role name within the same organization",
			orgID:       f.orgAID,
			request:     &roleModel.CreateRoleRequest{Name: "ep_editor", Description: "dup"},
			expectedErr: exception.ErrConflict,
		},
		{
			name:        "rejects reserved role name",
			orgID:       f.orgAID,
			request:     &roleModel.CreateRoleRequest{Name: "role:admin"},
			expectedErr: exception.ErrBadRequest,
		},
		{
			name:        "rejects empty organization id",
			orgID:       "",
			request:     &roleModel.CreateRoleRequest{Name: "no_org"},
			expectedErr: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := f.roleUC.CreateForOrganization(context.Background(), tt.orgID, tt.request)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.assertRole != nil {
				tt.assertRole(t, resp)
			}
		})
	}
}

func TestOrganizationRoleEndpoints_ListScoping(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newOrgRoleEndpointFixture(t, env, "eplist")

	_, err := f.roleUC.CreateForOrganization(context.Background(), f.orgAID, &roleModel.CreateRoleRequest{Name: "list_a"})
	require.NoError(t, err)
	_, err = f.roleUC.CreateForOrganization(context.Background(), f.orgBID, &roleModel.CreateRoleRequest{Name: "list_b"})
	require.NoError(t, err)

	tests := []struct {
		name          string
		orgID         string
		expectedErr   error
		mustContain   string
		mustNotContai string
	}{
		{
			name:          "lists only roles owned by organization A",
			orgID:         f.orgAID,
			mustContain:   "list_a",
			mustNotContai: "list_b",
		},
		{
			name:          "lists only roles owned by organization B",
			orgID:         f.orgBID,
			mustContain:   "list_b",
			mustNotContai: "list_a",
		},
		{
			name:        "rejects empty organization id",
			orgID:       "",
			expectedErr: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := f.roleUC.GetOrganizationRoles(context.Background(), tt.orgID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			names := make([]string, 0, len(roles))
			for _, r := range roles {
				require.NotNil(t, r.OrganizationID, "organization role must carry organization_id")
				assert.Equal(t, tt.orgID, *r.OrganizationID, "listing must not leak roles across tenants")
				names = append(names, r.Name)
			}

			assert.Contains(t, names, tt.mustContain)
			assert.NotContains(t, names, tt.mustNotContai)
		})
	}
}

func TestOrganizationRoleEndpoints_UpdateScopeIsolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newOrgRoleEndpointFixture(t, env, "epupdate")

	roleA, err := f.roleUC.CreateForOrganization(context.Background(), f.orgAID, &roleModel.CreateRoleRequest{
		Name: "upd_role", Description: "Original A",
	})
	require.NoError(t, err)
	roleB, err := f.roleUC.CreateForOrganization(context.Background(), f.orgBID, &roleModel.CreateRoleRequest{
		Name: "upd_role", Description: "Original B",
	})
	require.NoError(t, err)

	tests := []struct {
		name                string
		orgID               string
		roleID              string
		description         string
		expectedErr         error
		unchangedRoleID     string
		unchangedRoleOrgID  string
		unchangedRoleDetail string
	}{
		{
			name:                "cross-tenant update is rejected and leaves data untouched",
			orgID:               f.orgAID,
			roleID:              roleB.ID,
			description:         "hijacked",
			expectedErr:         exception.ErrNotFound,
			unchangedRoleID:     roleB.ID,
			unchangedRoleOrgID:  f.orgBID,
			unchangedRoleDetail: "Original B",
		},
		{
			name:        "update of unknown role id is rejected",
			orgID:       f.orgAID,
			roleID:      "00000000-0000-0000-0000-000000000000",
			description: "ghost",
			expectedErr: exception.ErrNotFound,
		},
		{
			name:        "rejects empty organization id",
			orgID:       "",
			roleID:      roleA.ID,
			description: "no org",
			expectedErr: exception.ErrBadRequest,
		},
		{
			name:        "rejects empty role id",
			orgID:       f.orgAID,
			roleID:      "",
			description: "no role",
			expectedErr: exception.ErrBadRequest,
		},
		{
			name:        "own-tenant update succeeds",
			orgID:       f.orgAID,
			roleID:      roleA.ID,
			description: "Editor Updated",
		},
		{
			name:        "own-tenant update with empty description persists empty string in DB",
			orgID:       f.orgAID,
			roleID:      roleA.ID,
			description: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := f.roleUC.UpdateForOrganization(context.Background(), tt.orgID, tt.roleID, &roleModel.UpdateRoleRequest{
				Description: tt.description,
			})

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)

				if tt.unchangedRoleID != "" {
					untouched, findErr := f.rRepo.FindOrganizationRoleByID(context.Background(), tt.unchangedRoleOrgID, tt.unchangedRoleID)
					require.NoError(t, findErr)
					assert.Equal(t, tt.unchangedRoleDetail, untouched.Description, "target tenant data must stay untouched")
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.description, resp.Description)

			persisted, findErr := f.rRepo.FindOrganizationRoleByID(context.Background(), tt.orgID, tt.roleID)
			require.NoError(t, findErr)
			assert.Equal(t, tt.description, persisted.Description, "updated description must be persisted in DB")
		})
	}
}

func TestOrganizationRoleEndpoints_DeleteScopeIsolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newOrgRoleEndpointFixture(t, env, "epdelete")

	roleA, err := f.roleUC.CreateForOrganization(context.Background(), f.orgAID, &roleModel.CreateRoleRequest{Name: "del_role"})
	require.NoError(t, err)
	roleB, err := f.roleUC.CreateForOrganization(context.Background(), f.orgBID, &roleModel.CreateRoleRequest{Name: "del_role"})
	require.NoError(t, err)

	tests := []struct {
		name           string
		orgID          string
		roleID         string
		expectedErr    error
		survivorOrgID  string
		survivorRoleID string
	}{
		{
			name:           "cross-tenant delete is rejected and role survives",
			orgID:          f.orgAID,
			roleID:         roleB.ID,
			expectedErr:    exception.ErrNotFound,
			survivorOrgID:  f.orgBID,
			survivorRoleID: roleB.ID,
		},
		{
			name:        "delete of unknown role id is rejected",
			orgID:       f.orgAID,
			roleID:      "00000000-0000-0000-0000-000000000000",
			expectedErr: exception.ErrNotFound,
		},
		{
			name:           "own-tenant delete succeeds and other tenant role survives",
			orgID:          f.orgAID,
			roleID:         roleA.ID,
			survivorOrgID:  f.orgBID,
			survivorRoleID: roleB.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.roleUC.DeleteForOrganization(context.Background(), tt.orgID, tt.roleID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tt.survivorRoleID != "" {
				_, findErr := f.rRepo.FindOrganizationRoleByID(context.Background(), tt.survivorOrgID, tt.survivorRoleID)
				require.NoError(t, findErr, "role of the other tenant must remain intact")
			}
		})
	}
}

func TestRoleRepository_DirectEdgeCases(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newOrgRoleEndpointFixture(t, env, "repoedge")

	roleA, err := f.roleUC.CreateForOrganization(context.Background(), f.orgAID, &roleModel.CreateRoleRequest{
		Name: "repo_edge_role", Description: "Initial Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "delete without org in context removes role and errors when missing",
			run: func(t *testing.T) {
				ctx := context.Background()
				require.NoError(t, f.rRepo.Delete(ctx, roleA.ID))

				err := f.rRepo.Delete(ctx, roleA.ID)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
		{
			name: "FindByNameInScope ignores conflicting org in context",
			run: func(t *testing.T) {
				roleB, err := f.roleUC.CreateForOrganization(context.Background(), f.orgBID, &roleModel.CreateRoleRequest{
					Name: "scoped_name",
				})
				require.NoError(t, err)

				ctx := database.SetOrganizationContext(context.Background(), f.orgAID)
				found, err := f.rRepo.FindByNameInScope(ctx, "scoped_name", &f.orgBID)
				require.NoError(t, err)
				assert.Equal(t, roleB.ID, found.ID)
			},
		},
		{
			name: "DeleteInOrg refuses role owned by another organization",
			run: func(t *testing.T) {
				roleB, err := f.roleUC.CreateForOrganization(context.Background(), f.orgBID, &roleModel.CreateRoleRequest{
					Name: "guard_role",
				})
				require.NoError(t, err)

				err = f.rRepo.DeleteInOrg(context.Background(), f.orgAID, roleB.ID)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

				_, findErr := f.rRepo.FindOrganizationRoleByID(context.Background(), f.orgBID, roleB.ID)
				require.NoError(t, findErr, "role must survive a cross-tenant delete attempt")
			},
		},
		{
			name: "DeleteInOrg succeeds for owning organization without org context",
			run: func(t *testing.T) {
				role, err := f.roleUC.CreateForOrganization(context.Background(), f.orgAID, &roleModel.CreateRoleRequest{
					Name: "guard_own",
				})
				require.NoError(t, err)

				require.NoError(t, f.rRepo.DeleteInOrg(context.Background(), f.orgAID, role.ID))
			},
		},
		{
			name: "DeleteInOrg returns not found when role is missing",
			run: func(t *testing.T) {
				err := f.rRepo.DeleteInOrg(context.Background(), f.orgAID, "00000000-0000-0000-0000-000000000000")
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
