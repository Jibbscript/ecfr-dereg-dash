package domain

import "time"

type Agency struct {
	ID   string
	Name string
}

type AgencyMetric struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ParentID        *string `json:"parent_id"`
	TotalWords      int     `json:"total_words"`
	AvgRSCS         float64 `json:"avg_rscs"`
	LSACounts       int     `json:"lsa_counts"`
	ContentChecksum string  `json:"content_checksum,omitempty"`
}

type Title struct {
	Title           string
	Name            string
	LatestAmendedOn time.Time
	LatestIssueDate time.Time
	UpToDateAsOf    time.Time
}

type Section struct {
	ID             string
	Title          string
	Part           string
	Section        string
	AgencyID       string
	Path           string
	Text           string
	RevDate        time.Time
	ChecksumSHA256 string
	WordCount      int
	DefCount       int
	XrefCount      int
	ModalCount     int
	RSCSRaw        int
	RSCSPer1K      float64
	SnapshotDate   string
}

type RawSection struct {
	ID       string
	Part     string
	Section  string
	AgencyID string
	Path     string
	Text     string
	RevDate  time.Time
}

type Summary struct {
	Kind      string    `json:"kind"` // agency|title|section
	Key       string    `json:"key"`
	Text      string    `json:"text"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

type LSAActivity struct {
	KeyKind         string // title|section
	Key             string
	SinceRevDate    time.Time
	ProposalsCount  int
	AmendmentsCount int
	FinalsCount     int
	CapturedAt      time.Time
	SourceHint      string
}

// AgencyLSA tracks regulatory activity (proposed rules, final rules, notices) per agency
type AgencyLSA struct {
	AgencyID       string    // Agency slug/ID (e.g., "environmental-protection-agency")
	AgencyName     string    // Human-readable agency name
	ProposedRules  int       // Count of proposed rules
	FinalRules     int       // Count of final rules
	Notices        int       // Count of notices
	TotalDocuments int       // Total document count
	SnapshotDate   string    // Date of data collection (YYYY-MM-DD)
	CapturedAt     time.Time // Timestamp when data was fetched
	SourceHint     string    // Data source identifier (e.g., "federalregister-api")
}

type Diff struct {
	SectionID      string
	DeltaWordCount int
	Changed        bool
}
