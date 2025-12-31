package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogActivity(t *testing.T) {
	mockRepo := new(mocks.MockAuditRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	uc := usecase.NewAuditUseCase(mockRepo, logger)

	t.Run("Success - Positive Case", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil // Ensure clean state
		req := model.CreateAuditLogRequest{
			UserID: "u1", Action: "CREATE", Entity: "User", EntityID: "u2",
			OldValues: map[string]string{"foo": "bar"},
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
			// Check if JSON marshaling worked
			return log.UserID == "u1" && log.Action == "CREATE" && log.OldValues != ""
		})).Return(nil)

		err := uc.LogActivity(context.Background(), req)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Edge - Nil JSON Values", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		req := model.CreateAuditLogRequest{
			UserID: "u1", Action: "DELETE", Entity: "User", EntityID: "u2",
			OldValues: nil, // Edge case: Nil value
			NewValues: nil,
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
			// json.Marshal(nil) returns "null" string
			return log.OldValues == "null" && log.NewValues == "null"
		})).Return(nil)

		err := uc.LogActivity(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("Negative - Repo Error", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		req := model.CreateAuditLogRequest{UserID: "u1"}
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := uc.LogActivity(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestGetLogsDynamic(t *testing.T) {
	mockRepo := new(mocks.MockAuditRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	uc := usecase.NewAuditUseCase(mockRepo, logger)

	t.Run("Success - Positive Case", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		now := time.Now().UnixMilli()
		entities := []*entity.AuditLog{
			{ID: "1", UserID: "u1", OldValues: `{"a":1}`, NewValues: `{"a":2}`, CreatedAt: now},
		}

		filter := &querybuilder.DynamicFilter{}
		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(entities, nil)

		res, err := uc.GetLogsDynamic(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "u1", res[0].UserID)

		// Verify JSON unmarshaling
		oldVal := res[0].OldValues.(map[string]interface{})
		assert.Equal(t, float64(1), oldVal["a"])
	})

	t.Run("Edge - Malformed JSON in DB", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		// Scenario where DB data is corrupted or not valid JSON
		entities := []*entity.AuditLog{
			{ID: "1", UserID: "u1", OldValues: `{broken_json`, NewValues: `null`},
		}

		filter := &querybuilder.DynamicFilter{}
		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(entities, nil)

		res, err := uc.GetLogsDynamic(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		// Should not panic, and OldValues should be nil/null because unmarshal failed
		assert.Nil(t, res[0].OldValues)
	})

	t.Run("Negative - Repo Error", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("FindAllDynamic", mock.Anything, mock.Anything).Return(nil, errors.New("db fail"))

		res, err := uc.GetLogsDynamic(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
