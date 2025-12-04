package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/duck"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
	delivery "github.com/Jibbscript/ecfr-dereg-dashboard/internal/delivery/http"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/platform"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/usecase"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

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
		// GCS backend (prod/staging)
		parquetRepo, err = parquet.NewRepo(ctx, config.ParquetBucket, config.ParquetPrefix)
		if err != nil {
			logger.Fatal("Failed to create Parquet repo", zap.Error(err))
		}
	}

	sqliteRepo, err := sqlite.NewRepo(config.DataDir + "/ecfr.db")
	if err != nil {
		logger.Fatal("Failed to create SQLite repo", zap.Error(err))
	}

	// Collect LSA data on startup if stale or missing
	hasRecent, err := sqliteRepo.HasRecentAgencyLSA(1)
	if err != nil || !hasRecent {
		logger.Info("Collecting fresh LSA data from Federal Register API")
		lsaCollector := lsa.NewCollector()
		records, lsaErr := lsaCollector.CollectAgencyLSABatch(ctx)
		if lsaErr != nil {
			logger.Warn("LSA collection failed", zap.Error(lsaErr))
		} else {
			if insertErr := sqliteRepo.InsertAgencyLSABatch(records); insertErr != nil {
				logger.Warn("LSA insert failed", zap.Error(insertErr))
			} else {
				logger.Info("LSA data refreshed", zap.Int("agencies", len(records)))
			}
		}
	}

	duckHelper, err := duck.NewHelper(parquetRepo, sqliteRepo, config.DuckDBUI)
	if err != nil {
		logger.Fatal("Failed to create DuckDB helper", zap.Error(err))
	}

	var govinfoClient *govinfo.Client
	if config.Env == "local" || config.Env == "dev" {
		logger.Warn("Skipping GovInfo Client initialization (local mode)")
		// Ideally mock this if needed, or ensure it's not used for read-only flows
	} else {
		govinfoClient, err = govinfo.NewClient(ctx, config.RawXMLBucket, config.RawXMLPrefix)
		if err != nil {
			logger.Fatal("Failed to create GovInfo client", zap.Error(err))
		}
	}

	usecases := delivery.Usecases{
		Ingest:    usecase.NewIngest(logger, govinfoClient, parquetRepo, sqliteRepo),
		Snapshot:  usecase.NewSnapshot(parquetRepo, sqliteRepo),
		Metrics:   usecase.NewMetrics(duckHelper, sqliteRepo),
		Summaries: usecase.NewSummariesReadOnly(logger, sqliteRepo),
	}

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		delivery.SetupHandlers(r, usecases, logger)
	})

	httpAddr := ":" + getEnv("PORT", "8080")
	logger.Info("Starting API server", zap.String("addr", httpAddr))
	if err := http.ListenAndServe(httpAddr, r); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
