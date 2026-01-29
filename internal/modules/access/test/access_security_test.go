package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSecurityAccessTest() (*mocks.MockAccessRepository, usecase.IAccessUseCase) {
	mockRepo := new(mocks.MockAccessRepository)
	logger := logrus.New()

	uc := usecase.NewAccessUseCase(mockRepo, logger)
	return mockRepo, uc
}

// TestCreateEndpoint_DuplicateDetection tests that duplicate endpoints are handled correctly.
// Since the UseCase relies on Repository to enforce uniqueness, we verify it propagates errors.
func TestCreateEndpoint_DuplicateDetection(t *testing.T) {
	repo, uc := setupSecurityAccessTest()

	req := model.CreateEndpointRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	// Case 1: Duplicate entry (Repo returns conflict/unique constraint error)
	// Assuming logic relies on DB unique constraint
	expectedErr := exception.ErrConflict // or standard error depending on repo impl
	
	// Mock repo returning error
	repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
		// Verify fields if needed, simplified for error check
		return true 
	})).Return(expectedErr).Once()

	resp, err := uc.CreateEndpoint(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	// Case 2: Success
	repo.On("CreateEndpoint", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1)
		// Simulate ID assignment
		// In real struct, fields are set. Mock just returns nil
		_ = arg
	})

	// To make asserting response easier, we'd need to mock assignment or check logic.
	// For this test, we care about error propagation logic.
}

// TestLinkEndpointToAccessRight_Duplicate tests linking same endpoint twice.
func TestLinkEndpointToAccessRight_Duplicate(t *testing.T) {
	repo, uc := setupSecurityAccessTest()

	req := model.LinkEndpointRequest{
		AccessRightID: uuid.New().String(),
		EndpointID:    uuid.New().String(),
	}

	// Case: Duplicate link
	repo.On("LinkEndpointToAccessRight", mock.Anything, req.AccessRightID, req.EndpointID).
		Return(errors.New("duplicate entry")) // Simulate DB error

	err := uc.LinkEndpointToAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}
