package test

import (
	"context"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuditTest() (*mocks.MockAuditRepository, usecase.AuditUseCase) {
	mockRepo := new(mocks.MockAuditRepository)
	mockWS := new(mocks.MockWebSocketManager)
	logger := logrus.New()

	// Mock broadcast call if needed (optional for robustness tests unless specifically testing WS)
	mockWS.On("BroadcastToChannel", mock.Anything, mock.Anything).Return()

	return mockRepo, usecase.NewAuditUseCase(mockRepo, logger, mockWS)
}

// CircularRefStruct simulates a structure with circular references
type CircularRefStruct struct {
	Self *CircularRefStruct
}

func TestAuditUseCase_LogActivity_CircularRef(t *testing.T) {
	mockRepo, uc := setupAuditTest()

	// Create a circular reference
	circular := &CircularRefStruct{}
	circular.Self = circular

	req := model.CreateAuditLogRequest{
		UserID:    "user-1",
		Action:    "UPDATE",
		Entity:    "User",
		EntityID:  "u-1",
		OldValues: circular, // This causes json.Marshal to fail
		NewValues: map[string]interface{}{"valid": "data"},
	}

	// Expectation: The usecase ignores jsonmarshal errors and stores "null" or empty string?
	// The implementation uses: oldValJSON, _ := json.Marshal(req.OldValues)
	// If error, it returns nil slice, so string(nil) is "".

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
		// OldValues should be empty string (or "null" if changed, but implementation implies "")
		// Verify other fields
		return log.UserID == "user-1" && log.OldValues == "" && log.NewValues != ""
	})).Return(nil)

	err := uc.LogActivity(context.Background(), req)
	assert.NoError(t, err)
}

func TestAuditUseCase_LogActivity_LargePayload(t *testing.T) {
	mockRepo, uc := setupAuditTest()

	// 1.5MB String
	largeVal := strings.Repeat("a", 1500*1024)

	req := model.CreateAuditLogRequest{
		UserID:    "user-1",
		Action:    "UPLOAD",
		Entity:    "File",
		EntityID:  "f-1",
		OldValues: nil,
		NewValues: map[string]interface{}{"data": largeVal},
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
		return len(log.NewValues) > 1500*1024
	})).Return(nil)

	err := uc.LogActivity(context.Background(), req)
	assert.NoError(t, err)
}
