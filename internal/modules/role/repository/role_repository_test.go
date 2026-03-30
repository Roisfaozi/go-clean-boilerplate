package repository_test

import (
	"context"
	"testing"

	"fmt"
	"io"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
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

	err = db.AutoMigrate(&entity.Role{})
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
	sqlDB.Close()

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
