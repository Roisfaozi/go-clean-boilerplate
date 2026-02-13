package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUseCase_GetTicket_Security_OrganizationMembership(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "user-123"
	targetOrgID := "org-target-456"
	sessionID := "session-789"
	role := "member"
	username := "testuser"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusActive,
	}

	// Mock UserRepo (User exists and is active)
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock FindUserOrganizations (User is NOT a member of targetOrgID)
	// Return empty list or list with other orgs
	deps.orgRepo.On("FindUserOrganizations", ctx, userID).Return([]*orgEntity.Organization{
		{ID: "other-org-789"},
	}, nil)

	// Execute GetTicket with targetOrgID
	result, err := authService.GetTicket(ctx, userID, targetOrgID, sessionID, role, username)

	// Assert Failure
	// We expect an error because the user is not in the organization.
	// NOTE: Before the fix, this test is expected to FAIL (i.e., GetTicket succeeds).
	assert.Error(t, err, "GetTicket should fail if user is not a member of the organization")
	if err != nil {
		assert.True(t, errors.Is(err, exception.ErrForbidden) || errors.Is(err, exception.ErrUnauthorized), "Error should be Forbidden or Unauthorized")
	}
	assert.Empty(t, result, "Ticket should not be generated")

	// Ensure CreateTicket was NOT called
	deps.ticketManager.AssertNotCalled(t, "CreateTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
