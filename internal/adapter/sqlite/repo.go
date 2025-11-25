package sqlite

import (
	"database/sql"

	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Repo struct {
	Path string
	db   *sql.DB
}

func NewRepo(path string) (*Repo, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sections (
			id TEXT PRIMARY KEY,
			title TEXT,
			part TEXT,
			section TEXT,
			agency_id TEXT,
			path TEXT,
			text TEXT,
			rev_date DATETIME,
			checksum_sha256 TEXT,
			word_count INTEGER,
			def_count INTEGER,
			xref_count INTEGER,
			modal_count INTEGER,
			rscs_raw INTEGER,
			rscs_per_1k REAL,
			snapshot_date TEXT
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Repo{Path: path, db: db}, nil
}

func (r *Repo) InsertSections(sections []domain.Section) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO sections (id, title, part, section, agency_id, path, text, rev_date, checksum_sha256, word_count, def_count, xref_count, modal_count, rscs_raw, rscs_per_1k, snapshot_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, s := range sections {
		_, err = stmt.Exec(s.ID, s.Title, s.Part, s.Section, s.AgencyID, s.Path, s.Text, s.RevDate, s.ChecksumSHA256, s.WordCount, s.DefCount, s.XrefCount, s.ModalCount, s.RSCSRaw, s.RSCSPer1K, s.SnapshotDate)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *Repo) GetAgencyTotals() ([]domain.AgencyMetric, error) {
	rows, err := r.db.Query(`
		SELECT 
			agency_id, 
			SUM(word_count) as total_words, 
			AVG(rscs_per_1k) as avg_rscs 
		FROM sections 
		GROUP BY agency_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.AgencyMetric
	for rows.Next() {
		var m domain.AgencyMetric
		if err := rows.Scan(&m.ID, &m.TotalWords, &m.AvgRSCS); err != nil {
			return nil, err
		}
		// For now, use ID as Name or map it if we have a mapping table.
		// Since we don't have an agencies table, we'll just use the ID.
		m.Name = m.ID
		metrics = append(metrics, m)
	}
	return metrics, nil
}
