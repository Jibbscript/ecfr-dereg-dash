package http

type AgencyDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TotalWords  int     `json:"total_words"`
	AvgRSCS     float64 `json:"avg_rscs"`
	LSACounts   int     `json:"lsa_counts"`
	LastUpdated string  `json:"last_updated"`
}
