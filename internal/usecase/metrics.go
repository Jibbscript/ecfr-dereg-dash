package usecase

import (
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/duck"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Metrics struct {
	duck *duck.Helper
}

func NewMetrics(duck *duck.Helper) *Metrics {
	return &Metrics{duck: duck}
}

func (u *Metrics) GetAgencyTotals() ([]domain.Agency, error) {
	query := `SELECT agency_id, SUM(word_count) as total_words, AVG(rscs_per_1k) as avg_rscs FROM v_sections_latest GROUP BY agency_id`
	return u.duck.QueryAgencies(query)
}
