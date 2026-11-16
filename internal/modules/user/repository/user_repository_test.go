package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	_ "github.com/glebarez/go-sqlite"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserRepo(t *testing.T) (repository.UserRepository, *sqlx.DB) {
	// Use unique DB name to prevent shared state issues in tests
	dbName := uuid.New().String()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)

	// Create a new Logrus logger that discards output
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(io.Discard)
	logrusLogger.SetLevel(logrus.FatalLevel)

	db, err := sqlx.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			name TEXT,
			slug TEXT,
			owner_id TEXT,
			status TEXT,
			created_at INTEGER DEFAULT 0,
			updated_at INTEGER DEFAULT 0,
			deleted_at INTEGER DEFAULT 0
		);
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			password TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			token TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			email_verified_at INTEGER,
			created_at INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL DEFAULT 0,
			deleted_at INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE organization_members (
			id TEXT,
			organization_id TEXT,
			user_id TEXT,
			role_id TEXT,
			created_at INTEGER DEFAULT 0,
			updated_at INTEGER DEFAULT 0
		);
		CREATE TABLE user_sso_identities (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			provider_id TEXT NOT NULL UNIQUE,
			created_at INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL DEFAULT 0
		);
	`)
	require.NoError(t, err)

	repo := repository.NewUserRepository(db, logrusLogger)
	return repo, db
}

func seedUser(t *testing.T, db *sqlx.DB, u *entity.User) {
	t.Helper()
	status := u.Status
	if status == "" {
		status = entity.UserStatusActive
	}
	_, err := db.Exec(`
		INSERT INTO users (id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, u.ID, u.OrganizationID, u.Password, u.Email, u.Username, u.Name, u.AvatarURL, u.Token, status, u.EmailVerifiedAt, u.CreatedAt, u.UpdatedAt, int64(u.DeletedAt))
	require.NoError(t, err)
}

func TestUserRepository_Create(t *testing.T) {
	repo, _ := setupUserRepo(t)
	ctx := context.Background()

	user := &entity.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Status:   entity.UserStatusActive,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	savedUser, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "testuser", savedUser.Username)
	assert.Equal(t, "test@example.com", savedUser.Email)
}

func TestUserRepository_Create_Error(t *testing.T) {
	repo, _ := setupUserRepo(t)
	ctx := context.Background()

	user1 := &entity.User{ID: "1", Username: "duplicate", Email: "dup@test.com"}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create another user with SAME Username (should fail due to unique constraint)
	user2 := &entity.User{ID: "2", Username: "duplicate", Email: "other@test.com"}
	err = repo.Create(ctx, user2)
	assert.Error(t, err)
}

func TestUserRepository_FindByUsername(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Username: "findme", Email: "findme@test.com"})

	t.Run("Found", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "findme")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "findme", user.Username)
	})

	t.Run("Not Found", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "unknown")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Email: "find@me.com", Username: "findme"})

	t.Run("Found", func(t *testing.T) {
		user, err := repo.FindByEmail(ctx, "find@me.com")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "find@me.com", user.Email)
	})

	t.Run("Not Found", func(t *testing.T) {
		user, err := repo.FindByEmail(ctx, "unknown@me.com")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_FindByToken(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Token: "valid-token", Username: "tokenuser", Email: "token@test.com"})

	t.Run("Found", func(t *testing.T) {
		user, err := repo.FindByToken(ctx, "valid-token")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "valid-token", user.Token)
	})

	t.Run("Not Found", func(t *testing.T) {
		user, err := repo.FindByToken(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_Update(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	user := &entity.User{ID: "1", Name: "Old Name", Username: "updateuser", Email: "update@test.com"}
	seedUser(t, db, user)

	user.Name = "New Name"
	err := repo.Update(ctx, user)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestUserRepository_Update_Error(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	// Create User 1
	seedUser(t, db, &entity.User{ID: "1", Username: "user1", Email: "user1@test.com"})

	// Create User 2
	seedUser(t, db, &entity.User{ID: "2", Username: "user2", Email: "user2@test.com"})

	// Try to update User 2 to have User 1's username (should fail)
	user2ToUpdate := &entity.User{ID: "2", Username: "user1", Email: "user2@test.com"}

	err := repo.Update(ctx, user2ToUpdate)
	assert.Error(t, err)
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Status: entity.UserStatusActive, Username: "statususer", Email: "status@test.com"})

	err := repo.UpdateStatus(ctx, "1", entity.UserStatusBanned)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, entity.UserStatusBanned, updated.Status)
}

func TestUserRepository_Delete(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Username: "deleteuser", Email: "delete@test.com"})

	err := repo.Delete(ctx, "1")
	require.NoError(t, err)

	// Check it's soft deleted
	_, err = repo.FindByID(ctx, "1")
	assert.Error(t, err) // Should be not found

	// Check DB directly for Unscoped
	var count int64
	err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = ?", "1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var deletedAt int64
	err = db.Get(&deletedAt, "SELECT deleted_at FROM users WHERE id = ?", "1")
	require.NoError(t, err)
	assert.True(t, deletedAt > 0, "DeletedAt should be > 0")
}

