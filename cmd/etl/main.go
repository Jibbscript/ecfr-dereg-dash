package main

import (
	"context"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/ecfr"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/xai/ecfr-dereg-dashboard/internal/platform"
	"github.com/xai/ecfr-dereg-dashboard/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

	logger.Info("Starting ETL Pipeline",
		zap.String("env", config.Env),
		zap.String("data_dir", config.DataDir))
	pipelineStart := time.Now()

	ctx := context.Background()

	parquetRepo := parquet.NewRepo(config.DataDir)
	sqlitePath := config.DataDir + "/ecfr.db"
	logger.Info("Initializing storage adapters", zap.String("sqlite_path", sqlitePath))
	sqliteRepo := sqlite.NewRepo(sqlitePath)

	ecfrClient := ecfr.NewClient()
	govinfoClient := govinfo.NewClient(config.DataDir)

	vertexClient, err := vertexai.NewClient(ctx, config.VertexProjectID, config.VertexLocation, config.VertexModelID)
	if err != nil {
		logger.Warn("Failed to create Vertex client, using mock", zap.Error(err))
		vertexClient = vertexai.NewMockClient()
	}

	lsaCollector := lsa.NewCollector(ecfrClient, vertexClient)

	ingestUseCase := usecase.NewIngest(logger, govinfoClient, lsaCollector, parquetRepo, sqliteRepo)
	summariesUseCase := usecase.NewSummaries(vertexClient, parquetRepo)

	snapshotDate := time.Now().Format("2006-01-02")

	// Step 1: Fetch title catalog
	logger.Info("Step 1/5: Fetching changed titles (Extract)")
	extractStart := time.Now()
	changedTitles, err := ingestUseCase.FetchChangedTitles(ctx)
	if err != nil {
		logger.Fatal("Failed to fetch changed titles", zap.Error(err))
	}
	logger.Info("Fetched changed titles",
		zap.Int("count", len(changedTitles)),
		zap.Duration("duration", time.Since(extractStart)))

	totalTitles := len(changedTitles)
	for i, title := range changedTitles {
		titleStart := time.Now()
		currentFileNum := i + 1

		// User-friendly progress message
		logger.Info(fmt.Sprintf("Processing file %d of %d: %s", currentFileNum, totalTitles, title.Title),
			zap.String("stage", "Processing"),
			zap.Int("current", currentFileNum),
			zap.Int("total", totalTitles),
			zap.String("title", title.Title))

		// Step 2: Pull sections (Extract)
		logger.Info("  > Extracting sections", zap.String("title", title.Title))
		opStart := time.Now()
		sections, err := ingestUseCase.IngestTitle(ctx, title)
		if err != nil {
			logger.Error("Ingest failed for title", zap.String("title", title.Title), zap.Error(err))
			continue
		}
		logger.Info("  > Extraction complete",
			zap.Int("sections_count", len(sections)),
			zap.Duration("duration", time.Since(opStart)))

		// Write to parquet and sqlite (Load)
		logger.Info("  > Loading sections to storage")
		opStart = time.Now()
		err = parquetRepo.WriteSections(snapshotDate, title.Title, sections)
		if err != nil {
			logger.Error("Parquet write failed", zap.Error(err))
		}
		err = sqliteRepo.InsertSections(sections)
		if err != nil {
			logger.Error("SQLite insert failed", zap.Error(err))
		}
		logger.Info("  > Sections loaded", zap.Duration("duration", time.Since(opStart)))

		// Step 3: Compute deltas (Transform)
		logger.Info("  > Computing diffs (Transform)")
		opStart = time.Now()
		diffs, err := usecase.NewSnapshot(parquetRepo, sqliteRepo).ComputeDiffs(snapshotDate, title.Title)
		if err != nil {
			logger.Error("Diff compute failed", zap.Error(err))
		}
		logger.Info("  > Diffs computed",
			zap.Int("diffs_count", len(diffs)),
			zap.Duration("duration", time.Since(opStart)))

		// Load Diffs
		err = parquetRepo.WriteDiffs(snapshotDate, title.Title, diffs)
		if err != nil {
			logger.Error("Diff write failed", zap.Error(err))
		}

		// Step 4: LSA metric (Transform)
		logger.Info("  > Collecting LSA metrics (Transform)")
		opStart = time.Now()
		lsaCounts, err := lsaCollector.CollectForTitle(ctx, title.Title, title.LatestAmendedOn)
		if err != nil {
			logger.Error("LSA collect failed", zap.Error(err))
		}
		logger.Info("  > LSA metrics collected",
			zap.Int("proposals", lsaCounts.ProposalsCount),
			zap.Int("amendments", lsaCounts.AmendmentsCount),
			zap.Int("finals", lsaCounts.FinalsCount),
			zap.Duration("duration", time.Since(opStart)))

		// Load LSA
		err = parquetRepo.WriteLSA(snapshotDate, title.Title, lsaCounts)
		if err != nil {
			logger.Error("LSA write failed", zap.Error(err))
		}

		// Step 5: Summaries (Transform)
		logger.Info("  > Generating summaries (Transform)")
		opStart = time.Now()
		summaries, err := summariesUseCase.GenerateForTitle(ctx, title, sections)
		if err != nil {
			logger.Error("Summaries generate failed", zap.Error(err))
		}
		logger.Info("  > Summaries generated",
			zap.Int("summaries_count", len(summaries)),
			zap.Duration("duration", time.Since(opStart)))

		// Load Summaries
		err = parquetRepo.WriteSummaries(snapshotDate, summaries)
		if err != nil {
			logger.Error("Summaries write failed", zap.Error(err))
		}

		logger.Info("Completed processing title",
			zap.String("title", title.Title),
			zap.Duration("duration", time.Since(titleStart)))
	}

	logger.Info("ETL Pipeline Completed Successfully",
		zap.String("snapshot", snapshotDate),
		zap.Duration("total_duration", time.Since(pipelineStart)))
}
