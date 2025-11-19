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

func NewRepo(path string) *Repo {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	return &Repo{Path: path, db: db}
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
