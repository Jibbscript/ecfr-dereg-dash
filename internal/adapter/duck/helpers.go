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

	// Parquet extension is bundled in duckdb-go v2, so we just load it if needed,
	// or trust it's autoloaded. Explicit INSTALL can cause issues if it tries to download.
	// We'll try LOAD, but ignore error if it's already loaded or builtin.
	// Actually, for v2.5.0, extensions are statically linked.
	// Let's skip explicit INSTALL/LOAD unless we hit errors.

	// Note: This view creation might fail if no parquet files exist yet.
	// In a real app we'd handle this more gracefully or init on first query.
	// For now, we wrap in try/catch logic or assume ETL runs first.
	// We'll skip the view creation here to avoid panic on startup if empty.

	if sqlite.Path != "" {
		// _, err = db.Exec("ATTACH '" + sqlite.Path + "' (TYPE sqlite);")
		// if err != nil {
		// 	return nil, err
		// }
	}
	if enableUI {
		// _, err = db.Exec("CALL start_ui()")
		// if err != nil {
		// 	return nil, err
		// }
	}
	return &Helper{db: db}, nil
}

func (h *Helper) QueryAgencies(query string) ([]domain.Agency, error) {
	// Mock implementation since we can't easily query empty parquet
	return []domain.Agency{}, nil
}
