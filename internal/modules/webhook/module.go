package webhook

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WebhookModule struct {
	Repo       repository.WebhookRepository
	UseCase    usecase.WebhookUseCase
	Controller *http.WebhookController
}

func NewWebhookModule(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	taskDistributor worker.TaskDistributor,
) *WebhookModule {
	repo := repository.NewWebhookRepository(db, log)
	uc := usecase.NewWebhookUseCase(repo, taskDistributor, log, validate)
	controller := http.NewWebhookController(uc)

	return &WebhookModule{
		Repo:       repo,
		UseCase:    uc,
		Controller: controller,
	}
}
