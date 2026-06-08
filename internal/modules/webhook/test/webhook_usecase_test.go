package test

import (
	"context"
	"io"
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

type webhookTestDeps struct {
	Repo        *mocks.MockWebhookRepository
	Distributor *mocking.MockTaskDistributor
}

func setupWebhookUseCaseTest() (*webhookTestDeps, usecase.WebhookUseCase) {
	deps := &webhookTestDeps{
		Repo:        new(mocks.MockWebhookRepository),
		Distributor: new(mocking.MockTaskDistributor),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	validate := validator.New()
	uc := usecase.NewWebhookUseCase(deps.Repo, deps.Distributor, log, validate)
	return deps, uc
}

func TestWebhookUseCase_Create(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()

	req := model.CreateWebhookRequest{
		Name:           "Test Webhook",
		OrganizationID: "org-1",
		URL:            "https://example.com/webhook",
		Events:         []string{"user.created"},
		Secret:         "supersecret",
	}

	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
		return w.Name == req.Name && w.OrganizationID == req.OrganizationID
	})).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Trigger(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()

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

	deps.Repo.On("FindByEvent", mock.Anything, orgID, eventType).Return(webhooks, nil)
	deps.Distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.Anything).Return(nil)

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
	deps.Distributor.AssertExpectations(t)
}

