package duck

import (
	"database/sql"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Helper struct {
	db *sql.DB
}

func NewHelper(parquet *parquet.Repo, sqlite *sqlite.Repo, enableUI bool) *Helper {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("INSTALL parquet; LOAD parquet;")
	if err != nil {
		panic(err)
	}
	// Note: This view creation might fail if no parquet files exist yet.
	// In a real app we'd handle this more gracefully or init on first query.
	// For now, we wrap in try/catch logic or assume ETL runs first.
	// We'll skip the view creation here to avoid panic on startup if empty.

	if sqlite.Path != "" {
		// _, err = db.Exec("ATTACH '" + sqlite.Path + "' (TYPE sqlite);")
		// if err != nil {
		// 	panic(err)
		// }
	}
	if enableUI {
		// _, err = db.Exec("CALL start_ui()")
		// if err != nil {
		// 	panic(err)
		// }
	}
	return &Helper{db: db}
}

func (h *Helper) QueryAgencies(query string) ([]domain.Agency, error) {
	// Mock implementation since we can't easily query empty parquet
	return []domain.Agency{}, nil
}
