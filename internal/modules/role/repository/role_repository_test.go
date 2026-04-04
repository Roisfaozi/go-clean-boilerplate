package repository_test

import (
	"context"
	"testing"

	"fmt"
	"io"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
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

func TestRoleRepository_Context(t *testing.T) {
	repo, db := setupRoleRepo(t)

	// Create context with txKey
	// txKey is private in tx package so we can't create it directly but we can use WithinTransaction
	tm := tx.NewTransactionManager(db, logrus.New())

	err := tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		role := &entity.Role{
			ID:          "role-ctx-1",
			Name:        "TestRoleCtx",
			Description: "Original Description",
		}

		err := repo.Create(ctx, role)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, role.Name, found.Name)

		return nil
	})

	require.NoError(t, err)
}

func TestRoleRepository_UpdateContext(t *testing.T) {
	// Not actually needed since we just need to hit the line returning txDB in getDB.
	// TestRoleRepository_Context covers that line.
}


func TestRoleRepository_Update(t *testing.T) {
	repo, _ := setupRoleRepo(t)
	ctx := context.Background()

	role := &entity.Role{
		ID:          "role-update-1",
		Name:        "TestRoleUpdate",
		Description: "Original Description",
	}

	err := repo.Create(ctx, role)
	require.NoError(t, err)

	updateRole := &entity.Role{
		ID:          "role-update-1",
		Name:        "TestRoleUpdateIgnored",
		Description: "Updated Description",
	}

	err = repo.Update(ctx, updateRole)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "TestRoleUpdate", found.Name) // Name should not be updated due to Omit
	assert.Equal(t, "Updated Description", found.Description)
}

func TestRoleRepository_Errors(t *testing.T) {
	repo, db := setupRoleRepo(t)
	ctx := context.Background()

	// Try dynamic filter error before closing DB to hit GenerateDynamicSort error
	// We can pass an invalid sort column
	_, err := repo.FindAllDynamic(ctx, &querybuilder.DynamicFilter{
		Sort: &[]querybuilder.SortModel{{ColId: "InvalidColumn", Sort: "desc"}},
	})
	assert.Error(t, err)

	// Try dynamic filter error to hit GenerateDynamicQuery error
	_, err = repo.FindAllDynamic(ctx, &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"InvalidColumn": {Type: "equals", From: "test"},
		},
	})
	assert.Error(t, err)


	sqlDB, err := db.DB()
	require.NoError(t, err)
	err = sqlDB.Close()
	require.NoError(t, err)

	role := &entity.Role{
		ID:          "role-err-1",
		Name:        "TestRoleErr",
		Description: "Test Description",
	}

	err = repo.Create(ctx, role)
	assert.Error(t, err)

	err = repo.Update(ctx, role)
	assert.Error(t, err)

	_, err = repo.FindByID(ctx, "role-err-1")
	assert.Error(t, err)

	_, err = repo.FindByName(ctx, "TestRoleErr")
	assert.Error(t, err)

	_, err = repo.FindAll(ctx)
	assert.Error(t, err)

	_, err = repo.FindAllDynamic(ctx, &querybuilder.DynamicFilter{})
	assert.Error(t, err)

	err = repo.Delete(ctx, "role-err-1")
	assert.Error(t, err)
}
