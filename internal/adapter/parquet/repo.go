package parquet

import (
	"os"
	"time"

	"github.com/parquet-go/parquet-go"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Repo struct {
	Root string
}

func NewRepo(root string) *Repo {
	return &Repo{Root: root}
}

func (r *Repo) WriteSections(snapshot, title string, sections []domain.Section) error {
	dir := r.Root + "/" + snapshot
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := dir + "/" + title + ".parquet"
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := parquet.NewGenericWriter[domain.Section](f)
	_, err = writer.Write(sections)
	if err != nil {
		return err
	}
	return writer.Close()
}

func (r *Repo) ReadSections(snapshot, title string) ([]domain.Section, error) {
	path := r.Root + "/" + snapshot + "/" + title + ".parquet"
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := parquet.NewGenericReader[domain.Section](f)
	rows := make([]domain.Section, reader.NumRows())
	_, err = reader.Read(rows)
	return rows, err
}

func (r *Repo) GetLatestSnapshot() (time.Time, error) {
	entries, err := os.ReadDir(r.Root)
	if err != nil {
		return time.Time{}, err
	}
	var maxTime time.Time
	for _, e := range entries {
		if e.IsDir() {
			t, err := time.Parse("2006-01-02", e.Name())
			if err == nil && t.After(maxTime) {
				maxTime = t
			}
		}
	}
	return maxTime, nil
}

func (r *Repo) GetPrevSnapshot(snapshot string) (string, error) {
	entries, err := os.ReadDir(r.Root)
	if err != nil {
		return "", err
	}
	currentTime, err := time.Parse("2006-01-02", snapshot)
	if err != nil {
		return "", err
	}
	var prevDate time.Time
	for _, e := range entries {
		if e.IsDir() {
			t, err := time.Parse("2006-01-02", e.Name())
			if err == nil && t.Before(currentTime) && t.After(prevDate) {
				prevDate = t
			}
		}
	}
	if prevDate.IsZero() {
		return "", os.ErrNotExist
	}
	return prevDate.Format("2006-01-02"), nil
}

func (r *Repo) WriteDiffs(snapshot, title string, diffs []domain.Diff) error {
	dir := r.Root + "/" + snapshot
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := dir + "/" + title + "_diffs.parquet"
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := parquet.NewGenericWriter[domain.Diff](f)
	_, err = writer.Write(diffs)
	if err != nil {
		return err
	}
	return writer.Close()
}

func (r *Repo) WriteLSA(snapshot, title string, lsa domain.LSAActivity) error {
	// Implementation for LSA writing
	return nil
}

func (r *Repo) WriteSummaries(snapshot string, summaries []domain.Summary) error {
	dir := r.Root + "/" + snapshot
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := dir + "/summaries.parquet"
	// Append mode logic would be needed here or separate files
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := parquet.NewGenericWriter[domain.Summary](f)
	_, err = writer.Write(summaries)
	if err != nil {
		return err
	}
	return writer.Close()
}
