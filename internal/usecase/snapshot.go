package usecase

import (
	"context"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Snapshot struct {
	parquetRepo *parquet.Repo
	sqliteRepo  *sqlite.Repo
}

func NewSnapshot(parquet *parquet.Repo, sqlite *sqlite.Repo) *Snapshot {
	return &Snapshot{parquetRepo: parquet, sqliteRepo: sqlite}
}

func (u *Snapshot) ComputeDiffs(ctx context.Context, snapshotDate, title string) ([]domain.Diff, error) {
	prevDate, err := u.parquetRepo.GetPrevSnapshot(ctx, snapshotDate)
	if err != nil {
		return nil, err
	}

	var prevSections []domain.Section
	if prevDate != "" {
		prevSections, err = u.parquetRepo.ReadSections(ctx, prevDate, title)
		if err != nil {
			return nil, err
		}
	}
	currSections, err := u.parquetRepo.ReadSections(ctx, snapshotDate, title)
	if err != nil {
		return nil, err
	}

	diffs := []domain.Diff{}
	prevMap := make(map[string]domain.Section)
	for _, p := range prevSections {
		prevMap[p.ID] = p
	}
	for _, c := range currSections {
		p, ok := prevMap[c.ID]
		if !ok {
			diffs = append(diffs, domain.Diff{SectionID: c.ID, DeltaWordCount: c.WordCount, Changed: true})
			continue
		}
		delta := c.WordCount - p.WordCount
		changed := c.ChecksumSHA256 != p.ChecksumSHA256
		diffs = append(diffs, domain.Diff{SectionID: c.ID, DeltaWordCount: delta, Changed: changed})
	}
	return diffs, nil
}
