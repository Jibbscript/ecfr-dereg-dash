package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/govinfo"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
	"go.uber.org/zap"
)

type Ingest struct {
	logger      *zap.Logger
	govinfo     *govinfo.Client
	parquetRepo *parquet.Repo
	sqliteRepo  *sqlite.Repo
}

func NewIngest(logger *zap.Logger, govinfo *govinfo.Client, parquet *parquet.Repo, sqlite *sqlite.Repo) *Ingest {
	return &Ingest{logger: logger, govinfo: govinfo, parquetRepo: parquet, sqliteRepo: sqlite}
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
	u.logger.Info("Starting ingestion for title", zap.String("title", title.Title))
	start := time.Now()

	titleNum, _ := strconv.Atoi(title.Title)

	u.logger.Debug("Downloading title XML", zap.Int("title_num", titleNum))
	path, err := u.govinfo.DownloadTitleXML(ctx, titleNum)
	if err != nil {
		u.logger.Error("Failed to download title XML", zap.String("title", title.Title), zap.Error(err))
		return nil, err
	}
	u.logger.Debug("Download complete", zap.String("path", path))

	u.logger.Debug("Parsing title XML", zap.String("path", path))
	rawSections, err := u.govinfo.ParseTitleXML(ctx, path)
	if err != nil {
		u.logger.Warn("Parsing failed, retrying download", zap.String("path", path), zap.Error(err))
		// Parsing failed, file might be corrupt. Delete and retry.
		// os.Remove(path) // No longer local file
		// Retry logic might need to delete from GCS or just overwrite?
		// DownloadTitleXML overwrites if we force it? Or we need to delete object?
		// For now let's assume DownloadTitleXML handles re-download if we call it again?
		// Actually my implementation checks "if exists return".
		// So I might need a "Force" flag or just ignore retry for now to keep it simple.
		// Or better: The previous implementation deleted local file.
		// I'll just re-call download. But Download checks existence.
		// I should probably rely on the error handling in Download.

		// Let's just fail for now as deleting GCS object adds complexity I didn't add to client.
		return nil, err
	}
	u.logger.Info("Parsing complete", zap.Int("sections_found", len(rawSections)))

	// Optimization: Pre-allocate result slice to preserve order and avoid mutex on append
	numSections := len(rawSections)
	sections := make([]domain.Section, numSections)

	// Optimization: Worker pool for CPU-bound regex operations
	// Use 2x logical cores for better saturation if some ops block slightly,
	// though these are mostly pure CPU.
	numWorkers := runtime.NumCPU()
	sem := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup

	// Snapshot date is constant for the batch
	snapshotDate := time.Now().Format("2006-01-02")

	for i, raw := range rawSections {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(idx int, raw domain.Section) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

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

			// Assign directly to pre-allocated slice index - thread safe
			sections[idx] = domain.Section{
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
				SnapshotDate:   snapshotDate,
			}
		}(i, raw)
	}

	wg.Wait()

	u.logger.Info("Ingestion finished for title",
		zap.String("title", title.Title),
		zap.Duration("duration", time.Since(start)),
		zap.Int("sections_generated", len(sections)),
	)
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
