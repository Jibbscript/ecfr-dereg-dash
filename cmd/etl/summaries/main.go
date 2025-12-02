package main

import (
	"context"
	"flag"
	"runtime"
	"sync"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/ecfr"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/platform"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/usecase"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

	titleFilter := flag.String("title", "", "Generate summary for a specific title (e.g., '1', '42'). If empty, processes all titles.")
	snapshotDate := flag.String("snapshot-date", time.Now().Format("2006-01-02"), "Snapshot date for output (YYYY-MM-DD)")
	maxConcurrency := flag.Int("concurrency", 2, "Maximum concurrent title processing (recommended: 2-4 for API rate limits)")
	flag.Parse()

	logger.Info("Starting Summary Generation Batch Job",
		zap.String("env", config.Env),
		zap.String("data_dir", config.DataDir),
		zap.String("title_filter", *titleFilter),
		zap.String("snapshot_date", *snapshotDate),
		zap.Int("max_concurrency", *maxConcurrency),
		zap.Int("gomaxprocs", runtime.GOMAXPROCS(0)))
	jobStart := time.Now()

	ctx := context.Background()

	var parquetRepo *parquet.Repo
	var err error

	if config.Env == "local" || config.Env == "dev" {
		parquetRepo, err = parquet.NewLocalRepo(config.DataDir, config.ParquetPrefix)
		if err != nil {
			logger.Fatal("Failed to create local Parquet repo", zap.Error(err))
		}
		logger.Info("Using local Parquet repo",
			zap.String("dir", config.DataDir),
			zap.String("prefix", config.ParquetPrefix),
		)
	} else {
		parquetRepo, err = parquet.NewRepo(ctx, config.ParquetBucket, config.ParquetPrefix)
		if err != nil {
			logger.Fatal("Failed to create Parquet repo", zap.Error(err))
		}
	}

	vertexClient, err := vertexai.NewClient(ctx, config.VertexProjectID, config.VertexLocation, config.VertexModelID, config.GCSBucket)
	if err != nil {
		logger.Fatal("Failed to create Vertex AI client", zap.Error(err))
	}

	ecfrClient := ecfr.NewClient()
	summariesUseCase := usecase.NewSummaries(logger, vertexClient, parquetRepo)

	logger.Info("Fetching title catalog from eCFR API")
	titles, err := ecfrClient.GetTitles()
	if err != nil {
		logger.Fatal("Failed to fetch title catalog", zap.Error(err))
	}

	if *titleFilter != "" {
		var filtered []domain.Title
		for _, t := range titles {
			if t.Title == *titleFilter {
				filtered = append(filtered, t)
				break
			}
		}
		if len(filtered) == 0 {
			logger.Fatal("Title not found in catalog", zap.String("title", *titleFilter))
		}
		titles = filtered
	}

	logger.Info("Processing titles for summary generation", zap.Int("count", len(titles)))

	sem := make(chan struct{}, *maxConcurrency)
	var wg sync.WaitGroup
	var summariesMutex sync.Mutex
	allSummaries := []domain.Summary{}

	for _, title := range titles {
		wg.Add(1)
		sem <- struct{}{}

		go func(t domain.Title) {
			defer wg.Done()
			defer func() { <-sem }()

			titleStart := time.Now()
			logger.Info("Generating summary for title", zap.String("title", t.Title), zap.String("name", t.Name))

			summaries, err := summariesUseCase.GenerateForTitle(ctx, t, nil)
			if err != nil {
				logger.Error("Summary generation failed", zap.String("title", t.Title), zap.Error(err))
				return
			}

			if len(summaries) > 0 {
				summariesMutex.Lock()
				allSummaries = append(allSummaries, summaries...)
				summariesMutex.Unlock()
				logger.Info("Summary generated", zap.String("title", t.Title), zap.Duration("duration", time.Since(titleStart)))
			}
		}(title)
	}

	wg.Wait()

	if len(allSummaries) > 0 {
		logger.Info("Writing summaries to Parquet", zap.Int("count", len(allSummaries)))
		if err := parquetRepo.WriteSummaries(ctx, *snapshotDate, allSummaries); err != nil {
			logger.Fatal("Failed to write summaries", zap.Error(err))
		}
	} else {
		logger.Warn("No summaries generated")
	}

	logger.Info("Summary Generation Batch Job Completed",
		zap.Int("summaries_generated", len(allSummaries)),
		zap.Duration("total_duration", time.Since(jobStart)))
}
