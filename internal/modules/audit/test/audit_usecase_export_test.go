package test_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportLogs(t *testing.T) {
	t.Run("Success - Export Logs", func(t *testing.T) {
		deps, uc := setupAuditTest()
		ctx := context.Background()
		fromDate := "2023-01-01"
		toDate := "2023-01-31"

		logs := []*entity.AuditLog{
			{
				ID:        "log-1",
				UserID:    "user-1",
				Action:    "LOGIN",
				Entity:    "User",
				EntityID:  "u-1",
				OldValues: `{"status": "inactive"}`,
				NewValues: `{"status": "active"}`,
				CreatedAt: time.Now().UnixMilli(),
			},
		}

		// Calculate expected start/end times
		startT, _ := time.Parse("2006-01-02", fromDate)
		endT, _ := time.Parse("2006-01-02", toDate)
		startTime := startT.UnixMilli()
		endTime := endT.Add(24 * time.Hour).UnixMilli()

		deps.Repo.On("FindAllInBatches", ctx, startTime, endTime, 1000, mock.AnythingOfType("func([]*entity.AuditLog) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(4).(func([]*entity.AuditLog) error)
				_ = fn(logs)
			}).Return(nil)

		processFunc := func(exportedLogs []model.AuditLogResponse) error {
			assert.Len(t, exportedLogs, 1)
			assert.Equal(t, "log-1", exportedLogs[0].ID)
			// Verify JSON unmarshalling happened
			oldVal, ok := exportedLogs[0].OldValues.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "inactive", oldVal["status"])
			return nil
		}

		exportErr := uc.ExportLogs(ctx, fromDate, toDate, processFunc)
		assert.NoError(t, exportErr)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Invalid From Date", func(t *testing.T) {
		_, uc := setupAuditTest()
		ctx := context.Background()
		err := uc.ExportLogs(ctx, "invalid-date", "2023-01-31", func([]model.AuditLogResponse) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid from_date")
	})

	t.Run("Error - Invalid To Date", func(t *testing.T) {
		_, uc := setupAuditTest()
		ctx := context.Background()
		err := uc.ExportLogs(ctx, "2023-01-01", "invalid-date", func([]model.AuditLogResponse) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid to_date")
	})

	t.Run("Error - Repo Error", func(t *testing.T) {
		deps, uc := setupAuditTest()
		ctx := context.Background()
		deps.Repo.On("FindAllInBatches", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("db error"))

		err := uc.ExportLogs(ctx, "2023-01-01", "2023-01-31", func([]model.AuditLogResponse) error { return nil })
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})

	t.Run("Success - JSON Unmarshal Error (Should log warning but proceed)", func(t *testing.T) {
		deps, uc := setupAuditTest()
		ctx := context.Background()
		logs := []*entity.AuditLog{
			{
				ID:        "log-bad-json",
				UserID:    "user-1",
				Action:    "LOGIN",
				Entity:    "User",
				EntityID:  "u-1",
				OldValues: `{broken-json`,
				NewValues: `{"status": "active"}`,
			},
		}

		deps.Repo.On("FindAllInBatches", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(4).(func([]*entity.AuditLog) error)
				_ = fn(logs)
			}).Return(nil)

		processFunc := func(exportedLogs []model.AuditLogResponse) error {
			assert.Len(t, exportedLogs, 1)
			assert.Equal(t, "log-bad-json", exportedLogs[0].ID)
			assert.Nil(t, exportedLogs[0].OldValues) // Should be nil due to error
			return nil
		}

		err := uc.ExportLogs(ctx, "2023-01-01", "2023-01-31", processFunc)
		assert.NoError(t, err)
	})

	t.Run("Error - Process Function Fails", func(t *testing.T) {
		deps, uc := setupAuditTest()
		ctx := context.Background()
		logs := []*entity.AuditLog{{ID: "log-1"}}

		deps.Repo.On("FindAllInBatches", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(4).(func([]*entity.AuditLog) error)
				// The mock simulates the repo calling the callback which returns an error
				err := fn(logs)
				// Repo usually propagates this error
				assert.Error(t, err)
			}).Return(errors.New("process error")) // Simulating propagation

		processFunc := func(exportedLogs []model.AuditLogResponse) error {
			return errors.New("process error")
		}

		err := uc.ExportLogs(ctx, "2023-01-01", "2023-01-31", processFunc)
		assert.Error(t, err)
		assert.Equal(t, "process error", err.Error())
	})
}
