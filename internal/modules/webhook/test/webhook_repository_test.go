package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWebhookRepoTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, repository.WebhookRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	log := logrus.New()
	repo := repository.NewWebhookRepository(sqlxDB, log)

	return sqlxDB, mock, repo
}

func TestWebhookRepository_Create(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully creates webhook", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID:             "wh-123",
			Name:           "Test Webhook",
			OrganizationID: "org-123",
			URL:            "https://test.com/hook",
			Events:         "[\"user.created\"]",
			Secret:         "secret",
			IsActive:       true,
		}

		mock.ExpectExec("INSERT INTO webhooks").
			WithArgs(
				webhook.ID, webhook.Name, webhook.OrganizationID, webhook.URL,
				webhook.Events, webhook.Secret, webhook.IsActive,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), webhook)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID: "wh-err",
		}

		mock.ExpectExec("INSERT INTO webhooks").
			WillReturnError(fmt.Errorf("db error"))

		err := repo.Create(context.Background(), webhook)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_Update(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully updates webhook", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID:             "wh-123",
			Name:           "Updated Webhook",
			OrganizationID: "org-123",
			URL:            "https://test.com/hook",
			Events:         "[\"user.created\"]",
			Secret:         "secret",
			IsActive:       true,
		}

		mock.ExpectExec("UPDATE webhooks SET").
			WithArgs(
				webhook.Name, webhook.URL, webhook.Events, webhook.Secret,
				webhook.IsActive, sqlmock.AnyArg(), webhook.ID, webhook.OrganizationID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(context.Background(), webhook)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID: "wh-err",
		}

		mock.ExpectExec("UPDATE webhooks SET").
			WillReturnError(fmt.Errorf("db error"))

		err := repo.Update(context.Background(), webhook)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_Delete(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully deletes webhook", func(t *testing.T) {
		mock.ExpectExec("UPDATE webhooks SET deleted_at = \\? WHERE (.+)").
			WithArgs(sqlmock.AnyArg(), "wh-123", "org-123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), "wh-123", "org-123")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectExec("UPDATE webhooks SET deleted_at = \\? WHERE (.+)").
			WithArgs(sqlmock.AnyArg(), "wh-123", "org-123").
			WillReturnError(fmt.Errorf("db error"))

		err := repo.Delete(context.Background(), "wh-123", "org-123")
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_FindByID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhook", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id", "url", "events", "secret", "is_active", "created_at", "updated_at"}).
			AddRow("wh-123", "Test Webhook", "org-123", "https://test.com/hook", "[]", "secret", true, 1000, 1000)

		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("wh-123", "org-123").
			WillReturnRows(rows)

		res, err := repo.FindByID(context.Background(), "wh-123", "org-123")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "wh-123", res.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("wh-123", "org-123").
			WillReturnError(fmt.Errorf("db error"))

		res, err := repo.FindByID(context.Background(), "wh-123", "org-123")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_FindByOrganizationID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhooks", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id", "url", "events", "secret", "is_active", "created_at", "updated_at"}).
			AddRow("wh-1", "W1", "org-123", "https://1.com", "[]", "s1", true, 1000, 1000).
			AddRow("wh-2", "W2", "org-123", "https://2.com", "[]", "s2", true, 2000, 2000)

		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("org-123").
			WillReturnRows(rows)

		res, err := repo.FindByOrganizationID(context.Background(), "org-123")
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("org-123").
			WillReturnError(fmt.Errorf("db error"))

		res, err := repo.FindByOrganizationID(context.Background(), "org-123")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_FindByEvent(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhooks", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id", "url", "events", "secret", "is_active", "created_at", "updated_at"}).
			AddRow("wh-1", "W1", "org-123", "https://1.com", "[\"user.created\"]", "s1", true, 1000, 1000)

		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("org-123", "user.created").
			WillReturnRows(rows)

		res, err := repo.FindByEvent(context.Background(), "org-123", "user.created")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE (.+)").
			WithArgs("org-123", "user.created").
			WillReturnError(fmt.Errorf("db error"))

		res, err := repo.FindByEvent(context.Background(), "org-123", "user.created")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_CreateLog(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully creates log", func(t *testing.T) {
		log := &entity.WebhookLog{
			ID:                 "log-123",
			WebhookID:          "wh-123",
			EventType:          "user.created",
			Payload:            "{}",
			ResponseStatusCode: 200,
			ResponseBody:       "",
			ExecutionTime:      100,
		}

		mock.ExpectExec("INSERT INTO webhook_logs").
			WithArgs(
				log.ID, log.WebhookID, log.EventType, log.Payload,
				log.ResponseStatusCode, log.ResponseBody, log.ExecutionTime,
				log.ErrorMessage, log.RetryCount, sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateLog(context.Background(), log)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		log := &entity.WebhookLog{
			ID: "log-err",
		}

		mock.ExpectExec("INSERT INTO webhook_logs").
			WillReturnError(fmt.Errorf("db error"))

		err := repo.CreateLog(context.Background(), log)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookRepository_FindLogsByWebhookID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds logs", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status_code", "response_body", "execution_time", "error_message", "retry_count", "created_at"}).
			AddRow("log-1", "wh-123", "user.created", "{}", 200, "", 100, "", 0, 1000).
			AddRow("log-2", "wh-123", "user.updated", "{}", 200, "", 150, "", 0, 2000)

		mock.ExpectQuery("SELECT (.+) FROM webhook_logs WHERE webhook_id = \\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs("wh-123", 10, 0).
			WillReturnRows(rows)

		logs, err := repo.FindLogsByWebhookID(context.Background(), "wh-123", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM webhook_logs WHERE webhook_id = \\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs("wh-123", 10, 0).
			WillReturnError(fmt.Errorf("db error"))

		logs, err := repo.FindLogsByWebhookID(context.Background(), "wh-123", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, logs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
