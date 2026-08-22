package repository_test

import (
	"context"
	"testing"

	"fmt"
	"io"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRoleRepo(t *testing.T) (repository.RoleRepository, *gorm.DB) {
	uid, _ := uuid.NewV7()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uid.String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Role{}, &orgEntity.Organization{})
	require.NoError(t, err)

	// Silent Logrus
	l := logrus.New()
	l.SetOutput(io.Discard)
	l.SetLevel(logrus.FatalLevel)

	repo := repository.NewRoleRepository(db, l)
	return repo, db
}

func TestRoleRepository_FindAllDynamic(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()

	roles := []entity.Role{
		{ID: "1", Name: "Admin", Description: "Administrator"},
		{ID: "2", Name: "Editor", Description: "Content Editor"},
		{ID: "3", Name: "Viewer", Description: "Read Only"},
	}
	db.Create(&roles)

	tests := []struct {
		name          string
		filter        *querybuilder.DynamicFilter
		expectedCount int
		expectedNames []string
	}{
		{
			name: "Contains Name 'd'",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Name": {Type: "contains", From: "d"},
				},
			},
			expectedCount: 2,
			expectedNames: []string{"Admin", "Editor"},
		},
		{
			name: "Sort Descending",
			filter: &querybuilder.DynamicFilter{
				Sort: &[]querybuilder.SortModel{{ColId: "Name", Sort: "desc"}},
			},
			expectedCount: 3,
			expectedNames: []string{"Viewer", "Editor", "Admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindAllDynamic(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			if len(tt.expectedNames) > 0 {
				var names []string
				for _, r := range result {
					names = append(names, r.Name)
				}

				if tt.name == "Sort Descending" {
					assert.Equal(t, tt.expectedNames, names)
				} else {
					assert.ElementsMatch(t, tt.expectedNames, names)
				}
			}
		})
	}
}

func TestRoleRepository_CRUD(t *testing.T) {
	repo, _ := setupRoleRepo(t)
	ctx := context.Background()

	role := &entity.Role{
		ID:          "role-1",
		Name:        "TestRole",
		Description: "Test Description",
	}

	// Create
	err := repo.Create(ctx, role)
	require.NoError(t, err)

	// FindByID
	found, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, role.Name, found.Name)

	// FindByName
	foundName, err := repo.FindByName(ctx, role.Name)
	require.NoError(t, err)
	assert.Equal(t, role.ID, foundName.ID)

	// FindAll
	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// Delete
	err = repo.Delete(ctx, role.ID)
	require.NoError(t, err)

	// Verify Delete
	_, err = repo.FindByID(ctx, role.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestRoleRepository_Update(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()

	role := &entity.Role{
		ID:          "role-2",
		Name:        "TestRole",
		Description: "Old Description",
	}

	err := repo.Create(ctx, role)
	require.NoError(t, err)

	role.Description = "New Description"
	err = repo.Update(ctx, role)
	require.NoError(t, err)

	var updatedRole entity.Role
	err = db.First(&updatedRole, "id = ?", role.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "New Description", updatedRole.Description)
}

func TestRoleRepository_FindAllDynamic_ErrorSort(t *testing.T) {
	repo, _ := setupRoleRepo(t)
	ctx := context.Background()

	filter := &querybuilder.DynamicFilter{
		Sort: &[]querybuilder.SortModel{{ColId: "NonExistentColumn", Sort: "desc"}},
	}

	// This will just get ignored or error depending on the underlying implementation. Let's verify behavior.
	_, err := repo.FindAllDynamic(ctx, filter)
	// According to querybuilder, unknown columns in sort might return an error. Let's assert an error.
	assert.Error(t, err)
}

func TestRoleRepository_FindAllDynamic_ErrorFilter(t *testing.T) {
	repo, _ := setupRoleRepo(t)
	ctx := context.Background()

	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"NonExistentColumn": {Type: "equals"},
		},
	}

	_, err := repo.FindAllDynamic(ctx, filter)
	assert.Error(t, err)
}

func TestRoleRepository_ErrorPath(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()

	// Close db to force error
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	// Create
	err := repo.Create(ctx, &entity.Role{})
	assert.Error(t, err)

	// Update
	err = repo.Update(ctx, &entity.Role{ID: "role-2"})
	assert.Error(t, err)

	// FindByID
	_, err = repo.FindByID(ctx, "role-2")
	assert.Error(t, err)

	// FindByName
	_, err = repo.FindByName(ctx, "test")
	assert.Error(t, err)

	// FindAll
	_, err = repo.FindAll(ctx)
	assert.Error(t, err)

	// FindAllDynamic
	_, err = repo.FindAllDynamic(ctx, &querybuilder.DynamicFilter{})
	assert.Error(t, err)

	// Delete
	err = repo.Delete(ctx, "role-2")
	assert.Error(t, err)
}

