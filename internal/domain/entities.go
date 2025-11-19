package domain

import "time"

type Agency struct {
	ID   string
	Name string
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
	Kind      string // agency|title|section
	Key       string
	Text      string
	Model     string
	CreatedAt time.Time
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

type Diff struct {
	SectionID      string
	DeltaWordCount int
	Changed        bool
}
