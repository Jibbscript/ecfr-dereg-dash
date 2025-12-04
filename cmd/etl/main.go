package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
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

	logger.Info("Starting ETL Pipeline (Optimized)",
		zap.String("env", config.Env),
		zap.String("data_dir", config.DataDir),
		zap.Int("gomaxprocs", runtime.GOMAXPROCS(0)))
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

	govinfoClient, err := govinfo.NewClient(ctx, config.RawXMLBucket, config.RawXMLPrefix)
	if err != nil {
		if config.Env == "local" {
			logger.Warn("Failed to create GovInfo client (skipping ETL steps that require it)", zap.Error(err))
			logger.Info("ETL Agency Ingestion Completed (Local Mode)")
			return
		}
		logger.Fatal("Failed to create GovInfo client", zap.Error(err))
	}

	lsaCollector := lsa.NewCollector()

	ingestUseCase := usecase.NewIngest(logger, govinfoClient, parquetRepo, sqliteRepo)
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

	// Step 4: Collect Agency-level LSA data from Federal Register API
	logger.Info("Step 4/5: Collecting agency-level LSA data (Transform)")
	agencyLSAStart := time.Now()

	agencyLSARecords, err := lsaCollector.CollectAgencyLSABatch(ctx)
	if err != nil {
		logger.Error("Agency LSA batch collection failed", zap.Error(err))
	} else {
		logger.Info("Collected agency LSA data",
			zap.Int("agency_count", len(agencyLSARecords)),
			zap.Duration("duration", time.Since(agencyLSAStart)))

		// Write to SQLite
		if err := sqliteRepo.InsertAgencyLSABatch(agencyLSARecords); err != nil {
			logger.Error("Agency LSA SQLite write failed", zap.Error(err))
		} else {
			logger.Info("Agency LSA data written to SQLite")
		}

		// Write to Parquet
		if err := parquetRepo.WriteAgencyLSA(ctx, snapshotDate, agencyLSARecords); err != nil {
			logger.Error("Agency LSA Parquet write failed", zap.Error(err))
		} else {
			logger.Info("Agency LSA data written to Parquet")
		}
	}

	logger.Info("ETL Pipeline Completed Successfully",
		zap.String("snapshot", snapshotDate),
		zap.Duration("total_duration", time.Since(pipelineStart)))
}
