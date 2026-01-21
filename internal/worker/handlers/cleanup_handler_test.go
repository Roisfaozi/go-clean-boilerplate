package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/test/mocks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type cleanupTestDeps struct {
	AuthRepo  *authMocks.MockTokenRepository
	UserRepo  *userMocks.MockUserRepository
	AuditRepo *mocks.MockAuditRepository
	Handler   *handlers.CleanupTaskHandler
}

func setupCleanupHandlerTest() *cleanupTestDeps {
	authRepo := new(authMocks.MockTokenRepository)
	userRepo := new(userMocks.MockUserRepository)
	auditRepo := new(mocks.MockAuditRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	handler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, logger)

	return &cleanupTestDeps{
		AuthRepo:  authRepo,
		UserRepo:  userRepo,
		AuditRepo: auditRepo,
		Handler:   handler,
	}
}

func TestProcessCleanupExpiredTokens(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		task := asynq.NewTask(tasks.TypeCleanupExpiredTokens, nil)

		deps.AuthRepo.On("DeleteExpiredResetTokens", mock.Anything).Return(nil)

		err := deps.Handler.ProcessCleanupExpiredTokens(context.Background(), task)
		assert.NoError(t, err)
		deps.AuthRepo.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		task := asynq.NewTask(tasks.TypeCleanupExpiredTokens, nil)

		deps.AuthRepo.On("DeleteExpiredResetTokens", mock.Anything).Return(errors.New("db error"))

		err := deps.Handler.ProcessCleanupExpiredTokens(context.Background(), task)
		assert.Error(t, err)
		deps.AuthRepo.AssertExpectations(t)
	})
}

func TestProcessCleanupSoftDeletedEntities(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		payload := tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, jsonPayload)

		deps.UserRepo.On("HardDeleteSoftDeletedUsers", mock.Anything, 30).Return(nil)

		err := deps.Handler.ProcessCleanupSoftDeletedEntities(context.Background(), task)
		assert.NoError(t, err)
		deps.UserRepo.AssertExpectations(t)
	})

	t.Run("Unmarshal Error", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		task := asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, []byte("invalid json"))

		err := deps.Handler.ProcessCleanupSoftDeletedEntities(context.Background(), task)
		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		payload := tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, jsonPayload)

		deps.UserRepo.On("HardDeleteSoftDeletedUsers", mock.Anything, 30).Return(errors.New("db error"))

		err := deps.Handler.ProcessCleanupSoftDeletedEntities(context.Background(), task)
		assert.Error(t, err)
		deps.UserRepo.AssertExpectations(t)
	})
}

func TestProcessPruneAuditLogs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		payload := tasks.PruneAuditLogsPayload{RetentionDays: 180}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypePruneAuditLogs, jsonPayload)

		// We can't easily match the exact timestamp, so we use mock.MatchedBy or just mock.AnythingOfType("int64")
		// Ideally we mock time.Now but that requires refactoring. For now assume it works if we check logic around it.
		// Actually, we can check if it is roughly correct.

		deps.AuditRepo.On("DeleteLogsOlderThan", mock.Anything, mock.AnythingOfType("int64")).Return(nil)

		err := deps.Handler.ProcessPruneAuditLogs(context.Background(), task)
		assert.NoError(t, err)
		deps.AuditRepo.AssertExpectations(t)
	})

	t.Run("Unmarshal Error", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		task := asynq.NewTask(tasks.TypePruneAuditLogs, []byte("invalid json"))

		err := deps.Handler.ProcessPruneAuditLogs(context.Background(), task)
		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		deps := setupCleanupHandlerTest()
		payload := tasks.PruneAuditLogsPayload{RetentionDays: 180}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypePruneAuditLogs, jsonPayload)

		deps.AuditRepo.On("DeleteLogsOlderThan", mock.Anything, mock.AnythingOfType("int64")).Return(errors.New("db error"))

		err := deps.Handler.ProcessPruneAuditLogs(context.Background(), task)
		assert.Error(t, err)
		deps.AuditRepo.AssertExpectations(t)
	})
}
