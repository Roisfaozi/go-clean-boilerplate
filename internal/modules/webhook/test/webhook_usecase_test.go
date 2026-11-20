package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhookUseCase_Create(t *testing.T) {
	repo := new(mocks.MockWebhookRepository)
	distributor := new(mocking.MockTaskDistributor)
	log := logrus.New()
	validate := validator.New()
	uc := usecase.NewWebhookUseCase(repo, distributor, log, validate)

	req := model.CreateWebhookRequest{
		Name:           "Test Webhook",
		OrganizationID: "org-1",
		URL:            "https://example.com/webhook",
		Events:         []string{"user.created"},
		Secret:         "supersecret",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
		return w.Name == req.Name && w.OrganizationID == req.OrganizationID
	})).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	repo.AssertExpectations(t)
}

func TestWebhookUseCase_Trigger(t *testing.T) {
	repo := new(mocks.MockWebhookRepository)
	distributor := new(mocking.MockTaskDistributor)
	log := logrus.New()
	validate := validator.New()
	uc := usecase.NewWebhookUseCase(repo, distributor, log, validate)

	orgID := "org-1"
	eventType := "user.created"
	payload := map[string]interface{}{"id": "user-1"}

	webhooks := []entity.Webhook{
		{
			ID:             "wh-1",
			Name:           "WH 1",
			URL:            "https://a.com",
			Secret:         "s1",
			Events:         `["user.created"]`,
			OrganizationID: orgID,
			IsActive:       true,
		},
		{
			ID:             "wh-2",
			Name:           "WH 2",
			URL:            "https://b.com",
			Secret:         "s2",
			Events:         `["user.created"]`,
			OrganizationID: orgID,
			IsActive:       true,
		},
	}

	var captured []tasks.WebhookTriggerPayload
	repo.On("FindByEvent", mock.Anything, orgID, eventType).Return(webhooks, nil)
	distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.MatchedBy(func(p tasks.WebhookTriggerPayload) bool {
		captured = append(captured, p)
		return p.EventID != ""
	})).Return(nil).Twice()

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	assert.NoError(t, err)
	assert.Len(t, captured, 2)
	assert.NotEmpty(t, captured[0].EventID)
	assert.Equal(t, captured[0].EventID, captured[1].EventID, "all webhooks triggered by one event must share EventID")
	repo.AssertExpectations(t)
	distributor.AssertExpectations(t)
}
