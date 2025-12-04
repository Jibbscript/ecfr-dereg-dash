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

	// Create agency_lsa table for per-agency LSA data from Federal Register API
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agency_lsa (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			agency_id     TEXT NOT NULL,
			agency_name   TEXT NOT NULL,
			proposed_rules INTEGER DEFAULT 0,
			final_rules   INTEGER DEFAULT 0,
			notices       INTEGER DEFAULT 0,
			total_documents INTEGER DEFAULT 0,
			snapshot_date TEXT NOT NULL,
			captured_at   DATETIME,
			source_hint   TEXT,
			UNIQUE(agency_id, snapshot_date)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_agency_lsa_agency ON agency_lsa(agency_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_agency_lsa_snapshot ON agency_lsa(snapshot_date)`)

	// Create summaries table for AI-generated summaries
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT NOT NULL DEFAULT 'title',
			key TEXT NOT NULL,
			text TEXT NOT NULL,
			model TEXT DEFAULT 'gemini-2.5-pro',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(kind, key)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_summaries_kind_key ON summaries(kind, key)`)

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
	// Build query with JOIN through agency_cfr_references
	// LSA counts now come directly from agency_lsa table (per-agency from Federal Register API)
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
			-- Get total words per agency
			SELECT agency_id, SUM(title_words) as total_words
			FROM agency_title_words
			GROUP BY agency_id
		),
		latest_agency_lsa AS (
			-- Get latest LSA data per agency from agency_lsa table
			SELECT agency_id, total_documents
			FROM agency_lsa
			WHERE snapshot_date = (SELECT MAX(snapshot_date) FROM agency_lsa)
		)
		SELECT
			a.id,
			a.name,
			a.parent_id,
			COALESCE(at.total_words, 0) as total_words,
			COALESCE(rscs.avg_rscs, 0) as avg_rscs,
			COALESCE(lsa.total_documents, 0) as lsa_counts
		FROM agencies a
		LEFT JOIN agency_totals at ON at.agency_id = a.id
		LEFT JOIN latest_agency_lsa lsa ON lsa.agency_id = a.id
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
		// Note: LSA counts are still per-agency (not filtered by title) since they come from Federal Register API
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
		latest_agency_lsa AS (
			SELECT agency_id, total_documents
			FROM agency_lsa
			WHERE snapshot_date = (SELECT MAX(snapshot_date) FROM agency_lsa)
		)
		SELECT
			a.id,
			a.name,
			a.parent_id,
			COALESCE(at.total_words, 0) as total_words,
			COALESCE(rscs.avg_rscs, 0) as avg_rscs,
			COALESCE(lsa.total_documents, 0) as lsa_counts
		FROM agencies a
		LEFT JOIN agency_totals at ON at.agency_id = a.id
		LEFT JOIN latest_agency_lsa lsa ON lsa.agency_id = a.id
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

// InsertAgencyLSA inserts or updates LSA activity data for an agency
func (r *Repo) InsertAgencyLSA(lsa domain.AgencyLSA) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO agency_lsa
		(agency_id, agency_name, proposed_rules, final_rules, notices, total_documents, snapshot_date, captured_at, source_hint)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		lsa.AgencyID, lsa.AgencyName, lsa.ProposedRules, lsa.FinalRules,
		lsa.Notices, lsa.TotalDocuments, lsa.SnapshotDate, lsa.CapturedAt, lsa.SourceHint)
	return err
}

// InsertAgencyLSABatch inserts multiple agency LSA records in a transaction
func (r *Repo) InsertAgencyLSABatch(lsaRecords []domain.AgencyLSA) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO agency_lsa
		(agency_id, agency_name, proposed_rules, final_rules, notices, total_documents, snapshot_date, captured_at, source_hint)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, lsa := range lsaRecords {
		_, err = stmt.Exec(lsa.AgencyID, lsa.AgencyName, lsa.ProposedRules, lsa.FinalRules,
			lsa.Notices, lsa.TotalDocuments, lsa.SnapshotDate, lsa.CapturedAt, lsa.SourceHint)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetAgencyLSA retrieves the latest LSA data for a specific agency
func (r *Repo) GetAgencyLSA(agencyID string) (*domain.AgencyLSA, error) {
	var lsa domain.AgencyLSA
	err := r.db.QueryRow(`
		SELECT agency_id, agency_name, proposed_rules, final_rules, notices, total_documents, snapshot_date, captured_at, source_hint
		FROM agency_lsa
		WHERE agency_id = ?
		ORDER BY snapshot_date DESC
		LIMIT 1`, agencyID).Scan(
		&lsa.AgencyID, &lsa.AgencyName, &lsa.ProposedRules, &lsa.FinalRules,
		&lsa.Notices, &lsa.TotalDocuments, &lsa.SnapshotDate, &lsa.CapturedAt, &lsa.SourceHint)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lsa, nil
}

// GetAllAgencyLSA retrieves the latest LSA data for all agencies
func (r *Repo) GetAllAgencyLSA() ([]domain.AgencyLSA, error) {
	rows, err := r.db.Query(`
		SELECT agency_id, agency_name, proposed_rules, final_rules, notices, total_documents, snapshot_date, captured_at, source_hint
		FROM agency_lsa
		WHERE snapshot_date = (SELECT MAX(snapshot_date) FROM agency_lsa)
		ORDER BY total_documents DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.AgencyLSA
	for rows.Next() {
		var lsa domain.AgencyLSA
		if err := rows.Scan(&lsa.AgencyID, &lsa.AgencyName, &lsa.ProposedRules, &lsa.FinalRules,
			&lsa.Notices, &lsa.TotalDocuments, &lsa.SnapshotDate, &lsa.CapturedAt, &lsa.SourceHint); err != nil {
			return nil, err
		}
		results = append(results, lsa)
	}
	return results, nil
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

// InsertSummary inserts or updates a single summary
func (r *Repo) InsertSummary(summary domain.Summary) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO summaries (kind, key, text, model, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		summary.Kind, summary.Key, summary.Text, summary.Model, summary.CreatedAt)
	return err
}

// InsertSummaries inserts or updates multiple summaries in a transaction
func (r *Repo) InsertSummaries(summaries []domain.Summary) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO summaries (kind, key, text, model, created_at)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range summaries {
		_, err = stmt.Exec(s.Kind, s.Key, s.Text, s.Model, s.CreatedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetAllSummaries retrieves all summaries from the database
func (r *Repo) GetAllSummaries() ([]domain.Summary, error) {
	rows, err := r.db.Query(`
		SELECT kind, key, text, model, created_at
		FROM summaries
		ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.Summary
	for rows.Next() {
		var s domain.Summary
		if err := rows.Scan(&s.Kind, &s.Key, &s.Text, &s.Model, &s.CreatedAt); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// GetSummaryByKey retrieves a single summary by kind and key
func (r *Repo) GetSummaryByKey(kind, key string) (*domain.Summary, error) {
	var s domain.Summary
	err := r.db.QueryRow(`
		SELECT kind, key, text, model, created_at
		FROM summaries
		WHERE kind = ? AND key = ?`, kind, key).Scan(&s.Kind, &s.Key, &s.Text, &s.Model, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
