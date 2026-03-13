package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/stretchr/testify/mock"
)

type MockWebhookUseCase struct {
	mock.Mock
}

func (m *MockWebhookUseCase) Create(ctx context.Context, req model.CreateWebhookRequest) (*model.WebhookResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebhookResponse), args.Error(1)
}

func (m *MockWebhookUseCase) Update(ctx context.Context, id string, organizationID string, req model.UpdateWebhookRequest) (*model.WebhookResponse, error) {
	args := m.Called(ctx, id, organizationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebhookResponse), args.Error(1)
}

func (m *MockWebhookUseCase) Delete(ctx context.Context, id string, organizationID string) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockWebhookUseCase) FindByID(ctx context.Context, id string, organizationID string) (*model.WebhookResponse, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebhookResponse), args.Error(1)
}

func (m *MockWebhookUseCase) FindByOrganizationID(ctx context.Context, organizationID string) ([]model.WebhookResponse, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]model.WebhookResponse), args.Error(1)
}

func (m *MockWebhookUseCase) Trigger(ctx context.Context, req model.TriggerWebhookRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockWebhookUseCase) FindLogs(ctx context.Context, webhookID string, organizationID string, limit int, offset int) ([]interface{}, error) {
	args := m.Called(ctx, webhookID, organizationID, limit, offset)
	return args.Get(0).([]interface{}), args.Error(1)
}
