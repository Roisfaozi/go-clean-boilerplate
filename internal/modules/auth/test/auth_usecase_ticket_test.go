package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/stretchr/testify/assert"
)

func TestAuthUseCase_GetTicket_Success(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "user-123"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"
	ticket := "generated-ticket-token"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusActive,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock TicketManager
	deps.ticketManager.On("CreateTicket", ctx, userID, orgID, sessionID, role, username).Return(ticket, nil)

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ticket, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertExpectations(t)
}

func TestAuthUseCase_GetTicket_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "non-existent-user"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to find user")
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertNotCalled(t, "CreateTicket")
}

func TestAuthUseCase_GetTicket_UserSuspended(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "suspended-user"
	orgID := "org-456"
	sessionID := "session-789"
	role := "user"
	username := "suspended"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusSuspended,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.ErrorIs(t, err, usecase.ErrAccountSuspended)
	assert.Empty(t, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertNotCalled(t, "CreateTicket")
}

func TestAuthUseCase_GetTicket_TicketManagerError(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "user-123"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusActive,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock TicketManager Error
	deps.ticketManager.On("CreateTicket", ctx, userID, orgID, sessionID, role, username).Return("", errors.New("redis error"))

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertExpectations(t)
}
