package test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserRepositoryTest(t *testing.T) (repository.UserRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	repo := repository.NewUserRepository(sqlxDB, logger)
	return repo, mock
}

func TestUserRepository_GetByOrganization(t *testing.T) {
	repo, mock := setupUserRepositoryTest(t)

	t.Run("Success", func(t *testing.T) {
		orgID := "org-1"
		rows := sqlmock.NewRows([]string{"id", "organization_id", "password", "email", "username", "name", "avatar_url", "token", "status", "email_verified_at", "created_at", "updated_at", "deleted_at"}).
			AddRow("user-1", "org-1", "pass", "user1@example.com", "user1", "User 1", "", "", "active", nil, 1000, 1000, 0).
			AddRow("user-2", "org-1", "pass", "user2@example.com", "user2", "User 2", "", "", "active", nil, 2000, 2000, 0)

		mock.ExpectQuery("SELECT (.+) FROM users (.+)").
			WithArgs(orgID).
			WillReturnRows(rows)

		users, err := repo.GetByOrganization(context.Background(), orgID)

		assert.NoError(t, err)
		if assert.Len(t, users, 2) {
			assert.Equal(t, "user-1", users[0].ID)
			assert.Equal(t, "user-2", users[1].ID)
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		orgID := "org-1"

		mock.ExpectQuery("SELECT (.+) FROM users (.+)").
			WithArgs(orgID).
			WillReturnError(assert.AnError)

		users, err := repo.GetByOrganization(context.Background(), orgID)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindBySSOIdentity(t *testing.T) {
	repo, mock := setupUserRepositoryTest(t)

	t.Run("Success", func(t *testing.T) {
		provider := "google"
		providerID := "12345"

		rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_id", "created_at", "updated_at"}).
			AddRow("sso-1", "user-1", provider, providerID, 1000, 1000)

		mock.ExpectQuery("SELECT (.+) FROM user_sso_identities WHERE provider = \\? AND provider_id = \\? LIMIT 1").
			WithArgs(provider, providerID).
			WillReturnRows(rows)

		identity, err := repo.FindBySSOIdentity(context.Background(), provider, providerID)

		assert.NoError(t, err)
		if assert.NotNil(t, identity) {
			assert.Equal(t, "sso-1", identity.ID)
			assert.Equal(t, "user-1", identity.UserID)
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		provider := "google"
		providerID := "12345"

		mock.ExpectQuery("SELECT (.+) FROM user_sso_identities WHERE provider = \\? AND provider_id = \\? LIMIT 1").
			WithArgs(provider, providerID).
			WillReturnError(exception.ErrNotFound)

		identity, err := repo.FindBySSOIdentity(context.Background(), provider, providerID)

		assert.Error(t, err)
		assert.Nil(t, identity)
		assert.Equal(t, exception.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_CreateSSOIdentity(t *testing.T) {
	repo, mock := setupUserRepositoryTest(t)

	t.Run("Success", func(t *testing.T) {
		identity := &entity.UserSSOIdentity{
			ID:         "sso-1",
			UserID:     "user-1",
			Provider:   "google",
			ProviderID: "12345",
		}

		mock.ExpectExec("INSERT INTO user_sso_identities").
			WithArgs(identity.ID, identity.UserID, identity.Provider, identity.ProviderID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateSSOIdentity(context.Background(), identity)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		identity := &entity.UserSSOIdentity{
			ID:         "sso-1",
			UserID:     "user-1",
			Provider:   "google",
			ProviderID: "12345",
		}

		mock.ExpectExec("INSERT INTO user_sso_identities").
			WithArgs(identity.ID, identity.UserID, identity.Provider, identity.ProviderID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(assert.AnError)

		err := repo.CreateSSOIdentity(context.Background(), identity)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
