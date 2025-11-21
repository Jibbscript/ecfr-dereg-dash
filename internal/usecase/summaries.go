package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"go.uber.org/zap"
)

type Summaries struct {
	logger      *zap.Logger
	vertex      *vertexai.Client
	parquetRepo *parquet.Repo
}

func NewSummaries(logger *zap.Logger, vertex *vertexai.Client, parquet *parquet.Repo) *Summaries {
	return &Summaries{logger: logger, vertex: vertex, parquetRepo: parquet}
}

func (u *Summaries) GenerateForTitle(ctx context.Context, title domain.Title, sections []domain.Section) ([]domain.Summary, error) {
	var summaries []domain.Summary
	var mu sync.Mutex

	// Limit concurrency for LLM calls (e.g., 8 concurrent calls per title worker)
	// Since we have 4 concurrent titles, total concurrent calls = 4 * 8 = 32.
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup

	for _, s := range sections {
		wg.Add(1)
		sem <- struct{}{}

		go func(sec domain.Section) {
			defer wg.Done()
			defer func() { <-sem }()

			text := sec.Text
			if len(text) > 10000 {
				text = text[:10000]
			}
			prompt := "Summarize this regulatory section: " + text

			resp, err := u.vertex.GenerateSummary(ctx, prompt)
			if err != nil {
				u.logger.Debug("Summary generation failed", zap.String("section", sec.ID), zap.Error(err))
				return
			}

			mu.Lock()
			summaries = append(summaries, domain.Summary{
				Kind:      "section",
				Key:       sec.ID,
				Text:      resp,
				Model:     "gemini-3-pro-preview",
				CreatedAt: time.Now(),
			})
			mu.Unlock()
		}(s)
	}
	wg.Wait()

	return summaries, nil
}
