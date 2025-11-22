package parquet

import (
	"bytes"
	"context"
	"io"
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

func (r *Repo) WriteSections(ctx context.Context, snapshot, title string, sections []domain.Section) error {
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
	path := r.objectPath(snapshot, title+".parquet")
	obj := r.client.Bucket(r.bucketName).Object(path)

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// parquet.NewGenericReader needs io.ReaderAt; buffer entire object in memory.
	buf, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	reader := parquet.NewGenericReader[domain.Section](bytes.NewReader(buf))
	defer reader.Close()

	rows := make([]domain.Section, reader.NumRows())
	_, err = reader.Read(rows)
	return rows, err
}

func (r *Repo) GetLatestSnapshot(ctx context.Context) (time.Time, error) {
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
	it := r.client.Bucket(r.bucketName).Objects(ctx, &storage.Query{
		Prefix:    r.rootPrefix + "/",
		Delimiter: "/",
	})

	currentTime, err := time.Parse("2006-01-02", snapshot)
	if err != nil {
		return "", err
	}

	var prevDate time.Time
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
	if prevDate.IsZero() {
		// Return empty if no previous snapshot found, caller handles it
		return "", nil 
	}
	return prevDate.Format("2006-01-02"), nil
}

func (r *Repo) WriteDiffs(ctx context.Context, snapshot, title string, diffs []domain.Diff) error {
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
