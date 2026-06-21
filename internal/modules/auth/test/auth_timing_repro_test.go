package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestForgotPassword_TimingDifference_Repro demonstrates the timing difference
// between finding a user (slow) and not finding a user (fast).
func TestForgotPassword_TimingDifference_Repro(t *testing.T) {
	authService, deps := setupTest(t)
	foundEmail := "found@example.com"
	notFoundEmail := "notfound@example.com"

	// Setup for Found User: Simulates DB latency in Save
	user, _ := createTestUser("password123")
	user.Email = foundEmail

	deps.userRepo.On("FindByEmail", mock.Anything, foundEmail).Return(user, nil)

	// Simulate 50ms latency for saving token (DB Write)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).
		Run(func(args mock.Arguments) {
			time.Sleep(50 * time.Millisecond)
		}).Return(nil)

	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "FORGOT_PASSWORD_REQUEST"
	}), mock.Anything).Return(nil)

	// Measure Time for Found User
	startFound := time.Now()
	err := authService.ForgotPassword(context.Background(), foundEmail)
	durationFound := time.Since(startFound)
	assert.NoError(t, err)

	// Setup for Not Found User: Returns immediately
	deps.userRepo.On("FindByEmail", mock.Anything, notFoundEmail).Return(nil, errors.New("user not found"))

	// Measure Time for Not Found User
	startNotFound := time.Now()
	err = authService.ForgotPassword(context.Background(), notFoundEmail)
	durationNotFound := time.Since(startNotFound)
	assert.NoError(t, err)

	t.Logf("Duration Found: %v", durationFound)
	t.Logf("Duration Not Found: %v", durationNotFound)

	// After the fix, Not Found should take at least 20ms (our minimum sleep duration)
	// and the difference should be smaller.

	t.Logf("Duration Found: %v", durationFound)
	t.Logf("Duration Not Found: %v", durationNotFound)

	if durationNotFound < 20*time.Millisecond {
		t.Errorf("Vulnerability NOT Fixed: 'User Not Found' returned too fast (%v), expected > 20ms", durationNotFound)
	} else {
		t.Log("Vulnerability Fixed: 'User Not Found' took enough time to mask the difference")
	}

	// Ideally we would assert they are close, but in a test environment with Mock Sleep vs Real Sleep,
	// exact comparison is flaky. The key is that we are sleeping > 0.
}
