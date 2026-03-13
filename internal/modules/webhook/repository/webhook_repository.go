package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type webhookRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewWebhookRepository(db *gorm.DB, log *logrus.Logger) WebhookRepository {
	return &webhookRepository{
		db:  db,
		log: log,
	}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *entity.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

func (r *webhookRepository) Update(ctx context.Context, webhook *entity.Webhook) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id string, organizationID string) error {
	return r.db.WithContext(ctx).Where("id = ? AND organization_id = ?", id, organizationID).Delete(&entity.Webhook{}).Error
}

func (r *webhookRepository) FindByID(ctx context.Context, id string, organizationID string) (*entity.Webhook, error) {
	var webhook entity.Webhook
	err := r.db.WithContext(ctx).Where("id = ? AND organization_id = ?", id, organizationID).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepository) FindByOrganizationID(ctx context.Context, organizationID string) ([]entity.Webhook, error) {
	var webhooks []entity.Webhook
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&webhooks).Error
	return webhooks, err
}

func (r *webhookRepository) FindByEvent(ctx context.Context, organizationID string, event string) ([]entity.Webhook, error) {
	var webhooks []entity.Webhook
	// Using JSON_CONTAINS for MySQL if events is a JSON array string
	// Or simple LIKE if it's stored as a comma-separated string.
	// In the migration I used TEXT, and the comment says JSON array.
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND is_active = ? AND JSON_CONTAINS(events, JSON_QUOTE(?))", organizationID, true, event).
		Find(&webhooks).Error
	return webhooks, err
}

func (r *webhookRepository) CreateLog(ctx context.Context, log *entity.WebhookLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *webhookRepository) FindLogsByWebhookID(ctx context.Context, webhookID string, limit int, offset int) ([]entity.WebhookLog, error) {
	var logs []entity.WebhookLog
	err := r.db.WithContext(ctx).
		Where("webhook_id = ?", webhookID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}
