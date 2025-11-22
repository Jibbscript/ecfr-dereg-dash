package usecase

import (
	"context"
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
	if len(sections) == 0 {
		return nil, nil
	}

	u.logger.Info("Starting batch summary generation", zap.String("title", title.Title), zap.Int("count", len(sections)))

	// 1. Prepare prompts
	prompts := make([]string, len(sections))
	for i, s := range sections {
		text := s.Text
		if len(text) > 10000 {
			text = text[:10000]
		}
		prompts[i] = "Summarize this regulatory section: " + text
	}

	// 2. Call Batch API
	responses, err := u.vertex.BatchGenerateSummaries(ctx, prompts)
	if err != nil {
		u.logger.Error("Batch summary generation failed", zap.String("title", title.Title), zap.Error(err))
		return nil, err
	}

	// 3. Map responses to summaries
	var summaries []domain.Summary
	now := time.Now()

	// We assume responses correspond 1:1 to prompts in order.
	// If length mismatch, we process what we have but log warning.
	count := len(responses)
	if count != len(sections) {
		u.logger.Warn("Batch response count mismatch", 
			zap.Int("expected", len(sections)), 
			zap.Int("got", count))
		if count > len(sections) {
			count = len(sections)
		}
	}

	for i := 0; i < count; i++ {
		respText := responses[i]
		if respText == "" {
			u.logger.Debug("Empty summary for section", zap.String("section", sections[i].ID))
			continue
		}

		summaries = append(summaries, domain.Summary{
			Kind:      "section",
			Key:       sections[i].ID,
			Text:      respText,
			Model:     "gemini-2.5-flash",
			CreatedAt: now,
		})
	}

	return summaries, nil
}
