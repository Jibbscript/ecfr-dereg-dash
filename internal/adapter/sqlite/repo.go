package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
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

	// Create agencies table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agencies (
			id            TEXT PRIMARY KEY,
			name          TEXT NOT NULL,
			short_name    TEXT,
			sortable_name TEXT,
			parent_id     TEXT,
			FOREIGN KEY(parent_id) REFERENCES agencies(id)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create agency CFR references table (N:N mapping - no PK)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agency_cfr_references (
			agency_id TEXT NOT NULL,
			title     INTEGER NOT NULL,
			chapter   TEXT NOT NULL,
			FOREIGN KEY (agency_id) REFERENCES agencies(id)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create indexes for fast joins
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_acr_title_chapter ON agency_cfr_references(title, chapter)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_acr_agency ON agency_cfr_references(agency_id)`)

	// Create LSA activity table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS lsa_activity (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			title         TEXT NOT NULL,
			snapshot_date TEXT NOT NULL,
			proposals     INTEGER DEFAULT 0,
			amendments    INTEGER DEFAULT 0,
			finals        INTEGER DEFAULT 0,
			captured_at   DATETIME,
			source_hint   TEXT,
			UNIQUE(title, snapshot_date)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_lsa_title ON lsa_activity(title)`)

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

func (r *Repo) GetAgencyTotals(titleFilter *string) ([]domain.AgencyMetric, error) {
	// Build query with JOIN through agency_cfr_references and weighted LSA
	query := `
		WITH agency_title_words AS (
			-- Get word count per agency per title
			SELECT
				acr.agency_id,
				acr.title,
				COALESCE(SUM(s.word_count), 0) as title_words
			FROM agency_cfr_references acr
			LEFT JOIN sections s
				ON s.title = CAST(acr.title AS TEXT)
				AND s.agency_id = acr.chapter
			GROUP BY acr.agency_id, acr.title
		),
		agency_totals AS (
			-- Get total words per agency (for weighting)
			SELECT agency_id, SUM(title_words) as total_words
			FROM agency_title_words
			GROUP BY agency_id
		),
		weighted_lsa AS (
			-- Compute weighted LSA: (title_words / total_words) * lsa_count
			SELECT
				atw.agency_id,
				SUM(
					CASE WHEN at.total_words > 0
					THEN (CAST(atw.title_words AS REAL) / at.total_words)
						 * (COALESCE(l.proposals, 0) + COALESCE(l.amendments, 0) + COALESCE(l.finals, 0))
					ELSE 0 END
				) as weighted_lsa
			FROM agency_title_words atw
			JOIN agency_totals at ON at.agency_id = atw.agency_id
			LEFT JOIN lsa_activity l ON l.title = CAST(atw.title AS TEXT)
			GROUP BY atw.agency_id
		)
		SELECT
			a.id,
			a.name,
			a.parent_id,
			COALESCE(at.total_words, 0) as total_words,
			COALESCE(rscs.avg_rscs, 0) as avg_rscs,
			COALESCE(CAST(wl.weighted_lsa AS INTEGER), 0) as lsa_counts
		FROM agencies a
		LEFT JOIN agency_totals at ON at.agency_id = a.id
		LEFT JOIN weighted_lsa wl ON wl.agency_id = a.id
		LEFT JOIN (
			-- Compute avg RSCS per agency
			SELECT
				acr.agency_id,
				AVG(s.rscs_per_1k) as avg_rscs
			FROM agency_cfr_references acr
			LEFT JOIN sections s
				ON s.title = CAST(acr.title AS TEXT)
				AND s.agency_id = acr.chapter
			GROUP BY acr.agency_id
		) rscs ON rscs.agency_id = a.id
	`

	args := []any{}
	if titleFilter != nil && *titleFilter != "" {
		// For title filter, we need to filter the CTEs
		query = `
		WITH agency_title_words AS (
			SELECT
				acr.agency_id,
				acr.title,
				COALESCE(SUM(s.word_count), 0) as title_words
			FROM agency_cfr_references acr
			LEFT JOIN sections s
				ON s.title = CAST(acr.title AS TEXT)
				AND s.agency_id = acr.chapter
			WHERE acr.title = CAST(? AS INTEGER)
			GROUP BY acr.agency_id, acr.title
		),
		agency_totals AS (
			SELECT agency_id, SUM(title_words) as total_words
			FROM agency_title_words
			GROUP BY agency_id
		),
		weighted_lsa AS (
			SELECT
				atw.agency_id,
				SUM(
					CASE WHEN at.total_words > 0
					THEN (CAST(atw.title_words AS REAL) / at.total_words)
						 * (COALESCE(l.proposals, 0) + COALESCE(l.amendments, 0) + COALESCE(l.finals, 0))
					ELSE 0 END
				) as weighted_lsa
			FROM agency_title_words atw
			JOIN agency_totals at ON at.agency_id = atw.agency_id
			LEFT JOIN lsa_activity l ON l.title = CAST(atw.title AS TEXT)
			GROUP BY atw.agency_id
		)
		SELECT
			a.id,
			a.name,
			a.parent_id,
			COALESCE(at.total_words, 0) as total_words,
			COALESCE(rscs.avg_rscs, 0) as avg_rscs,
			COALESCE(CAST(wl.weighted_lsa AS INTEGER), 0) as lsa_counts
		FROM agencies a
		LEFT JOIN agency_totals at ON at.agency_id = a.id
		LEFT JOIN weighted_lsa wl ON wl.agency_id = a.id
		LEFT JOIN (
			SELECT
				acr.agency_id,
				AVG(s.rscs_per_1k) as avg_rscs
			FROM agency_cfr_references acr
			LEFT JOIN sections s
				ON s.title = CAST(acr.title AS TEXT)
				AND s.agency_id = acr.chapter
			WHERE acr.title = CAST(? AS INTEGER)
			GROUP BY acr.agency_id
		) rscs ON rscs.agency_id = a.id
		WHERE at.total_words > 0
		`
		args = append(args, *titleFilter, *titleFilter)
	}

	query += " ORDER BY total_words DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.AgencyMetric
	for rows.Next() {
		var m domain.AgencyMetric
		if err := rows.Scan(&m.ID, &m.Name, &m.ParentID, &m.TotalWords, &m.AvgRSCS, &m.LSACounts); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// InsertLSA inserts or updates LSA activity data for a title
func (r *Repo) InsertLSA(lsa domain.LSAActivity, snapshotDate string) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO lsa_activity
		(title, snapshot_date, proposals, amendments, finals, captured_at, source_hint)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		lsa.Key, snapshotDate, lsa.ProposalsCount, lsa.AmendmentsCount,
		lsa.FinalsCount, lsa.CapturedAt, lsa.SourceHint)
	return err
}

// GetAgencyChecksum computes a SHA256 hash of all section content for an agency
func (r *Repo) GetAgencyChecksum(agencyID string) (string, error) {
	query := `
		SELECT s.text
		FROM sections s
		JOIN agency_cfr_references acr
			ON s.title = CAST(acr.title AS TEXT)
			AND s.agency_id = acr.chapter
		WHERE acr.agency_id = ?
		ORDER BY s.id
	`
	rows, err := r.db.Query(query, agencyID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var combined strings.Builder
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return "", err
		}
		combined.WriteString(text)
	}

	if combined.Len() == 0 {
		return "", nil
	}

	hash := sha256.Sum256([]byte(combined.String()))
	return hex.EncodeToString(hash[:]), nil
}
