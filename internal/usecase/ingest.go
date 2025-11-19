package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/lsa"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Ingest struct {
	govinfo     *govinfo.Client
	lsa         *lsa.Collector
	parquetRepo *parquet.Repo
	sqliteRepo  *sqlite.Repo
}

func NewIngest(govinfo *govinfo.Client, lsa *lsa.Collector, parquet *parquet.Repo, sqlite *sqlite.Repo) *Ingest {
	return &Ingest{govinfo: govinfo, lsa: lsa, parquetRepo: parquet, sqliteRepo: sqlite}
}

func (u *Ingest) FetchChangedTitles(ctx context.Context) ([]domain.Title, error) {
	// For MVP, we might iterate all 50 titles or check a manifest.
	// Simplified: Return list of 50 titles.
	titles := []domain.Title{}
	for i := 1; i <= 50; i++ {
		titles = append(titles, domain.Title{
			Title: strconv.Itoa(i),
			Name:  "Title " + strconv.Itoa(i),
		})
	}
	return titles, nil
}

func (u *Ingest) IngestTitle(ctx context.Context, title domain.Title) ([]domain.Section, error) {
	titleNum, _ := strconv.Atoi(title.Title)
	path, err := u.govinfo.DownloadTitleXML(titleNum)
	if err != nil {
		return nil, err
	}

	rawSections, err := u.govinfo.ParseTitleXML(path)
	if err != nil {
		// Parsing failed, file might be corrupt. Delete and retry.
		os.Remove(path)
		path, err = u.govinfo.DownloadTitleXML(titleNum)
		if err != nil {
			return nil, err
		}
		rawSections, err = u.govinfo.ParseTitleXML(path)
		if err != nil {
			return nil, err
		}
	}

	sections := []domain.Section{}
	for _, raw := range rawSections {
		text := normalizeText(raw.Text)
		checksum := sha256.Sum256([]byte(text))
		wordCount := len(strings.Fields(text))

		defCount := countDefs(text)
		xrefCount := countXrefs(text)
		modalCount := countModals(text)

		rscsRaw := wordCount + 20*defCount + 50*xrefCount + 100*modalCount
		rscsPer1K := 0.0
		if wordCount > 0 {
			rscsPer1K = 1000.0 * float64(rscsRaw) / float64(wordCount)
		}

		sections = append(sections, domain.Section{
			ID:             raw.ID,
			Title:          title.Title,
			Part:           raw.Part,
			Section:        raw.Section,
			AgencyID:       raw.AgencyID,
			Path:           raw.Path,
			Text:           raw.Text,
			RevDate:        raw.RevDate,
			ChecksumSHA256: hex.EncodeToString(checksum[:]),
			WordCount:      wordCount,
			DefCount:       defCount,
			XrefCount:      xrefCount,
			ModalCount:     modalCount,
			RSCSRaw:        rscsRaw,
			RSCSPer1K:      rscsPer1K,
			SnapshotDate:   time.Now().Format("2006-01-02"),
		})
	}
	return sections, nil
}

func normalizeText(text string) string {
	text = strings.ToLower(text)
	text = regexp.MustCompile(`\p{P}`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func countDefs(text string) int {
	reHead := regexp.MustCompile(`(?i)^(definitions\.?|as used in this (part|subpart|section))`)
	reMeans := regexp.MustCompile(`(?i)\b[a-z][\w\- ]{1,80}\b\s+means\b`)
	return len(reHead.FindAllString(text, -1)) + len(reMeans.FindAllString(text, -1))
}

func countXrefs(text string) int {
	reCfr := regexp.MustCompile(`(?i)(§\s*\d+(?:\.\d+)*|\b\d+\s*cfr\s*\d+(?:\.\d+)*)`)
	return len(reCfr.FindAllString(text, -1))
}

func countModals(text string) int {
	reModal := regexp.MustCompile(`(?i)\b(shall|must|may not|must not)\b`)
	return len(reModal.FindAllString(text, -1))
}
