package usecase

import (
	"context"
	"encoding/json"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type webhookUseCase struct {
	repo            repository.WebhookRepository
	taskDistributor worker.TaskDistributor
	log             *logrus.Logger
	validate        *validator.Validate
}

func NewWebhookUseCase(
	repo repository.WebhookRepository,
	taskDistributor worker.TaskDistributor,
	log *logrus.Logger,
	validate *validator.Validate,
) WebhookUseCase {
	return &webhookUseCase{
		repo:            repo,
		taskDistributor: taskDistributor,
		log:             log,
		validate:        validate,
	}
}

func (u *webhookUseCase) Create(ctx context.Context, req model.CreateWebhookRequest) (*model.WebhookResponse, error) {
	if err := u.validate.Struct(req); err != nil {
		return nil, err
	}

	eventsJSON, _ := json.Marshal(req.Events)

	webhook := &entity.Webhook{
		ID:             uuid.New().String(),
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
		URL:            req.URL,
		Events:         string(eventsJSON),
		Secret:         req.Secret,
		IsActive:       true,
	}

	if err := u.repo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return u.toResponse(webhook), nil
}

func (u *webhookUseCase) Update(ctx context.Context, id string, organizationID string, req model.UpdateWebhookRequest) (*model.WebhookResponse, error) {
	if err := u.validate.Struct(req); err != nil {
		return nil, err
	}

	webhook, err := u.repo.FindByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		eventsJSON, _ := json.Marshal(*req.Events)
		webhook.Events = string(eventsJSON)
	}
	if req.Secret != nil {
		webhook.Secret = *req.Secret
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return u.toResponse(webhook), nil
}

func (u *webhookUseCase) Delete(ctx context.Context, id string, organizationID string) error {
	return u.repo.Delete(ctx, id, organizationID)
}

func (u *webhookUseCase) FindByID(ctx context.Context, id string, organizationID string) (*model.WebhookResponse, error) {
	webhook, err := u.repo.FindByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}
	return u.toResponse(webhook), nil
}

func (u *webhookUseCase) FindByOrganizationID(ctx context.Context, organizationID string) ([]model.WebhookResponse, error) {
	webhooks, err := u.repo.FindByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.WebhookResponse, len(webhooks))
	for i, w := range webhooks {
		responses[i] = *u.toResponse(&w)
	}
	return responses, nil
}

func (u *webhookUseCase) Trigger(ctx context.Context, req model.TriggerWebhookRequest) error {
	webhooks, err := u.repo.FindByEvent(ctx, req.OrganizationID, req.EventType)
	if err != nil {
		return err
	}

	payloadJSON, _ := json.Marshal(req.Payload)
	payloadStr := string(payloadJSON)

	for _, w := range webhooks {
		taskPayload := tasks.WebhookTriggerPayload{
			WebhookID: w.ID,
			URL:       w.URL,
			Secret:    w.Secret,
			EventType: req.EventType,
			Payload:   payloadStr,
		}

		if err := u.taskDistributor.DistributeTaskWebhookTrigger(ctx, taskPayload); err != nil {
			u.log.WithError(err).Errorf("Failed to distribute webhook task for %s", w.ID)
		}
	}

	return nil
}

func (u *webhookUseCase) FindLogs(ctx context.Context, webhookID string, organizationID string, limit int, offset int) ([]interface{}, error) {
	// Verify ownership
	_, err := u.repo.FindByID(ctx, webhookID, organizationID)
	if err != nil {
		return nil, err
	}

	logs, err := u.repo.FindLogsByWebhookID(ctx, webhookID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(logs))
	for i, l := range logs {
		result[i] = l
	}
	return result, nil
}

func (u *webhookUseCase) toResponse(w *entity.Webhook) *model.WebhookResponse {
	var events []string
	_ = json.Unmarshal([]byte(w.Events), &events)

	return &model.WebhookResponse{
		ID:             w.ID,
		Name:           w.Name,
		OrganizationID: w.OrganizationID,
		URL:            w.URL,
		Events:         events,
		IsActive:       w.IsActive,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}
