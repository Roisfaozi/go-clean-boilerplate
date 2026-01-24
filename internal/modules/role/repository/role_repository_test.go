package repository_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

import (
	"fmt"
	"github.com/google/uuid"
)

func setupRoleRepo(t *testing.T) (repository.RoleRepository, *gorm.DB) {
	uid, _ := uuid.NewV7()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uid.String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Role{})
	require.NoError(t, err)

	logger := logrus.New()
	repo := repository.NewRoleRepository(db, logger)
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
