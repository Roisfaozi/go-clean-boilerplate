package repository_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupUserRepo(t *testing.T) (repository.UserRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.User{})
	require.NoError(t, err)

	logger := logrus.New()
	repo := repository.NewUserRepository(db, logger)
	return repo, db
}

func TestUserRepository_FindAllDynamic(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	users := []entity.User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Username: "alice"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Username: "bob"},
		{ID: "3", Name: "Charlie", Email: "charlie@example.com", Username: "charlie"},
	}
	db.Create(&users)

	tests := []struct {
		name          string
		filter        *querybuilder.DynamicFilter
		expectedCount int
		expectedNames []string
	}{
		{
			name: "Contains Name 'a'",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Name": {Type: "contains", From: "a"},
				},
				Sort: &[]querybuilder.SortModel{{ColId: "Name", Sort: "asc"}},
			},
			expectedCount: 2,
			expectedNames: []string{"Alice", "Charlie"},
		},
		{
			name: "Equals Username 'bob'",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Username": {Type: "equals", From: "bob"},
				},
			},
			expectedCount: 1,
			expectedNames: []string{"Bob"},
		},
		{
			name:          "No Filter (All)",
			filter:        &querybuilder.DynamicFilter{},
			expectedCount: 3,
			expectedNames: []string{"Alice", "Bob", "Charlie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindAllDynamic(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			if len(tt.expectedNames) > 0 {
				var names []string
				for _, u := range result {
					names = append(names, u.Name)
				}
				assert.ElementsMatch(t, tt.expectedNames, names)
			}
		})
	}
}
