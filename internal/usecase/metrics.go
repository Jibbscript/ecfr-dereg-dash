package usecase

import (
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/duck"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Metrics struct {
	duck   *duck.Helper
	sqlite *sqlite.Repo
}

func NewMetrics(duck *duck.Helper, sqlite *sqlite.Repo) *Metrics {
	return &Metrics{duck: duck, sqlite: sqlite}
}

func (u *Metrics) GetAgencyTotals() ([]domain.AgencyMetric, error) {
	return u.sqlite.GetAgencyTotals()
}
