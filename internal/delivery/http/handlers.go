package http

import (
	"encoding/json"
	"net/http"
	"sync"

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

		// Populate checksums if requested (use pre-computed if available, compute on-demand otherwise)
		if includeChecksum {
			// Find agencies without pre-computed checksums
			var needsChecksum []int
			for i := range agencies {
				if agencies[i].ContentChecksum == "" {
					needsChecksum = append(needsChecksum, i)
				}
			}

			// Compute missing checksums concurrently
			if len(needsChecksum) > 0 {
				var wg sync.WaitGroup
				var mu sync.Mutex
				sem := make(chan struct{}, 10) // Limit concurrency to 10

				for _, idx := range needsChecksum {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						sem <- struct{}{}        // Acquire
						defer func() { <-sem }() // Release

						checksum, err := usecases.Metrics.GetAgencyChecksum(agencies[i].ID)
						if err != nil {
							logger.Warn("Failed to get checksum for agency",
								zap.String("agency_id", agencies[i].ID),
								zap.Error(err))
							return
						}

						mu.Lock()
						agencies[i].ContentChecksum = checksum
						mu.Unlock()
					}(idx)
				}
				wg.Wait()
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

	r.Get("/summaries", func(w http.ResponseWriter, req *http.Request) {
		summaries, err := usecases.Summaries.GetAllSummaries(req.Context())
		if err != nil {
			logger.Error("Failed to get summaries", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summaries); err != nil {
			logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	})
}
