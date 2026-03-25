package test

import (
	"context"
	"errors"
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

func setupWebhookTest() (*webhookTestDeps, usecase.WebhookUseCase) {
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

func TestWebhookUseCase_Create_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

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

func TestWebhookUseCase_Create_ValidationError(t *testing.T) {
	_, uc := setupWebhookTest()

	req := model.CreateWebhookRequest{
		Name:           "", // invalid name
		OrganizationID: "org-1",
		URL:            "https://example.com/webhook",
		Events:         []string{"user.created"},
		Secret:         "supersecret",
	}

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestWebhookUseCase_Create_RepoError(t *testing.T) {
	deps, uc := setupWebhookTest()

	req := model.CreateWebhookRequest{
		Name:           "Test Webhook",
		OrganizationID: "org-1",
		URL:            "https://example.com/webhook",
		Events:         []string{"user.created"},
		Secret:         "supersecret",
	}

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func ptrStr(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrArrStr(a []string) *[]string {
	return &a
}

func TestWebhookUseCase_Update_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	req := model.UpdateWebhookRequest{
		Name:     ptrStr("Updated Webhook"),
		URL:      ptrStr("https://new.com/webhook"),
		Events:   ptrArrStr([]string{"user.updated"}),
		Secret:   ptrStr("newsecret"),
		IsActive: ptrBool(false),
	}

	existing := &entity.Webhook{
		ID:             id,
		Name:           "Old Webhook",
		OrganizationID: orgID,
		URL:            "https://old.com/webhook",
		Events:         `["user.created"]`,
		Secret:         "oldsecret",
		IsActive:       true,
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
		return w.Name == *req.Name && w.URL == *req.URL && w.Secret == *req.Secret && w.IsActive == *req.IsActive
	})).Return(nil)

	res, err := uc.Update(context.Background(), id, orgID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, *req.Name, res.Name)
	assert.Equal(t, *req.URL, res.URL)
	assert.Equal(t, *req.IsActive, res.IsActive)
	assert.Equal(t, []string{"user.updated"}, res.Events)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Update_PartialFields(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	req := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook Only"),
	}

	existing := &entity.Webhook{
		ID:             id,
		Name:           "Old Webhook",
		OrganizationID: orgID,
		URL:            "https://old.com/webhook",
		Events:         `["user.created"]`,
		Secret:         "oldsecret",
		IsActive:       true,
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(w *entity.Webhook) bool {
		return w.Name == *req.Name && w.URL == "https://old.com/webhook"
	})).Return(nil)

	res, err := uc.Update(context.Background(), id, orgID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, *req.Name, res.Name)
	assert.Equal(t, existing.URL, res.URL) // unchanged
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Update_ValidationError(t *testing.T) {
	_, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	req := model.UpdateWebhookRequest{
		URL: ptrStr("not-a-url"),
	}

	res, err := uc.Update(context.Background(), id, orgID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestWebhookUseCase_Update_FindByIDError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	req := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook"),
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(nil, errors.New("not found"))

	res, err := uc.Update(context.Background(), id, orgID, req)

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Update_RepoError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	req := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook"),
	}

	existing := &entity.Webhook{
		ID:             id,
		Name:           "Old Webhook",
		OrganizationID: orgID,
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)
	deps.Repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	res, err := uc.Update(context.Background(), id, orgID, req)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Delete_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"

	deps.Repo.On("Delete", mock.Anything, id, orgID).Return(nil)

	err := uc.Delete(context.Background(), id, orgID)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Delete_RepoError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"

	deps.Repo.On("Delete", mock.Anything, id, orgID).Return(errors.New("db error"))

	err := uc.Delete(context.Background(), id, orgID)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindByID_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"

	existing := &entity.Webhook{
		ID:             id,
		Name:           "Webhook 1",
		OrganizationID: orgID,
		Events:         `["user.created"]`,
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)

	res, err := uc.FindByID(context.Background(), id, orgID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
	assert.Equal(t, existing.Name, res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindByID_RepoError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(nil, errors.New("not found"))

	res, err := uc.FindByID(context.Background(), id, orgID)

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindByOrganizationID_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

	orgID := "org-1"

	webhooks := []entity.Webhook{
		{ID: "wh-1", Name: "Webhook 1", OrganizationID: orgID, Events: `["event1"]`},
		{ID: "wh-2", Name: "Webhook 2", OrganizationID: orgID, Events: `["event2"]`},
	}

	deps.Repo.On("FindByOrganizationID", mock.Anything, orgID).Return(webhooks, nil)

	res, err := uc.FindByOrganizationID(context.Background(), orgID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)
	assert.Equal(t, webhooks[0].ID, res[0].ID)
	assert.Equal(t, webhooks[1].ID, res[1].ID)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindByOrganizationID_RepoError(t *testing.T) {
	deps, uc := setupWebhookTest()

	orgID := "org-1"

	var webhooks []entity.Webhook
	deps.Repo.On("FindByOrganizationID", mock.Anything, orgID).Return(webhooks, errors.New("db error"))

	res, err := uc.FindByOrganizationID(context.Background(), orgID)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_Trigger_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

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

	deps.Repo.On("FindByEvent", mock.Anything, orgID, eventType).Return(webhooks, nil)
	deps.Distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.MatchedBy(func(p interface{}) bool {
		return true // matches all calls
	})).Return(nil).Twice() // Expect two calls

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
	deps.Distributor.AssertExpectations(t)
}

func TestWebhookUseCase_Trigger_DistributionErrorContinues(t *testing.T) {
	deps, uc := setupWebhookTest()

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
	deps.Distributor.On("DistributeTaskWebhookTrigger", mock.Anything, mock.Anything).Return(errors.New("distribution error"))

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	// Does not return error if distribution fails
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
	deps.Distributor.AssertExpectations(t)
}

func TestWebhookUseCase_Trigger_FindByEventError(t *testing.T) {
	deps, uc := setupWebhookTest()

	orgID := "org-1"
	eventType := "user.created"
	payload := map[string]interface{}{"id": "user-1"}

	var webhooks []entity.Webhook
	deps.Repo.On("FindByEvent", mock.Anything, orgID, eventType).Return(webhooks, errors.New("db error"))

	err := uc.Trigger(context.Background(), model.TriggerWebhookRequest{
		OrganizationID: orgID,
		EventType:      eventType,
		Payload:        payload,
	})

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindLogs_Success(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	limit := 10
	offset := 0

	existing := &entity.Webhook{
		ID:             id,
		OrganizationID: orgID,
	}

	logs := []entity.WebhookLog{
		{ID: "log-1", WebhookID: id, ResponseStatusCode: 200},
		{ID: "log-2", WebhookID: id, ResponseStatusCode: 500},
	}

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)
	deps.Repo.On("FindLogsByWebhookID", mock.Anything, id, limit, offset).Return(logs, nil)

	res, err := uc.FindLogs(context.Background(), id, orgID, limit, offset)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindLogs_FindByIDError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	limit := 10
	offset := 0

	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(nil, errors.New("unauthorized"))

	res, err := uc.FindLogs(context.Background(), id, orgID, limit, offset)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestWebhookUseCase_FindLogs_FindLogsByWebhookIDError(t *testing.T) {
	deps, uc := setupWebhookTest()

	id := "wh-1"
	orgID := "org-1"
	limit := 10
	offset := 0

	existing := &entity.Webhook{
		ID:             id,
		OrganizationID: orgID,
	}

	var logs []entity.WebhookLog
	deps.Repo.On("FindByID", mock.Anything, id, orgID).Return(existing, nil)
	deps.Repo.On("FindLogsByWebhookID", mock.Anything, id, limit, offset).Return(logs, errors.New("db error"))

	res, err := uc.FindLogs(context.Background(), id, orgID, limit, offset)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}
