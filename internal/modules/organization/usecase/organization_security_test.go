// [NEW TEST FILE]
package usecase_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ===============================================
// Security Hardening Tests
// ===============================================

// TestCreateOrganization_XSS verifies that input is sanitized before persistence
func TestCreateOrganization_XSS(t *testing.T) {
	orgRepo, _, tm, enforcer, uc := setupOrganizationUseCase()
	ctx := context.Background()

	xssPayload := "<script>alert('xss')</script>"
	rawName := "Acme " + xssPayload
	expectedName := pkg.SanitizeString(rawName)

	request := &model.CreateOrganizationRequest{
		Name: rawName,
		Slug: "acme-xss",
	}
	userID := "user-123"

	// Mock successful creating
	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("SlugExists", ctx, "acme-xss").Return(false, nil)
	orgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
		return org.Name == expectedName // verifying sanitized persistence
	}), usecase.DefaultOwnerRoleID).Return(nil)

	enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	response, err := uc.CreateOrganization(ctx, userID, request)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, response.Name)
}

// TestUpdateOrganization_Settings verifies that arbitrary JSON settings can be persisted correctly.
func TestUpdateOrganization_Settings(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgID := "org-123"
	settings := map[string]interface{}{
		"theme":       "dark",
		"mfa_enabled": true,
		"max_users":   100,
	}

	existingOrg := &entity.Organization{
		ID:       orgID,
		Name:     "Acme Corp",
		Settings: nil,
	}

	request := &model.UpdateOrganizationRequest{
		Settings: settings,
	}

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
	orgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
		// Verify settings map matches
		return org.Settings["theme"] == "dark" && org.Settings["mfa_enabled"] == true
	})).Return(nil)

	response, err := uc.UpdateOrganization(ctx, orgID, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "dark", response.Settings["theme"])
}

// TestUpdateOrganization_ConcurrentModification Edge Case
// Verifies that updates process cleanly even if data unchanged
func TestUpdateOrganization_NoChange(t *testing.T) {
	orgRepo, _, tm, _, uc := setupOrganizationUseCase()
	ctx := context.Background()

	orgID := "org-123"
	existingOrg := &entity.Organization{
		ID:   orgID,
		Name: "Acme Corp",
	}

	request := &model.UpdateOrganizationRequest{} // Empty request

	tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(ctx)
	}).Return(nil)

	orgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
	orgRepo.On("Update", ctx, mock.Anything).Return(nil)

	response, err := uc.UpdateOrganization(ctx, orgID, request)

	assert.NoError(t, err)
	assert.Equal(t, "Acme Corp", response.Name)
}
