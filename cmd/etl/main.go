package main

import (
	"context"
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

	logger.Info("Config loaded", zap.String("DataDir", config.DataDir))

	ctx := context.Background()

	parquetRepo := parquet.NewRepo(config.DataDir)
	sqlitePath := config.DataDir + "/ecfr.db"
	logger.Info("Initializing SQLite", zap.String("path", sqlitePath))
	sqliteRepo := sqlite.NewRepo(sqlitePath)

	ecfrClient := ecfr.NewClient()
	govinfoClient := govinfo.NewClient(config.DataDir)

	vertexClient, err := vertexai.NewClient(ctx, config.VertexProjectID, config.VertexLocation, config.VertexModelID)
	if err != nil {
		logger.Warn("Failed to create Vertex client, using mock", zap.Error(err))
		vertexClient = vertexai.NewMockClient()
	}

	lsaCollector := lsa.NewCollector(ecfrClient, vertexClient)

	ingestUseCase := usecase.NewIngest(govinfoClient, lsaCollector, parquetRepo, sqliteRepo)
	summariesUseCase := usecase.NewSummaries(vertexClient, parquetRepo)

	snapshotDate := time.Now().Format("2006-01-02")

	// Step 1: Fetch title catalog, mark changed
	changedTitles, err := ingestUseCase.FetchChangedTitles(ctx)
	if err != nil {
		logger.Fatal("Failed to fetch changed titles", zap.Error(err))
	}

	for _, title := range changedTitles {
		// Step 2: Pull sections, compute metrics (incl RSCS)
		sections, err := ingestUseCase.IngestTitle(ctx, title)
		if err != nil {
			logger.Error("Ingest failed for title", zap.String("title", title.Title), zap.Error(err))
			continue
		}

		// Write to parquet and sqlite
		err = parquetRepo.WriteSections(snapshotDate, title.Title, sections)
		if err != nil {
			logger.Error("Parquet write failed", zap.Error(err))
		}
		err = sqliteRepo.InsertSections(sections)
		if err != nil {
			logger.Error("SQLite insert failed", zap.Error(err))
		}

		// Step 3: Compute deltas
		diffs, err := usecase.NewSnapshot(parquetRepo, sqliteRepo).ComputeDiffs(snapshotDate, title.Title)
		if err != nil {
			logger.Error("Diff compute failed", zap.Error(err))
		}
		err = parquetRepo.WriteDiffs(snapshotDate, title.Title, diffs)
		if err != nil {
			logger.Error("Diff write failed", zap.Error(err))
		}

		// Step 4: LSA metric
		lsaCounts, err := lsaCollector.CollectForTitle(ctx, title.Title, title.LatestAmendedOn)
		if err != nil {
			logger.Error("LSA collect failed", zap.Error(err))
		}
		err = parquetRepo.WriteLSA(snapshotDate, title.Title, lsaCounts)
		if err != nil {
			logger.Error("LSA write failed", zap.Error(err))
		}

		// Step 5: Summaries
		summaries, err := summariesUseCase.GenerateForTitle(ctx, title, sections)
		if err != nil {
			logger.Error("Summaries generate failed", zap.Error(err))
		}
		err = parquetRepo.WriteSummaries(snapshotDate, summaries)
		if err != nil {
			logger.Error("Summaries write failed", zap.Error(err))
		}
	}

	logger.Info("ETL completed", zap.String("snapshot", snapshotDate))
}
