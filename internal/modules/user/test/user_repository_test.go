package test

import (
	"context"
	"testing"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupUserRepositoryTest(t *testing.T) (repository.UserRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	logger := logrus.New()
	repo := repository.NewUserRepository(gormDB, logger)
	return repo, mock
}

func TestUserRepository_GetByOrganization(t *testing.T) {
	repo, mock := setupUserRepositoryTest(t)

	t.Run("Success", func(t *testing.T) {
		orgID := "org-1"
		rows := sqlmock.NewRows([]string{"id", "email", "username"}).
			AddRow("user-1", "user1@example.com", "user1").
			AddRow("user-2", "user2@example.com", "user2")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE users.id IN (SELECT user_id FROM "organization_members" WHERE organization_id = $1) AND "users"."deleted_at" = $2`)).
			WithArgs(orgID, 0).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE users.id IN (SELECT user_id FROM "organization_members" WHERE organization_id = $1) AND "users"."deleted_at" = $2`)).
			WithArgs(orgID, 0).
			WillReturnError(gorm.ErrInvalidDB)

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

		rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_id"}).
			AddRow("sso-1", "user-1", provider, providerID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sso_identities" WHERE provider = $1 AND provider_id = $2 ORDER BY "user_sso_identities"."id" LIMIT $3`)).
			WithArgs(provider, providerID, 1).
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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sso_identities" WHERE provider = $1 AND provider_id = $2 ORDER BY "user_sso_identities"."id" LIMIT $3`)).
			WithArgs(provider, providerID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		identity, err := repo.FindBySSOIdentity(context.Background(), provider, providerID)

		assert.Error(t, err)
		assert.Nil(t, identity)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
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

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "user_sso_identities" ("id","user_id","provider","provider_id","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6)`)).
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

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "user_sso_identities"`)).
			WithArgs(identity.ID, identity.UserID, identity.Provider, identity.ProviderID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(gorm.ErrInvalidDB)

		err := repo.CreateSSOIdentity(context.Background(), identity)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