func TestWebhookUseCase_Create_Negative(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()

	t.Run("Validation Error", func(t *testing.T) {
		req := model.CreateWebhookRequest{
			// Missing required fields
		}
		res, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("DB Error", func(t *testing.T) {
		req := model.CreateWebhookRequest{
			Name:           "Test Webhook",
			OrganizationID: "org-1",
			URL:            "https://example.com/webhook",
			Events:         []string{"user.created"},
			Secret:         "supersecret",
		}
		deps.Repo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()
		res, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestWebhookUseCase_Update(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()
	webhookID := "wh-1"
	orgID := "org-1"

	existingWebhook := &entity.Webhook{
		ID:             webhookID,
		Name:           "Old Name",
		OrganizationID: orgID,
		URL:            "https://old.com",
		Events:         `["user.created"]`,
		Secret:         "oldsecret",
		IsActive:       true,
	}

	t.Run("Positive - Selective update", func(t *testing.T) {
		newName := "New Name"
		newIsActive := false
		req := model.UpdateWebhookRequest{
			Name:     &newName,
			IsActive: &newIsActive,
		}

		deps.Repo.On("FindByID", mock.Anything, webhookID, orgID).Return(existingWebhook, nil).Once()
		deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
			return w.Name == newName && w.IsActive == newIsActive && w.URL == "https://old.com"
		})).Return(nil).Once()

		res, err := uc.Update(ctx, webhookID, orgID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, newName, res.Name)
		assert.False(t, res.IsActive)
	})

	t.Run("Validation Error", func(t *testing.T) {
		invalidURL := "not-a-url"
		req := model.UpdateWebhookRequest{
			URL: &invalidURL,
		}
		res, err := uc.Update(ctx, webhookID, orgID, req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Not Found", func(t *testing.T) {
		req := model.UpdateWebhookRequest{}
		deps.Repo.On("FindByID", mock.Anything, webhookID, orgID).Return(nil, assert.AnError).Once()
		res, err := uc.Update(ctx, webhookID, orgID, req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Update DB Error", func(t *testing.T) {
		req := model.UpdateWebhookRequest{}
		deps.Repo.On("FindByID", mock.Anything, webhookID, orgID).Return(existingWebhook, nil).Once()
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError).Once()
		res, err := uc.Update(ctx, webhookID, orgID, req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

    t.Run("Positive - All fields update", func(t *testing.T) {
		newName := "New Name 2"
        newUrl := "https://new.com"
        newEvents := []string{"user.deleted"}
        newSecret := "newsecret"
		newIsActive := true
		req := model.UpdateWebhookRequest{
			Name:     &newName,
            URL: &newUrl,
            Events: &newEvents,
            Secret: &newSecret,
			IsActive: &newIsActive,
		}

		deps.Repo.On("FindByID", mock.Anything, webhookID, orgID).Return(existingWebhook, nil).Once()
		deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
			return w.Name == newName && w.IsActive == newIsActive && w.URL == newUrl && w.Secret == newSecret && w.Events == "[\"user.deleted\"]"
		})).Return(nil).Once()

		res, err := uc.Update(ctx, webhookID, orgID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestWebhookUseCase_Delete(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()

	t.Run("Positive", func(t *testing.T) {
		deps.Repo.On("Delete", mock.Anything, "wh-1", "org-1").Return(nil).Once()
		err := uc.Delete(ctx, "wh-1", "org-1")
		assert.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		deps.Repo.On("Delete", mock.Anything, "wh-1", "org-1").Return(assert.AnError).Once()
		err := uc.Delete(ctx, "wh-1", "org-1")
		assert.Error(t, err)
	})
}

func TestWebhookUseCase_FindByID(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()

	webhook := &entity.Webhook{
		ID:             "wh-1",
		Events:         `["user.created"]`,
	}

	t.Run("Positive", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "wh-1", "org-1").Return(webhook, nil).Once()
		res, err := uc.FindByID(ctx, "wh-1", "org-1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "wh-1", res.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "wh-1", "org-1").Return(nil, assert.AnError).Once()
		res, err := uc.FindByID(ctx, "wh-1", "org-1")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestWebhookUseCase_FindByOrganizationID(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()

	webhooks := []entity.Webhook{
		{ID: "wh-1", Events: `["user.created"]`},
		{ID: "wh-2", Events: `[]`},
	}

	t.Run("Positive", func(t *testing.T) {
		deps.Repo.On("FindByOrganizationID", mock.Anything, "org-1").Return(webhooks, nil).Once()
		res, err := uc.FindByOrganizationID(ctx, "org-1")
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("DB Error", func(t *testing.T) {
		deps.Repo.On("FindByOrganizationID", mock.Anything, "org-1").Return([]entity.Webhook(nil), assert.AnError).Once()
		res, err := uc.FindByOrganizationID(ctx, "org-1")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestWebhookUseCase_Trigger_Negative(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()

	req := model.TriggerWebhookRequest{
		OrganizationID: "org-1",
		EventType:      "user.created",
	}

	t.Run("Repo Find Error", func(t *testing.T) {
		deps.Repo.On("FindByEvent", mock.Anything, "org-1", "user.created").Return([]entity.Webhook(nil), assert.AnError).Once()
		err := uc.Trigger(ctx, req)
		assert.Error(t, err)
	})

	t.Run("Distributor Error Logs But Continues", func(t *testing.T) {
		webhooks := []entity.Webhook{{ID: "wh-1", URL: "http://a.com"}}
		deps.Repo.On("FindByEvent", mock.Anything, "org-1", "user.created").Return(webhooks, nil).Once()
		deps.Distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		// Should not return error, just logs it
		err := uc.Trigger(ctx, req)
		assert.NoError(t, err)
	})
}

func TestWebhookUseCase_FindLogs(t *testing.T) {
	deps, uc := setupWebhookUseCaseTest()
	ctx := context.Background()

	t.Run("Positive", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "wh-1", "org-1").Return(&entity.Webhook{}, nil).Once()
		logs := []entity.WebhookLog{
			{ID: "log-1"}, {ID: "log-2"},
		}
		deps.Repo.On("FindLogsByWebhookID", mock.Anything, "wh-1", 10, 0).Return(logs, nil).Once()

		res, err := uc.FindLogs(ctx, "wh-1", "org-1", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("Webhook Not Found or Unauthorized", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "wh-1", "org-1").Return(nil, assert.AnError).Once()
		res, err := uc.FindLogs(ctx, "wh-1", "org-1", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Find Logs DB Error", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "wh-1", "org-1").Return(&entity.Webhook{}, nil).Once()
		deps.Repo.On("FindLogsByWebhookID", mock.Anything, "wh-1", 10, 0).Return([]entity.WebhookLog(nil), assert.AnError).Once()

		res, err := uc.FindLogs(ctx, "wh-1", "org-1", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
