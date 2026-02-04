package repository_test

import (
	"context"
	"testing"

	"io"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAccessRepo(t *testing.T) (repository.AccessRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Endpoint{}, &entity.AccessRight{})
	require.NoError(t, err)

	l := logrus.New()
	l.SetOutput(io.Discard)
	l.SetLevel(logrus.FatalLevel)
	
	repo := repository.NewAccessRepository(db, l)
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

// =============================================================================
// Endpoint CRUD Tests
// =============================================================================

func TestAccessRepository_CreateEndpoint(t *testing.T) {
	repo, _ := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		endpoint := &entity.Endpoint{
			ID:     "ep-1",
			Path:   "/api/test",
			Method: "GET",
		}
		err := repo.CreateEndpoint(ctx, endpoint)
		require.NoError(t, err)

		// Verify it was created
		found, err := repo.GetEndpointByID(ctx, "ep-1")
		require.NoError(t, err)
		assert.Equal(t, "/api/test", found.Path)
		assert.Equal(t, "GET", found.Method)
	})

	t.Run("Duplicate ID error", func(t *testing.T) {
		endpoint := &entity.Endpoint{
			ID:     "ep-dup",
			Path:   "/api/first",
			Method: "GET",
		}
		err := repo.CreateEndpoint(ctx, endpoint)
		require.NoError(t, err)

		// Try to create with same ID
		duplicate := &entity.Endpoint{
			ID:     "ep-dup",
			Path:   "/api/second",
			Method: "POST",
		}
		err = repo.CreateEndpoint(ctx, duplicate)
		require.Error(t, err)
	})
}

func TestAccessRepository_GetEndpoints(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Returns all endpoints", func(t *testing.T) {
		// Seed data
		endpoints := []entity.Endpoint{
			{ID: "get-1", Path: "/api/users", Method: "GET"},
			{ID: "get-2", Path: "/api/roles", Method: "POST"},
		}
		db.Create(&endpoints)

		result, err := repo.GetEndpoints(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})

	t.Run("Empty table returns empty slice", func(t *testing.T) {
		// Clean the table first
		db.Exec("DELETE FROM endpoints")

		result, err := repo.GetEndpoints(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestAccessRepository_GetEndpointByID(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Found", func(t *testing.T) {
		endpoint := entity.Endpoint{ID: "find-1", Path: "/api/find", Method: "GET"}
		db.Create(&endpoint)

		result, err := repo.GetEndpointByID(ctx, "find-1")
		require.NoError(t, err)
		assert.Equal(t, "/api/find", result.Path)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := repo.GetEndpointByID(ctx, "non-existent-id")
		require.Error(t, err)
	})
}

func TestAccessRepository_DeleteEndpoint(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		endpoint := entity.Endpoint{ID: "del-1", Path: "/api/delete", Method: "DELETE"}
		db.Create(&endpoint)

		err := repo.DeleteEndpoint(ctx, "del-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.GetEndpointByID(ctx, "del-1")
		require.Error(t, err)
	})

	t.Run("Delete non-existent does not error", func(t *testing.T) {
		err := repo.DeleteEndpoint(ctx, "never-existed")
		require.NoError(t, err) // GORM soft deletes don't error for missing records
	})
}

// =============================================================================
// AccessRight CRUD Tests
// =============================================================================

func TestAccessRepository_CreateAccessRight(t *testing.T) {
	repo, _ := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		ar := &entity.AccessRight{
			ID:          "ar-1",
			Name:        "Test Access",
			Description: "Test description",
		}
		err := repo.CreateAccessRight(ctx, ar)
		require.NoError(t, err)

		// Verify
		found, err := repo.GetAccessRightByID(ctx, "ar-1")
		require.NoError(t, err)
		assert.Equal(t, "Test Access", found.Name)
	})

	t.Run("Duplicate ID error", func(t *testing.T) {
		ar := &entity.AccessRight{ID: "ar-dup", Name: "First"}
		err := repo.CreateAccessRight(ctx, ar)
		require.NoError(t, err)

		duplicate := &entity.AccessRight{ID: "ar-dup", Name: "Second"}
		err = repo.CreateAccessRight(ctx, duplicate)
		require.Error(t, err)
	})
}

func TestAccessRepository_GetAccessRights(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Returns all with preloaded endpoints", func(t *testing.T) {
		// Create endpoint
		endpoint := entity.Endpoint{ID: "preload-ep", Path: "/api/preload", Method: "GET"}
		db.Create(&endpoint)

		// Create access right with endpoint
		ar := entity.AccessRight{
			ID:        "ar-preload",
			Name:      "Preload Test",
			Endpoints: []entity.Endpoint{endpoint},
		}
		db.Create(&ar)

		result, err := repo.GetAccessRights(ctx)
		require.NoError(t, err)

		// Find the created one
		var found *entity.AccessRight
		for _, r := range result {
			if r.ID == "ar-preload" {
				found = r
				break
			}
		}
		require.NotNil(t, found)
		assert.NotEmpty(t, found.Endpoints)
	})
}

func TestAccessRepository_GetAccessRightByID(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Found with preload", func(t *testing.T) {
		endpoint := entity.Endpoint{ID: "ar-ep-1", Path: "/api/ar", Method: "GET"}
		db.Create(&endpoint)

		ar := entity.AccessRight{
			ID:        "ar-find",
			Name:      "Find Test",
			Endpoints: []entity.Endpoint{endpoint},
		}
		db.Create(&ar)

		result, err := repo.GetAccessRightByID(ctx, "ar-find")
		require.NoError(t, err)
		assert.Equal(t, "Find Test", result.Name)
		assert.Len(t, result.Endpoints, 1)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := repo.GetAccessRightByID(ctx, "non-existent")
		require.Error(t, err)
	})
}

func TestAccessRepository_DeleteAccessRight(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		ar := entity.AccessRight{ID: "ar-del", Name: "Delete Me"}
		db.Create(&ar)

		err := repo.DeleteAccessRight(ctx, "ar-del")
		require.NoError(t, err)

		_, err = repo.GetAccessRightByID(ctx, "ar-del")
		require.Error(t, err)
	})
}

func TestAccessRepository_LinkEndpointToAccessRight(t *testing.T) {
	repo, db := setupAccessRepo(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Create endpoint and access right
		endpoint := entity.Endpoint{ID: "link-ep", Path: "/api/link", Method: "GET"}
		db.Create(&endpoint)

		ar := entity.AccessRight{ID: "link-ar", Name: "Link Test"}
		db.Create(&ar)

		// Link them
		err := repo.LinkEndpointToAccessRight(ctx, "link-ar", "link-ep")
		require.NoError(t, err)

		// Verify
		result, err := repo.GetAccessRightByID(ctx, "link-ar")
		require.NoError(t, err)
		assert.Len(t, result.Endpoints, 1)
		assert.Equal(t, "link-ep", result.Endpoints[0].ID)
	})
}