func TestRoleRepository_FindOrganizationRoleByID(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()
	orgID := "org-1"
	roleID := "role-org-1"

	// Create test org
	err := db.Create(&orgEntity.Organization{ID: orgID, Name: "Test Org 1", Slug: "org-1"}).Error
	require.NoError(t, err)

	role := &entity.Role{
		ID:             roleID,
		Name:           "OrgAdmin",
		Description:    "Org Admin",
		OrganizationID: &orgID,
	}
	err = repo.Create(ctx, role)
	require.NoError(t, err)

	found, err := repo.FindOrganizationRoleByID(ctx, orgID, roleID)
	require.NoError(t, err)
	assert.Equal(t, role.Name, found.Name)

	// Not found test
	_, err = repo.FindOrganizationRoleByID(ctx, "other-org", roleID)
	assert.Error(t, err)
}

func TestRoleRepository_FindOrganizationRoles(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()
	orgID := "org-2"

	// Create test org
	err := db.Create(&orgEntity.Organization{ID: orgID, Name: "Test Org 2", Slug: "org-2"}).Error
	require.NoError(t, err)

	roles := []entity.Role{
		{ID: "role-org-2-a", Name: "RoleA", OrganizationID: &orgID},
		{ID: "role-org-2-b", Name: "RoleB", OrganizationID: &orgID},
		{ID: "role-global", Name: "RoleGlobal", OrganizationID: nil}, // Should also be found because of "OR organization_id IS NULL"
	}
	db.Create(&roles)

	found, err := repo.FindOrganizationRoles(ctx, orgID)
	require.NoError(t, err)
	assert.Len(t, found, 3)

	names := []string{}
	for _, r := range found {
		names = append(names, r.Name)
	}
	assert.ElementsMatch(t, []string{"RoleA", "RoleB", "RoleGlobal"}, names)
}

func TestRoleRepository_FindByNameInScope(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()
	orgID := "org-3"
	otherOrgID := "org-4"

	// Create test orgs
	err := db.Create(&orgEntity.Organization{ID: orgID, Name: "Test Org 3", Slug: "org-3"}).Error
	require.NoError(t, err)
	err = db.Create(&orgEntity.Organization{ID: otherOrgID, Name: "Test Org 4", Slug: "org-4"}).Error
	require.NoError(t, err)

	roles := []entity.Role{
		{ID: "role-scoped", Name: "ScopedRole", OrganizationID: &orgID},
		{ID: "role-global-scoped", Name: "GlobalRole", OrganizationID: nil},
	}
	db.Create(&roles)

	// Test found in org
	found, err := repo.FindByNameInScope(ctx, "ScopedRole", &orgID)
	require.NoError(t, err)
	assert.Equal(t, "role-scoped", found.ID)

	// Test found global
	emptyOrg := ""
	foundGlobal, err := repo.FindByNameInScope(ctx, "GlobalRole", &emptyOrg)
	require.NoError(t, err)
	assert.Equal(t, "role-global-scoped", foundGlobal.ID)

	// Test not found in wrong org
	_, err = repo.FindByNameInScope(ctx, "ScopedRole", &otherOrgID)
	assert.Error(t, err)
}

func TestRoleRepository_DeleteInOrg(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()
	orgID := "org-5"
	roleID := "role-org-5"

	// Create test org
	err := db.Create(&orgEntity.Organization{ID: orgID, Name: "Test Org 5", Slug: "org-5"}).Error
	require.NoError(t, err)

	role := &entity.Role{
		ID:             roleID,
		Name:           "RoleToDelete",
		OrganizationID: &orgID,
	}
	err = repo.Create(ctx, role)
	require.NoError(t, err)

	// Delete missing orgID
	err = repo.DeleteInOrg(ctx, "", roleID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Delete wrong org
	err = repo.DeleteInOrg(ctx, "other-org", roleID)
	assert.Error(t, err)

	// Delete success
	err = repo.DeleteInOrg(ctx, orgID, roleID)
	require.NoError(t, err)

	// Verify
	var count int64
	db.Model(&entity.Role{}).Where("id = ?", roleID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRoleRepository_GetDBWithTx(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()

	// Inject transaction db into context
	txDB := db.Session(&gorm.Session{})
	ctxWithTx := tx.ContextWithDB(ctx, txDB)

	// Should still work without errors
	role := &entity.Role{
		ID:   "role-tx",
		Name: "TxRole",
	}
	err := repo.Create(ctxWithTx, role)
	require.NoError(t, err)
}

func TestRoleRepository_AdditionalErrorPath(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()
	orgID := "org-err"

	// Close db to force error
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	_, err := repo.FindOrganizationRoleByID(ctx, orgID, "role-err")
	assert.Error(t, err)

	_, err = repo.FindOrganizationRoles(ctx, orgID)
	assert.Error(t, err)

	_, err = repo.FindByNameInScope(ctx, "err-role", &orgID)
	assert.Error(t, err)

	err = repo.DeleteInOrg(ctx, orgID, "role-err")
	assert.Error(t, err)
}
