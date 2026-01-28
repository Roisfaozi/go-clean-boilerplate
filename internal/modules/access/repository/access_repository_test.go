package repository_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAccessRepo(t *testing.T) (repository.AccessRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Endpoint{}, &entity.AccessRight{})
	require.NoError(t, err)

	logger := logrus.New()
	repo := repository.NewAccessRepository(db, logger)
	return repo, db
}

func TestAccessRepository_FindEndpointsDynamic(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	// Seed Endpoints
	endpoints := []entity.Endpoint{
		{ID: "1", Path: "/api/users", Method: "GET"},
		{ID: "2", Path: "/api/users", Method: "POST"},
		{ID: "3", Path: "/api/roles", Method: "GET"},
	}
	db.Create(&endpoints)

	tests := []struct {
		name          string
		filter        *querybuilder.DynamicFilter
		expectedCount int
	}{
		{
			name: "Method GET",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Method": {Type: "equals", From: "GET"},
				},
			},
			expectedCount: 2,
		},
		{
			name: "Path contains 'users'",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Path": {Type: "contains", From: "users"},
				},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, total, err := repo.FindEndpointsDynamic(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, res, tt.expectedCount)
			assert.Equal(t, int64(tt.expectedCount), total)
		})
	}
}

func TestAccessRepository_FindAccessRightsDynamic(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	// Seed AccessRights
	ars := []entity.AccessRight{
		{ID: "1", Name: "User Management", Description: "Manage users"},
		{ID: "2", Name: "Role Management", Description: "Manage roles"},
	}
	db.Create(&ars)

	tests := []struct {
		name          string
		filter        *querybuilder.DynamicFilter
		expectedCount int
	}{
		{
			name: "Name contains 'User'",
			filter: &querybuilder.DynamicFilter{
				Filter: map[string]querybuilder.Filter{
					"Name": {Type: "contains", From: "User"},
				},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, total, err := repo.FindAccessRightsDynamic(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, res, tt.expectedCount)
			assert.Equal(t, int64(tt.expectedCount), total)
		})
	}
}
