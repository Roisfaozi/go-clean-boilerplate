package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupWebhookRepoTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, repository.WebhookRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a gorm connection", err)
	}

	log := logrus.New()
	repo := repository.NewWebhookRepository(gormDB, log)

	return gormDB, mock, repo
}

const webhookVisibilityClause = "((webhooks.organization_id IS NULL OR NOT EXISTS (SELECT 1 FROM organizations WHERE organizations.id = webhooks.organization_id AND organizations.deleted_at IS NOT NULL AND organizations.deleted_at <> 0)))"

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

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhooks`")).
			WithArgs(
				webhook.ID, webhook.Name, webhook.OrganizationID, webhook.URL,
				webhook.Events, webhook.Secret, webhook.IsActive,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), webhook)
		assert.NoError(t, err)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID: "wh-err",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhooks`")).
			WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), webhook)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
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

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks` SET `name`=?,`organization_id`=?,`url`=?,`events`=?,`secret`=?,`is_active`=?,`created_at`=?,`updated_at`=?,`deleted_at`=? WHERE `webhooks`.`deleted_at` IS NULL AND `id` = ?")).
			WithArgs(
				webhook.Name, webhook.OrganizationID, webhook.URL,
				webhook.Events, webhook.Secret, webhook.IsActive,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), webhook.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), webhook)
		assert.NoError(t, err)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		webhook := &entity.Webhook{
			ID: "wh-err",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks`")).
			WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), webhook)
		assert.Error(t, err)
	})
}

func TestWebhookRepository_Delete(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully deletes webhook", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks` SET `deleted_at`=? WHERE (id = ? AND organization_id = ?) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs(sqlmock.AnyArg(), "wh-123", "org-123").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "wh-123", "org-123")
		assert.NoError(t, err)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks` SET `deleted_at`=? WHERE (id = ? AND organization_id = ?) AND " + webhookVisibilityClause + " AND `webhooks`.`deleted_at` IS NULL")).
			WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "wh-123", "org-123")
		assert.Error(t, err)
	})
}

func TestWebhookRepository_FindByID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhook", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id", "url"}).
			AddRow("wh-123", "Test Webhook", "org-123", "https://test.com/hook")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (id = ? AND organization_id = ?) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL ORDER BY `webhooks`.`id` LIMIT ?")).
			WithArgs("wh-123", "org-123", 1).
			WillReturnRows(rows)

		webhook, err := repo.FindByID(context.Background(), "wh-123", "org-123")
		assert.NoError(t, err)
		assert.NotNil(t, webhook)
		if webhook != nil {
			assert.Equal(t, "wh-123", webhook.ID)
		}
	})

	t.Run("Negative - Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (id = ? AND organization_id = ?) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL ORDER BY `webhooks`.`id` LIMIT ?")).
			WithArgs("wh-123", "org-123", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		webhook, err := repo.FindByID(context.Background(), "wh-123", "org-123")
		assert.Error(t, err)
		assert.Nil(t, webhook)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

func TestWebhookRepository_FindByOrganizationID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhooks", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id"}).
			AddRow("wh-123", "Test 1", "org-123").
			AddRow("wh-124", "Test 2", "org-123")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE organization_id = ? AND " + webhookVisibilityClause + " AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs("org-123").
			WillReturnRows(rows)

		webhooks, err := repo.FindByOrganizationID(context.Background(), "org-123")
		assert.NoError(t, err)
		assert.Len(t, webhooks, 2)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE organization_id = ? AND " + webhookVisibilityClause + " AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs("org-123").
			WillReturnError(fmt.Errorf("db error"))

		webhooks, err := repo.FindByOrganizationID(context.Background(), "org-123")
		assert.Error(t, err)
		assert.Nil(t, webhooks)
	})
}

func TestWebhookRepository_FindByEvent(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds webhooks", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "organization_id"}).
			AddRow("wh-123", "Test 1", "org-123")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (organization_id = ? AND is_active = ? AND JSON_CONTAINS(events, JSON_QUOTE(?))) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs("org-123", true, "user.created").
			WillReturnRows(rows)

		webhooks, err := repo.FindByEvent(context.Background(), "org-123", "user.created")
		assert.NoError(t, err)
		assert.Len(t, webhooks, 1)
	})

	t.Run("Vulnerability - SQL Injection attempt via Event string", func(t *testing.T) {
		// Test that dangerous input strings are handled properly by GORM parameterization
		rows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (organization_id = ? AND is_active = ? AND JSON_CONTAINS(events, JSON_QUOTE(?))) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs("org-123", true, "'); DROP TABLE webhooks;--").
			WillReturnRows(rows)

		webhooks, err := repo.FindByEvent(context.Background(), "org-123", "'); DROP TABLE webhooks;--")
		assert.NoError(t, err)
		assert.Len(t, webhooks, 0)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (organization_id = ? AND is_active = ? AND JSON_CONTAINS(events, JSON_QUOTE(?))) AND "+webhookVisibilityClause+" AND `webhooks`.`deleted_at` IS NULL")).
			WithArgs("org-123", true, "user.created").
			WillReturnError(fmt.Errorf("db error"))

		webhooks, err := repo.FindByEvent(context.Background(), "org-123", "user.created")
		assert.Error(t, err)
		assert.Nil(t, webhooks)
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
			ExecutionTime:      100,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhook_logs`")).
			WithArgs(
				log.ID, log.WebhookID, log.EventType, log.Payload,
				log.ResponseStatusCode, log.ResponseBody, log.ExecutionTime,
				log.ErrorMessage, log.RetryCount, sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateLog(context.Background(), log)
		assert.NoError(t, err)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		log := &entity.WebhookLog{
			ID: "log-err",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhook_logs`")).
			WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		err := repo.CreateLog(context.Background(), log)
		assert.Error(t, err)
	})
}

func TestWebhookRepository_FindLogsByWebhookID(t *testing.T) {
	_, mock, repo := setupWebhookRepoTest(t)

	t.Run("Positive - Successfully finds logs", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "webhook_id"}).
			AddRow("log-1", "wh-123").
			AddRow("log-2", "wh-123")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhook_logs` WHERE webhook_id = ? ORDER BY created_at DESC LIMIT ?")).
			WithArgs("wh-123", 10). // Note offset not sent if 0 by gorm usually, but let's check exact args.
			WillReturnRows(rows)

		logs, err := repo.FindLogsByWebhookID(context.Background(), "wh-123", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
	})

	t.Run("Positive - With Offset", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "webhook_id"}).
			AddRow("log-1", "wh-123")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhook_logs` WHERE webhook_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?")).
			WithArgs("wh-123", 10, 5).
			WillReturnRows(rows)

		logs, err := repo.FindLogsByWebhookID(context.Background(), "wh-123", 10, 5)
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
	})

	t.Run("Negative - DB Error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhook_logs` WHERE webhook_id = ? ORDER BY created_at DESC LIMIT ?")).
			WithArgs("wh-123", 10).
			WillReturnError(fmt.Errorf("db error"))

		logs, err := repo.FindLogsByWebhookID(context.Background(), "wh-123", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, logs)
	})
}
