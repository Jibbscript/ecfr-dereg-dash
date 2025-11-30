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

func (u *Metrics) GetAgencyTotals(titleFilter *string) ([]domain.AgencyMetric, error) {
	return u.sqlite.GetAgencyTotals(titleFilter)
}

// GetAgencyChecksum returns the SHA256 hash of all section content for an agency
func (u *Metrics) GetAgencyChecksum(agencyID string) (string, error) {
	return u.sqlite.GetAgencyChecksum(agencyID)
}
