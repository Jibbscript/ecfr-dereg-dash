package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
	"go.uber.org/zap"
)

type Summaries struct {
	logger      *zap.Logger
	vertex      *vertexai.Client
	parquetRepo *parquet.Repo
	sqliteRepo  *sqlite.Repo

	mu        sync.Mutex
	cache     []domain.Summary
	cacheTime time.Time
}

func NewSummaries(logger *zap.Logger, vertex *vertexai.Client, parquet *parquet.Repo, sqliteRepo *sqlite.Repo) *Summaries {
	return &Summaries{logger: logger, vertex: vertex, parquetRepo: parquet, sqliteRepo: sqliteRepo}
}

// NewSummariesReadOnly creates a Summaries usecase for read-only operations (API).
// This constructor does not require Vertex AI client since GetAllSummaries only reads from SQLite.
func NewSummariesReadOnly(logger *zap.Logger, sqliteRepo *sqlite.Repo) *Summaries {
	return &Summaries{logger: logger, sqliteRepo: sqliteRepo}
}

func (u *Summaries) GetAllSummaries(ctx context.Context) ([]domain.Summary, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Cache for 1 hour
	if !u.cacheTime.IsZero() && time.Since(u.cacheTime) < 1*time.Hour {
		return u.cache, nil
	}

	// Fetch summaries from SQLite database
	summaries, err := u.sqliteRepo.GetAllSummaries()
	if err != nil {
		u.logger.Error("Failed to fetch summaries from SQLite", zap.Error(err))
		return nil, err
	}

	u.cache = summaries
	u.cacheTime = time.Now()

	u.logger.Info("Loaded summaries from SQLite", zap.Int("count", len(summaries)))
	return summaries, nil
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
