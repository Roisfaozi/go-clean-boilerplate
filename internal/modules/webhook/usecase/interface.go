package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
)

type WebhookUseCase interface {
	Create(ctx context.Context, req model.CreateWebhookRequest) (*model.WebhookResponse, error)
	Update(ctx context.Context, id string, organizationID string, req model.UpdateWebhookRequest) (*model.WebhookResponse, error)
	Delete(ctx context.Context, id string, organizationID string) error
	FindByID(ctx context.Context, id string, organizationID string) (*model.WebhookResponse, error)
	FindByOrganizationID(ctx context.Context, organizationID string) ([]model.WebhookResponse, error)
	Trigger(ctx context.Context, req model.TriggerWebhookRequest) error

	// Logs
	FindLogs(ctx context.Context, webhookID string, organizationID string, limit int, offset int) ([]interface{}, error)
}
