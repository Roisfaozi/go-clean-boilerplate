package test

import (
	"context"
	"errors"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupRepositoryTest(t *testing.T) (repository.WebhookRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	log := logrus.New()
	log.SetOutput(io.Discard)

	repo := repository.NewWebhookRepository(gormDB, log)
	return repo, mock
}

func TestWebhookRepository_Create_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	webhook := &entity.Webhook{
		ID:             "wh-1",
		Name:           "Test",
		OrganizationID: "org-1",
		URL:            "http://a.com",
		Events:         `["user.created"]`,
		Secret:         "secret",
		IsActive:       true,
		CreatedAt:      time.Now().UnixMilli(),
		UpdatedAt:      time.Now().UnixMilli(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhooks` (`id`,`name`,`organization_id`,`url`,`events`,`secret`,`is_active`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(webhook.ID, webhook.Name, webhook.OrganizationID, webhook.URL, webhook.Events, webhook.Secret, webhook.IsActive, webhook.CreatedAt, webhook.UpdatedAt, webhook.DeletedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), webhook)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Create_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)
	webhook := &entity.Webhook{ID: "wh-1", Name: "Test"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `webhooks`").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), webhook)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Update_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	webhook := &entity.Webhook{
		ID:             "wh-1",
		Name:           "Test",
		OrganizationID: "org-1",
		URL:            "http://a.com",
		Events:         `["user.created"]`,
		Secret:         "secret",
		IsActive:       true,
		CreatedAt:      time.Now().UnixMilli(),
		UpdatedAt:      time.Now().UnixMilli(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks` SET `name`=?,`organization_id`=?,`url`=?,`events`=?,`secret`=?,`is_active`=?,`created_at`=?,`updated_at`=?,`deleted_at`=? WHERE `webhooks`.`deleted_at` IS NULL AND `id` = ?")).
		WithArgs(webhook.Name, webhook.OrganizationID, webhook.URL, webhook.Events, webhook.Secret, webhook.IsActive, webhook.CreatedAt, sqlmock.AnyArg(), webhook.DeletedAt, webhook.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), webhook)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Update_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)
	webhook := &entity.Webhook{ID: "wh-1"}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `webhooks`").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), webhook)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Delete_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `webhooks` SET `deleted_at`=? WHERE (id = ? AND organization_id = ?) AND `webhooks`.`deleted_at` IS NULL")).
		WithArgs(sqlmock.AnyArg(), "wh-1", "org-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), "wh-1", "org-1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Delete_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `webhooks` SET `deleted_at`").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), "wh-1", "org-1")
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByID_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	rows := sqlmock.NewRows([]string{"id", "name", "organization_id"}).
		AddRow("wh-1", "Test Webhook", "org-1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (id = ? AND organization_id = ?) AND `webhooks`.`deleted_at` IS NULL ORDER BY `webhooks`.`id` LIMIT ?")).
		WithArgs("wh-1", "org-1", 1).
		WillReturnRows(rows)

	res, err := repo.FindByID(context.Background(), "wh-1", "org-1")
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "wh-1", res.ID)
		assert.Equal(t, "Test Webhook", res.Name)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByID_NotFound(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectQuery("SELECT \\* FROM `webhooks`").
		WillReturnError(gorm.ErrRecordNotFound)

	res, err := repo.FindByID(context.Background(), "wh-1", "org-1")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByOrganizationID_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("wh-1", "Webhook 1").
		AddRow("wh-2", "Webhook 2")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE organization_id = ? AND `webhooks`.`deleted_at` IS NULL")).
		WithArgs("org-1").
		WillReturnRows(rows)

	res, err := repo.FindByOrganizationID(context.Background(), "org-1")
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "wh-1", res[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByOrganizationID_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectQuery("SELECT \\* FROM `webhooks`").
		WillReturnError(errors.New("db error"))

	res, err := repo.FindByOrganizationID(context.Background(), "org-1")
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByEvent_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	rows := sqlmock.NewRows([]string{"id", "events"}).
		AddRow("wh-1", `["user.created"]`)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE (organization_id = ? AND is_active = ? AND JSON_CONTAINS(events, JSON_QUOTE(?))) AND `webhooks`.`deleted_at` IS NULL")).
		WithArgs("org-1", true, "user.created").
		WillReturnRows(rows)

	res, err := repo.FindByEvent(context.Background(), "org-1", "user.created")
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindByEvent_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectQuery("SELECT \\* FROM `webhooks`").
		WillReturnError(errors.New("db error"))

	res, err := repo.FindByEvent(context.Background(), "org-1", "user.created")
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_CreateLog_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	log := &entity.WebhookLog{
		ID:                 "log-1",
		WebhookID:          "wh-1",
		EventType:          "user.created",
		Payload:            "{}",
		ResponseStatusCode: 200,
		ResponseBody:       "ok",
		ExecutionTime:      100,
		CreatedAt:          time.Now().UnixMilli(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `webhook_logs` (`id`,`webhook_id`,`event_type`,`payload`,`response_status_code`,`response_body`,`execution_time`,`error_message`,`retry_count`,`created_at`) VALUES (?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(log.ID, log.WebhookID, log.EventType, log.Payload, log.ResponseStatusCode, log.ResponseBody, log.ExecutionTime, log.ErrorMessage, log.RetryCount, log.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreateLog(context.Background(), log)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_CreateLog_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)
	log := &entity.WebhookLog{ID: "log-1"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `webhook_logs`").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.CreateLog(context.Background(), log)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindLogsByWebhookID_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	rows := sqlmock.NewRows([]string{"id", "webhook_id"}).
		AddRow("log-1", "wh-1").
		AddRow("log-2", "wh-1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhook_logs` WHERE webhook_id = ? ORDER BY created_at DESC LIMIT ?")).
		WithArgs("wh-1", 10).
		WillReturnRows(rows)

	res, err := repo.FindLogsByWebhookID(context.Background(), "wh-1", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindLogsByWebhookID_Offset_Success(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	rows := sqlmock.NewRows([]string{"id", "webhook_id"}).
		AddRow("log-1", "wh-1").
		AddRow("log-2", "wh-1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `webhook_logs` WHERE webhook_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?")).
		WithArgs("wh-1", 10, 5).
		WillReturnRows(rows)

	res, err := repo.FindLogsByWebhookID(context.Background(), "wh-1", 10, 5)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_FindLogsByWebhookID_Error(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectQuery("SELECT \\* FROM `webhook_logs`").
		WillReturnError(errors.New("db error"))

	res, err := repo.FindLogsByWebhookID(context.Background(), "wh-1", 10, 0)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}
