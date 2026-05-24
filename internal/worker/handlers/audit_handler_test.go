package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuditHandlerTest() (*mocks.MockAuditUseCase, *handlers.AuditTaskHandler) {
	uc := new(mocks.MockAuditUseCase)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := handlers.NewAuditTaskHandler(logger, uc)
	return uc, handler
}

func TestAuditTaskHandler_ProcessTaskAuditLog(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uc, handler := setupAuditHandlerTest()

		payload := auditModel.CreateAuditLogRequest{
			UserID: "user123",
			Action: "CREATE",
		}
		b, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeAuditLogCreate, b)

		uc.EXPECT().LogActivity(mock.Anything, payload).Return(nil)

		err := handler.ProcessTaskAuditLog(context.Background(), task)
		assert.NoError(t, err)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		_, handler := setupAuditHandlerTest()
		task := asynq.NewTask(tasks.TypeAuditLogCreate, []byte("invalid json"))

		err := handler.ProcessTaskAuditLog(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	t.Run("UseCase Error", func(t *testing.T) {
		uc, handler := setupAuditHandlerTest()

		payload := auditModel.CreateAuditLogRequest{
			UserID: "user123",
			Action: "CREATE",
		}
		b, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeAuditLogCreate, b)

		expectedErr := errors.New("db error")
		uc.EXPECT().LogActivity(mock.Anything, payload).Return(expectedErr)

		err := handler.ProcessTaskAuditLog(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to log audit activity")
		uc.AssertExpectations(t)
	})
}

func TestAuditTaskHandler_ProcessTaskAuditLogExport(t *testing.T) {
	// Clean up exports directory after tests
	defer func() { _ = os.RemoveAll("exports") }()

	t.Run("Success", func(t *testing.T) {
		uc, handler := setupAuditHandlerTest()

		payload := auditModel.AuditLogExportPayload{
			UserID:         "user123",
			OrganizationID: "org123",
			FromDate:       "2023-01-01",
			ToDate:         "2023-01-31",
		}
		b, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeAuditLogExport, b)

		uc.EXPECT().ExportLogs(mock.Anything, payload.FromDate, payload.ToDate, mock.AnythingOfType("func([]model.AuditLogResponse) error")).
			RunAndReturn(func(ctx context.Context, from, to string, process func([]auditModel.AuditLogResponse) error) error {
				return process([]auditModel.AuditLogResponse{
					{
						ID:        "log1",
						UserID:    "user123",
						Action:    "LOGIN",
						Entity:    "User",
						EntityID:  "user123",
						CreatedAt: 1672531200,
					},
				})
			})

		err := handler.ProcessTaskAuditLogExport(context.Background(), task)
		assert.NoError(t, err)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		_, handler := setupAuditHandlerTest()
		task := asynq.NewTask(tasks.TypeAuditLogExport, []byte("invalid json"))

		err := handler.ProcessTaskAuditLogExport(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	t.Run("UseCase Error", func(t *testing.T) {
		uc, handler := setupAuditHandlerTest()

		payload := auditModel.AuditLogExportPayload{
			UserID:   "user123",
			FromDate: "2023-01-01",
			ToDate:   "2023-01-31",
		}
		b, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeAuditLogExport, b)

		expectedErr := errors.New("export failed")
		uc.EXPECT().ExportLogs(mock.Anything, payload.FromDate, payload.ToDate, mock.AnythingOfType("func([]model.AuditLogResponse) error")).Return(expectedErr)

		err := handler.ProcessTaskAuditLogExport(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to export logs")
		uc.AssertExpectations(t)
	})

	t.Run("Process Callback Error", func(t *testing.T) {
		uc, handler := setupAuditHandlerTest()

		payload := auditModel.AuditLogExportPayload{
			UserID:   "user123",
		}
		b, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeAuditLogExport, b)

		expectedErr := errors.New("write error")
		uc.EXPECT().ExportLogs(mock.Anything, payload.FromDate, payload.ToDate, mock.AnythingOfType("func([]model.AuditLogResponse) error")).
			RunAndReturn(func(ctx context.Context, from, to string, process func([]auditModel.AuditLogResponse) error) error {
				return expectedErr
			})

		err := handler.ProcessTaskAuditLogExport(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to export logs: write error")
		uc.AssertExpectations(t)
	})
}
