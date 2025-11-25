package usecase

import (
	"context"
	"fmt"
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
	// Requirement: Eliminate summary per section. Generate ONE summary per Title.
	// Configure batch processing so that one eCFR Title is processed per batch job.
	
	u.logger.Info("Starting title-level summary generation", zap.String("title", title.Title))

	// Construct a single prompt for the Title.
	// We leverage Google Search Grounding, so we can ask about the Title without passing all text.
	// However, passing some high-level context or the list of parts might be useful if available.
	// Since we have 'sections', we can extract unique Parts or just use the Title metadata.
	
	prompt := fmt.Sprintf("Generate a comprehensive summary for US Code of Federal Regulations Title %s: %s. "+
		"Include the agencies involved, the scope of regulations, and recent major changes. "+
		"Use Google Search to find the most recent context and details.", title.Title, title.Name)

	// Call Batch API with just one prompt
	responses, err := u.vertex.BatchGenerateSummaries(ctx, []string{prompt})
	if err != nil {
		u.logger.Error("Batch summary generation failed", zap.String("title", title.Title), zap.Error(err))
		return nil, err
	}

	if len(responses) == 0 {
		u.logger.Warn("No summary returned for title", zap.String("title", title.Title))
		return nil, nil
	}

	// Map the single response to a Summary object
	respText := responses[0]
	if respText == "" {
		u.logger.Warn("Empty summary content for title", zap.String("title", title.Title))
		return nil, nil
	}

	summary := domain.Summary{
		Kind:      "title",
		Key:       title.Title,
		Text:      respText,
		Model:     "gemini-2.5-pro", // Updated to valid model
		CreatedAt: time.Now(),
	}

	return []domain.Summary{summary}, nil
}
