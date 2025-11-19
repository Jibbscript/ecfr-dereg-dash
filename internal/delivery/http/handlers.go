package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xai/ecfr-dereg-dashboard/internal/usecase"
	"go.uber.org/zap"
)

type Usecases struct {
	Ingest    *usecase.Ingest
	Snapshot  *usecase.Snapshot
	Metrics   *usecase.Metrics
	Summaries *usecase.Summaries
}

func SetupHandlers(r chi.Router, usecases Usecases, logger *zap.Logger) {
	r.Get("/agencies", func(w http.ResponseWriter, req *http.Request) {
		_, err := usecases.Metrics.GetAgencyTotals()
		if err != nil {
			logger.Error("Get agencies failed", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		// Placeholder JSON with dummy data for E2E
		dummyAgencies := `[
		{"id": "1", "name": "Department of Agriculture", "slug": "usda", "metric": 100},
		{"id": "2", "name": "Department of Commerce", "slug": "doc", "metric": 200}
	]`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(dummyAgencies))
	})

	r.Get("/titles/{id}", func(w http.ResponseWriter, req *http.Request) {
		titleID := chi.URLParam(req, "id")
		// Dummy data for E2E testing
		dummyTitle := `{
			"id": "` + titleID + `",
			"title": "Title ` + titleID + `",
			"total_words": 50000,
			"avg_rscs": 15.5,
			"summary": "This is a summary for Title ` + titleID + `. It contains regulatory information about various topics."
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(dummyTitle))
	})

	r.Get("/sections/{id}", func(w http.ResponseWriter, req *http.Request) {
		sectionID := chi.URLParam(req, "id")
		// Dummy data for E2E testing
		dummySection := `{
			"id": "` + sectionID + `",
			"section": "§ ` + sectionID + `",
			"text": "This is the full text of section ` + sectionID + `. It contains detailed regulatory requirements and compliance information that would typically be much longer in a real application.",
			"rscs_per_1k": 12.3,
			"summary": "Summary of section ` + sectionID + ` discussing key regulatory requirements."
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(dummySection))
	})
}
