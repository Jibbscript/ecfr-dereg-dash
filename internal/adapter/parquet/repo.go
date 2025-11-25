package parquet

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/parquet-go/parquet-go"
	"google.golang.org/api/iterator"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Repo struct {
	client     *storage.Client
	bucketName string
	rootPrefix string
	
	// Local mode
	localDir string // if non-empty, use filesystem backend
}

func NewRepo(ctx context.Context, bucketName, rootPrefix string) (*Repo, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Repo{
		client:     c,
		bucketName: bucketName,
		rootPrefix: rootPrefix,
	}, nil
}

func NewLocalRepo(rootDir, rootPrefix string) (*Repo, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(filepath.Join(rootDir, rootPrefix), 0o755); err != nil {
		return nil, err
	}
	return &Repo{
		localDir:   rootDir,
		rootPrefix: rootPrefix,
	}, nil
}

func (r *Repo) objectPath(parts ...string) string {
	// join with "/": <root>/<snapshot>/<file>
	p := r.rootPrefix
	for _, part := range parts {
		if p == "" {
			p = part
		} else {
			p = p + "/" + part
		}
	}
	return p
}

func (r *Repo) localPath(parts ...string) string {
	// <localDir>/<rootPrefix>/<snapshot>/<file>
	rel := r.objectPath(parts...)
	if r.localDir == "" {
		return rel
	}
	return filepath.Join(r.localDir, rel)
}

func (r *Repo) isLocal() bool {
	return r.localDir != ""
}

func (r *Repo) WriteSections(ctx context.Context, snapshot, title string, sections []domain.Section) error {
	if r.isLocal() {
		path := r.localPath(snapshot, title+".parquet")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		writer := parquet.NewGenericWriter[domain.Section](f)
		if _, err := writer.Write(sections); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		return nil
	}

	path := r.objectPath(snapshot, title+".parquet")
	w := r.client.Bucket(r.bucketName).Object(path).NewWriter(ctx)
	defer w.Close()

	writer := parquet.NewGenericWriter[domain.Section](w)
	if _, err := writer.Write(sections); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return w.Close()
}

func (r *Repo) ReadSections(ctx context.Context, snapshot, title string) ([]domain.Section, error) {
	var reader io.Reader

	if r.isLocal() {
		path := r.localPath(snapshot, title+".parquet")
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	} else {
		path := r.objectPath(snapshot, title+".parquet")
		obj := r.client.Bucket(r.bucketName).Object(path)

		rc, err := obj.NewReader(ctx)
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		reader = rc
	}

	// parquet.NewGenericReader needs io.ReaderAt; buffer entire object in memory.
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	pr := parquet.NewGenericReader[domain.Section](bytes.NewReader(buf))
	defer pr.Close()

	rows := make([]domain.Section, pr.NumRows())
	_, err = pr.Read(rows)
	return rows, err
}

func (r *Repo) GetLatestSnapshot(ctx context.Context) (time.Time, error) {
	if r.isLocal() {
		base := filepath.Join(r.localDir, r.rootPrefix)
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				return time.Time{}, nil
			}
			return time.Time{}, err
		}

		var maxTime time.Time
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name() // expect "YYYY-MM-DD"
			t, err := time.Parse("2006-01-02", name)
			if err == nil && t.After(maxTime) {
				maxTime = t
			}
		}
		return maxTime, nil
	}

	// List "directories" under rootPrefix by using Delimiter="/"
	it := r.client.Bucket(r.bucketName).Objects(ctx, &storage.Query{
		Prefix:    r.rootPrefix + "/",
		Delimiter: "/",
	})

	var maxTime time.Time
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return time.Time{}, err
		}
		// When Delimiter is set, prefixes are returned as ObjectAttrs with Prefix set
		if attrs.Prefix != "" {
			// strip rootPrefix + "/"
			snap := strings.TrimPrefix(attrs.Prefix, r.rootPrefix+"/")
			snap = strings.TrimSuffix(snap, "/")
			t, err := time.Parse("2006-01-02", snap)
			if err == nil && t.After(maxTime) {
				maxTime = t
			}
		}
	}
	return maxTime, nil
}

func (r *Repo) GetPrevSnapshot(ctx context.Context, snapshot string) (string, error) {
	currentTime, err := time.Parse("2006-01-02", snapshot)
	if err != nil {
		return "", err
	}

	var prevDate time.Time

	if r.isLocal() {
		base := filepath.Join(r.localDir, r.rootPrefix)
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			t, err := time.Parse("2006-01-02", name)
			if err == nil && t.Before(currentTime) && t.After(prevDate) {
				prevDate = t
			}
		}
	} else {
		it := r.client.Bucket(r.bucketName).Objects(ctx, &storage.Query{
			Prefix:    r.rootPrefix + "/",
			Delimiter: "/",
		})

		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return "", err
			}
			if attrs.Prefix != "" {
				snap := strings.TrimPrefix(attrs.Prefix, r.rootPrefix+"/")
				snap = strings.TrimSuffix(snap, "/")
				t, err := time.Parse("2006-01-02", snap)
				if err == nil && t.Before(currentTime) && t.After(prevDate) {
					prevDate = t
				}
			}
		}
	}

	if prevDate.IsZero() {
		// Return empty if no previous snapshot found, caller handles it
		return "", nil 
	}
	return prevDate.Format("2006-01-02"), nil
}

func (r *Repo) WriteDiffs(ctx context.Context, snapshot, title string, diffs []domain.Diff) error {
	if r.isLocal() {
		path := r.localPath(snapshot, title+"_diffs.parquet")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		writer := parquet.NewGenericWriter[domain.Diff](f)
		if _, err := writer.Write(diffs); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		return nil
	}

	path := r.objectPath(snapshot, title+"_diffs.parquet")
	w := r.client.Bucket(r.bucketName).Object(path).NewWriter(ctx)
	defer w.Close()

	writer := parquet.NewGenericWriter[domain.Diff](w)
	if _, err := writer.Write(diffs); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return w.Close()
}

func (r *Repo) WriteLSA(ctx context.Context, snapshot, title string, lsa domain.LSAActivity) error {
	// Implementation for LSA writing
	return nil
}

func (r *Repo) WriteSummaries(ctx context.Context, snapshot string, summaries []domain.Summary) error {
	if r.isLocal() {
		path := r.localPath(snapshot, "summaries.parquet")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		writer := parquet.NewGenericWriter[domain.Summary](f)
		if _, err := writer.Write(summaries); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		return nil
	}

	path := r.objectPath(snapshot, "summaries.parquet")
	w := r.client.Bucket(r.bucketName).Object(path).NewWriter(ctx)
	defer w.Close()

	writer := parquet.NewGenericWriter[domain.Summary](w)
	if _, err := writer.Write(summaries); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return w.Close()
}
