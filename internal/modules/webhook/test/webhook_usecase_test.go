package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
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
	}

	repo.On("FindByEvent", mock.Anything, orgID, eventType).Return(webhooks, nil)
	distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.Anything).Return(nil)

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	distributor.AssertExpectations(t)
}
