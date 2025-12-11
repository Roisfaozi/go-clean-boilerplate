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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleRepo(t *testing.T) (repository.RoleRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
