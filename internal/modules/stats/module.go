package stats

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/sirupsen/logrus"
)

type StatsModule struct {
	UseCase         usecase.StatsUseCase
	StatsController *http.StatsController
}

func NewStatsModule(db any, log *logrus.Logger) *StatsModule {
	uc := usecase.NewStatsUseCase(db, log)
	ctrl := http.NewStatsController(uc)

	return &StatsModule{
		UseCase:         uc,
		StatsController: ctrl,
	}
}
