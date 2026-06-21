package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
)

type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.Webhook) error
	Update(ctx context.Context, webhook *entity.Webhook) error
	Delete(ctx context.Context, id string, organizationID string) error
	FindByID(ctx context.Context, id string, organizationID string) (*entity.Webhook, error)
	FindByOrganizationID(ctx context.Context, organizationID string) ([]entity.Webhook, error)
	FindByEvent(ctx context.Context, organizationID string, event string) ([]entity.Webhook, error)

	// Webhook Logs
	CreateLog(ctx context.Context, log *entity.WebhookLog) error
	FindLogsByWebhookID(ctx context.Context, webhookID string, limit int, offset int) ([]entity.WebhookLog, error)
}
