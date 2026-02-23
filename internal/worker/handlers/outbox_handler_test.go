package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	auditEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOutboxTaskHandler_ProcessAuditOutbox_Robustness(t *testing.T) {
	logger := logrus.New() // Note: assuming logrus is available in package scope or I'll import it

	t.Run("Success Path - Moves entry to main log and deletes from outbox", func(t *testing.T) {
		mockRepo := new(mocks.MockAuditRepository)
		handler := NewOutboxTaskHandler(mockRepo, logger)
		ctx := context.Background()

		orgID := "org-1"
		entries := []*auditEntity.AuditOutbox{
			{
				ID:             "outbox-1",
				OrganizationID: &orgID,
				UserID:         "user-1",
				Action:         "UPDATE",
				Entity:         "Profile",
				EntityID:       "user-1",
				CreatedAt:      123456789,
			},
		}

		mockRepo.On("FindPendingOutbox", ctx, 50).Return(entries, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AuditLog")).Return(nil)
		mockRepo.On("DeleteOutbox", ctx, "outbox-1").Return(nil)

		err := handler.ProcessAuditOutbox(ctx, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure Path - Update status to failed when main log creation fails", func(t *testing.T) {
		mockRepo := new(mocks.MockAuditRepository)
		handler := NewOutboxTaskHandler(mockRepo, logger)
		ctx := context.Background()

		entries := []*auditEntity.AuditOutbox{
			{ID: "outbox-2", UserID: "user-1", Action: "LOGIN"},
		}

		dbErr := errors.New("database connection lost")

		mockRepo.On("FindPendingOutbox", ctx, 50).Return(entries, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(dbErr)
		// Crucial: Verify that UpdateOutboxStatus is called with 'failed' status and the error message
		mockRepo.On("UpdateOutboxStatus", ctx, "outbox-2", auditEntity.OutboxStatusFailed, mock.MatchedBy(func(err string) bool {
			return strings.Contains(err, dbErr.Error())
		})).Return(nil)

		err := handler.ProcessAuditOutbox(ctx, nil)

		// Handler should not return error to asynq (we handled it via failed status)
		// or it could return error if we want asynq to retry the whole batch. 
		// In our implementation, we continue to next entry, so we return nil.
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Outbox - Does nothing gracefully", func(t *testing.T) {
		mockRepo := new(mocks.MockAuditRepository)
		handler := NewOutboxTaskHandler(mockRepo, logger)
		ctx := context.Background()

		mockRepo.On("FindPendingOutbox", ctx, 50).Return([]*auditEntity.AuditOutbox{}, nil)

		err := handler.ProcessAuditOutbox(ctx, nil)

		assert.NoError(t, err)
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}