func TestUserRepository_FindAll(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, &entity.User{ID: "1", Username: "alpha", Email: "alpha@test.com", Name: "Alpha User"})
	seedUser(t, db, &entity.User{ID: "2", Username: "beta", Email: "beta@test.com", Name: "Beta User"})

	// Test 1: No Filter
	users, total, err := repo.FindAll(ctx, &model.GetUserListRequest{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)

	// Test 2: Filter by Username (LIKE)
	users, total, err = repo.FindAll(ctx, &model.GetUserListRequest{Username: "alp", Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "alpha", users[0].Username)

	// Test 3: Pagination
	users, total, err = repo.FindAll(ctx, &model.GetUserListRequest{Page: 1, Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total) // Total is still 2
	assert.Len(t, users, 1)          // But we got 1
}

func TestUserRepository_FindAllDynamic(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	users := []entity.User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Username: "alice"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Username: "bob"},
		{ID: "3", Name: "Charlie", Email: "charlie@example.com", Username: "charlie"},
	}
	for i := range users {
		seedUser(t, db, &users[i])
	}

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
			result, total, err := repo.FindAllDynamic(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
			assert.Equal(t, int64(tt.expectedCount), total)

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

func TestUserRepository_HardDeleteSoftDeletedUsers(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	// User 1: Deleted 31 days ago (Should be hard deleted if retention is 30)
	deletedOld := time.Now().Add(-31 * 24 * time.Hour).UnixMilli()
	db.MustExec(`INSERT INTO users (id, username, email, deleted_at) VALUES (?, ?, ?, ?)`, "1", "old_deleted", "old@test.com", deletedOld)

	// User 2: Deleted 10 days ago (Should stay)
	deletedRecent := time.Now().Add(-10 * 24 * time.Hour).UnixMilli()
	db.MustExec(`INSERT INTO users (id, username, email, deleted_at) VALUES (?, ?, ?, ?)`, "2", "recent_deleted", "recent@test.com", deletedRecent)

	// User 3: Active (Should stay)
	db.MustExec(`INSERT INTO users (id, username, email, deleted_at) VALUES (?, ?, ?, 0)`, "3", "active", "active@test.com")

	err := repo.HardDeleteSoftDeletedUsers(ctx, 30)
	require.NoError(t, err)

	// Verify User 1 is GONE (hard deleted)
	var count1 int64
	err = db.Get(&count1, "SELECT COUNT(*) FROM users WHERE id = ?", "1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count1)

	// Verify User 2 exists (soft deleted)
	var count2 int64
	err = db.Get(&count2, "SELECT COUNT(*) FROM users WHERE id = ?", "2")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count2)

	// Verify User 3 exists (active)
	var count3 int64
	err = db.Get(&count3, "SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at = 0", "3")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count3)
}

func TestUserRepository_Create_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	user := &entity.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
	}

	err = repo.Create(ctx, user)
	assert.Error(t, err)
}

func TestUserRepository_Update_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	user := &entity.User{ID: "1", Name: "Old Name", Username: "updateuser", Email: "update@test.com"}

	err := db.Close()
	require.NoError(t, err)

	user.Name = "New Name"
	err = repo.Update(ctx, user)
	assert.Error(t, err)
}

func TestUserRepository_UpdateStatus_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, "1", entity.UserStatusBanned)
	assert.Error(t, err)
}

func TestUserRepository_FindByID_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	user, err := repo.FindByID(ctx, "1")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_FindByEmail_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	user, err := repo.FindByEmail(ctx, "test@example.com")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_FindByToken_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	user, err := repo.FindByToken(ctx, "sometoken")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_Delete_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	err = repo.Delete(ctx, "1")
	assert.Error(t, err)
}

func TestUserRepository_FindByUsername_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	user, err := repo.FindByUsername(ctx, "someuser")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_getDB_TransactionContext(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	logrusLogger := logrus.New()
	logrusLogger.SetOutput(io.Discard)
	tm := tx.NewSQLXTransactionManager(db, logrusLogger)

	err := tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		user := &entity.User{
			ID:       "txuser",
			Username: "txuser",
			Email:    "tx@example.com",
		}
		return repo.Create(txCtx, user)
	})
	require.NoError(t, err)

	// Verify the user exists (since it commits directly in sqlite mode)
	saved, err := repo.FindByID(ctx, "txuser")
	require.NoError(t, err)
	assert.Equal(t, "txuser", saved.Username)
}

func TestUserRepository_FindAll_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	users, total, err := repo.FindAll(ctx, &model.GetUserListRequest{Page: 1, Limit: 10})
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, int64(0), total)
}

