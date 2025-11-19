package usecase

import (
	"context"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Summaries struct {
	vertex      *vertexai.Client
	parquetRepo *parquet.Repo
}

func NewSummaries(vertex *vertexai.Client, parquet *parquet.Repo) *Summaries {
	return &Summaries{vertex: vertex, parquetRepo: parquet}
}

func (u *Summaries) GenerateForTitle(ctx context.Context, title domain.Title, sections []domain.Section) ([]domain.Summary, error) {
	summaries := []domain.Summary{}
	for _, s := range sections {
		text := s.Text
		if len(text) > 4096 {
			text = text[:4096]
		}
		prompt := "Summarize this regulatory section: " + text
		resp, err := u.vertex.GenerateSummary(ctx, prompt)
		if err != nil {
			continue
		}
		summaries = append(summaries, domain.Summary{
			Kind:      "section",
			Key:       s.ID,
			Text:      resp,
			Model:     "gemini-pro",
			CreatedAt: time.Now(),
		})
	}
	return summaries, nil
}
