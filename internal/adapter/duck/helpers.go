package duck

import (
	"database/sql"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
)

type Helper struct {
	db *sql.DB
}

func NewHelper(parquet *parquet.Repo, sqlite *sqlite.Repo, enableUI bool) (*Helper, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, err
	}

	// Attach SQLite database for combined queries across SQLite and Parquet
	if sqlite.Path != "" {
		_, err = db.Exec("ATTACH '" + sqlite.Path + "' (TYPE sqlite);")
		if err != nil {
			return nil, err
		}
	}

	// Create a view for Parquet data if files exist (ignore errors if no files yet)
	// This allows analysts to query sections_parquet directly
	_, _ = db.Exec(`
		CREATE OR REPLACE VIEW sections_parquet AS
		SELECT * FROM read_parquet('data/parquet/*/*.parquet')
	`)

	// Start DuckDB web UI on port 4213 for interactive analytics
	if enableUI {
		_, err = db.Exec("CALL start_ui()")
		if err != nil {
			return nil, err
		}
	}

	return &Helper{db: db}, nil
}

func (h *Helper) QueryAgencies(query string) ([]domain.Agency, error) {
	// Mock implementation since we can't easily query empty parquet
	return []domain.Agency{}, nil
}