func TestUserRepository_FindAll_WithOrganizationContext(t *testing.T) {
	repo, db := setupUserRepo(t)

	seedUser(t, db, &entity.User{ID: "user1", Username: "user1", Email: "u1@test.com"})
	seedUser(t, db, &entity.User{ID: "user2", Username: "user2", Email: "u2@test.com"})
	seedUser(t, db, &entity.User{ID: "user3", Username: "user3", Email: "u3@test.com"})

	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org1', 'user1')`)
	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org1', 'user2')`)
	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org2', 'user3')`)

	ctx := context.Background()
	ctxWithOrg := database.SetOrganizationContext(ctx, "org1")

	users, total, err := repo.FindAll(ctxWithOrg, &model.GetUserListRequest{Page: 1, Limit: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(2), total)
	require.Len(t, users, 2)
	var usernames []string
	for _, u := range users {
		usernames = append(usernames, u.Username)
	}
	assert.ElementsMatch(t, []string{"user1", "user2"}, usernames)
}

func TestUserRepository_FindAllDynamic_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	users, total, err := repo.FindAllDynamic(ctx, &querybuilder.DynamicFilter{})
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, int64(0), total)
}

func TestUserRepository_FindAllDynamic_WithOrganizationContextAndSkipCount(t *testing.T) {
	repo, db := setupUserRepo(t)

	seedUser(t, db, &entity.User{ID: "user1", Username: "user1", Email: "u1@test.com"})
	seedUser(t, db, &entity.User{ID: "user2", Username: "user2", Email: "u2@test.com"})

	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org1', 'user1')`)
	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org2', 'user2')`)

	ctx := context.Background()
	ctxWithOrg := database.SetOrganizationContext(ctx, "org1")

	filter := &querybuilder.DynamicFilter{
		SkipCount: true,
		Page:      1,
		PageSize:  10,
	}

	users, total, err := repo.FindAllDynamic(ctxWithOrg, filter)
	require.NoError(t, err)

	assert.Equal(t, int64(-1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}

func TestUserRepository_HardDeleteSoftDeletedUsers_DBError(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	err := db.Close()
	require.NoError(t, err)

	err = repo.HardDeleteSoftDeletedUsers(ctx, 30)
	assert.Error(t, err)
}

func TestUserRepository_GetByOrganization(t *testing.T) {
	repo, db := setupUserRepo(t)

	seedUser(t, db, &entity.User{ID: "user1", Username: "user1", Email: "u1@test.com"})
	seedUser(t, db, &entity.User{ID: "user2", Username: "user2", Email: "u2@test.com"})

	db.MustExec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('org1', 'user1')`)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		users, err := repo.GetByOrganization(ctx, "org1")
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "user1", users[0].Username)
	})

	t.Run("Empty", func(t *testing.T) {
		users, err := repo.GetByOrganization(ctx, "org_nonexistent")
		require.NoError(t, err)
		assert.Len(t, users, 0)
	})

	t.Run("DBError", func(t *testing.T) {
		err := db.Close()
		require.NoError(t, err)

		users, err := repo.GetByOrganization(ctx, "org1")
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}

func TestUserRepository_FindBySSOIdentity(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	identity := &entity.UserSSOIdentity{
		ID:         "sso1",
		UserID:     "user1",
		Provider:   "google",
		ProviderID: "google-123",
	}
	require.NoError(t, repo.CreateSSOIdentity(ctx, identity))

	t.Run("Found", func(t *testing.T) {
		res, err := repo.FindBySSOIdentity(ctx, "google", "google-123")
		require.NoError(t, err)
		assert.Equal(t, "user1", res.UserID)
	})

	t.Run("NotFound", func(t *testing.T) {
		res, err := repo.FindBySSOIdentity(ctx, "google", "nonexistent")
		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, res)
	})

	t.Run("DBError", func(t *testing.T) {
		err := db.Close()
		require.NoError(t, err)

		res, err := repo.FindBySSOIdentity(ctx, "google", "google-123")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, res)
	})
}

func TestUserRepository_CreateSSOIdentity(t *testing.T) {
	repo, db := setupUserRepo(t)
	ctx := context.Background()

	identity := &entity.UserSSOIdentity{
		ID:         "sso1",
		UserID:     "user1",
		Provider:   "google",
		ProviderID: "google-123",
	}

	t.Run("Success", func(t *testing.T) {
		err := repo.CreateSSOIdentity(ctx, identity)
		require.NoError(t, err)

		var saved entity.UserSSOIdentity
		err = db.Get(&saved, "SELECT id, user_id, provider, provider_id, created_at, updated_at FROM user_sso_identities WHERE id = ?", "sso1")
		require.NoError(t, err)
		assert.Equal(t, "user1", saved.UserID)
	})

	t.Run("DBError", func(t *testing.T) {
		err := db.Close()
		require.NoError(t, err)

		err = repo.CreateSSOIdentity(ctx, &entity.UserSSOIdentity{ID: "sso2"})
		assert.Error(t, err)
	})
}
