package http

import (
	"encoding/json"
	"net/http"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/usecase"
	"github.com/go-chi/chi/v5"
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
		// Parse optional title filter from query params
		titleFilter := req.URL.Query().Get("title")
		var tf *string
		if titleFilter != "" {
			tf = &titleFilter
		}

		// Check if checksums are requested
		includeChecksum := req.URL.Query().Get("include_checksum") == "true"

		agencies, err := usecases.Metrics.GetAgencyTotals(tf)
		if err != nil {
			logger.Error("Get agencies failed", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Populate checksums if requested
		if includeChecksum {
			for i := range agencies {
				checksum, err := usecases.Metrics.GetAgencyChecksum(agencies[i].ID)
				if err != nil {
					logger.Warn("Failed to get checksum for agency",
						zap.String("agency_id", agencies[i].ID),
						zap.Error(err))
					continue
				}
				agencies[i].ContentChecksum = checksum
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(agencies); err != nil {
			logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
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
