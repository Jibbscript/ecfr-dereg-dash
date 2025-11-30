package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/ecfr"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"github.com/xai/ecfr-dereg-dashboard/internal/platform"
	"github.com/xai/ecfr-dereg-dashboard/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

	// Parse command line flags
	skipSummary := flag.Bool("skip-summary", false, "Skip the summary generation step")
	flag.Parse()

	logger.Info("Starting ETL Pipeline (Optimized)",
		zap.String("env", config.Env),
		zap.String("data_dir", config.DataDir),
		zap.Int("gomaxprocs", runtime.GOMAXPROCS(0)),
		zap.Bool("skip_summary", *skipSummary))
	pipelineStart := time.Now()

	ctx := context.Background()

	var parquetRepo *parquet.Repo
	var err error

	// GCS backend (prod/staging)
	if config.Env == "local" || config.Env == "dev" {
		// Local filesystem backend
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

	sqlitePath := config.DataDir + "/ecfr.db"
	logger.Info("Initializing storage adapters", zap.String("sqlite_path", sqlitePath))
	sqliteRepo, err := sqlite.NewRepo(sqlitePath)
	if err != nil {
		logger.Fatal("Failed to initialize SQLite repo", zap.String("path", sqlitePath), zap.Error(err))
	}

	// Ingest agency data from JSON (refreshes every ETL run)
	agenciesPath := "ecfr_agencies.json"
	logger.Info("Ingesting agency data", zap.String("path", agenciesPath))
	if err := sqliteRepo.IngestAgencies(agenciesPath); err != nil {
		logger.Warn("Agency ingestion failed (continuing without agency mapping)", zap.Error(err))
	} else {
		logger.Info("Agency data ingested successfully")
	}

	ecfrClient := ecfr.NewClient()
	
	govinfoClient, err := govinfo.NewClient(ctx, config.RawXMLBucket, config.RawXMLPrefix)
	if err != nil {
		if config.Env == "local" {
			logger.Warn("Failed to create GovInfo client (skipping ETL steps that require it)", zap.Error(err))
			logger.Info("ETL Agency Ingestion Completed (Local Mode)")
			return
		}
		logger.Fatal("Failed to create GovInfo client", zap.Error(err))
	}

	vertexClient, err := vertexai.NewClient(ctx, config.VertexProjectID, config.VertexLocation, config.VertexModelID, config.GCSBucket)
	if err != nil {
		logger.Warn("Failed to create Vertex client, using mock", zap.Error(err))
		vertexClient = vertexai.NewMockClient()
	}


	lsaCollector := lsa.NewCollector(ecfrClient, vertexClient)

	ingestUseCase := usecase.NewIngest(logger, govinfoClient, lsaCollector, parquetRepo, sqliteRepo)
	summariesUseCase := usecase.NewSummaries(logger, vertexClient, parquetRepo)
	snapshotUseCase := usecase.NewSnapshot(parquetRepo, sqliteRepo)

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

	// --- OPTIMIZATION: SQLite Writer Actor ---
	// A dedicated goroutine for SQLite writes to prevent lock contention
	sqliteCh := make(chan []domain.Section, 100) // Buffered channel
	var sqliteWg sync.WaitGroup
	sqliteWg.Add(1)
	go func() {
		defer sqliteWg.Done()
		for sections := range sqliteCh {
			if err := sqliteRepo.InsertSections(sections); err != nil {
				logger.Error("SQLite insert failed", zap.Error(err))
			}
		}
	}()

	// --- OPTIMIZATION: Parallel Title Processing ---
	// Limit concurrency to avoid FD exhaustion (e.g., 4 concurrent titles)
	// While we have many cores, we don't want to hammer external APIs too hard
	maxConcurrentTitles := 4 
	sem := make(chan struct{}, maxConcurrentTitles)
	var wg sync.WaitGroup

	// Mutex for collecting results that must be aggregated centrally
	var summariesMutex sync.Mutex
	allSummaries := []domain.Summary{}

	totalTitles := len(changedTitles)

	for i, title := range changedTitles {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(t domain.Title, idx int) {
			defer wg.Done()
			defer func() { <-sem }() // Release token
			
			titleStart := time.Now()
			currentFileNum := idx + 1

			logger.Info(fmt.Sprintf("Processing title %s (%d/%d)", t.Title, currentFileNum, totalTitles),
				zap.String("title", t.Title))

			// Step 2: Pull sections (Extract)
			// Note: IngestTitle is now internally parallelized for regex ops
			sections, err := ingestUseCase.IngestTitle(ctx, t)
			if err != nil {
				if err == domain.ErrNotFound {
					logger.Warn("Title not found (skipping)", zap.String("title", t.Title))
					return
				}
				logger.Error("Ingest failed for title", zap.String("title", t.Title), zap.Error(err))
				return
			}

			// Write to Parquet (Thread-safe for different titles)
			if err := parquetRepo.WriteSections(ctx, snapshotDate, t.Title, sections); err != nil {
				logger.Error("Parquet write failed", zap.String("title", t.Title), zap.Error(err))
			}

			// Send to SQLite Writer (Non-blocking if buffer space exists)
			select {
			case sqliteCh <- sections:
			case <-ctx.Done():
				return
			}

			// Step 3: Compute deltas (Transform)
			diffs, err := snapshotUseCase.ComputeDiffs(ctx, snapshotDate, t.Title)
			if err != nil {
				logger.Error("Diff compute failed", zap.String("title", t.Title), zap.Error(err))
			} else {
				if err := parquetRepo.WriteDiffs(ctx, snapshotDate, t.Title, diffs); err != nil {
					logger.Error("Diff write failed", zap.String("title", t.Title), zap.Error(err))
				}
			}

			// Step 4: LSA metric (Transform)
			lsaCounts, err := lsaCollector.CollectForTitle(ctx, t.Title, t.LatestAmendedOn)
			if err != nil {
				logger.Error("LSA collect failed", zap.String("title", t.Title), zap.Error(err))
			} else {
				// Write to SQLite (for API queries)
				if err := sqliteRepo.InsertLSA(lsaCounts, snapshotDate); err != nil {
					logger.Error("LSA SQLite write failed", zap.String("title", t.Title), zap.Error(err))
				}
				// Write to Parquet (for analytics)
				if err := parquetRepo.WriteLSA(ctx, snapshotDate, t.Title, lsaCounts); err != nil {
					logger.Error("LSA Parquet write failed", zap.String("title", t.Title), zap.Error(err))
				}
			}

	// Step 5: Summaries (Transform)
	if !*skipSummary {
		logger.Info("Generating title summary", zap.String("title", t.Title))
		summaries, err := summariesUseCase.GenerateForTitle(ctx, t, sections)
		if err != nil {
			logger.Error("Summaries generate failed", zap.String("title", t.Title), zap.Error(err))
		} else {
			logger.Info("Generated summary", zap.Int("count", len(summaries)))
			// Thread-safe aggregation
			summariesMutex.Lock()
			allSummaries = append(allSummaries, summaries...)
			summariesMutex.Unlock()
		}
	} else {
		logger.Info("Skipping summary generation", zap.String("title", t.Title))
	}

			logger.Info("Completed title",
				zap.String("title", t.Title),
				zap.Duration("duration", time.Since(titleStart)))

		}(title, i)
	}

	// Wait for all title workers to finish
	wg.Wait()

	// Close SQLite channel and wait for writer to drain
	close(sqliteCh)
	sqliteWg.Wait()

	// Load Summaries (Single batch write to avoid file contention)
	if len(allSummaries) > 0 {
		logger.Info("Writing gathered summaries", zap.Int("count", len(allSummaries)))
		if err := parquetRepo.WriteSummaries(ctx, snapshotDate, allSummaries); err != nil {
			logger.Error("Summaries write failed", zap.Error(err))
		}
	}

	logger.Info("ETL Pipeline Completed Successfully",
		zap.String("snapshot", snapshotDate),
		zap.Duration("total_duration", time.Since(pipelineStart)))
}
