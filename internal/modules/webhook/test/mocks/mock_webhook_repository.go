package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/stretchr/testify/mock"
)

type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) Create(ctx context.Context, webhook *entity.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookRepository) Update(ctx context.Context, webhook *entity.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookRepository) Delete(ctx context.Context, id string, organizationID string) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockWebhookRepository) FindByID(ctx context.Context, id string, organizationID string) (*entity.Webhook, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) FindByOrganizationID(ctx context.Context, organizationID string) ([]entity.Webhook, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]entity.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) FindByEvent(ctx context.Context, organizationID string, event string) ([]entity.Webhook, error) {
	args := m.Called(ctx, organizationID, event)
	return args.Get(0).([]entity.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) CreateLog(ctx context.Context, log *entity.WebhookLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockWebhookRepository) FindLogsByWebhookID(ctx context.Context, webhookID string, limit int, offset int) ([]entity.WebhookLog, error) {
	args := m.Called(ctx, webhookID, limit, offset)
	return args.Get(0).([]entity.WebhookLog), args.Error(1)
}
