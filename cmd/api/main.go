package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/duck"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/ecfr"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	delivery "github.com/xai/ecfr-dereg-dashboard/internal/delivery/http"
	"github.com/xai/ecfr-dereg-dashboard/internal/platform"
	"github.com/xai/ecfr-dereg-dashboard/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

	ctx := context.Background()

	parquetRepo := parquet.NewRepo(config.DataDir)
	sqliteRepo := sqlite.NewRepo(config.DataDir + "/ecfr.db")
	duckHelper := duck.NewHelper(parquetRepo, sqliteRepo, config.DuckDBUI)

	ecfrClient := ecfr.NewClient() // Still kept for LSA if needed, or remove if unused
	govinfoClient := govinfo.NewClient(config.DataDir)

	var vertexClient *vertexai.Client
	if os.Getenv("SKIP_VERTEX") == "true" {
		logger.Warn("Skipping Vertex Client initialization (using mock)")
		vertexClient = vertexai.NewMockClient()
	} else {
		var err error
		vertexClient, err = vertexai.NewClient(ctx, config.VertexProjectID, config.VertexLocation, config.VertexModelID)
		if err != nil {
			logger.Fatal("Failed to create Vertex client", zap.Error(err))
		}
	}

	lsaCollector := lsa.NewCollector(ecfrClient, vertexClient)

	usecases := delivery.Usecases{
		Ingest:    usecase.NewIngest(govinfoClient, lsaCollector, parquetRepo, sqliteRepo),
		Snapshot:  usecase.NewSnapshot(parquetRepo, sqliteRepo),
		Metrics:   usecase.NewMetrics(duckHelper),
		Summaries: usecase.NewSummaries(vertexClient, parquetRepo),
	}

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		delivery.SetupHandlers(r, usecases, logger)
	})

	httpAddr := ":8080"
	logger.Info("Starting API server", zap.String("addr", httpAddr))
	if err := http.ListenAndServe(httpAddr, r); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
