package stats

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StatsModule struct {
	StatsController *http.StatsController
}

func NewStatsModule(db *gorm.DB, log *logrus.Logger) *StatsModule {
	uc := usecase.NewStatsUseCase(db, log)
	ctrl := http.NewStatsController(uc)

	return &StatsModule{
		StatsController: ctrl,
	}
}
